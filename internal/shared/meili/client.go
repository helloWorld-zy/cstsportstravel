// Package meili provides a Meilisearch client wrapper for product search indexing.
// It handles index creation, document CRUD, and search queries with facet filtering.
package meili

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/meilisearch/meilisearch-go"
)

// Client wraps the Meilisearch client with product-specific operations.
type Client struct {
	client meilisearch.ServiceManager
}

// Config holds Meilisearch connection settings.
type Config struct {
	Host   string // Meilisearch host, e.g. "http://localhost:7700"
	APIKey string // Master API key
}

// NewClient creates a new Meilisearch client.
func NewClient(cfg Config) (*Client, error) {
	if cfg.Host == "" {
		cfg.Host = "http://localhost:7700"
	}

	client := meilisearch.New(cfg.Host, meilisearch.WithAPIKey(cfg.APIKey))

	// Verify connectivity
	health, err := client.Health()
	if err != nil {
		return nil, fmt.Errorf("connect to Meilisearch: %w", err)
	}
	if health.Status != "available" {
		return nil, fmt.Errorf("Meilisearch not available: status=%s", health.Status)
	}

	log.Printf("meili: connected to %s", cfg.Host)
	return &Client{client: client}, nil
}

// IndexConfig defines the searchable/filterable/sortable attributes for an index.
type IndexConfig struct {
	UID                string
	PrimaryKey         string
	SearchableAttrs    []string
	FilterableAttrs    []string
	SortableAttrs      []string
	RankingRules       []string
	StopWords          []string
	Synonyms           map[string][]string
}

// EnsureIndex creates or updates a Meilisearch index with the given configuration.
func (c *Client) EnsureIndex(cfg IndexConfig) error {
	// Create index if not exists
	_, err := c.client.GetIndex(cfg.UID)
	if err != nil {
		// Index doesn't exist, create it
		task, createErr := c.client.CreateIndex(&meilisearch.IndexConfig{
			Uid:        cfg.UID,
			PrimaryKey: cfg.PrimaryKey,
		})
		if createErr != nil {
			return fmt.Errorf("create index %s: %w", cfg.UID, createErr)
		}
		if waitErr := c.waitForTask(task.TaskUID); waitErr != nil {
			return waitErr
		}
		log.Printf("meili: created index %s", cfg.UID)
	}

	// Update settings
	settings := &meilisearch.Settings{
		SearchableAttributes: cfg.SearchableAttrs,
		FilterableAttributes: cfg.FilterableAttrs,
		SortableAttributes:   cfg.SortableAttrs,
	}
	if len(cfg.RankingRules) > 0 {
		settings.RankingRules = cfg.RankingRules
	}
	if len(cfg.StopWords) > 0 {
		settings.StopWords = cfg.StopWords
	}
	if len(cfg.Synonyms) > 0 {
		settings.Synonyms = cfg.Synonyms
	}

	task, err := c.client.Index(cfg.UID).UpdateSettings(settings)
	if err != nil {
		return fmt.Errorf("update settings for %s: %w", cfg.UID, err)
	}
	if waitErr := c.waitForTask(task.TaskUID); waitErr != nil {
		return waitErr
	}

	log.Printf("meili: configured index %s", cfg.UID)
	return nil
}

// AddDocuments adds or updates documents in the specified index.
func (c *Client) AddDocuments(indexUID string, documents interface{}) error {
	task, err := c.client.Index(indexUID).AddDocuments(documents, nil)
	if err != nil {
		return fmt.Errorf("add documents to %s: %w", indexUID, err)
	}
	return c.waitForTask(task.TaskUID)
}

// DeleteDocument deletes a document by its primary key value.
func (c *Client) DeleteDocument(indexUID string, documentID string) error {
	task, err := c.client.Index(indexUID).DeleteDocument(documentID, nil)
	if err != nil {
		return fmt.Errorf("delete document %s from %s: %w", documentID, indexUID, err)
	}
	return c.waitForTask(task.TaskUID)
}

// DeleteDocumentsByFilter deletes documents matching a filter expression.
func (c *Client) DeleteDocumentsByFilter(indexUID string, filter interface{}) error {
	task, err := c.client.Index(indexUID).DeleteDocumentsByFilter(filter, nil)
	if err != nil {
		return fmt.Errorf("delete documents by filter from %s: %w", indexUID, err)
	}
	return c.waitForTask(task.TaskUID)
}

// SearchResult represents a Meilisearch search response.
type SearchResult struct {
	Hits               []map[string]interface{} `json:"hits"`
	EstimatedTotalHits  int64                   `json:"estimatedTotalHits"`
	Offset              int                     `json:"offset"`
	Limit               int                     `json:"limit"`
	ProcessingTimeMs    int                     `json:"processingTimeMs"`
	FacetDistribution   map[string]interface{}   `json:"facetDistribution,omitempty"`
}

// SearchRequest holds search query parameters.
type SearchRequest struct {
	Limit                int
	Offset               int
	Filter               string
	Facets               []string
	Sort                 []string
	AttributesToRetrieve []string
}

// Search performs a search query with optional facet filters.
func (c *Client) Search(indexUID string, query string, opts *SearchRequest) (*SearchResult, error) {
	if opts == nil {
		opts = &SearchRequest{}
	}

	req := &meilisearch.SearchRequest{
		Limit:  int64(opts.Limit),
		Offset: int64(opts.Offset),
	}
	if opts.Filter != "" {
		req.Filter = opts.Filter
	}
	if len(opts.Facets) > 0 {
		req.Facets = opts.Facets
	}
	if len(opts.Sort) > 0 {
		req.Sort = opts.Sort
	}
	if opts.AttributesToRetrieve != nil {
		req.AttributesToRetrieve = opts.AttributesToRetrieve
	}

	resp, err := c.client.Index(indexUID).Search(query, req)
	if err != nil {
		return nil, fmt.Errorf("search %s: %w", indexUID, err)
	}

	result := &SearchResult{
		EstimatedTotalHits: resp.EstimatedTotalHits,
		Offset:             int(resp.Offset),
		Limit:              int(resp.Limit),
		ProcessingTimeMs:   int(resp.ProcessingTimeMs),
	}

	// Convert hits
	if resp.Hits != nil {
		hits := make([]map[string]interface{}, 0, len(resp.Hits))
		for _, hit := range resp.Hits {
			m := make(map[string]interface{})
			for k, v := range hit {
				var val interface{}
				if err := json.Unmarshal(v, &val); err != nil {
					val = string(v)
				}
				m[k] = val
			}
			hits = append(hits, m)
		}
		result.Hits = hits
	}

	return result, nil
}

// waitForTask polls a Meilisearch task until it completes or fails.
func (c *Client) waitForTask(taskUID int64) error {
	for i := 0; i < 50; i++ {
		task, err := c.client.GetTask(taskUID)
		if err != nil {
			return fmt.Errorf("get task %d: %w", taskUID, err)
		}
		switch task.Status {
		case "succeeded":
			return nil
		case "failed":
			return fmt.Errorf("task %d failed: %v", taskUID, task.Error)
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("task %d timed out", taskUID)
}

// RawClient returns the underlying Meilisearch client for advanced operations.
func (c *Client) RawClient() meilisearch.ServiceManager {
	return c.client
}
