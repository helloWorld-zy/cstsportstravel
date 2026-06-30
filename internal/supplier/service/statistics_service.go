package service

import (
	"time"

	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/supplier/repository"
)

// StatisticsService provides aggregated statistics for supplier workspace.
// Uses Asynq pre-aggregation for performance.
type StatisticsService struct {
	supplierRepo   *repository.SupplierRepository
	settlementRepo *repository.SettlementRepository
	logger         *zap.Logger
}

// NewStatisticsService creates a new StatisticsService.
func NewStatisticsService(
	supplierRepo *repository.SupplierRepository,
	settlementRepo *repository.SettlementRepository,
	logger *zap.Logger,
) *StatisticsService {
	return &StatisticsService{
		supplierRepo:   supplierRepo,
		settlementRepo: settlementRepo,
		logger:         logger,
	}
}

// SupplierStats contains aggregated statistics for a supplier.
type SupplierStats struct {
	TotalOrders     int64   `json:"total_orders"`
	TotalRevenue    float64 `json:"total_revenue"`
	TotalCommission float64 `json:"total_commission"`
	AvgOrderValue   float64 `json:"avg_order_value"`
	TotalRefunds    float64 `json:"total_refunds"`
	RatingScore     float64 `json:"rating_score"`
	Period          string  `json:"period"`
}

// SalesTrend represents a data point in the sales trend chart.
type SalesTrend struct {
	Date     string  `json:"date"`
	Orders   int     `json:"orders"`
	Revenue  float64 `json:"revenue"`
	Refunds  float64 `json:"refunds"`
}

// GetSupplierStats returns aggregated statistics for a supplier.
func (s *StatisticsService) GetSupplierStats(tenantID, supplierID int64, period string) (*SupplierStats, error) {
	// TODO: query from pre-aggregated statistics table (Asynq)
	// For now, query directly from orders
	var startDate time.Time
	now := time.Now()

	switch period {
	case "today":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	default:
		startDate = now.AddDate(0, -1, 0)
	}

	_ = startDate

	// TODO: aggregate from order table filtered by supplier_id
	return &SupplierStats{
		TotalOrders:     0,
		TotalRevenue:    0,
		TotalCommission: 0,
		AvgOrderValue:   0,
		TotalRefunds:    0,
		RatingScore:     0,
		Period:          period,
	}, nil
}

// GetSalesTrend returns sales trend data for charts.
func (s *StatisticsService) GetSalesTrend(tenantID, supplierID int64, days int) ([]SalesTrend, error) {
	// TODO: query from pre-aggregated statistics table
	trends := make([]SalesTrend, days)
	now := time.Now()
	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -(days - 1 - i))
		trends[i] = SalesTrend{
			Date:    date.Format("2006-01-02"),
			Orders:  0,
			Revenue: 0,
			Refunds: 0,
		}
	}
	return trends, nil
}
