// Package repository provides data access for the Marketing domain.
package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/marketing/domain"
)

// CouponRepository provides data access for Coupon with tenant isolation.
type CouponRepository struct {
	db *gorm.DB
}

// NewCouponRepository creates a new CouponRepository.
func NewCouponRepository(db *gorm.DB) *CouponRepository {
	return &CouponRepository{db: db}
}

// Create inserts a new coupon record.
func (r *CouponRepository) Create(c *domain.Coupon) error {
	return r.db.Create(c).Error
}

// FindByID returns a coupon by ID with tenant isolation.
func (r *CouponRepository) FindByID(tenantID, id int64) (*domain.Coupon, error) {
	var c domain.Coupon
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Update updates a coupon record.
func (r *CouponRepository) Update(c *domain.Coupon) error {
	c.UpdatedAt = time.Now()
	return r.db.Save(c).Error
}

// UpdateStatus updates the coupon status.
func (r *CouponRepository) UpdateStatus(tenantID, id int64, targetStatus string) error {
	coupon, err := r.FindByID(tenantID, id)
	if err != nil {
		return err
	}
	if !coupon.CanTransitionTo(targetStatus) {
		return fmt.Errorf("invalid status transition from %s to %s", coupon.Status, targetStatus)
	}
	coupon.Status = targetStatus
	coupon.UpdatedAt = time.Now()
	return r.db.Save(coupon).Error
}

// List returns coupons with filtering and pagination.
func (r *CouponRepository) List(tenantID int64, couponType, status string, page, pageSize int) ([]domain.Coupon, int64, error) {
	var coupons []domain.Coupon
	var total int64

	query := r.db.Model(&domain.Coupon{}).Where("tenant_id = ?", tenantID)
	if couponType != "" {
		query = query.Where("coupon_type = ?", couponType)
	}
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&coupons).Error
	return coupons, total, err
}

// ListActive returns active coupons available in the coupon center.
func (r *CouponRepository) ListActive(tenantID int64, page, pageSize int) ([]domain.Coupon, int64, error) {
	var coupons []domain.Coupon
	var total int64

	query := r.db.Model(&domain.Coupon{}).
		Where("tenant_id = ? AND status = ?", tenantID, domain.CouponStatusActive)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&coupons).Error
	return coupons, total, err
}

