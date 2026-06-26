package repository

import (
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/product/model"
)

// ReviewRepository provides data access for ProductReview.
type ReviewRepository struct {
	db *gorm.DB
}

// NewReviewRepository creates a new ReviewRepository.
func NewReviewRepository(db *gorm.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

// FindByProductID returns reviews for a product with pagination.
func (r *ReviewRepository) FindByProductID(productID int64, rating *int, page, pageSize int) ([]model.ProductReview, int64, error) {
	query := r.db.Model(&model.ProductReview{}).Where("product_id = ?", productID)

	if rating != nil {
		query = query.Where("rating = ?", *rating)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var reviews []model.ProductReview
	offset := (page - 1) * pageSize
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&reviews).Error
	return reviews, total, err
}

// ReviewSummary holds aggregated review statistics.
type ReviewSummary struct {
	TotalCount       int64            `json:"total_count"`
	AverageRating    float64          `json:"average_rating"`
	RatingDistribution map[int]int64  `json:"rating_distribution"`
}

// GetReviewSummary returns review statistics for a product.
func (r *ReviewRepository) GetReviewSummary(productID int64) (*ReviewSummary, error) {
	var summary ReviewSummary

	// Total count and average
	var result struct {
		Count   int64
		Average *float64
	}
	err := r.db.Model(&model.ProductReview{}).
		Where("product_id = ?", productID).
		Select("COUNT(*) as count, AVG(rating) as average").
		Scan(&result).Error
	if err != nil {
		return nil, err
	}
	summary.TotalCount = result.Count
	if result.Average != nil {
		summary.AverageRating = *result.Average
	}

	// Rating distribution
	summary.RatingDistribution = make(map[int]int64)
	var rows []struct {
		Rating int
		Count  int64
	}
	err = r.db.Model(&model.ProductReview{}).
		Where("product_id = ?", productID).
		Select("rating, COUNT(*) as count").
		Group("rating").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		summary.RatingDistribution[row.Rating] = row.Count
	}

	return &summary, nil
}

// FindByOrderID returns a review by order ID (for duplicate check).
func (r *ReviewRepository) FindByOrderID(orderID int64) (*model.ProductReview, error) {
	var review model.ProductReview
	err := r.db.Where("order_id = ?", orderID).First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

// Create creates a new product review.
func (r *ReviewRepository) Create(review *model.ProductReview) error {
	return r.db.Create(review).Error
}
