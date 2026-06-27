// Package service provides business logic for the Admin domain.
package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	adminrepo "github.com/travel-booking/server/internal/admin/repository"
	productmodel "github.com/travel-booking/server/internal/product/model"
)

// Sentinel errors for admin product operations.
var (
	ErrProductNotFound       = errors.New("product not found")
	ErrInvalidStatus         = errors.New("invalid product status for this operation")
	ErrMissingRequiredFields = errors.New("missing required fields")
	ErrKeyFieldsChanged      = errors.New("key fields changed, requires re-review")
)

// Key fields that trigger re-review when changed on an approved product.
var keyReviewFields = []string{
	"days", "nights", "transport_mode",
	"fee_included", "fee_excluded",
}

// AdminProductService provides business logic for admin product management.
type AdminProductService struct {
	productRepo *adminrepo.AdminProductRepository
	logger      *zap.Logger
}

// NewAdminProductService creates a new AdminProductService.
func NewAdminProductService(productRepo *adminrepo.AdminProductRepository, logger *zap.Logger) *AdminProductService {
	return &AdminProductService{
		productRepo: productRepo,
		logger:      logger,
	}
}

// --- Request/Response DTOs ---

// CreateProductRequest is the request body for creating a product.
type CreateProductRequest struct {
	ProductName       string   `json:"product_name" binding:"required,max=200"`
	CategoryID        int64    `json:"category_id" binding:"required"`
	OriginCity        string   `json:"origin_city" binding:"required"`
	DestinationCities []string `json:"destination_cities" binding:"required,min=1"`
	Days              int      `json:"days" binding:"required,min=1"`
	Nights            int      `json:"nights" binding:"min=0"`
	TransportMode     string   `json:"transport_mode"`
	ProductGrade      string   `json:"product_grade"`
	MinGroupSize      int      `json:"min_group_size"`
	MaxGroupSize      int      `json:"max_group_size"`
	CoverImage        string   `json:"cover_image"`
	Images            []string `json:"images"`
	Summary           string   `json:"summary" binding:"max=500"`
	Description        string   `json:"description"`
	FeeIncluded       string   `json:"fee_included"`
	FeeExcluded       string   `json:"fee_excluded"`
	BookingNotes      string   `json:"booking_notes"`
}

// UpdateProductRequest is the request body for updating a product.
type UpdateProductRequest struct {
	ProductName       *string   `json:"product_name"`
	CategoryID        *int64    `json:"category_id"`
	OriginCity        *string   `json:"origin_city"`
	DestinationCities []string  `json:"destination_cities"`
	Days              *int      `json:"days"`
	Nights            *int      `json:"nights"`
	TransportMode     *string   `json:"transport_mode"`
	ProductGrade      *string   `json:"product_grade"`
	MinGroupSize      *int      `json:"min_group_size"`
	MaxGroupSize      *int      `json:"max_group_size"`
	CoverImage        *string   `json:"cover_image"`
	Images            []string  `json:"images"`
	Summary           *string   `json:"summary"`
	Description        *string   `json:"description"`
	FeeIncluded       *string   `json:"fee_included"`
	FeeExcluded       *string   `json:"fee_excluded"`
	BookingNotes      *string   `json:"booking_notes"`
}