// IncrementClaimedCount atomically increments the claimed count and checks stock.
func (r *CouponRepository) IncrementClaimedCount(tenantID, couponID int64) error {
	result := r.db.Model(&domain.Coupon{}).
		Where("tenant_id = ? AND id = ? AND claimed_count < total_stock", tenantID, couponID).
		Update("claimed_count", gorm.Expr("claimed_count + 1"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrCouponOutOfStock
	}
	return nil
}

// IncrementUsedCount atomically increments the used count.
func (r *CouponRepository) IncrementUsedCount(tenantID, couponID int64) error {
	return r.db.Model(&domain.Coupon{}).
		Where("tenant_id = ? AND id = ?", tenantID, couponID).
		Update("used_count", gorm.Expr("used_count + 1")).Error
}

// DecrementUsedCount atomically decrements the used count (on refund/return).
func (r *CouponRepository) DecrementUsedCount(tenantID, couponID int64) error {
	return r.db.Model(&domain.Coupon{}).
		Where("tenant_id = ? AND id = ? AND used_count > 0", tenantID, couponID).
		Update("used_count", gorm.Expr("used_count - 1")).Error
}

// CouponClaimRepository provides data access for CouponClaim.
type CouponClaimRepository struct {
	db *gorm.DB
}

// NewCouponClaimRepository creates a new CouponClaimRepository.
func NewCouponClaimRepository(db *gorm.DB) *CouponClaimRepository {
	return &CouponClaimRepository{db: db}
}

// Create inserts a new coupon claim record.
func (r *CouponClaimRepository) Create(claim *domain.CouponClaim) error {
	return r.db.Create(claim).Error
}

// FindByID returns a coupon claim by ID.
func (r *CouponClaimRepository) FindByID(id int64) (*domain.CouponClaim, error) {
	var claim domain.CouponClaim
	err := r.db.Where("id = ?", id).First(&claim).Error
	if err != nil {
		return nil, err
	}
	return &claim, nil
}

// FindByUserAndCoupon returns a claim by user and coupon (for duplicate check).
func (r *CouponClaimRepository) FindByUserAndCoupon(userID, couponID int64) (*domain.CouponClaim, error) {
	var claim domain.CouponClaim
	err := r.db.Where("user_id = ? AND coupon_id = ?", userID, couponID).First(&claim).Error
	if err != nil {
		return nil, err
	}
	return &claim, nil
}

// CountByUserAndCoupon counts how many times a user has claimed a specific coupon.
func (r *CouponClaimRepository) CountByUserAndCoupon(userID, couponID int64) (int64, error) {
	var count int64
	err := r.db.Model(&domain.CouponClaim{}).
		Where("user_id = ? AND coupon_id = ?", userID, couponID).
		Count(&count).Error
	return count, err
}

// CountByDeviceAndCoupon counts how many times a device has claimed a specific coupon.
func (r *CouponClaimRepository) CountByDeviceAndCoupon(deviceID string, couponID int64) (int64, error) {
	var count int64
	err := r.db.Model(&domain.CouponClaim{}).
		Where("device_id = ? AND coupon_id = ?", deviceID, couponID).
		Count(&count).Error
	return count, err
}

// UpdateStatus updates the claim status.
func (r *CouponClaimRepository) UpdateStatus(id int64, targetStatus string) error {
	claim, err := r.FindByID(id)
	if err != nil {
		return err
	}
	if !claim.CanTransitionTo(targetStatus) {
		return fmt.Errorf("invalid claim status transition from %s to %s", claim.Status, targetStatus)
	}
	claim.Status = targetStatus
	return r.db.Save(claim).Error
}

// OccupyClaim marks a claim as occupied (when order is placed).
func (r *CouponClaimRepository) OccupyClaim(id int64, orderID int64) error {
	claim, err := r.FindByID(id)
	if err != nil {
		return err
	}
	if !claim.CanTransitionTo(domain.ClaimStatusOccupied) {
		return fmt.Errorf("cannot occupy claim in status %s", claim.Status)
	}
	now := time.Now()
	return r.db.Model(&domain.CouponClaim{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     domain.ClaimStatusOccupied,
			"order_id":   orderID,
			"used_at":    now, // Will be overwritten when actually used
		}).Error
}

// UseClaim marks a claim as used (after payment).
func (r *CouponClaimRepository) UseClaim(id int64) error {
	now := time.Now()
	return r.db.Model(&domain.CouponClaim{}).
		Where("id = ? AND status = ?", id, domain.ClaimStatusOccupied).
		Updates(map[string]interface{}{
			"status": domain.ClaimStatusUsed,
			"used_at": now,
		}).Error
}

// ReturnClaim marks a claim as returned (on refund).
func (r *CouponClaimRepository) ReturnClaim(id int64) error {
	now := time.Now()
	return r.db.Model(&domain.CouponClaim{}).
		Where("id = ? AND status IN ?", id, []string{domain.ClaimStatusOccupied, domain.ClaimStatusUsed}).
		Updates(map[string]interface{}{
			"status":      domain.ClaimStatusReturned,
			"returned_at": now,
		}).Error
}

// VoidClaim marks a claim as voided.
func (r *CouponClaimRepository) VoidClaim(id int64) error {
	return r.db.Model(&domain.CouponClaim{}).
		Where("id = ? AND status = ?", id, domain.ClaimStatusAvailable).
		Update("status", domain.ClaimStatusVoided).Error
}

// ListByUser returns coupon claims for a user, optionally filtered by status.
func (r *CouponClaimRepository) ListByUser(userID int64, status string, page, pageSize int) ([]domain.CouponClaim, int64, error) {
	var claims []domain.CouponClaim
	var total int64

	query := r.db.Model(&domain.CouponClaim{}).Where("user_id = ?", userID)
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("claimed_at DESC").Find(&claims).Error
	return claims, total, err
}

// ExpireStaleClaims expires claims that have passed their expiry time.
func (r *CouponClaimRepository) ExpireStaleClaims() (int64, error) {
	result := r.db.Model(&domain.CouponClaim{}).
		Where("status = ? AND expired_at IS NOT NULL AND expired_at < ?", domain.ClaimStatusAvailable, time.Now()).
		Update("status", domain.ClaimStatusExpired)
	return result.RowsAffected, result.Error
}

// PromotionActivityRepository provides data access for PromotionActivity.
type PromotionActivityRepository struct {
	db *gorm.DB
}

// NewPromotionActivityRepository creates a new PromotionActivityRepository.
func NewPromotionActivityRepository(db *gorm.DB) *PromotionActivityRepository {
	return &PromotionActivityRepository{db: db}
}

// Create inserts a new promotion activity record.
func (r *PromotionActivityRepository) Create(a *domain.PromotionActivity) error {
	return r.db.Create(a).Error
}

// FindByID returns a promotion activity by ID with tenant isolation.
func (r *PromotionActivityRepository) FindByID(tenantID, id int64) (*domain.PromotionActivity, error) {
	var a domain.PromotionActivity
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// Update updates a promotion activity record.
func (r *PromotionActivityRepository) Update(a *domain.PromotionActivity) error {
	a.UpdatedAt = time.Now()
	return r.db.Save(a).Error
}

// UpdateStatus updates the activity status.
func (r *PromotionActivityRepository) UpdateStatus(tenantID, id int64, targetStatus string) error {
	activity, err := r.FindByID(tenantID, id)
	if err != nil {
		return err
	}
	if !activity.CanTransitionTo(targetStatus) {
		return fmt.Errorf("invalid activity status transition from %s to %s", activity.Status, targetStatus)
	}
	activity.Status = targetStatus
	activity.UpdatedAt = time.Now()
	return r.db.Save(activity).Error
}

// List returns promotion activities with filtering and pagination.
func (r *PromotionActivityRepository) List(tenantID int64, activityType, status string, page, pageSize int) ([]domain.PromotionActivity, int64, error) {
	var activities []domain.PromotionActivity
	var total int64

	query := r.db.Model(&domain.PromotionActivity{}).Where("tenant_id = ?", tenantID)
	if activityType != "" {
		query = query.Where("activity_type = ?", activityType)
	}
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&activities).Error
	return activities, total, err
}

// FindActiveByProduct returns active activities that apply to a specific product.
func (r *PromotionActivityRepository) FindActiveByProduct(tenantID, productID int64, now time.Time) ([]domain.PromotionActivity, error) {
	var activities []domain.PromotionActivity
	err := r.db.Where("tenant_id = ? AND status = ? AND start_time <= ? AND end_time >= ?",
		tenantID, domain.ActivityStatusActive, now, now).
		Find(&activities).Error
	if err != nil {
		return nil, err
	}

	// Filter by applicable products
	var result []domain.PromotionActivity
	for _, a := range activities {
		if len(a.ApplicableProducts) == 0 {
			// No product restriction = applies to all
			result = append(result, a)
			continue
		}
		for _, pid := range a.ApplicableProducts {
			if pid == productID {
				result = append(result, a)
				break
			}
		}
	}
	return result, nil
}
