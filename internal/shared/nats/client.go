// Package nats provides a NATS JetStream client wrapper with connection management,
// stream/consumer setup, and typed publish/subscribe helpers for inter-service events.
package nats

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

// Client wraps the NATS connection and JetStream context.
type Client struct {
	conn *nats.Conn
	js   nats.JetStreamContext
}

// Config holds NATS connection settings.
type Config struct {
	URLs         []string // NATS server URLs, e.g. ["nats://localhost:4222"]
	ClusterName  string   // Cluster name for logging
	MaxReconnect int      // Max reconnect attempts, -1 for infinite
	ReconnectWait int     // Seconds between reconnect attempts
}

// NewClient creates a new NATS client with JetStream enabled.
// It establishes a connection and verifies JetStream availability.
func NewClient(cfg Config) (*Client, error) {
	if len(cfg.URLs) == 0 {
		cfg.URLs = []string{nats.DefaultURL}
	}
	if cfg.MaxReconnect == 0 {
		cfg.MaxReconnect = -1
	}
	if cfg.ReconnectWait == 0 {
		cfg.ReconnectWait = 2
	}

	opts := []nats.Option{
		nats.MaxReconnects(cfg.MaxReconnect),
		nats.ReconnectWait(time.Duration(cfg.ReconnectWait) * time.Second),
		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			log.Printf("nats: disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(conn *nats.Conn) {
			log.Printf("nats: reconnected to %s", conn.ConnectedUrl())
		}),
		nats.ClosedHandler(func(conn *nats.Conn) {
			log.Printf("nats: connection closed")
		}),
	}

	conn, err := nats.Connect(cfg.URLs[0], opts...)
	if err != nil {
		return nil, fmt.Errorf("connect to NATS: %w", err)
	}

	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("enable JetStream: %w", err)
	}

	log.Printf("nats: connected to %s with JetStream enabled", conn.ConnectedUrl())
	return &Client{conn: conn, js: js}, nil
}

// EnsureStream creates a JetStream stream if it doesn't exist.
func (c *Client) EnsureStream(name string, subjects []string) error {
	info, err := c.js.StreamInfo(name)
	if err == nil && info != nil {
		return nil // stream already exists
	}

	_, err = c.js.AddStream(&nats.StreamConfig{
		Name:     name,
		Subjects: subjects,
		Storage:  nats.FileStorage,
		MaxAge:   7 * 24 * time.Hour, // 7 days retention
		Replicas: 1,
	})
	if err != nil {
		return fmt.Errorf("create stream %s: %w", name, err)
	}

	log.Printf("nats: created stream %s with subjects %v", name, subjects)
	return nil
}

// Publish publishes a message to a NATS subject. The message is JSON-encoded.
func (c *Client) Publish(subject string, msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	_, err = c.js.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("publish to %s: %w", subject, err)
	}

	return nil
}

// Subscribe creates a durable subscriber on a JetStream subject.
// The handler receives the raw message bytes. Ack/Nack is handled automatically.
func (c *Client) Subscribe(subject string, durable string, handler func(data []byte) error) (*nats.Subscription, error) {
	sub, err := c.js.Subscribe(subject, func(msg *nats.Msg) {
		if err := handler(msg.Data); err != nil {
			log.Printf("nats: handler error on %s: %v", subject, err)
			// Nak the message for retry
			if nakErr := msg.Nak(); nakErr != nil {
				log.Printf("nats: nak error: %v", nakErr)
			}
			return
		}
		if ackErr := msg.Ack(); ackErr != nil {
			log.Printf("nats: ack error: %v", ackErr)
		}
	}, nats.Durable(durable), nats.ManualAck(), nats.AckWait(30*time.Second))

	if err != nil {
		return nil, fmt.Errorf("subscribe to %s: %w", subject, err)
	}

	log.Printf("nats: subscribed to %s with durable=%s", subject, durable)
	return sub, nil
}

// QueueSubscribe creates a queue-based durable subscriber for load balancing
// across multiple instances of the same service.
func (c *Client) QueueSubscribe(subject string, queue string, durable string, handler func(data []byte) error) (*nats.Subscription, error) {
	sub, err := c.js.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		if err := handler(msg.Data); err != nil {
			log.Printf("nats: queue handler error on %s: %v", subject, err)
			if nakErr := msg.Nak(); nakErr != nil {
				log.Printf("nats: nak error: %v", nakErr)
			}
			return
		}
		if ackErr := msg.Ack(); ackErr != nil {
			log.Printf("nats: ack error: %v", ackErr)
		}
	}, nats.Durable(durable), nats.ManualAck(), nats.AckWait(30*time.Second))

	if err != nil {
		return nil, fmt.Errorf("queue subscribe to %s: %w", subject, err)
	}

	log.Printf("nats: queue subscribed to %s (queue=%s, durable=%s)", subject, queue, durable)
	return sub, nil
}

// Close drains the NATS connection gracefully.
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Drain()
	}
}

// Conn returns the underlying NATS connection for advanced operations.
func (c *Client) Conn() *nats.Conn {
	return c.conn
}

// JS returns the JetStream context for advanced operations.
func (c *Client) JS() nats.JetStreamContext {
	return c.js
}
