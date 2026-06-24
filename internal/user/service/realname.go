package service

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/common/encrypt"
	"github.com/travel-booking/server/internal/user/model"
	"github.com/travel-booking/server/internal/user/repository"
)

const (
	maxDailyVerificationAttempts = 3
)

// RealNameService handles real-name verification logic.
type RealNameService struct {
	repo      *repository.UserRepository
	rnvRepo   *RealNameVerificationRepo
	encryptor *encrypt.Encryptor
	logger    *zap.Logger
}

// NewRealNameService creates a new RealNameService.
func NewRealNameService(
	repo *repository.UserRepository,
	rnvRepo *RealNameVerificationRepo,
	encryptor *encrypt.Encryptor,
	logger *zap.Logger,
) *RealNameService {
	return &RealNameService{
		repo:      repo,
		rnvRepo:   rnvRepo,
		encryptor: encryptor,
		logger:    logger,
	}
}

// SubmitVerificationRequest is the request body for real-name verification.
type SubmitVerificationRequest struct {
	RealName string `json:"real_name" binding:"required"`
	IDCardNo string `json:"id_card_no" binding:"required"`
}

// VerificationResponse is the response for real-name verification.
type VerificationResponse struct {
	Status string `json:"status"`
}

// SubmitVerification submits a real-name verification request.
func (s *RealNameService) SubmitVerification(userID int64, req SubmitVerificationRequest) (*VerificationResponse, error) {
	// Validate ID card format (ISO 7064:1983.MOD 11-2)
	if !validateIDCard(req.IDCardNo) {
		return nil, ErrInvalidIDCard
	}

	// Check daily attempt limit
	today := time.Now().Truncate(24 * time.Hour)
	attempts, err := s.rnvRepo.CountTodayAttempts(userID, today)
	if err != nil {
		return nil, fmt.Errorf("count attempts: %w", err)
	}
	if attempts >= maxDailyVerificationAttempts {
		return nil, ErrDailyLimitExceeded
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

	// In MVP, auto-verify (skip public security API call)
	// In production, call public security API for name+ID verification
	status := model.RNStatusVerified
	verifiedAt := time.Now()

	verification := &model.RealNameVerification{
		UserID:     userID,
		RealName:   encryptedName,
		IDCardNo:   encryptedIDCard,
		Status:     status,
		VerifiedAt: &verifiedAt,
	}
	if err := s.rnvRepo.Create(verification); err != nil {
		return nil, fmt.Errorf("create verification: %w", err)
	}

	// Update user's real-name status
	if err := s.repo.UpdateRealNameStatus(userID, status, encryptedName, encryptedIDCard); err != nil {
		return nil, fmt.Errorf("update user status: %w", err)
	}

	s.logger.Info("real-name verification completed",
		zap.Int64("user_id", userID),
		zap.String("status", status),
	)

	return &VerificationResponse{Status: status}, nil
}

// validateIDCard validates an 18-digit Chinese ID card number using ISO 7064:1983.MOD 11-2.
func validateIDCard(idCard string) bool {
	if len(idCard) != 18 {
		return false
	}

	// Check all digits (last can be X/x)
	for i := 0; i < 17; i++ {
		if idCard[i] < '0' || idCard[i] > '9' {
			return false
		}
	}
	last := idCard[17]
	if last != 'X' && last != 'x' && (last < '0' || last > '9') {
		return false
	}

	// Weights for each position
	weights := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	// Check codes
	checkCodes := []byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}

	sum := 0
	for i := 0; i < 17; i++ {
		sum += int(idCard[i]-'0') * weights[i]
	}
	expected := checkCodes[sum%11]

	return idCard[17] == expected || (idCard[17] == 'x' && expected == 'X') || (idCard[17] == 'X' && expected == 'X')
}

// RealNameVerificationRepo provides CRUD for RealNameVerification.
type RealNameVerificationRepo struct {
	db *gorm.DB
}

// NewRealNameVerificationRepo creates a new RealNameVerificationRepo.
func NewRealNameVerificationRepo(db *gorm.DB) *RealNameVerificationRepo {
	return &RealNameVerificationRepo{db: db}
}

// Create inserts a new verification record.
func (r *RealNameVerificationRepo) Create(v *model.RealNameVerification) error {
	return r.db.Create(v).Error
}

// CountTodayAttempts counts verification attempts by user today.
func (r *RealNameVerificationRepo) CountTodayAttempts(userID int64, today time.Time) (int, error) {
	var count int64
	err := r.db.Model(&model.RealNameVerification{}).
		Where("user_id = ? AND created_at >= ?", userID, today).
		Count(&count).Error
	return int(count), err
}

// Domain errors.
var (
	ErrInvalidIDCard = fmt.Errorf("invalid ID card number format")
)
