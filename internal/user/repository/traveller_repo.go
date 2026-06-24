package repository

import (
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/user/model"
)

// TravellerRepository provides CRUD operations for FrequentTraveller.
type TravellerRepository struct {
	db *gorm.DB
}

// NewTravellerRepository creates a new TravellerRepository.
func NewTravellerRepository(db *gorm.DB) *TravellerRepository {
	return &TravellerRepository{db: db}
}

// Create inserts a new frequent traveller.
func (r *TravellerRepository) Create(t *model.FrequentTraveller) error {
	return r.db.Create(t).Error
}

// FindByUserID returns all travellers for a user, ordered by default first then creation time.
func (r *TravellerRepository) FindByUserID(userID int64) ([]model.FrequentTraveller, error) {
	var travellers []model.FrequentTraveller
	err := r.db.Where("user_id = ?", userID).
		Order("is_default DESC, created_at ASC").
		Find(&travellers).Error
	return travellers, err
}

// FindByID returns a traveller by ID, ensuring it belongs to the user.
func (r *TravellerRepository) FindByID(id, userID int64) (*model.FrequentTraveller, error) {
	var t model.FrequentTraveller
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Update updates a traveller record.
func (r *TravellerRepository) Update(t *model.FrequentTraveller) error {
	return r.db.Save(t).Error
}

// Delete deletes a traveller by ID, ensuring it belongs to the user.
func (r *TravellerRepository) Delete(id, userID int64) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).
		Delete(&model.FrequentTraveller{}).Error
}

// CountByUserID returns the number of travellers for a user.
func (r *TravellerRepository) CountByUserID(userID int64) (int, error) {
	var count int64
	err := r.db.Model(&model.FrequentTraveller{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return int(count), err
}

// ClearDefault clears the default flag for all travellers of a user.
func (r *TravellerRepository) ClearDefault(userID int64) error {
	return r.db.Model(&model.FrequentTraveller{}).
		Where("user_id = ? AND is_default = true", userID).
		Update("is_default", false).Error
}
