package handler

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/product/model"
	"github.com/travel-booking/server/internal/product/repository"
	"github.com/travel-booking/server/internal/product/service"
)

// HomepageHandler handles HTTP requests for homepage data.
type HomepageHandler struct {
	productRepo  *repository.ProductRepository
	categoryRepo *repository.CategoryRepository
	logger       *zap.Logger
}

// NewHomepageHandler creates a new HomepageHandler.
func NewHomepageHandler(
	productRepo *repository.ProductRepository,
	categoryRepo *repository.CategoryRepository,
	logger *zap.Logger,
) *HomepageHandler {
	return &HomepageHandler{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		logger:       logger,
	}
}

// Banner represents a homepage banner.
type Banner struct {
	ID       int64  `json:"id"`
	ImageURL string `json:"image_url"`
	Title    string `json:"title"`
	Link     string `json:"link"`
	SortOrder int   `json:"sort_order"`
}

// HomepageResponse is the homepage data payload.
type HomepageResponse struct {
	Banners            []Banner                        `json:"banners"`
	Categories         []CategoryResponse              `json:"categories"`
	PopularDestinations []PopularDestinationResponse    `json:"popular_destinations"`
	RecommendedProducts []service.ProductSummaryResponse `json:"recommended_products"`
}

// CategoryResponse is a category item.
type CategoryResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	IconURL  string `json:"icon_url,omitempty"`
	ParentID *int64 `json:"parent_id,omitempty"`
}

// PopularDestinationResponse is a popular destination.
type PopularDestinationResponse struct {
	Name        string `json:"name"`
	ImageURL    string `json:"image_url,omitempty"`
	ProductCount int   `json:"product_count"`
	MinPrice    int    `json:"min_price"`
}

// GetHomepageData handles GET /api/v1/homepage.
func (h *HomepageHandler) GetHomepageData(c *gin.Context) {
	// Banners are static for MVP (could be from DB in future)
	banners := []Banner{
		{ID: 1, ImageURL: "/static/images/banner1.jpg", Title: "暑期特惠·云南6日游", Link: "/products?destination=云南", SortOrder: 1},
		{ID: 2, ImageURL: "/static/images/banner2.jpg", Title: "亲子游·北京5日研学之旅", Link: "/products?destination=北京", SortOrder: 2},
		{ID: 3, ImageURL: "/static/images/banner3.jpg", Title: "海岛度假·海南三亚4日游", Link: "/products?destination=海南", SortOrder: 3},
	}

	// Categories
	categories, err := h.categoryRepo.FindAll()
	if err != nil {
		h.logger.Error("failed to get categories", zap.Error(err))
		categories = []model.Category{}
	}
	catResp := make([]CategoryResponse, len(categories))
	for i, c := range categories {
		catResp[i] = CategoryResponse{
			ID:       c.ID,
			Name:     c.Name,
			IconURL:  c.IconURL,
			ParentID: c.ParentID,
		}
	}

	// Popular destinations (static for MVP)
	popular := []PopularDestinationResponse{
		{Name: "云南", ProductCount: 25, MinPrice: 2999},
		{Name: "海南", ProductCount: 18, MinPrice: 1999},
		{Name: "北京", ProductCount: 20, MinPrice: 2599},
		{Name: "四川", ProductCount: 15, MinPrice: 3299},
		{Name: "广西", ProductCount: 12, MinPrice: 2199},
	}

	// Recommended products
	products, err := h.productRepo.FindRecommended(10)
	if err != nil {
		h.logger.Error("failed to get recommended products", zap.Error(err))
		products = []model.Product{}
	}
	recResp := make([]service.ProductSummaryResponse, len(products))
	for i, p := range products {
		destCities := parseStringArraySafe(p.DestinationCities)
		recResp[i] = service.ProductSummaryResponse{
			ID:                p.ID,
			ProductNo:         p.ProductNo,
			ProductName:       p.ProductName,
			CoverImage:        p.CoverImage,
			OriginCity:        p.OriginCity,
			DestinationCities: destCities,
			Days:              p.Days,
			Nights:            p.Nights,
			ProductGrade:      p.ProductGrade,
			SatisfactionRate:  p.SatisfactionRate,
			OrderCount:        p.OrderCount,
		}
	}

	response.OK(c, HomepageResponse{
		Banners:             banners,
		Categories:          catResp,
		PopularDestinations: popular,
		RecommendedProducts: recResp,
	})
}

func parseStringArraySafe(raw json.RawMessage) []string {
	if raw == nil {
		return nil
	}
	var arr []string
	_ = json.Unmarshal(raw, &arr)
	return arr
}