// AdminProductResponse is the admin product detail response.
type AdminProductResponse struct {
	ID                int64    `json:"id"`
	ProductNo         string   `json:"product_no"`
	ProductName       string   `json:"product_name"`
	CategoryID        int64    `json:"category_id"`
	CategoryName      string   `json:"category_name,omitempty"`
	OriginCity        string   `json:"origin_city"`
	DestinationCities []string `json:"destination_cities"`
	Days              int      `json:"days"`
	Nights            int      `json:"nights"`
	TransportMode     string   `json:"transport_mode,omitempty"`
	ProductGrade      string   `json:"product_grade,omitempty"`
	MinGroupSize      int      `json:"min_group_size"`
	MaxGroupSize      int      `json:"max_group_size"`
	CoverImage        string   `json:"cover_image,omitempty"`
	Images            []string `json:"images,omitempty"`
	Summary           string   `json:"summary,omitempty"`
	Description        string   `json:"description,omitempty"`
	FeeIncluded       string   `json:"fee_included,omitempty"`
	FeeExcluded       string   `json:"fee_excluded,omitempty"`
	BookingNotes      string   `json:"booking_notes,omitempty"`
	Status            string   `json:"status"`
	RejectReason      string   `json:"reject_reason,omitempty"`
	SupplierID        *int64   `json:"supplier_id,omitempty"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
}

// PaginatedAdminProductsResponse is the paginated product list for admin.
type PaginatedAdminProductsResponse struct {
	Items    []AdminProductResponse `json:"items"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

// AdminListProductsRequest holds query parameters for admin product listing.
type AdminListProductsRequest struct {
	Status      string `form:"status"`
	Keyword     string `form:"keyword"`
	Destination string `form:"destination"`
	SupplierID  *int64 `form:"supplier_id"`
	Page        int    `form:"page" binding:"min=1"`
	PageSize    int    `form:"page_size" binding:"min=1,max=100"`
}

// --- Service Methods ---

// CreateProduct creates a new product in draft status.
func (s *AdminProductService) CreateProduct(supplierID *int64, req CreateProductRequest) (*AdminProductResponse, error) {
	// Validate required fields
	if req.ProductName == "" || req.CategoryID == 0 || req.OriginCity == "" || len(req.DestinationCities) == 0 || req.Days < 1 {
		return nil, ErrMissingRequiredFields
	}

	// Generate product number: DOM-{category_id}-{YYYYMMDD}-{seq}
	dateStr := time.Now().Format("20060102")
	seq, err := s.productRepo.NextProductSeq(dateStr)
	if err != nil {
		return nil, fmt.Errorf("generate product seq: %w", err)
	}
	productNo := fmt.Sprintf("DOM-%d-%s-%04d", req.CategoryID, dateStr, seq)

	// Set defaults
	nights := req.Nights
	if nights == 0 {
		nights = req.Days - 1
	}
	minGroup := req.MinGroupSize
	if minGroup == 0 {
		minGroup = 2
	}
	maxGroup := req.MaxGroupSize
	if maxGroup == 0 {
		maxGroup = 50
	}

	// Marshal JSONB fields
	destCitiesJSON, _ := json.Marshal(req.DestinationCities)
	var imagesJSON json.RawMessage
	if req.Images != nil {
		imagesJSON, _ = json.Marshal(req.Images)
	}

	product := &productmodel.Product{
		ProductNo:         productNo,
		ProductName:       req.ProductName,
		CategoryID:        req.CategoryID,
		ProductType:       "group_tour",
		OriginCity:        req.OriginCity,
		DestinationCities: destCitiesJSON,
		Days:              req.Days,
		Nights:            nights,
		TransportMode:     req.TransportMode,
		MinGroupSize:      minGroup,
		MaxGroupSize:      maxGroup,
		ProductGrade:      req.ProductGrade,
		CoverImage:        req.CoverImage,
		Images:            imagesJSON,
		Summary:           req.Summary,
		Description:       req.Description,
		FeeIncluded:       req.FeeIncluded,
		FeeExcluded:       req.FeeExcluded,
		BookingNotes:      req.BookingNotes,
		Status:            productmodel.ProductStatusDraft,
		SupplierID:        supplierID,
	}

	if err := s.productRepo.Create(product); err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}

	s.logger.Info("product created",
		zap.Int64("id", product.ID),
		zap.String("product_no", productNo),
	)

	return s.toAdminProductResponse(product), nil
}

