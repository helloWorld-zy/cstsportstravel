// Package service provides business logic for the Admin domain.
package service

import (
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	adminrepo "github.com/travel-booking/server/internal/admin/repository"
	productmodel "github.com/travel-booking/server/internal/product/model"
)

// ReviewService provides business logic for product review workflow.
type ReviewService struct {
	productRepo *adminrepo.AdminProductRepository
	logger      *zap.Logger
}

// NewReviewService creates a new ReviewService.
func NewReviewService(productRepo *adminrepo.AdminProductRepository, logger *zap.Logger) *ReviewService {
	return &ReviewService{
		productRepo: productRepo,
		logger:      logger,
	}
}

// --- Request/Response DTOs ---

// ReviewListRequest holds query parameters for review queue listing.
type ReviewListRequest struct {
	Status   string `form:"status"`
	Keyword  string `form:"keyword"`
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
}

// ReviewListResponse is the paginated review list.
type ReviewListResponse struct {
	Items    []AdminProductResponse `json:"items"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

// ReviewActionRequest is the request for approve/reject actions.
type ReviewActionRequest struct {
	Note string `json:"note"`
}

// RejectReviewRequest is the request for rejecting a review.
type RejectReviewRequest struct {
	Reason string `json:"reason" binding:"required,max=500"`
}

// --- Service Methods ---

// ListPendingReviews returns products pending review (pending_review or change_pending_review).
func (s *ReviewService) ListPendingReviews(req ReviewListRequest) (*ReviewListResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}

	// Default to showing both pending_review and change_pending_review
	status := req.Status
	if status == "" {
		status = "pending"
	}

	// We'll query for both statuses when "pending" is specified
	filter := adminrepo.AdminProductFilter{
		Keyword: req.Keyword,
	}

	// For "pending" status, we need to handle specially
	// Since the repo only supports single status filter, we'll use a custom approach
	products, total, err := s.findPendingReviews(filter, status, req.Page, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("list pending reviews: %w", err)
	}

	items := make([]AdminProductResponse, len(products))
	for i, p := range products {
		items[i] = *toAdminProductResponseFromModel(&p)
	}

	return &ReviewListResponse{
		Items:    items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// findPendingReviews is a custom query for pending review products.
func (s *ReviewService) findPendingReviews(filter adminrepo.AdminProductFilter, status string, page, pageSize int) ([]productmodel.Product, int64, error) {
	db := s.productRepo.DB()

	query := db.Model(&productmodel.Product{})

	switch status {
	case "pending":
		query = query.Where("status IN ?", []string{
			productmodel.ProductStatusPendingReview,
			productmodel.ProductStatusChangePendingReview,
		})
	case productmodel.ProductStatusPendingReview, productmodel.ProductStatusChangePendingReview:
		query = query.Where("status = ?", status)
	default:
		query = query.Where("status IN ?", []string{
			productmodel.ProductStatusPendingReview,
			productmodel.ProductStatusChangePendingReview,
		})
	}

	if filter.Keyword != "" {
		kw := "%" + filter.Keyword + "%"
		query = query.Where("(product_name ILIKE ? OR product_no ILIKE ?)", kw, kw)
	}

	if filter.SupplierID != nil {
		query = query.Where("supplier_id = ?", *filter.SupplierID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var products []productmodel.Product
	err := query.
		Preload("Category").
		Order("updated_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&products).Error

	return products, total, err
}

// ApproveReview approves a product review (pending_review → approved, or change_pending_review → approved).
func (s *ReviewService) ApproveReview(productID int64, operatorID int64, note string) (*AdminProductResponse, error) {
	product, err := s.productRepo.FindByIDBasic(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("find product: %w", err)
	}

	// Can approve pending_review or change_pending_review
	if product.Status != productmodel.ProductStatusPendingReview &&
		product.Status != productmodel.ProductStatusChangePendingReview {
		return nil, ErrInvalidStatus
	}

	prevStatus := product.Status

	if err := s.productRepo.UpdateStatus(productID, productmodel.ProductStatusApproved, ""); err != nil {
		return nil, fmt.Errorf("approve product: %w", err)
	}

	s.logger.Info("product review approved",
		zap.Int64("product_id", productID),
		zap.Int64("operator_id", operatorID),
		zap.String("prev_status", prevStatus),
		zap.String("note", note),
	)

	product.Status = productmodel.ProductStatusApproved
	product.RejectReason = ""
	return toAdminProductResponseFromModel(product), nil
}

// RejectReview rejects a pending_review product and returns it to draft.
func (s *ReviewService) RejectReview(productID int64, operatorID int64, reason string) (*AdminProductResponse, error) {
	if reason == "" {
		return nil, ErrMissingRequiredFields
	}

	product, err := s.productRepo.FindByIDBasic(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("find product: %w", err)
	}

	// Can only reject pending_review products
	if product.Status != productmodel.ProductStatusPendingReview {
		return nil, ErrInvalidStatus
	}

	if err := s.productRepo.UpdateStatus(productID, productmodel.ProductStatusDraft, reason); err != nil {
		return nil, fmt.Errorf("reject product: %w", err)
	}

	s.logger.Info("product review rejected",
		zap.Int64("product_id", productID),
		zap.Int64("operator_id", operatorID),
		zap.String("reason", reason),
	)

	product.Status = productmodel.ProductStatusDraft
	product.RejectReason = reason
	return toAdminProductResponseFromModel(product), nil
}

// RejectChangeReview rejects a change_pending_review, keeping the product in approved status.
func (s *ReviewService) RejectChangeReview(productID int64, operatorID int64, reason string) (*AdminProductResponse, error) {
	product, err := s.productRepo.FindByIDBasic(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("find product: %w", err)
	}

	// Can only reject change_pending_review products
	if product.Status != productmodel.ProductStatusChangePendingReview {
		return nil, ErrInvalidStatus
	}

	// Revert to approved (keep the product online with original content)
	if err := s.productRepo.UpdateStatus(productID, productmodel.ProductStatusApproved, reason); err != nil {
		return nil, fmt.Errorf("reject change review: %w", err)
	}

	s.logger.Info("product change review rejected",
		zap.Int64("product_id", productID),
		zap.Int64("operator_id", operatorID),
		zap.String("reason", reason),
	)

	product.Status = productmodel.ProductStatusApproved
	product.RejectReason = reason
	return toAdminProductResponseFromModel(product), nil
}

// toAdminProductResponseFromModel converts a product model to admin response DTO.
// This is a standalone function to avoid circular dependency with the service package.
func toAdminProductResponseFromModel(p *productmodel.Product) *AdminProductResponse {
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
		CreatedAt:         p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if p.Category != nil {
		resp.CategoryName = p.Category.Name
	}

	return resp
}
