// Package service provides business logic for the Admin domain.
package service

import (
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/admin/model"
	"github.com/travel-booking/server/internal/admin/repository"
)

// Sentinel errors for banner operations.
var (
	ErrBannerNotFound = errors.New("banner not found")
)

// BannerService handles banner management business logic.
type BannerService struct {
	bannerRepo *repository.BannerRepository
	logger     *zap.Logger
}

// NewBannerService creates a new BannerService.
func NewBannerService(bannerRepo *repository.BannerRepository, logger *zap.Logger) *BannerService {
	return &BannerService{
		bannerRepo: bannerRepo,
		logger:     logger,
	}
}

// CreateBannerRequest is the request to create a banner.
type CreateBannerRequest struct {
	Title     string     `json:"title" binding:"required,max=200"`
	ImageURL  string     `json:"image_url" binding:"required,max=500"`
	LinkURL   string     `json:"link_url" binding:"omitempty,max=500"`
	Position  string     `json:"position" binding:"omitempty,max=50"`
	SortOrder int        `json:"sort_order"`
	StartAt   *time.Time `json:"start_at,omitempty"`
	EndAt     *time.Time `json:"end_at,omitempty"`
}

// UpdateBannerRequest is the request to update a banner.
type UpdateBannerRequest struct {
	Title     *string    `json:"title,omitempty" binding:"omitempty,max=200"`
	ImageURL  *string    `json:"image_url,omitempty" binding:"omitempty,max=500"`
	LinkURL   *string    `json:"link_url,omitempty" binding:"omitempty,max=500"`
	Position  *string    `json:"position,omitempty" binding:"omitempty,max=50"`
	SortOrder *int       `json:"sort_order,omitempty"`
	Status    *string    `json:"status,omitempty" binding:"omitempty,oneof=active inactive"`
	StartAt   *time.Time `json:"start_at,omitempty"`
	EndAt     *time.Time `json:"end_at,omitempty"`
}

// BannerResponse is the response for a banner.
type BannerResponse struct {
	ID        int64      `json:"id"`
	Title     string     `json:"title"`
	ImageURL  string     `json:"image_url"`
	LinkURL   string     `json:"link_url,omitempty"`
	Position  string     `json:"position"`
	SortOrder int        `json:"sort_order"`
	Status    string     `json:"status"`
	StartAt   *time.Time `json:"start_at,omitempty"`
	EndAt     *time.Time `json:"end_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// CreateBanner creates a new banner.
func (s *BannerService) CreateBanner(createdBy int64, req CreateBannerRequest) (*BannerResponse, error) {
	position := req.Position
	if position == "" {
		position = model.BannerPositionHomeTop
	}

	banner := &model.HomepageBanner{
		Title:     req.Title,
		ImageURL:  req.ImageURL,
		LinkURL:   req.LinkURL,
		Position:  position,
		SortOrder: req.SortOrder,
		Status:    model.BannerStatusActive,
		StartAt:   req.StartAt,
		EndAt:     req.EndAt,
		CreatedBy: &createdBy,
	}

	if err := s.bannerRepo.Create(banner); err != nil {
		return nil, fmt.Errorf("create banner: %w", err)
	}

	return toBannerResponse(banner), nil
}

// UpdateBanner updates an existing banner.
func (s *BannerService) UpdateBanner(id int64, req UpdateBannerRequest) (*BannerResponse, error) {
	banner, err := s.bannerRepo.FindByID(id)
	if err != nil {
		return nil, ErrBannerNotFound
	}

	if req.Title != nil {
		banner.Title = *req.Title
	}
	if req.ImageURL != nil {
		banner.ImageURL = *req.ImageURL
	}
	if req.LinkURL != nil {
		banner.LinkURL = *req.LinkURL
	}
	if req.Position != nil {
		banner.Position = *req.Position
	}
	if req.SortOrder != nil {
		banner.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		banner.Status = *req.Status
	}
	if req.StartAt != nil {
		banner.StartAt = req.StartAt
	}
	if req.EndAt != nil {
		banner.EndAt = req.EndAt
	}

	if err := s.bannerRepo.Update(banner); err != nil {
		return nil, fmt.Errorf("update banner: %w", err)
	}

	return toBannerResponse(banner), nil
}

// DeleteBanner deletes a banner by ID.
func (s *BannerService) DeleteBanner(id int64) error {
	_, err := s.bannerRepo.FindByID(id)
	if err != nil {
		return ErrBannerNotFound
	}
	return s.bannerRepo.Delete(id)
}

// GetBanner returns a banner by ID.
func (s *BannerService) GetBanner(id int64) (*BannerResponse, error) {
	banner, err := s.bannerRepo.FindByID(id)
	if err != nil {
		return nil, ErrBannerNotFound
	}
	return toBannerResponse(banner), nil
}

// ListBanners returns banners for admin management.
func (s *BannerService) ListBanners(position, status string) ([]BannerResponse, error) {
	banners, err := s.bannerRepo.ListBanners(position, status)
	if err != nil {
		return nil, err
	}

	results := make([]BannerResponse, len(banners))
	for i, b := range banners {
		results[i] = *toBannerResponse(&b)
	}
	return results, nil
}

// FindActiveBanners returns currently active banners for public display.
func (s *BannerService) FindActiveBanners(position string) ([]BannerResponse, error) {
	banners, err := s.bannerRepo.FindActiveBanners(position)
	if err != nil {
		return nil, err
	}

	results := make([]BannerResponse, len(banners))
	for i, b := range banners {
		results[i] = *toBannerResponse(&b)
	}
	return results, nil
}

func toBannerResponse(b *model.HomepageBanner) *BannerResponse {
	return &BannerResponse{
		ID:        b.ID,
		Title:     b.Title,
		ImageURL:  b.ImageURL,
		LinkURL:   b.LinkURL,
		Position:  b.Position,
		SortOrder: b.SortOrder,
		Status:    b.Status,
		StartAt:   b.StartAt,
		EndAt:     b.EndAt,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
}
