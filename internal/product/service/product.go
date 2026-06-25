// Package service provides business logic for the Product domain.
package service

import (
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/product/model"
	"github.com/travel-booking/server/internal/product/repository"
)

// ProductService provides business logic for product listing, detail, and search.
type ProductService struct {
	productRepo  *repository.ProductRepository
	categoryRepo *repository.CategoryRepository
	reviewRepo   *repository.ReviewRepository
	reviewSvc    *ReviewService
	logger       *zap.Logger
}

// NewProductService creates a new ProductService.
func NewProductService(
	productRepo *repository.ProductRepository,
	categoryRepo *repository.CategoryRepository,
	reviewRepo *repository.ReviewRepository,
	reviewSvc *ReviewService,
	logger *zap.Logger,
) *ProductService {
	return &ProductService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		reviewRepo:   reviewRepo,
		reviewSvc:    reviewSvc,
		logger:       logger,
	}
}

// --- Request/Response DTOs ---

// ListProductsRequest holds query parameters for product listing.
type ListProductsRequest struct {
	Destination  string `form:"destination"`
	Origin       string `form:"origin"`
	DaysMin      *int   `form:"days_min"`
	DaysMax      *int   `form:"days_max"`
	PriceMin     *int   `form:"price_min"`
	PriceMax     *int   `form:"price_max"`
	DepartureDate string `form:"departure_date"`
	CategoryID   *int64 `form:"category_id"`
	ProductGrade string `form:"product_grade"`
	Keyword      string `form:"keyword"`
	Sort         string `form:"sort" binding:"omitempty,oneof=recommended price_asc price_desc satisfaction sales days_asc days_desc"`
	Page         int    `form:"page" binding:"min=1"`
	PageSize     int    `form:"page_size" binding:"min=1,max=100"`
}

// ProductSummaryResponse is the product card data for list view.
type ProductSummaryResponse struct {
	ID                int64    `json:"id"`
	ProductNo         string   `json:"product_no"`
	ProductName       string   `json:"product_name"`
	CoverImage        string   `json:"cover_image,omitempty"`
	OriginCity        string   `json:"origin_city"`
	DestinationCities []string `json:"destination_cities"`
	Days              int      `json:"days"`
	Nights            int      `json:"nights"`
	MinPrice          int      `json:"min_price"` // yuan
	ProductGrade      string   `json:"product_grade,omitempty"`
	SatisfactionRate  *float64 `json:"satisfaction_rate,omitempty"`
	OrderCount        int      `json:"order_count"`
	Tags              []string `json:"tags,omitempty"`
}