// UpdateProduct updates a product. If the product is approved and key fields change,
// status transitions to change_pending_review.
func (s *AdminProductService) UpdateProduct(id int64, req UpdateProductRequest) (*AdminProductResponse, error) {
	product, err := s.productRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("find product: %w", err)
	}

	// Only draft, approved, suspended, and change_pending_review products can be edited
	switch product.Status {
	case productmodel.ProductStatusDraft,
		productmodel.ProductStatusApproved,
		productmodel.ProductStatusSuspended,
		productmodel.ProductStatusChangePendingReview:
		// OK to edit
	default:
		return nil, ErrInvalidStatus
	}

	// Track if key fields changed (for approved products)
	keyFieldChanged := false

	// Apply updates
	if req.ProductName != nil {
		product.ProductName = *req.ProductName
	}
	if req.CategoryID != nil {
		product.CategoryID = *req.CategoryID
	}
	if req.OriginCity != nil {
		product.OriginCity = *req.OriginCity
	}
	if req.DestinationCities != nil {
		product.DestinationCities, _ = json.Marshal(req.DestinationCities)
	}
	if req.Days != nil {
		if product.Days != *req.Days {
			keyFieldChanged = true
		}
		product.Days = *req.Days
	}
	if req.Nights != nil {
		if product.Nights != *req.Nights {
			keyFieldChanged = true
		}
		product.Nights = *req.Nights
	}
	if req.TransportMode != nil {
		if product.TransportMode != *req.TransportMode {
			keyFieldChanged = true
		}
		product.TransportMode = *req.TransportMode
	}
	if req.ProductGrade != nil {
		product.ProductGrade = *req.ProductGrade
	}
	if req.MinGroupSize != nil {
		product.MinGroupSize = *req.MinGroupSize
	}
	if req.MaxGroupSize != nil {
		product.MaxGroupSize = *req.MaxGroupSize
	}
	if req.CoverImage != nil {
		product.CoverImage = *req.CoverImage
	}
	if req.Images != nil {
		product.Images, _ = json.Marshal(req.Images)
	}
	if req.Summary != nil {
		product.Summary = *req.Summary
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.FeeIncluded != nil {
		if product.FeeIncluded != *req.FeeIncluded {
			keyFieldChanged = true
		}
		product.FeeIncluded = *req.FeeIncluded
	}
	if req.FeeExcluded != nil {
		if product.FeeExcluded != *req.FeeExcluded {
			keyFieldChanged = true
		}
		product.FeeExcluded = *req.FeeExcluded
	}
	if req.BookingNotes != nil {
		product.BookingNotes = *req.BookingNotes
	}

	// If approved product had key fields changed, move to change_pending_review
	if product.Status == productmodel.ProductStatusApproved && keyFieldChanged {
		product.Status = productmodel.ProductStatusChangePendingReview
		s.logger.Info("approved product key fields changed, moving to change_pending_review",
			zap.Int64("product_id", product.ID),
		)
	}

	if err := s.productRepo.Update(product); err != nil {
		return nil, fmt.Errorf("update product: %w", err)
	}

	return s.toAdminProductResponse(product), nil
}

// SubmitForReview transitions a product from draft to pending_review.
func (s *AdminProductService) SubmitForReview(id int64) (*AdminProductResponse, error) {
	product, err := s.productRepo.FindByIDBasic(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("find product: %w", err)
	}

	// Can only submit draft or suspended products
	if product.Status != productmodel.ProductStatusDraft && product.Status != productmodel.ProductStatusSuspended {
		return nil, ErrInvalidStatus
	}

	// Validate required fields are present
	if product.ProductName == "" || product.CategoryID == 0 || product.OriginCity == "" || product.Days < 1 {
		return nil, ErrMissingRequiredFields
	}

	if err := s.productRepo.UpdateStatus(id, productmodel.ProductStatusPendingReview, ""); err != nil {
		return nil, fmt.Errorf("submit for review: %w", err)
	}

	s.logger.Info("product submitted for review", zap.Int64("product_id", id))

	product.Status = productmodel.ProductStatusPendingReview
	return s.toAdminProductResponse(product), nil
}

// ApproveProduct approves a pending_review or change_pending_review product.
func (s *AdminProductService) ApproveProduct(id int64, note string) (*AdminProductResponse, error) {
	product, err := s.productRepo.FindByIDBasic(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("find product: %w", err)
	}

	if product.Status != productmodel.ProductStatusPendingReview && product.Status != productmodel.ProductStatusChangePendingReview {
		return nil, ErrInvalidStatus
	}

	if err := s.productRepo.UpdateStatus(id, productmodel.ProductStatusApproved, ""); err != nil {
		return nil, fmt.Errorf("approve product: %w", err)
	}

	s.logger.Info("product approved",
		zap.Int64("product_id", id),
		zap.String("note", note),
	)

	product.Status = productmodel.ProductStatusApproved
	product.RejectReason = ""
	return s.toAdminProductResponse(product), nil
}

