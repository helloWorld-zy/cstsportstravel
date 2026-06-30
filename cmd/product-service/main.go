// product-service handles product catalog, categories, destinations, and supplier management.
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
	port := flag.Int("port", 8082, "HTTP port")
	consulAddr := flag.String("consul", "localhost:8500", "Consul agent address")
	flag.Parse()

	hostname, _ := os.Hostname()
	svcID := fmt.Sprintf("product-service-%s-%d", hostname, *port)

	s := service.New(service.Config{
		Name:       "product-service",
		ID:         svcID,
		Address:    "0.0.0.0",
		Port:       *port,
		Tags:       []string{"microservice", "product", "v2"},
		Meta:       map[string]string{"version": "2.0.0"},
		ConsulAddr: *consulAddr,
	})

	r := s.Engine()
	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			// Legacy MVP routes (backward compatible)
			v1.GET("/products", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "product-service", "endpoint": "list"})
			})
			v1.GET("/products/:id", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "product-service", "endpoint": "detail"})
			})
			v1.GET("/categories", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "product-service", "endpoint": "categories"})
			})
			v1.GET("/destinations", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "product-service", "endpoint": "destinations"})
			})
		}
		v2 := api.Group("/v2")
		{
			// Outbound product routes
			v2.GET("/products/outbound", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "product-service", "endpoint": "outbound-list"})
			})
			v2.GET("/products/outbound/:id", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "product-service", "endpoint": "outbound-detail"})
			})
			// Supplier routes
			v2.POST("/suppliers/apply", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "product-service", "endpoint": "supplier-apply"})
			})
			v2.GET("/suppliers/:id", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "product-service", "endpoint": "supplier-detail"})
			})
			// Country routes
			v2.GET("/countries", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "product-service", "endpoint": "countries"})
			})
		}
	}

	s.SetReady()

	if err := s.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}
