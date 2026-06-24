package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
)

// ConsulClient manages dynamic configuration from Consul KV.
type ConsulClient struct {
	client    *api.Client
	keyPrefix string
	mu        sync.RWMutex
	values    map[string]string
	onChange  func(key string, value string)
}

// NewConsulClient creates a new Consul KV client.
func NewConsulClient(cfg ConsulConfig) (*ConsulClient, error) {
	consulCfg := api.DefaultConfig()
	consulCfg.Address = cfg.Addr

	client, err := api.NewClient(consulCfg)
	if err != nil {
		return nil, fmt.Errorf("create consul client: %w", err)
	}

	c := &ConsulClient{
		client:    client,
		keyPrefix: cfg.KeyPrefix,
		values:    make(map[string]string),
	}

	// Load initial values
	if err := c.loadAll(); err != nil {
		// Log warning but don't fail — fallback to static config
		log.Printf("WARN: consul unavailable, using static config: %v", err)
	}

	return c, nil
}

// SetOnChange registers a callback for configuration changes.
func (c *ConsulClient) SetOnChange(fn func(key string, value string)) {
	c.onChange = fn
}

// GetString returns a string value from Consul KV, or empty string if not found.
func (c *ConsulClient) GetString(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.values[c.keyPrefix+key]
}

// GetInt returns an int value from Consul KV, or 0 if not found.
func (c *ConsulClient) GetInt(key string) int {
	v := c.GetString(key)
	if v == "" {
		return 0
	}
	var result int
	if _, err := fmt.Sscanf(v, "%d", &result); err != nil {
		return 0
	}
	return result
}

// GetBool returns a bool value from Consul KV, or false if not found.
func (c *ConsulClient) GetBool(key string) bool {
	return c.GetString(key) == "true"
}

// GetJSON unmarshals a JSON value from Consul KV.
func (c *ConsulClient) GetJSON(key string, dest interface{}) error {
	v := c.GetString(key)
	if v == "" {
		return nil
	}
	return json.Unmarshal([]byte(v), dest)
}

// Watch starts watching Consul KV for changes. Blocks until ctx is cancelled.
func (c *ConsulClient) Watch(ctx context.Context) {
	var lastIndex uint64

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		pairs, meta, err := c.client.KV().List(c.keyPrefix, &api.QueryOptions{
			WaitIndex: lastIndex,
			WaitTime:  5 * time.Minute,
		})
		if err != nil {
			log.Printf("WARN: consul watch error: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if meta.LastIndex <= lastIndex {
			continue
		}
		lastIndex = meta.LastIndex

		c.mu.Lock()
		for _, p := range pairs {
			old := c.values[p.Key]
			c.values[p.Key] = string(p.Value)
			if old != string(p.Value) && c.onChange != nil {
				c.onChange(p.Key, string(p.Value))
			}
		}
		c.mu.Unlock()
	}
}

// loadAll fetches all KV pairs under the prefix.
func (c *ConsulClient) loadAll() error {
	pairs, _, err := c.client.KV().List(c.keyPrefix, nil)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, p := range pairs {
		c.values[p.Key] = string(p.Value)
	}
	return nil
}
