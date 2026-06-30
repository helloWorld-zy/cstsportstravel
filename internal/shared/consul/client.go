// Package consul provides service registration, deregistration, and health check
// wrappers for Consul service discovery. All microservices register with Consul
// on startup and deregister on graceful shutdown.
package consul

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/hashicorp/consul/api"
)

// ServiceRegistration holds the metadata for registering a service with Consul.
type ServiceRegistration struct {
	ServiceName string
	ServiceID   string
	Address     string
	Port        int
	Tags        []string
	Meta        map[string]string
}

// Client wraps the Consul API client with service registration capabilities.
type Client struct {
	client *api.Client
	reg    *ServiceRegistration
}

// Config holds Consul connection settings.
type Config struct {
	Addr      string // Consul agent address, e.g. "localhost:8500"
	Token     string // ACL token (optional)
	Namespace string // Consul namespace (optional)
}

// NewClient creates a new Consul client.
func NewClient(cfg Config) (*Client, error) {
	consulCfg := api.DefaultConfig()
	consulCfg.Address = cfg.Addr
	if cfg.Token != "" {
		consulCfg.Token = cfg.Token
	}
	if cfg.Namespace != "" {
		consulCfg.Namespace = cfg.Namespace
	}

	client, err := api.NewClient(consulCfg)
	if err != nil {
		return nil, fmt.Errorf("create consul client: %w", err)
	}

	return &Client{client: client}, nil
}

// Register registers the service with Consul, including HTTP health checks.
// The health check hits /health every 10 seconds with a 5-second timeout.
// Services are deregistered after 30 seconds of critical state.
func (c *Client) Register(reg ServiceRegistration) error {
	c.reg = &reg

	portStr := strconv.Itoa(reg.Port)
	checkAddr := net.JoinHostPort(reg.Address, portStr)

	registration := &api.AgentServiceRegistration{
		ID:      reg.ServiceID,
		Name:    reg.ServiceName,
		Address: reg.Address,
		Port:    reg.Port,
		Tags:    reg.Tags,
		Meta:    reg.Meta,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s/health", checkAddr),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	if err := c.client.Agent().ServiceRegister(registration); err != nil {
		return fmt.Errorf("register service %s: %w", reg.ServiceName, err)
	}

	log.Printf("consul: registered service %s (id=%s) at %s:%d",
		reg.ServiceName, reg.ServiceID, reg.Address, reg.Port)
	return nil
}

// Deregister removes the service from Consul. Call this on graceful shutdown.
func (c *Client) Deregister() error {
	if c.reg == nil {
		return nil
	}
	if err := c.client.Agent().ServiceDeregister(c.reg.ServiceID); err != nil {
		return fmt.Errorf("deregister service %s: %w", c.reg.ServiceID, err)
	}
	log.Printf("consul: deregistered service %s", c.reg.ServiceID)
	return nil
}

// Discover finds healthy instances of a service by name.
func (c *Client) Discover(serviceName string) ([]*api.ServiceEntry, error) {
	entries, _, err := c.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("discover service %s: %w", serviceName, err)
	}
	return entries, nil
}

// GetServiceAddress returns the address:port of a healthy service instance.
// Returns an error if no healthy instances are found.
func (c *Client) GetServiceAddress(serviceName string) (string, error) {
	entries, err := c.Discover(serviceName)
	if err != nil {
		return "", err
	}
	if len(entries) == 0 {
		return "", fmt.Errorf("no healthy instances of service %s", serviceName)
	}

	// Simple round-robin: pick a random entry based on current time
	entry := entries[time.Now().UnixNano()%int64(len(entries))]
	return net.JoinHostPort(entry.Service.Address, strconv.Itoa(entry.Service.Port)), nil
}

// KV returns the Consul KV client for configuration management.
func (c *Client) KV() *api.KV {
	return c.client.KV()
}

// Client returns the underlying Consul API client for advanced operations.
func (c *Client) RawClient() *api.Client {
	return c.client
}