// RejectProduct rejects a pending_review product and returns it to draft.
func (s *AdminProductService) RejectProduct(id int64, reason string) (*AdminProductResponse, error) {
	if reason == "" {
		return nil, ErrMissingRequiredFields
	}

	product, err := s.productRepo.FindByIDBasic(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("find product: %w", err)
	}

	if product.Status != productmodel.ProductStatusPendingReview {
		return nil, ErrInvalidStatus
	}

	if err := s.productRepo.UpdateStatus(id, productmodel.ProductStatusDraft, reason); err != nil {
		return nil, fmt.Errorf("reject product: %w", err)
	}

	s.logger.Info("product rejected",
		zap.Int64("product_id", id),
		zap.String("reason", reason),
	)

	product.Status = productmodel.ProductStatusDraft
	product.RejectReason = reason
	return s.toAdminProductResponse(product), nil
}

// SuspendProduct suspends an approved product.
func (s *AdminProductService) SuspendProduct(id int64, reason string) (*AdminProductResponse, error) {
	product, err := s.productRepo.FindByIDBasic(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("find product: %w", err)
	}

	if product.Status != productmodel.ProductStatusApproved {
		return nil, ErrInvalidStatus
	}

	if err := s.productRepo.UpdateStatus(id, productmodel.ProductStatusSuspended, reason); err != nil {
		return nil, fmt.Errorf("suspend product: %w", err)
	}

	s.logger.Info("product suspended",
		zap.Int64("product_id", id),
		zap.String("reason", reason),
	)

	product.Status = productmodel.ProductStatusSuspended
	return s.toAdminProductResponse(product), nil
}

// ListProducts returns a paginated admin product list with filters.
func (s *AdminProductService) ListProducts(req AdminListProductsRequest) (*PaginatedAdminProductsResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}

	filter := adminrepo.AdminProductFilter{
		Status:      req.Status,
		Keyword:     req.Keyword,
		Destination: req.Destination,
		SupplierID:  req.SupplierID,
	}

	products, total, err := s.productRepo.FindWithFilters(filter, req.Page, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}

	items := make([]AdminProductResponse, len(products))
	for i, p := range products {
		items[i] = *s.toAdminProductResponse(&p)
	}

	return &PaginatedAdminProductsResponse{
		Items:    items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetProduct returns the full admin product detail.
func (s *AdminProductService) GetProduct(id int64) (*AdminProductResponse, error) {
	product, err := s.productRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("find product: %w", err)
	}
	return s.toAdminProductResponse(product), nil
}

// toAdminProductResponse converts a product model to admin response DTO.
func (s *AdminProductService) toAdminProductResponse(p *productmodel.Product) *AdminProductResponse {
	destCities := parseJSONStringArray(p.DestinationCities)
	images := parseJSONStringArray(p.Images)

	resp := &AdminProductResponse{
		ID:                p.ID,
		ProductNo:         p.ProductNo,
		ProductName:       p.ProductName,
		CategoryID:        p.CategoryID,
		OriginCity:        p.OriginCity,
		DestinationCities: destCities,
		Days:              p.Days,
		Nights:            p.Nights,
		TransportMode:     p.TransportMode,
		ProductGrade:      p.ProductGrade,
		MinGroupSize:      p.MinGroupSize,
		MaxGroupSize:      p.MaxGroupSize,
		CoverImage:        p.CoverImage,
		Images:            images,
		Summary:           p.Summary,
		Description:       p.Description,
		FeeIncluded:       p.FeeIncluded,
		FeeExcluded:       p.FeeExcluded,
		BookingNotes:      p.BookingNotes,
		Status:            p.Status,
		RejectReason:      p.RejectReason,
		SupplierID:        p.SupplierID,
		CreatedAt:         p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         p.UpdatedAt.Format(time.RFC3339),
	}

	if p.Category != nil {
		resp.CategoryName = p.Category.Name
	}

	return resp
}

// parseJSONStringArray parses a JSON raw message to string slice.
func parseJSONStringArray(raw json.RawMessage) []string {
	if raw == nil {
		return nil
	}
	var arr []string
	_ = json.Unmarshal(raw, &arr)
	return arr
}
