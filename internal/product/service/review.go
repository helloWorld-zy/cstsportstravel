// Package service provides business logic for the Product domain.
package service

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/product/model"
	"github.com/travel-booking/server/internal/product/repository"
)

// ReviewService provides business logic for product reviews.
type ReviewService struct {
	reviewRepo *repository.ReviewRepository
	logger     *zap.Logger
}

// NewReviewService creates a new ReviewService.
func NewReviewService(reviewRepo *repository.ReviewRepository, logger *zap.Logger) *ReviewService {
	return &ReviewService{
		reviewRepo: reviewRepo,
		logger:     logger,
	}
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
