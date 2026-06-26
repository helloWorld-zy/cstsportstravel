// Package service provides business logic for the Product domain.
package service

import (
	"encoding/json"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/product/model"
	"github.com/travel-booking/server/internal/product/repository"
)

// Review errors.
var (
	ErrReviewAlreadyExists = errors.New("review already exists for this order")
	ErrOrderNotCompleted   = errors.New("order must be completed before submitting a review")
	ErrInvalidRating       = errors.New("rating must be between 1 and 5")
)

// ReviewService provides business logic for product reviews.
type ReviewService struct {
	reviewRepo *repository.ReviewRepository
	orderDB    *gorm.DB // for checking order status
	logger     *zap.Logger
}

// NewReviewService creates a new ReviewService.
func NewReviewService(reviewRepo *repository.ReviewRepository, logger *zap.Logger) *ReviewService {
	return &ReviewService{
		reviewRepo: reviewRepo,
		logger:     logger,
	}
}

// SetOrderDB sets the order database for order validation.
func (s *ReviewService) SetOrderDB(db *gorm.DB) {
	s.orderDB = db
}

// ReviewResponse represents a single review in API responses.
type ReviewResponse struct {
	ID          int64    `json:"id"`
	UserID      int64    `json:"user_id"`
	Rating      int      `json:"rating"`
	Content     string   `json:"content,omitempty"`
	Images      []string `json:"images,omitempty"`
	IsAnonymous bool     `json:"is_anonymous"`
	CreatedAt   string   `json:"created_at"`
}

// ReviewListResponse holds paginated reviews.
type ReviewListResponse struct {
	Items   []ReviewResponse       `json:"items"`
	Total   int64                  `json:"total"`
	Summary *ReviewSummaryResponse `json:"summary"`
}

// ReviewSummaryResponse holds review statistics.
type ReviewSummaryResponse struct {
	TotalCount        int64            `json:"total_count"`
	AverageRating     float64          `json:"average_rating"`
	RatingDistribution map[string]int64 `json:"rating_distribution"`
}

// ListReviews returns paginated reviews for a product.
func (s *ReviewService) ListReviews(productID int64, rating *int, page, pageSize int) (*ReviewListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	reviews, total, err := s.reviewRepo.FindByProductID(productID, rating, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("find reviews: %w", err)
	}

	// Get summary
	summaryData, err := s.reviewRepo.GetReviewSummary(productID)
	if err != nil {
		s.logger.Warn("failed to get review summary", zap.Int64("product_id", productID), zap.Error(err))
	}

	items := make([]ReviewResponse, len(reviews))
	for i, r := range reviews {
		items[i] = toReviewResponse(r)
	}

	var summary *ReviewSummaryResponse
	if summaryData != nil {
		dist := make(map[string]int64)
		for k, v := range summaryData.RatingDistribution {
			dist[fmt.Sprintf("%d", k)] = v
		}
		summary = &ReviewSummaryResponse{
			TotalCount:         summaryData.TotalCount,
			AverageRating:      summaryData.AverageRating,
			RatingDistribution: dist,
		}
	}

	return &ReviewListResponse{
		Items:   items,
		Total:   total,
		Summary: summary,
	}, nil
}

// toReviewResponse converts a model to API response.
func toReviewResponse(r model.ProductReview) ReviewResponse {
	var images []string
	if r.Images != nil {
		// json.RawMessage → parse to []string
		_ = json.Unmarshal(r.Images, &images)
	}

	displayContent := r.Content
	if r.IsAnonymous {
		// mask reviewer identity but keep content
	}

	return ReviewResponse{
		ID:          r.ID,
		UserID:      r.UserID,
		Rating:      r.Rating,
		Content:     displayContent,
		Images:      images,
		IsAnonymous: r.IsAnonymous,
		CreatedAt:   r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// SubmitReviewInput is the request body for submitting a review.
type SubmitReviewInput struct {
	Rating      int      `json:"rating" binding:"required,min=1,max=5"`
	Content     string   `json:"content"`
	Images      []string `json:"images"`
	IsAnonymous bool     `json:"is_anonymous"`
}

// SubmitReview creates a new product review.
// Validates: user has a completed order for this product, no duplicate review per order.
func (s *ReviewService) SubmitReview(userID, productID, orderID int64, input SubmitReviewInput) (*ReviewResponse, error) {
	// Validate rating
	if input.Rating < 1 || input.Rating > 5 {
		return nil, ErrInvalidRating
	}

	// Validate content length
	if len(input.Content) < 10 {
		return nil, fmt.Errorf("review content must be at least 10 characters")
	}

	// Check if review already exists for this order
	existing, _ := s.reviewRepo.FindByOrderID(orderID)
	if existing != nil {
		return nil, ErrReviewAlreadyExists
	}

	// Validate order is completed and belongs to user
	if s.orderDB != nil {
		var order struct {
			ID          int64  `gorm:"column:id"`
			UserID      int64  `gorm:"column:user_id"`
			OrderStatus string `gorm:"column:order_status"`
		}
		err := s.orderDB.Table("main_order").
			Where("id = ? AND user_id = ?", orderID, userID).
			First(&order).Error
		if err != nil {
			return nil, fmt.Errorf("order not found or access denied")
		}
		if order.OrderStatus != "completed" {
			return nil, ErrOrderNotCompleted
		}
	}

	// Marshal images
	var imagesJSON json.RawMessage
	if len(input.Images) > 0 {
		b, err := json.Marshal(input.Images)
		if err != nil {
			return nil, fmt.Errorf("marshal images: %w", err)
		}
		imagesJSON = b
	}

	// Create review
	review := &model.ProductReview{
		ProductID:   productID,
		UserID:      userID,
		OrderID:     orderID,
		Rating:      input.Rating,
		Content:     input.Content,
		Images:      imagesJSON,
		IsAnonymous: input.IsAnonymous,
	}

	if err := s.reviewRepo.Create(review); err != nil {
		return nil, fmt.Errorf("create review: %w", err)
	}

	s.logger.Info("review submitted",
		zap.Int64("review_id", review.ID),
		zap.Int64("product_id", productID),
		zap.Int64("order_id", orderID),
		zap.Int("rating", input.Rating),
	)

	resp := toReviewResponse(*review)
	return &resp, nil
}
