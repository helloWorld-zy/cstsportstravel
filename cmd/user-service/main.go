// user-service handles user authentication, profile management, and traveller info.
// Extracted from monolith as part of Phase 1 service decomposition.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/travel-booking/server/internal/shared/service"
)

func main() {
	port := flag.Int("port", 8081, "HTTP port")
	consulAddr := flag.String("consul", "localhost:8500", "Consul agent address")
	flag.Parse()

	hostname, _ := os.Hostname()
	svcID := fmt.Sprintf("user-service-%s-%d", hostname, *port)

	s := service.New(service.Config{
		Name:       "user-service",
		ID:         svcID,
		Address:    "0.0.0.0",
		Port:       *port,
		Tags:       []string{"microservice", "user", "v2"},
		Meta:       map[string]string{"version": "2.0.0"},
		ConsulAddr: *consulAddr,
	})

	// Register user-service routes
	r := s.Engine()
	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			// Legacy MVP routes (backward compatible)
			v1.GET("/users/profile", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "user-service", "endpoint": "profile"})
			})
		}
		v2 := api.Group("/v2")
		{
			// Phase 2 routes
			v2.GET("/users/profile", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "user-service", "endpoint": "profile-v2"})
			})
			v2.POST("/users/passport", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "user-service", "endpoint": "passport-create"})
			})
		}
	}

	// Mark as ready after route registration
	s.SetReady()

	if err := s.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}
