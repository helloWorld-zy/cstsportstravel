// Package service provides business logic for the Marketing domain.
package service

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/marketing/domain"
	"github.com/travel-booking/server/internal/marketing/repository"
)

// ActivityEngine provides promotion activity calculation and matching.
type ActivityEngine struct {
	activityRepo *repository.PromotionActivityRepository
	db           *gorm.DB
	logger       *zap.Logger
}

// NewActivityEngine creates a new ActivityEngine.
func NewActivityEngine(
	activityRepo *repository.PromotionActivityRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *ActivityEngine {
	return &ActivityEngine{
		activityRepo: activityRepo,
		db:           db,
		logger:       logger,
	}
}

// DiscountResult represents the result of activity discount calculation.
type DiscountResult struct {
	ActivityID     int64   `json:"activity_id"`
	ActivityName   string  `json:"activity_name"`
	ActivityType   string  `json:"activity_type"`
	DiscountAmount float64 `json:"discount_amount"`
	FlashPrice     float64 `json:"flash_price,omitempty"`     // For flash sale
	Message        string  `json:"message"`
}

// CalculateFlashSaleDiscount calculates the discount for a flash sale activity.
func (e *ActivityEngine) CalculateFlashSaleDiscount(
	activity *domain.PromotionActivity,
	originalPrice float64,
) (*DiscountResult, error) {
	rule, err := activity.ParseFlashSaleRule()
	if err != nil {
		return nil, err
	}

	discount := originalPrice - rule.FlashPrice
	if discount < 0 {
		discount = 0
	}

	return &DiscountResult{
		ActivityID:     activity.ID,
		ActivityName:   activity.ActivityName,
		ActivityType:   activity.ActivityType,
		DiscountAmount: discount,
		FlashPrice:     rule.FlashPrice,
		Message:        "Flash sale discount applied",
	}, nil
}

// CalculateFullReductionDiscount calculates the discount for a full-reduction activity.
func (e *ActivityEngine) CalculateFullReductionDiscount(
	activity *domain.PromotionActivity,
	orderAmount float64,
) (*DiscountResult, error) {
	rule, err := activity.ParseFullReductionRule()
	if err != nil {
		return nil, err
	}

	// Find the best applicable tier (highest discount for the order amount)
	var bestDiscount float64
	for _, tier := range rule.Tiers {
		if orderAmount >= tier.Threshold && tier.Discount > bestDiscount {
			bestDiscount = tier.Discount
		}
	}

	if bestDiscount == 0 {
		return &DiscountResult{
			ActivityID:   activity.ID,
			ActivityName: activity.ActivityName,
			ActivityType: activity.ActivityType,
			Message:      "Order does not meet any reduction tier",
		}, nil
	}

	return &DiscountResult{
		ActivityID:     activity.ID,
		ActivityName:   activity.ActivityName,
		ActivityType:   activity.ActivityType,
		DiscountAmount: bestDiscount,
		Message:        "Full reduction discount applied",
	}, nil
}

// CalculateEarlyBirdDiscount calculates the discount for an early-bird activity.
func (e *ActivityEngine) CalculateEarlyBirdDiscount(
	activity *domain.PromotionActivity,
	orderAmount float64,
	departureDate time.Time,
) (*DiscountResult, error) {
	rule, err := activity.ParseEarlyBirdRule()
	if err != nil {
		return nil, err
	}

	// Calculate days before departure
	now := time.Now()
	daysBefore := int(departureDate.Sub(now).Hours() / 24)
	if daysBefore < 0 {
		return &DiscountResult{
			ActivityID:   activity.ID,
			ActivityName: activity.ActivityName,
			ActivityType: activity.ActivityType,
			Message:      "Departure date has passed",
		}, nil
	}

	// Find the best matching tier (most days before departure that applies)
	var bestRate float64
	var bestDays int
	for _, tier := range rule.Tiers {
		if daysBefore >= tier.DaysBeforeDeparture && tier.DaysBeforeDeparture > bestDays {
			bestRate = tier.Rate
			bestDays = tier.DaysBeforeDeparture
		}
	}

	if bestRate == 0 || bestRate >= 100 {
		return &DiscountResult{
			ActivityID:   activity.ID,
			ActivityName: activity.ActivityName,
			ActivityType: activity.ActivityType,
			Message:      "No early bird tier matches",
		}, nil
	}

	discount := orderAmount * (100 - bestRate) / 100

	return &DiscountResult{
		ActivityID:     activity.ID,
		ActivityName:   activity.ActivityName,
		ActivityType:   activity.ActivityType,
		DiscountAmount: discount,
		Message:        "Early bird discount applied",
	}, nil
}

// CalculateBestDiscount finds the best discount across all active activities for a product.
func (e *ActivityEngine) CalculateBestDiscount(
	tenantID, productID int64,
	orderAmount float64,
	departureDate *time.Time,
) (*DiscountResult, error) {
	now := time.Now()
	activities, err := e.activityRepo.FindActiveByProduct(tenantID, productID, now)
	if err != nil {
		return nil, err
	}

	var bestResult *DiscountResult
	for _, activity := range activities {
		var result *DiscountResult

		switch activity.ActivityType {
		case domain.ActivityTypeFlashSale:
			result, err = e.CalculateFlashSaleDiscount(&activity, orderAmount)
		case domain.ActivityTypeFullReduction:
			result, err = e.CalculateFullReductionDiscount(&activity, orderAmount)
		case domain.ActivityTypeEarlyBird:
			if departureDate != nil {
				result, err = e.CalculateEarlyBirdDiscount(&activity, orderAmount, *departureDate)
			}
		}

		if err != nil {
			e.logger.Warn("failed to calculate activity discount",
				zap.Int64("activity_id", activity.ID),
				zap.Error(err))
			continue
		}

		if result != nil && result.DiscountAmount > 0 {
			if bestResult == nil || result.DiscountAmount > bestResult.DiscountAmount {
				bestResult = result
			}
		}
	}

	return bestResult, nil
}
