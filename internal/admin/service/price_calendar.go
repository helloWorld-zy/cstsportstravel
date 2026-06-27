// Package service provides business logic for the Admin domain.
package service

import (
	"errors"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	adminrepo "github.com/travel-booking/server/internal/admin/repository"
	productmodel "github.com/travel-booking/server/internal/product/model"
)

// PriceCalendarService provides business logic for price calendar management.
type PriceCalendarService struct {
	productRepo *adminrepo.AdminProductRepository
	logger      *zap.Logger
}

// NewPriceCalendarService creates a new PriceCalendarService.
func NewPriceCalendarService(productRepo *adminrepo.AdminProductRepository, logger *zap.Logger) *PriceCalendarService {
	return &PriceCalendarService{
		productRepo: productRepo,
		logger:      logger,
	}
}

// --- Request/Response DTOs ---

// BatchPriceUpdateRequest is the request body for batch price update.
type BatchPriceUpdateRequest struct {
	Mode          string   `json:"mode" binding:"required,oneof=fixed percentage amount follow"`
	TargetDates   []string `json:"target_dates" binding:"required,min=1"`
	AdultPrice    *int     `json:"adult_price"`
	ChildPrice    *int     `json:"child_price"`
	InfantPrice   *int     `json:"infant_price"`
	SingleSupplement *int  `json:"single_supplement"`
	Percentage    *float64 `json:"percentage"`
	Amount        *int     `json:"amount"`
	ReferenceDates []string `json:"reference_dates"`
}

// BatchPriceUpdateResponse is the response for batch price update.
type BatchPriceUpdateResponse struct {
	UpdatedCount int `json:"updated_count"`
}

// SetDailyPriceRequest is the request for setting a single day's price.
type SetDailyPriceRequest struct {
	DepartureDate    string `json:"departure_date" binding:"required"`
	ReturnDate       string `json:"return_date" binding:"required"`
	AdultPrice       int    `json:"adult_price" binding:"required,min=1"`
	ChildPrice       int    `json:"child_price" binding:"required,min=0"`
	InfantPrice      int    `json:"infant_price"`
	SingleSupplement int    `json:"single_supplement"`
	TotalStock       int    `json:"total_stock" binding:"required,min=1"`
	CutoffDays       int    `json:"cutoff_days"`
}

// PriceCalendarResponse is the response for price calendar queries.
type PriceCalendarResponse struct {
	Departures []DepartureDetailResponse `json:"departures"`
}

// DepartureDetailResponse is a single departure with full detail.
type DepartureDetailResponse struct {
	ID               int64  `json:"id"`
	DepartureDate    string `json:"departure_date"`
	ReturnDate       string `json:"return_date"`
	AdultPrice       int    `json:"adult_price"`
	ChildPrice       int    `json:"child_price"`
	InfantPrice      int    `json:"infant_price"`
	SingleSupplement int    `json:"single_supplement"`
	TotalStock       int    `json:"total_stock"`
	SoldCount        int    `json:"sold_count"`
	LockedCount      int    `json:"locked_count"`
	AvailableStock   int    `json:"available_stock"`
	CutoffDays       int    `json:"cutoff_days"`
	Status           string `json:"status"`
}

// --- Service Methods ---

