// payment-service handles payment processing, refunds, and reconciliation.
// Supports Alipay, WeChat Pay, and UnionPay channels.
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
	port := flag.Int("port", 8084, "HTTP port")
	consulAddr := flag.String("consul", "localhost:8500", "Consul agent address")
	flag.Parse()

	hostname, _ := os.Hostname()
	svcID := fmt.Sprintf("payment-service-%s-%d", hostname, *port)

	s := service.New(service.Config{
		Name:       "payment-service",
		ID:         svcID,
		Address:    "0.0.0.0",
		Port:       *port,
		Tags:       []string{"microservice", "payment", "v2"},
		Meta:       map[string]string{"version": "2.0.0"},
		ConsulAddr: *consulAddr,
	})

	r := s.Engine()
	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			// Legacy MVP routes (backward compatible)
			v1.POST("/payments/create", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "payment-service", "endpoint": "create"})
			})
			v1.POST("/payments/notify/alipay", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "payment-service", "endpoint": "alipay-notify"})
			})
			v1.POST("/payments/notify/wechat", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "payment-service", "endpoint": "wechat-notify"})
			})
		}
		v2 := api.Group("/v2")
		{
			// Payment creation
			v2.POST("/payments", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "payment-service", "endpoint": "create-v2"})
			})
			// UnionPay callbacks
			v2.POST("/payments/notify/unionpay", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "payment-service", "endpoint": "unionpay-notify"})
			})
			v2.POST("/payments/notify/unionpay/front", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "payment-service", "endpoint": "unionpay-front"})
			})
			// Deposit + balance
			v2.POST("/payments/deposit", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "payment-service", "endpoint": "deposit"})
			})
			v2.POST("/payments/balance", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "payment-service", "endpoint": "balance"})
			})
			// Refund
			v2.POST("/payments/:id/refund", func(c *gin.Context) {
				c.JSON(200, gin.H{"service": "payment-service", "endpoint": "refund"})
			})
		}
	}

	s.SetReady()

	if err := s.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}
