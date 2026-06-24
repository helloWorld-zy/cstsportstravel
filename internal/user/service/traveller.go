package service

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/encrypt"
	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/user/model"
	"github.com/travel-booking/server/internal/user/repository"
)

const (
	maxTravellersPerUser = 20
)

// TravellerService provides business logic for frequent traveller management.
type TravellerService struct {
	repo      *repository.TravellerRepository
	encryptor *encrypt.Encryptor
	logger    *zap.Logger
}

// NewTravellerService creates a new TravellerService.
func NewTravellerService(
	repo *repository.TravellerRepository,
	encryptor *encrypt.Encryptor,
	logger *zap.Logger,
) *TravellerService {
	return &TravellerService{
		repo:      repo,
		encryptor: encryptor,
		logger:    logger,
	}
}

// TravellerResponse is the public representation of a traveller.
type TravellerResponse struct {
	ID        int64  `json:"id"`
	RealName  string `json:"real_name"`
	IDCardNo  string `json:"id_card_no"`
	Phone     string `json:"phone"`
	BirthDate string `json:"birth_date"`
	Gender    string `json:"gender"`
	IsDefault bool   `json:"is_default"`
	CreatedAt string `json:"created_at"`
}

// CreateTravellerRequest is the request body for creating a traveller.
type CreateTravellerRequest struct {
	RealName  string `json:"real_name" binding:"required"`
	IDCardNo  string `json:"id_card_no" binding:"required"`
	Phone     string `json:"phone"`
	BirthDate string `json:"birth_date"`
	Gender    string `json:"gender" binding:"omitempty,oneof=male female"`
	IsDefault bool   `json:"is_default"`
}

// UpdateTravellerRequest is the request body for updating a traveller.
type UpdateTravellerRequest struct {
	RealName  string `json:"real_name"`
	IDCardNo  string `json:"id_card_no"`
	Phone     string `json:"phone"`
	BirthDate string `json:"birth_date"`
	Gender    string `json:"gender" binding:"omitempty,oneof=male female"`
	IsDefault *bool  `json:"is_default"`
}

// List returns all travellers for a user.
func (s *TravellerService) List(userID int64) ([]TravellerResponse, error) {
	travellers, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	result := make([]TravellerResponse, 0, len(travellers))
	for _, t := range travellers {
		result = append(result, s.toResponse(&t))
	}
	return result, nil
}

// Create adds a new frequent traveller for the user.
func (s *TravellerService) Create(userID int64, req CreateTravellerRequest) (*TravellerResponse, error) {
	// Validate ID card format
	if !validateIDCard(req.IDCardNo) {
		return nil, ErrInvalidIDCard
	}

	// Check max limit
	count, err := s.repo.CountByUserID(userID)
	if err != nil {
		return nil, err
	}
	if count >= maxTravellersPerUser {
		return nil, ErrMaxTravellersReached
	}

	// Encrypt sensitive fields
	encryptedName, err := s.encryptor.Encrypt(req.RealName)
	if err != nil {
		return nil, fmt.Errorf("encrypt name: %w", err)
	}
	encryptedIDCard, err := s.encryptor.Encrypt(req.IDCardNo)
	if err != nil {
		return nil, fmt.Errorf("encrypt id card: %w", err)
	}

	// Handle default flag
	if req.IsDefault {
		s.repo.ClearDefault(userID)
	}

	// Parse birth date
	var birthDate *time.Time
	if req.BirthDate != "" {
		t, parseErr := time.Parse("2006-01-02", req.BirthDate)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid birth_date format, expected YYYY-MM-DD")
		}
		birthDate = &t
	}

	traveller := &model.FrequentTraveller{
		UserID:    userID,
		RealName:  encryptedName,
		IDCardNo:  encryptedIDCard,
		Phone:     req.Phone,
		BirthDate: birthDate,
		Gender:    req.Gender,
		IsDefault: req.IsDefault,
	}
	if err := s.repo.Create(traveller); err != nil {
		return nil, err
	}

	result := s.toResponse(traveller)
	result.RealName = response.MaskName(req.RealName)
	result.IDCardNo = response.MaskIDCard(req.IDCardNo)
	return &result, nil
}

// Update updates an existing traveller.
func (s *TravellerService) Update(userID, travellerID int64, req UpdateTravellerRequest) (*TravellerResponse, error) {
	traveller, err := s.repo.FindByID(travellerID, userID)
	if err != nil {
		return nil, ErrTravellerNotFound
	}

	if req.RealName != "" {
		encrypted, encErr := s.encryptor.Encrypt(req.RealName)
		if encErr != nil {
			return nil, fmt.Errorf("encrypt name: %w", encErr)
		}
		traveller.RealName = encrypted
	}
	if req.IDCardNo != "" {
		if !validateIDCard(req.IDCardNo) {
			return nil, ErrInvalidIDCard
		}
		encrypted, encErr := s.encryptor.Encrypt(req.IDCardNo)
		if encErr != nil {
			return nil, fmt.Errorf("encrypt id card: %w", encErr)
		}
		traveller.IDCardNo = encrypted
	}
	if req.Phone != "" {
		traveller.Phone = req.Phone
	}
	if req.BirthDate != "" {
		t, parseErr := time.Parse("2006-01-02", req.BirthDate)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid birth_date format")
		}
		traveller.BirthDate = &t
	}
	if req.Gender != "" {
		traveller.Gender = req.Gender
	}
	if req.IsDefault != nil {
		if *req.IsDefault {
			s.repo.ClearDefault(userID)
		}
		traveller.IsDefault = *req.IsDefault
	}

	if err := s.repo.Update(traveller); err != nil {
		return nil, err
	}

	result := s.toResponse(traveller)
	return &result, nil
}

// Delete removes a traveller.
func (s *TravellerService) Delete(userID, travellerID int64) error {
	_, err := s.repo.FindByID(travellerID, userID)
	if err != nil {
		return ErrTravellerNotFound
	}
	return s.repo.Delete(travellerID, userID)
}

// toResponse converts a model to a response with masked fields.
func (s *TravellerService) toResponse(t *model.FrequentTraveller) TravellerResponse {
	resp := TravellerResponse{
		ID:        t.ID,
		Phone:     t.Phone,
		Gender:    t.Gender,
		IsDefault: t.IsDefault,
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
	}
	if t.BirthDate != nil {
		resp.BirthDate = t.BirthDate.Format("2006-01-02")
	}
	// Decrypt and mask sensitive fields
	if t.RealName != "" {
		if name, err := s.encryptor.Decrypt(t.RealName); err == nil {
			resp.RealName = response.MaskName(name)
		}
	}
	if t.IDCardNo != "" {
		if idCard, err := s.encryptor.Decrypt(t.IDCardNo); err == nil {
			resp.IDCardNo = response.MaskIDCard(idCard)
		}
	}
	return resp
}

// Domain errors.
var (
	ErrMaxTravellersReached = fmt.Errorf("maximum %d travellers per user reached", maxTravellersPerUser)
	ErrTravellerNotFound    = fmt.Errorf("traveller not found")
)