// PaginatedProductsResponse is the paginated product list.
type PaginatedProductsResponse struct {
	Items    []ProductSummaryResponse `json:"items"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
}

// ProductDetailResponse is the full product detail.
type ProductDetailResponse struct {
	ProductSummaryResponse
	Summary           string                `json:"summary,omitempty"`
	Description       string                `json:"description,omitempty"`
	TransportMode     string                `json:"transport_mode,omitempty"`
	MinGroupSize      int                   `json:"min_group_size"`
	MaxGroupSize      int                   `json:"max_group_size"`
	FeeIncluded       string                `json:"fee_included,omitempty"`
	FeeExcluded       string                `json:"fee_excluded,omitempty"`
	BookingNotes      string                `json:"booking_notes,omitempty"`
	Itinerary         []ItineraryDayResponse `json:"itinerary,omitempty"`
	CancellationRules []CancellationRuleResp  `json:"cancellation_rules,omitempty"`
	Images            []string              `json:"images,omitempty"`
	ReviewSummary     *ReviewSummaryResponse `json:"review_summary,omitempty"`
}

// ItineraryDayResponse is a single day in the itinerary.
type ItineraryDayResponse struct {
	DayNo       int             `json:"day_no"`
	Title       string          `json:"title"`
	Description string          `json:"description,omitempty"`
	Meals       json.RawMessage `json:"meals,omitempty"`
	Hotel       string          `json:"hotel,omitempty"`
	Transport   string          `json:"transport,omitempty"`
	Spots       json.RawMessage `json:"spots,omitempty"`
}

// CancellationRuleResp is a cancellation rule.
type CancellationRuleResp struct {
	ID               int64   `json:"id"`
	RuleName         string  `json:"rule_name"`
	DaysBeforeMin    int     `json:"days_before_min"`
	DaysBeforeMax    *int    `json:"days_before_max,omitempty"`
	RefundPercentage float64 `json:"refund_percentage"`
	Description      string  `json:"description,omitempty"`
}

// DepartureCalendarResponse is a departure date entry.
type DepartureCalendarResponse struct {
	ID              int64  `json:"id"`
	DepartureDate   string `json:"departure_date"`
	ReturnDate      string `json:"return_date"`
	AdultPrice      int    `json:"adult_price"` // yuan
	ChildPrice      int    `json:"child_price"`
	InfantPrice     int    `json:"infant_price"`
	SingleSupplement int   `json:"single_supplement"`
	AvailableStock  int    `json:"available_stock"`
	StockStatus     string `json:"stock_status"` // sufficient/tight/sold_out
	CutoffDays      int    `json:"cutoff_days"`
}

// SearchSuggestResponse is an autocomplete suggestion.
type SearchSuggestResponse struct {
	Text string `json:"text"`
	Type string `json:"type"` // destination/product/spot
	ID   *int64 `json:"id,omitempty"`
}

// --- Service Methods ---

// ListProducts returns a filtered, sorted, paginated product list.
func (s *ProductService) ListProducts(req ListProductsRequest) (*PaginatedProductsResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}

	filter := repository.ProductFilter{
		Destination:  req.Destination,
		Origin:       req.Origin,
		DaysMin:      req.DaysMin,
		DaysMax:      req.DaysMax,
		PriceMin:     req.PriceMin,
		PriceMax:     req.PriceMax,
		CategoryID:   req.CategoryID,
		ProductGrade: req.ProductGrade,
		Keyword:      req.Keyword,
	}

	sort := repository.ProductSort(req.Sort)
	if sort == "" {
		sort = repository.SortRecommended
	}

	products, total, err := s.productRepo.FindWithFilters(filter, sort, req.Page, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("find products: %w", err)
	}

	items := make([]ProductSummaryResponse, len(products))
	for i, p := range products {
		items[i] = s.toProductSummary(p)
	}

	return &PaginatedProductsResponse{
		Items:    items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetProductDetail returns complete product information.
func (s *ProductService) GetProductDetail(id int64) (*ProductDetailResponse, error) {
	product, err := s.productRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("find product: %w", err)
	}

	// Increment view count (best-effort)
	go s.productRepo.IncrementViewCount(id)

	// Get min price from departures
	minPrice := s.getMinPrice(product.DepartureDates)

	// Build itinerary
	itinerary := make([]ItineraryDayResponse, len(product.Itineraries))
	for i, it := range product.Itineraries {
		itinerary[i] = ItineraryDayResponse{
			DayNo:       it.DayNo,
			Title:       it.Title,
			Description: it.Description,
			Meals:       it.Meals,
			Hotel:       it.Hotel,
			Transport:   it.Transport,
			Spots:       it.Spots,
		}
	}

	// Build cancellation rules
	rules := make([]CancellationRuleResp, len(product.RefundRules))
	for i, r := range product.RefundRules {
		rules[i] = CancellationRuleResp{
			ID:               r.ID,
			RuleName:         r.RuleName,
			DaysBeforeMin:    r.DaysBeforeMin,
			DaysBeforeMax:    r.DaysBeforeMax,
			RefundPercentage: r.RefundPercentage,
			Description:      r.Description,
		}
	}

	// Parse images
	var images []string
	if product.Images != nil {
		_ = json.Unmarshal(product.Images, &images)
	}

	// Get review summary
	reviewSummary, _ := s.reviewSvc.ListReviews(id, nil, 1, 1)

	// Build destination cities
	destCities := parseStringArray(product.DestinationCities)

	detail := &ProductDetailResponse{
		ProductSummaryResponse: ProductSummaryResponse{
			ID:                product.ID,
			ProductNo:         product.ProductNo,
			ProductName:       product.ProductName,
			CoverImage:        product.CoverImage,
			OriginCity:        product.OriginCity,
			DestinationCities: destCities,
			Days:              product.Days,
			Nights:            product.Nights,
			MinPrice:          minPrice,
			ProductGrade:      product.ProductGrade,
			SatisfactionRate:  product.SatisfactionRate,
			OrderCount:        product.OrderCount,
		},
		Summary:           product.Summary,
		Description:       product.Description,
		TransportMode:     product.TransportMode,
		MinGroupSize:      product.MinGroupSize,
		MaxGroupSize:      product.MaxGroupSize,
		FeeIncluded:       product.FeeIncluded,
		FeeExcluded:       product.FeeExcluded,
		BookingNotes:      product.BookingNotes,
		Itinerary:         itinerary,
		CancellationRules: rules,
		Images:            images,
	}

	if reviewSummary != nil {
		detail.ReviewSummary = reviewSummary.Summary
	}

	return detail, nil
}

// GetDepartureCalendar returns departure dates for a product within a month range.
func (s *ProductService) GetDepartureCalendar(productID int64, month string, months int) ([]DepartureCalendarResponse, error) {
	startDate, err := time.Parse("2006-01", month)
	if err != nil {
		return nil, fmt.Errorf("invalid month format: %w", err)
	}
	if months < 1 || months > 6 {
		months = 3
	}
	endDate := startDate.AddDate(0, months, 0)

	// Load product with departures
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return nil, fmt.Errorf("find product: %w", err)
	}

	var result []DepartureCalendarResponse
	for _, d := range product.DepartureDates {
		if d.DepartureDate.Before(startDate) || d.DepartureDate.Equal(endDate) || d.DepartureDate.After(endDate) {
			continue
		}
		avail := d.AvailableStock()
		stockStatus := "sufficient"
		if avail <= 0 {
			stockStatus = "sold_out"
		} else if avail < 10 {
			stockStatus = "tight"
		}

		result = append(result, DepartureCalendarResponse{
			ID:               d.ID,
			DepartureDate:    d.DepartureDate.Format("2006-01-02"),
			ReturnDate:       d.ReturnDate.Format("2006-01-02"),
			AdultPrice:       d.AdultPrice / 100, // cents → yuan
			ChildPrice:       d.ChildPrice / 100,
			InfantPrice:      d.InfantPrice / 100,
			SingleSupplement: d.SingleSupplement / 100,
			AvailableStock:   avail,
			StockStatus:      stockStatus,
			CutoffDays:       d.CutoffDays,
		})
	}

	return result, nil
}

// GetItinerary returns the itinerary for a product.
func (s *ProductService) GetItinerary(productID int64) ([]ItineraryDayResponse, error) {
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return nil, fmt.Errorf("find product: %w", err)
	}

	result := make([]ItineraryDayResponse, len(product.Itineraries))
	for i, it := range product.Itineraries {
		result[i] = ItineraryDayResponse{
			DayNo:       it.DayNo,
			Title:       it.Title,
			Description: it.Description,
			Meals:       it.Meals,
			Hotel:       it.Hotel,
			Transport:   it.Transport,
			Spots:       it.Spots,
		}
	}
	return result, nil
}

// SearchAutocomplete returns search suggestions.
func (s *ProductService) SearchAutocomplete(query string, limit int) ([]SearchSuggestResponse, error) {
	if limit < 1 || limit > 20 {
		limit = 10
	}

	names, err := s.productRepo.SearchSuggest(query, limit)
	if err != nil {
		return nil, fmt.Errorf("search suggest: %w", err)
	}

	result := make([]SearchSuggestResponse, len(names))
	for i, n := range names {
		result[i] = SearchSuggestResponse{
			Text: n,
			Type: "product",
		}
	}
	return result, nil
}

// toProductSummary converts a product model to summary response.
func (s *ProductService) toProductSummary(p model.Product) ProductSummaryResponse {
	destCities := parseStringArray(p.DestinationCities)

	// Get min price from departure dates (if loaded)
	minPrice := 0
	if len(p.DepartureDates) > 0 {
		minPrice = s.getMinPrice(p.DepartureDates)
	}

	return ProductSummaryResponse{
		ID:                p.ID,
		ProductNo:         p.ProductNo,
		ProductName:       p.ProductName,
		CoverImage:        p.CoverImage,
		OriginCity:        p.OriginCity,
		DestinationCities: destCities,
		Days:              p.Days,
		Nights:            p.Nights,
		MinPrice:          minPrice,
		ProductGrade:      p.ProductGrade,
		SatisfactionRate:  p.SatisfactionRate,
		OrderCount:        p.OrderCount,
	}
}

// getMinPrice finds the lowest adult price among departures.
func (s *ProductService) getMinPrice(departures []model.DepartureDate) int {
	min := 0
	for _, d := range departures {
		if d.Status != model.DepartureStatusOpen {
			continue
		}
		priceYuan := d.AdultPrice / 100
		if min == 0 || priceYuan < min {
			min = priceYuan
		}
	}
	return min
}

// parseStringArray parses a JSON raw message to string slice.
func parseStringArray(raw json.RawMessage) []string {
	if raw == nil {
		return nil
	}
	var arr []string
	_ = json.Unmarshal(raw, &arr)
	return arr
}
