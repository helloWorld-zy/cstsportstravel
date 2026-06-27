// Package service provides business logic for the Admin domain.
package service

import (
	"encoding/json"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	adminrepo "github.com/travel-booking/server/internal/admin/repository"
	productmodel "github.com/travel-booking/server/internal/product/model"
)

// ItineraryService provides business logic for itinerary management.
type ItineraryService struct {
	productRepo *adminrepo.AdminProductRepository
	logger      *zap.Logger
}

// NewItineraryService creates a new ItineraryService.
func NewItineraryService(productRepo *adminrepo.AdminProductRepository, logger *zap.Logger) *ItineraryService {
	return &ItineraryService{
		productRepo: productRepo,
		logger:      logger,
	}
}

// --- Request/Response DTOs ---

// SaveItineraryRequest is the request body for saving itineraries.
type SaveItineraryRequest struct {
	Itineraries []ItineraryDayRequest `json:"itineraries" binding:"required,min=1"`
}

// ItineraryDayRequest is a single day in the itinerary.
type ItineraryDayRequest struct {
	DayNo       int             `json:"day_no" binding:"required,min=1"`
	Title       string          `json:"title" binding:"required,max=200"`
	Description string          `json:"description"`
	Meals       *MealsRequest   `json:"meals"`
	Hotel       string          `json:"hotel"`
	Transport   string          `json:"transport"`
	Spots       []SpotRequest   `json:"spots"`
	Images      []string        `json:"images"`
}

// MealsRequest represents meal plan for a day.
type MealsRequest struct {
	Breakfast bool `json:"breakfast"`
	Lunch     bool `json:"lunch"`
	Dinner    bool `json:"dinner"`
}

// SpotRequest represents a tourist spot visit.
type SpotRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Duration    string `json:"duration"`
	Image       string `json:"image"`
}

// ItineraryResponse is the response for saved itineraries.
type ItineraryResponse struct {
	ProductID  int64                `json:"product_id"`
	Itineraries []ItineraryDayResponse `json:"itineraries"`
}

// ItineraryDayResponse is a single day in the itinerary response.
type ItineraryDayResponse struct {
	ID          int64           `json:"id"`
	DayNo       int             `json:"day_no"`
	Title       string          `json:"title"`
	Description string          `json:"description,omitempty"`
	Meals       json.RawMessage `json:"meals,omitempty"`
	Hotel       string          `json:"hotel,omitempty"`
	Transport   string          `json:"transport,omitempty"`
	Spots       json.RawMessage `json:"spots,omitempty"`
	Images      json.RawMessage `json:"images,omitempty"`
}

// --- Service Methods ---

// SaveItinerary replaces the itinerary for a product.
func (s *ItineraryService) SaveItinerary(productID int64, req SaveItineraryRequest) (*ItineraryResponse, error) {
	// Verify product exists
	product, err := s.productRepo.FindByIDBasic(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("find product: %w", err)
	}

	// Build itinerary models
	itineraries := make([]productmodel.Itinerary, len(req.Itineraries))
	for i, day := range req.Itineraries {
		var mealsJSON json.RawMessage
		if day.Meals != nil {
			mealsJSON, _ = json.Marshal(day.Meals)
		}

		var spotsJSON json.RawMessage
		if day.Spots != nil {
			spotsJSON, _ = json.Marshal(day.Spots)
		}

		var imagesJSON json.RawMessage
		if day.Images != nil {
			imagesJSON, _ = json.Marshal(day.Images)
		}

		itineraries[i] = productmodel.Itinerary{
			ProductID:   productID,
			DayNo:       day.DayNo,
			Title:       day.Title,
			Description: day.Description,
			Meals:       mealsJSON,
			Hotel:       day.Hotel,
			Transport:   day.Transport,
			Spots:       spotsJSON,
			Images:      imagesJSON,
		}
	}

	if err := s.productRepo.SaveItineraries(productID, itineraries); err != nil {
		return nil, fmt.Errorf("save itineraries: %w", err)
	}

	s.logger.Info("itinerary saved",
		zap.Int64("product_id", productID),
		zap.Int("days", len(itineraries)),
	)

	// If product is approved, mark as change_pending_review
	if product.Status == productmodel.ProductStatusApproved {
		if err := s.productRepo.UpdateStatus(productID, productmodel.ProductStatusChangePendingReview, ""); err != nil {
			s.logger.Warn("failed to update product status after itinerary change", zap.Error(err))
		}
	}

	// Build response
	respDays := make([]ItineraryDayResponse, len(itineraries))
	for i, it := range itineraries {
		respDays[i] = ItineraryDayResponse{
			ID:          it.ID,
			DayNo:       it.DayNo,
			Title:       it.Title,
			Description: it.Description,
			Meals:       it.Meals,
			Hotel:       it.Hotel,
			Transport:   it.Transport,
			Spots:       it.Spots,
			Images:      it.Images,
		}
	}

	return &ItineraryResponse{
		ProductID:  productID,
		Itineraries: respDays,
	}, nil
}

// GetItinerary returns the itinerary for a product.
func (s *ItineraryService) GetItinerary(productID int64) (*ItineraryResponse, error) {
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("find product: %w", err)
	}

	days := make([]ItineraryDayResponse, len(product.Itineraries))
	for i, it := range product.Itineraries {
		days[i] = ItineraryDayResponse{
			ID:          it.ID,
			DayNo:       it.DayNo,
			Title:       it.Title,
			Description: it.Description,
			Meals:       it.Meals,
			Hotel:       it.Hotel,
			Transport:   it.Transport,
			Spots:       it.Spots,
			Images:      it.Images,
		}
	}

	return &ItineraryResponse{
		ProductID:  productID,
		Itineraries: days,
	}, nil
}
