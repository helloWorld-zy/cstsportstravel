// distribution-service handles the two-level distribution system:
// distributor management, commission calculation, promotion links, and withdrawals.
// This is a new service created as part of Phase 1 service decomposition.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/travel-booking/server/internal/shared/service"
)

func main() {
	port := flag.Int("port", 8085, "HTTP port")
	consulAddr := flag.String("consul", "localhost:8500", "Consul agent address")
	flag.Parse()

	hostname, _ := os.Hostname()
	svcID := fmt.Sprintf("distribution-service-%s-%d", hostname, *port)

	s := service.New(service.Config{
		Name:       "distribution-service",
		ID:         svcID,
		Address:    "0.0.0.0",
		Port:       *port,
		Tags:       []string{"microservice", "distribution", "v2"},
		Meta:       map[string]string{"version": "2.0.0"},
		ConsulAddr: *consulAddr,
	})

	r := s.Engine()
	api := r.Group("/api")
	{
		v2 := api.Group("/v2")
		{
			// Distributor management
			v2.POST("/distributors/apply", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "distribution-service", "endpoint": "apply"})
			})
			v2.GET("/distributors/:id", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "distribution-service", "endpoint": "detail"})
			})
			v2.GET("/distributors/overview", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "distribution-service", "endpoint": "overview"})
			})

			// Promotion links
			v2.POST("/distributors/promotion-links", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "distribution-service", "endpoint": "create-link"})
			})
			v2.GET("/distributors/promotion-links", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "distribution-service", "endpoint": "list-links"})
			})

			// Commission
			v2.GET("/commissions", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "distribution-service", "endpoint": "commission-list"})
			})
			v2.GET("/commissions/rules", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "distribution-service", "endpoint": "commission-rules"})
			})

			// Withdrawal
			v2.POST("/withdrawals", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "distribution-service", "endpoint": "withdraw"})
			})
			v2.GET("/withdrawals", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "distribution-service", "endpoint": "withdrawal-list"})
			})

			// Team
			v2.GET("/distributors/team", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "distribution-service", "endpoint": "team"})
			})

			// Performance
			v2.GET("/distributors/performance", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "distribution-service", "endpoint": "performance"})
			})
		}
	}

	s.SetReady()

	if err := s.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}
