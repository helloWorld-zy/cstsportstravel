// order-service handles order lifecycle, visa orders, and booking flows.
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
	port := flag.Int("port", 8083, "HTTP port")
	consulAddr := flag.String("consul", "localhost:8500", "Consul agent address")
	flag.Parse()

	hostname, _ := os.Hostname()
	svcID := fmt.Sprintf("order-service-%s-%d", hostname, *port)

	s := service.New(service.Config{
		Name:       "order-service",
		ID:         svcID,
		Address:    "0.0.0.0",
		Port:       *port,
		Tags:       []string{"microservice", "order", "v2"},
		Meta:       map[string]string{"version": "2.0.0"},
		ConsulAddr: *consulAddr,
	})

	r := s.Engine()
	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			// Legacy MVP routes (backward compatible)
			v1.POST("/orders", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "order-service", "endpoint": "create"})
			})
			v1.GET("/orders", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "order-service", "endpoint": "list"})
			})
			v1.GET("/orders/:id", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "order-service", "endpoint": "detail"})
			})
		}
		v2 := api.Group("/v2")
		{
			// Outbound booking routes
			v2.POST("/orders/outbound", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "order-service", "endpoint": "outbound-booking"})
			})
			// Visa routes
			v2.POST("/visa/orders", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "order-service", "endpoint": "visa-create"})
			})
			v2.GET("/visa/orders/:id", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "order-service", "endpoint": "visa-detail"})
			})
			v2.POST("/visa/orders/:id/materials", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "order-service", "endpoint": "visa-material-upload"})
			})
			v2.GET("/visa/orders/:id/progress", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "order-service", "endpoint": "visa-progress"})
			})
			// Passport routes
			v2.POST("/passports", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "order-service", "endpoint": "passport-create"})
			})
		}
	}

	s.SetReady()

	if err := s.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}
