// Package service provides a shared bootstrap for all microservices.
// It handles health/ready endpoints, Consul registration, graceful shutdown,
// and common middleware setup.
//
// Constitution: All services MUST register with Consul and expose /health and /ready.
package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/travel-booking/server/internal/shared/consul"
)

// Config holds the service bootstrap configuration.
type Config struct {
	Name       string // Service name, e.g. "user-service"
	ID         string // Unique instance ID, e.g. "user-service-1"
	Address    string // Bind address, e.g. "0.0.0.0"
	Port       int    // HTTP port
	Tags       []string
	Meta       map[string]string
	ConsulAddr string // Consul agent address
}

// Service represents a running microservice instance.
type Service struct {
	cfg          Config
	engine       *gin.Engine
	consulClient *consul.Client
	ready        bool
}

// New creates a new Service with health/ready endpoints pre-configured.
func New(cfg Config) *Service {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	svc := &Service{
		cfg:    cfg,
		engine: engine,
	}

	// Health check endpoint (always healthy if process is running)
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": cfg.Name,
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Readiness check endpoint (healthy only after initialization is complete)
	engine.GET("/ready", func(c *gin.Context) {
		if !svc.ready {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "not_ready",
				"service": cfg.Name,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  "ready",
			"service": cfg.Name,
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	return svc
}

// Engine returns the Gin engine for registering routes.
func (s *Service) Engine() *gin.Engine {
	return s.engine
}

// SetReady marks the service as ready to accept traffic.
// Call this after all initialization (DB, cache, NATS, etc.) is complete.
func (s *Service) SetReady() {
	s.ready = true
}

// Run starts the service: registers with Consul, starts HTTP server,
// and blocks until a shutdown signal is received.
func (s *Service) Run() error {
	// Register with Consul
	if s.cfg.ConsulAddr != "" {
		client, err := consul.NewClient(consul.Config{Addr: s.cfg.ConsulAddr})
		if err != nil {
			log.Printf("WARN: consul client creation failed: %v (continuing without Consul)", err)
		} else {
			s.consulClient = client
			reg := consul.ServiceRegistration{
				ServiceName: s.cfg.Name,
				ServiceID:   s.cfg.ID,
				Address:     s.cfg.Address,
				Port:        s.cfg.Port,
				Tags:        s.cfg.Tags,
				Meta:        s.cfg.Meta,
			}
			if err := client.Register(reg); err != nil {
				log.Printf("WARN: consul registration failed: %v (continuing without Consul)", err)
			}
		}
	}

	// HTTP server
	addr := fmt.Sprintf("%s:%d", s.cfg.Address, s.cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      s.engine,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("%s: listening on %s", s.cfg.Name, addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("%s: server error: %v", s.cfg.Name, err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("%s: shutting down (signal: %s)", s.cfg.Name, sig)

	// Deregister from Consul
	if s.consulClient != nil {
		if err := s.consulClient.Deregister(); err != nil {
			log.Printf("WARN: consul deregistration failed: %v", err)
		}
	}

	// Shutdown HTTP server with 30s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("%s: forced shutdown: %v", s.cfg.Name, err)
	}

	log.Printf("%s: stopped", s.cfg.Name)
	return nil
}
