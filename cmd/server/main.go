package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/cache"
	"github.com/travel-booking/server/internal/common/config"
	"github.com/travel-booking/server/internal/common/database"
	"github.com/travel-booking/server/internal/common/logger"
	"github.com/travel-booking/server/internal/common/middleware"
	"github.com/travel-booking/server/internal/common/router"
)

func main() {
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.New(cfg.Log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("starting server",
		zap.Int("port", cfg.Server.Port),
		zap.String("mode", cfg.Server.Mode),
	)

	// Initialize database
	db, err := database.NewPostgres(cfg.Database)
	if err != nil {
		log.Fatal("failed to connect database", zap.Error(err))
	}
	log.Info("database connected")

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Initialize Redis
	rdb, err := cache.NewRedis(cfg.Redis)
	if err != nil {
		log.Fatal("failed to connect redis", zap.Error(err))
	}
	log.Info("redis connected")
	defer rdb.Close()

	// Initialize JWT manager with Redis token revocation (CHK039)
	tokenRevoker := middleware.NewRedisTokenRevoker(rdb.Client())
	jwtManager, err := middleware.NewJWTManager(cfg.JWT, tokenRevoker)
	if err != nil {
		log.Fatal("failed to init JWT manager", zap.Error(err))
	}
	log.Info("JWT manager initialized", zap.Bool("token_revocation", true))

	// Setup router
	r := router.New(cfg, db, rdb, jwtManager, log)

	// HTTP server with TLS 1.3 support (Constitution Principle III)
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r.Engine,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// Configure TLS 1.3 minimum version (Constitution §III: TLS 1.3 mandatory)
	if cfg.Server.TLS.Enabled {
		srv.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS13,
		}
		log.Info("TLS 1.3 enabled",
			zap.String("cert", cfg.Server.TLS.CertFile),
			zap.String("min_version", "1.3"),
		)
	}

	// Start server in goroutine
	go func() {
		if cfg.Server.TLS.Enabled {
			log.Info("server listening (TLS)", zap.String("addr", srv.Addr))
			if err := srv.ListenAndServeTLS(cfg.Server.TLS.CertFile, cfg.Server.TLS.KeyFile); err != nil && err != http.ErrServerClosed {
				log.Fatal("server error", zap.Error(err))
			}
		} else {
			log.Info("server listening (HTTP — TLS disabled)", zap.String("addr", srv.Addr))
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal("server error", zap.Error(err))
			}
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Info("shutting down", zap.String("signal", sig.String()))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server forced shutdown", zap.Error(err))
	}

	log.Info("server stopped")
}
