// Package service provides business logic for the Admin domain.
package service

import (
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	adminrepo "github.com/travel-booking/server/internal/admin/repository"
	productmodel "github.com/travel-booking/server/internal/product/model"
)

// DepartureService provides business logic for departure and inventory management.
type DepartureService struct {
	productRepo *adminrepo.AdminProductRepository
	logger      *zap.Logger
}

// NewDepartureService creates a new DepartureService.
func NewDepartureService(productRepo *adminrepo.AdminProductRepository, logger *zap.Logger) *DepartureService {
	return &DepartureService{
		productRepo: productRepo,
		logger:      logger,
	}
}

// --- Request/Response DTOs ---

// CreateDepartureRequest is the request body for creating/updating departures.
type CreateDepartureRequest struct {
	Departures []DepartureEntryRequest `json:"departures" binding:"required,min=1"`
}

// DepartureEntryRequest is a single departure entry.
type DepartureEntryRequest struct {
	DepartureDate    string `json:"departure_date" binding:"required"`
	ReturnDate       string `json:"return_date" binding:"required"`
	AdultPrice       int    `json:"adult_price" binding:"required,min=1"`
	ChildPrice       int    `json:"child_price" binding:"required,min=0"`
	InfantPrice      int    `json:"infant_price"`
	SingleSupplement int    `json:"single_supplement"`
	TotalStock       int    `json:"total_stock" binding:"required,min=1"`
	CutoffDays       int    `json:"cutoff_days"`
}

// UpdateStockRequest is the request body for manual stock adjustment.
type UpdateStockRequest struct {
	TotalStock int    `json:"total_stock" binding:"required,min=0"`
	Reason     string `json:"reason" binding:"required"`
}

// DepartureListResponse is the response for departure listing.
type DepartureListResponse struct {
	Departures []DepartureDetailResponse `json:"departures"`
}

// --- Service Methods ---

// CreateOrUpdateDepartures batch-creates or updates departure dates.
func (s *DepartureService) CreateOrUpdateDepartures(productID int64, req CreateDepartureRequest) (*DepartureListResponse, error) {
	// Verify product exists
	_, err := s.productRepo.FindByIDBasic(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("find product: %w", err)
	}

	departures := make([]productmodel.DepartureDate, len(req.Departures))
	for i, d := range req.Departures {
		depDate, err := time.Parse("2006-01-02", d.DepartureDate)
		if err != nil {
			return nil, fmt.Errorf("invalid departure_date at index %d: %w", i, err)
		}
		retDate, err := time.Parse("2006-01-02", d.ReturnDate)
		if err != nil {
			return nil, fmt.Errorf("invalid return_date at index %d: %w", i, err)
		}

		cutoffDays := d.CutoffDays
		if cutoffDays == 0 {
			cutoffDays = 1
		}

		departures[i] = productmodel.DepartureDate{
			ProductID:        productID,
			DepartureDate:    depDate,
			ReturnDate:       retDate,
			AdultPrice:       d.AdultPrice,
			ChildPrice:       d.ChildPrice,
			InfantPrice:      d.InfantPrice,
			SingleSupplement: d.SingleSupplement,
			TotalStock:       d.TotalStock,
			CutoffDays:       cutoffDays,
			Status:           productmodel.DepartureStatusOpen,
		}
	}

	if err := s.productRepo.SaveDepartures(departures); err != nil {
		return nil, fmt.Errorf("save departures: %w", err)
	}

	s.logger.Info("departures saved",
		zap.Int64("product_id", productID),
		zap.Int("count", len(departures)),
	)

	// Build response
	items := make([]DepartureDetailResponse, len(departures))
	for i, d := range departures {
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
			AvailableStock:   d.TotalStock - d.SoldCount - d.LockedCount,
			CutoffDays:       d.CutoffDays,
			Status:           d.Status,
		}
	}

	return &DepartureListResponse{Departures: items}, nil
}

// ListDepartures returns all departures for a product.
func (s *DepartureService) ListDepartures(productID int64, month string) (*DepartureListResponse, error) {
	_, err := s.productRepo.FindByIDBasic(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("find product: %w", err)
	}

	var departures []productmodel.DepartureDate
	if month != "" {
		departures, err = s.productRepo.FindDeparturesByProductAndMonth(productID, month)
	} else {
		departures, err = s.productRepo.FindDeparturesByProduct(productID)
	}
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

	return &DepartureListResponse{Departures: items}, nil
}

// UpdateStock manually adjusts the stock of a departure date.
func (s *DepartureService) UpdateStock(productID int64, departureID int64, req UpdateStockRequest) error {
	// Verify product exists
	_, err := s.productRepo.FindByIDBasic(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProductNotFound
		}
		return fmt.Errorf("find product: %w", err)
	}

	if err := s.productRepo.UpdateDepartureStock(departureID, req.TotalStock); err != nil {
		return fmt.Errorf("update stock: %w", err)
	}

	s.logger.Info("stock manually adjusted",
		zap.Int64("product_id", productID),
		zap.Int64("departure_id", departureID),
		zap.Int("new_total_stock", req.TotalStock),
		zap.String("reason", req.Reason),
	)

	return nil
}

// GetStockStatus returns the stock status summary for a product.
func (s *DepartureService) GetStockStatus(productID int64) (map[string]interface{}, error) {
	departures, err := s.productRepo.FindDeparturesByProduct(productID)
	if err != nil {
		return nil, fmt.Errorf("find departures: %w", err)
	}

	totalStock := 0
	totalSold := 0
	totalLocked := 0
	openCount := 0
	fullCount := 0

	for _, d := range departures {
		totalStock += d.TotalStock
		totalSold += d.SoldCount
		totalLocked += d.LockedCount
		if d.Status == productmodel.DepartureStatusOpen {
			openCount++
		}
		if d.Status == productmodel.DepartureStatusFull {
			fullCount++
		}
	}

	return map[string]interface{}{
		"total_departures": len(departures),
		"open_departures":  openCount,
		"full_departures":  fullCount,
		"total_stock":      totalStock,
		"total_sold":       totalSold,
		"total_locked":     totalLocked,
		"available_stock":  totalStock - totalSold - totalLocked,
	}, nil
}
