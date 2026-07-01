// Package meili provides Meilisearch index schema definitions for product search.
// This file defines searchable/filterable/sortable attributes covering domestic, outbound, and cruise products.
package meili

// ProductIndexConfig returns the Meilisearch index configuration for the product search index.
// Covers domestic tour (group_tour), outbound travel (outbound_group), and cruise products.
func ProductIndexConfig() IndexConfig {
	return IndexConfig{
		UID:        "products",
		PrimaryKey: "id",
		SearchableAttrs: []string{
			"product_name",
			"summary",
			"destination_cities",
			"destination_tags",
			"origin_city",
		},
		FilterableAttrs: []string{
			// Product type: group_tour, outbound_group, cruise
			"product_type",
			// Category
			"category_id",
			// Outbound dimensions
			"continent",
			"country_id",
			"visa_type",
			"origin_city",
			"days",
			// Common dimensions
			"status",
			"supplier_id",
			"price_range",
			"transport_mode",
			"product_grade",
		},
		SortableAttrs: []string{
			"adult_price",
			"days",
			"order_count",
			"view_count",
			"satisfaction_rate",
			"created_at",
		},
		RankingRules: []string{
			"words",
			"typo",
			"proximity",
			"attribute",
			"sort",
			"exactness",
		},
		StopWords: []string{
			"的", "了", "在", "是", "和", "有",
			"游", "旅", "行", "团",
		},
		Synonyms: map[string][]string{
			"出境游": {"出境旅游", "出国游", "海外游"},
			"跟团游": {"团体游", "团队游", "组团游"},
			"邮轮":   {"游轮", "邮轮旅游", "邮轮游"},
			"自由行": {"自助游", "半自助"},
			"签证":   {"旅游签证", "签证办理"},
		},
	}
}

// SuggestIndexConfig returns the Meilisearch index configuration for the search suggestion index.
// Supports three suggestion types: hot_destinations, product_names, attractions.
func SuggestIndexConfig() IndexConfig {
	return IndexConfig{
		UID:        "suggestions",
		PrimaryKey: "id",
		SearchableAttrs: []string{
			"text",
		},
		FilterableAttrs: []string{
			"type",
			"continent",
			"country_id",
		},
		SortableAttrs: []string{
			"weight",
		},
		RankingRules: []string{
			"sort",
			"words",
			"exactness",
		},
	}
}

// AllIndexConfigs returns all Meilisearch index configurations for the application.
func AllIndexConfigs() []IndexConfig {
	return []IndexConfig{
		ProductIndexConfig(),
		SuggestIndexConfig(),
	}
}
