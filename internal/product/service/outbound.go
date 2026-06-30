// Package service provides business logic for the Product domain.
package service

import (
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/product/model"
	"github.com/travel-booking/server/internal/product/repository"
)

// OutboundProductService provides business logic for outbound travel products.
type OutboundProductService struct {
	productRepo *repository.ProductRepository
	countryRepo *repository.CountryRepository
	templateRepo *repository.VisaMaterialTemplateRepository
	logger      *zap.Logger
}

// NewOutboundProductService creates a new OutboundProductService.
func NewOutboundProductService(
	productRepo *repository.ProductRepository,
	countryRepo *repository.CountryRepository,
	templateRepo *repository.VisaMaterialTemplateRepository,
	logger *zap.Logger,
) *OutboundProductService {
	return &OutboundProductService{
		productRepo:  productRepo,
		countryRepo:  countryRepo,
		templateRepo: templateRepo,
		logger:       logger,
	}
}

// --- Request/Response DTOs ---

// ListOutboundProductsRequest holds query parameters for outbound product listing.
type ListOutboundProductsRequest struct {
	Continent   string `form:"continent"`
	CountryID   *int64 `form:"country_id"`
	VisaType    string `form:"visa_type" binding:"omitempty,oneof=visa_free visa_on_arrival e_visa visa_required"`
	OriginCity  string `form:"origin_city"`
	DaysMin     *int   `form:"days_min"`
	DaysMax     *int   `form:"days_max"`
	PriceMin    *int   `form:"price_min"`
	PriceMax    *int   `form:"price_max"`
	Keyword     string `form:"keyword"`
	Sort        string `form:"sort" binding:"omitempty,oneof=recommended price_asc price_desc days_asc days_desc"`
	Page        int    `form:"page" binding:"min=1"`
	PageSize    int    `form:"page_size" binding:"min=1,max=100"`
}

// OutboundProductDetail represents a detailed outbound product with visa info.
type OutboundProductDetail struct {
	*model.Product
	Country             *model.Country              `json:"country,omitempty"`
	VisaInfo            *model.VisaInfo              `json:"visa_info_parsed,omitempty"`
	InsuranceReqs       *model.InsuranceRequirements `json:"insurance_requirements_parsed,omitempty"`
	FlightInfo          *model.InternationalFlightInfo `json:"flight_info_parsed,omitempty"`
	PreTripSvc          *model.PreTripServices       `json:"pre_trip_services_parsed,omitempty"`
	MaterialPreview     []MaterialPreviewByOccupation `json:"material_preview,omitempty"`
}

// MaterialPreviewByOccupation groups material templates by occupation type.
type MaterialPreviewByOccupation struct {
	OccupationType string                       `json:"occupation_type"`
	OccupationName string                       `json:"occupation_name"`
	Materials      []model.VisaMaterialTemplate `json:"materials"`
}

// PreTripInfo contains pre-trip service information for a destination country.
type PreTripInfo struct {
	Country           *model.Country `json:"country"`
	EntryPolicy       json.RawMessage `json:"entry_policy,omitempty"`
	CashRegulation    json.RawMessage `json:"cash_regulation,omitempty"`
	ProhibitedItems   json.RawMessage `json:"prohibited_items,omitempty"`
	EntryCardGuide    json.RawMessage `json:"entry_card_guide,omitempty"`
	CustomsGuide      json.RawMessage `json:"customs_guide,omitempty"`
	EmergencyContacts json.RawMessage `json:"emergency_contacts,omitempty"`
}

// ContinentTreeResponse represents the continent→country hierarchy.
type ContinentTreeResponse struct {
	Continent     string          `json:"continent"`
	ContinentName string          `json:"continent_name"`
	Countries     []model.Country `json:"countries"`
}

// continentNameMap maps continent codes to Chinese names.
var continentNameMap = map[string]string{
	model.ContinentAsia:         "亚洲",
	model.ContinentEurope:       "欧洲",
	model.ContinentNorthAmerica: "北美洲",
	model.ContinentSouthAmerica: "南美洲",
	model.ContinentOceania:      "大洋洲",
	model.ContinentAfrica:       "非洲",
}

// OccupationNameMap maps occupation codes to Chinese names.
var OccupationNameMap = map[string]string{
	model.OccupationEmployed:  "在职人员",
	model.OccupationFreelance: "自由职业",
	model.OccupationRetired:   "退休人员",
	model.OccupationStudent:   "学生",
	model.OccupationChild:     "儿童",
}

// ListOutboundProducts returns a paginated list of outbound products with filters.
func (s *OutboundProductService) ListOutboundProducts(req ListOutboundProductsRequest) (interface{}, error) {
	filter := repository.ProductFilter{
		Status: model.ProductStatusApproved,
	}
	if req.OriginCity != "" {
		filter.Origin = req.OriginCity
	}
	if req.DaysMin != nil {
		filter.DaysMin = req.DaysMin
	}
	if req.DaysMax != nil {
		filter.DaysMax = req.DaysMax
	}
	if req.PriceMin != nil {
		filter.PriceMin = req.PriceMin
	}
	if req.PriceMax != nil {
		filter.PriceMax = req.PriceMax
	}
	if req.Keyword != "" {
		filter.Keyword = req.Keyword
	}

	// Override product type to outbound
	sort := repository.SortRecommended
	if req.Sort != "" {
		sort = repository.ProductSort(req.Sort)
	}

	page := req.Page
	if page == 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 20
	}

	// Use the product repo with outbound type filter
	products, total, err := s.productRepo.FindOutboundWithFilters(filter, req.Continent, req.CountryID, req.VisaType, sort, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("list outbound products: %w", err)
	}

	return map[string]interface{}{
		"items":      products,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	}, nil
}