// BatchUpdatePrices batch-updates prices for selected departure dates.
func (s *PriceCalendarService) BatchUpdatePrices(productID int64, req BatchPriceUpdateRequest) (*BatchPriceUpdateResponse, error) {
	// Verify product exists
	_, err := s.productRepo.FindByIDBasic(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("find product: %w", err)
	}

	// Get target departures
	targetDepartures, err := s.productRepo.FindDeparturesByProductAndDates(productID, req.TargetDates)
	if err != nil {
		return nil, fmt.Errorf("find target departures: %w", err)
	}

	if len(targetDepartures) == 0 {
		return &BatchPriceUpdateResponse{UpdatedCount: 0}, nil
	}

	var refDepartures []productmodel.DepartureDate

	// For follow mode, load reference departures
	if req.Mode == "follow" {
		if len(req.ReferenceDates) == 0 {
			return nil, fmt.Errorf("reference_dates required for follow mode")
		}
		refDepartures, err = s.productRepo.FindDeparturesByProductAndDates(productID, req.ReferenceDates)
		if err != nil {
			return nil, fmt.Errorf("find reference departures: %w", err)
		}
		if len(refDepartures) == 0 {
			return nil, fmt.Errorf("no reference departures found")
		}
	}

	// Calculate average from reference dates for follow mode
	var refAdultAvg, refChildAvg, refInfantAvg, refSupplementAvg int
	if req.Mode == "follow" && len(refDepartures) > 0 {
		for _, rd := range refDepartures {
			refAdultAvg += rd.AdultPrice
			refChildAvg += rd.ChildPrice
			refInfantAvg += rd.InfantPrice
			refSupplementAvg += rd.SingleSupplement
		}
		n := len(refDepartures)
		refAdultAvg /= n
		refChildAvg /= n
		refInfantAvg /= n
		refSupplementAvg /= n
	}

	updatedCount := 0
	for i := range targetDepartures {
		d := &targetDepartures[i]

		switch req.Mode {
		case "fixed":
			if req.AdultPrice != nil {
				d.AdultPrice = *req.AdultPrice
			}
			if req.ChildPrice != nil {
				d.ChildPrice = *req.ChildPrice
			}
			if req.InfantPrice != nil {
				d.InfantPrice = *req.InfantPrice
			}
			if req.SingleSupplement != nil {
				d.SingleSupplement = *req.SingleSupplement
			}

		case "percentage":
			pct := 1.0
			if req.Percentage != nil {
				pct = 1.0 + (*req.Percentage / 100.0)
			}
			d.AdultPrice = applyPercentage(d.AdultPrice, pct)
			d.ChildPrice = applyPercentage(d.ChildPrice, pct)
			d.InfantPrice = applyPercentage(d.InfantPrice, pct)
			d.SingleSupplement = applyPercentage(d.SingleSupplement, pct)

		case "amount":
			amt := 0
			if req.Amount != nil {
				amt = *req.Amount
			}
			d.AdultPrice = d.AdultPrice + amt
			d.ChildPrice = d.ChildPrice + amt
			d.InfantPrice = d.InfantPrice + amt
			d.SingleSupplement = d.SingleSupplement + amt
			// Ensure non-negative
			if d.AdultPrice < 0 {
				d.AdultPrice = 0
			}
			if d.ChildPrice < 0 {
				d.ChildPrice = 0
			}
			if d.InfantPrice < 0 {
				d.InfantPrice = 0
			}
			if d.SingleSupplement < 0 {
				d.SingleSupplement = 0
			}

		case "follow":
			d.AdultPrice = refAdultAvg
			d.ChildPrice = refChildAvg
			d.InfantPrice = refInfantAvg
			d.SingleSupplement = refSupplementAvg
		}

		updatedCount++
	}

	// Save all departures
	if err := s.productRepo.SaveDepartures(targetDepartures); err != nil {
		return nil, fmt.Errorf("save departures: %w", err)
	}

	s.logger.Info("batch price update completed",
		zap.Int64("product_id", productID),
		zap.String("mode", req.Mode),
		zap.Int("updated", updatedCount),
	)

	return &BatchPriceUpdateResponse{UpdatedCount: updatedCount}, nil
}

// GetPriceCalendar returns the price calendar for a product by month.
func (s *PriceCalendarService) GetPriceCalendar(productID int64, month string) (*PriceCalendarResponse, error) {
	_, err := s.productRepo.FindByIDBasic(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("find product: %w", err)
	}

	departures, err := s.productRepo.FindDeparturesByProductAndMonth(productID, month)
	if err != nil {
		return nil, fmt.Errorf("find departures: %w", err)
	}

	items := make([]DepartureDetailResponse, len(departures))
	for i, d := range departures {
		avail := d.AvailableStock()
		status := d.Status
		if avail <= 0 && status == productmodel.DepartureStatusOpen {
			status = productmodel.DepartureStatusFull
		}

		items[i] = DepartureDetailResponse{
			ID:               d.ID,
			DepartureDate:    d.DepartureDate.Format("2006-01-02"),
			ReturnDate:       d.ReturnDate.Format("2006-01-02"),
			AdultPrice:       d.AdultPrice,
			ChildPrice:       d.ChildPrice,
			InfantPrice:      d.InfantPrice,
			SingleSupplement: d.SingleSupplement,
			TotalStock:       d.TotalStock,
			SoldCount:        d.SoldCount,
			LockedCount:      d.LockedCount,
			AvailableStock:   avail,
			CutoffDays:       d.CutoffDays,
			Status:           status,
		}
	}

	return &PriceCalendarResponse{Departures: items}, nil
}

// applyPercentage applies a percentage multiplier to a price in cents.
func applyPercentage(price int, multiplier float64) int {
	return int(math.Round(float64(price) * multiplier))
}

// ParseDate parses a date string in YYYY-MM-DD format.
func ParseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}