// GetContinentTree returns the continent→country hierarchy for filtering.
func (s *OutboundProductService) GetContinentTree() ([]ContinentTreeResponse, error) {
	tree, err := s.countryRepo.ContinentTree()
	if err != nil {
		return nil, fmt.Errorf("get continent tree: %w", err)
	}

	var result []ContinentTreeResponse
	// Maintain consistent order
	continents := []string{
		model.ContinentAsia, model.ContinentEurope, model.ContinentNorthAmerica,
		model.ContinentSouthAmerica, model.ContinentOceania, model.ContinentAfrica,
	}
	for _, c := range continents {
		if countries, ok := tree[c]; ok {
			result = append(result, ContinentTreeResponse{
				Continent:     c,
				ContinentName: continentNameMap[c],
				Countries:     countries,
			})
		}
	}
	return result, nil
}

// GetOutboundProductDetail returns detailed outbound product info including visa card.
func (s *OutboundProductService) GetOutboundProductDetail(productID int64) (*OutboundProductDetail, error) {
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return nil, fmt.Errorf("get product: %w", err)
	}

	if product.ProductType != model.ProductTypeOutbound {
		return nil, fmt.Errorf("product is not an outbound product")
	}

	detail := &OutboundProductDetail{
		Product: product,
	}

	// Parse visa info
	if product.VisaInfo != nil {
		var visaInfo model.VisaInfo
		if err := json.Unmarshal(product.VisaInfo, &visaInfo); err == nil {
			detail.VisaInfo = &visaInfo
		}
	}

	// Parse insurance requirements
	if product.InsuranceRequirements != nil {
		var ins model.InsuranceRequirements
		if err := json.Unmarshal(product.InsuranceRequirements, &ins); err == nil {
			detail.InsuranceReqs = &ins
		}
	}

	// Parse flight info
	if product.InternationalFlightInfo != nil {
		var flight model.InternationalFlightInfo
		if err := json.Unmarshal(product.InternationalFlightInfo, &flight); err == nil {
			detail.FlightInfo = &flight
		}
	}

	// Parse pre-trip services
	if product.PreTripServices != nil {
		var svc model.PreTripServices
		if err := json.Unmarshal(product.PreTripServices, &svc); err == nil {
			detail.PreTripSvc = &svc
		}
	}

	// Load country info
	if product.DestinationCountryID != nil {
		country, err := s.countryRepo.FindByID(*product.DestinationCountryID)
		if err != nil {
			s.logger.Warn("failed to load destination country",
				zap.Int64("country_id", *product.DestinationCountryID),
				zap.Error(err))
		} else {
			detail.Country = country
		}

		// Load material preview by occupation type
		templates, err := s.templateRepo.FindByCountry(*product.DestinationCountryID)
		if err != nil {
			s.logger.Warn("failed to load visa material templates",
				zap.Int64("country_id", *product.DestinationCountryID),
				zap.Error(err))
		} else {
			detail.MaterialPreview = groupByOccupation(templates)
		}
	}

	return detail, nil
}

// GetPreTripInfo returns pre-trip service information for a country.
func (s *OutboundProductService) GetPreTripInfo(countryID int64) (*PreTripInfo, error) {
	country, err := s.countryRepo.FindByID(countryID)
	if err != nil {
		return nil, fmt.Errorf("get country: %w", err)
	}

	return &PreTripInfo{
		Country:           country,
		EntryPolicy:       country.EntryPolicy,
		CashRegulation:    country.CashRegulation,
		ProhibitedItems:   country.ProhibitedItems,
		EntryCardGuide:    country.EntryCardGuide,
		CustomsGuide:      country.CustomsGuide,
		EmergencyContacts: country.EmergencyContacts,
	}, nil
}

// groupByOccupation groups visa material templates by occupation type.
func groupByOccupation(templates []model.VisaMaterialTemplate) []MaterialPreviewByOccupation {
	grouped := make(map[string][]model.VisaMaterialTemplate)
	for _, t := range templates {
		grouped[t.OccupationType] = append(grouped[t.OccupationType], t)
	}

	var result []MaterialPreviewByOccupation
	// Maintain consistent order
	order := []string{
		model.OccupationEmployed, model.OccupationFreelance,
		model.OccupationRetired, model.OccupationStudent, model.OccupationChild,
	}
	for _, occ := range order {
		if mats, ok := grouped[occ]; ok {
			result = append(result, MaterialPreviewByOccupation{
				OccupationType: occ,
				OccupationName: OccupationNameMap[occ],
				Materials:      mats,
			})
		}
	}
	return result
}

// ValidatePassportExpiry checks if passport validity covers return date + 6 months.
// Returns true if valid, false if insufficient.
func ValidatePassportExpiry(passportExpiry, returnDate time.Time) bool {
	// Passport must be valid for at least 6 months after return date
	minValidity := returnDate.AddDate(0, 6, 0)
	return passportExpiry.After(minValidity) || passportExpiry.Equal(minValidity)
}
