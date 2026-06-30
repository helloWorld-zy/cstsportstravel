// Package model defines GORM models for the Product domain.
package model

import (
	"encoding/json"
	"time"
)

// Continent constants for country classification.
const (
	ContinentAsia         = "asia"
	ContinentEurope       = "europe"
	ContinentNorthAmerica = "north_america"
	ContinentSouthAmerica = "south_america"
	ContinentOceania      = "oceania"
	ContinentAfrica       = "africa"
)

// Visa type constants.
const (
	VisaTypeFreeOnArrival = "visa_free"
	VisaTypeOnArrival     = "visa_on_arrival"
	VisaTypeEVisa         = "e_visa"
	VisaTypeRequired      = "visa_required"
)

// Country represents a destination country/region for outbound travel.
type Country struct {
	ID                     int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID               int64           `gorm:"column:tenant_id;not null;index:idx_country_tenant_continent" json:"tenant_id"`
	NameCN                 string          `gorm:"column:name_cn;size:100;not null" json:"name_cn"`
	NameEN                 string          `gorm:"column:name_en;size:100;not null" json:"name_en"`
	Continent              string          `gorm:"column:continent;size:20;not null;index:idx_country_tenant_continent" json:"continent"`
	VisaType               string          `gorm:"column:visa_type;size:20;not null;index:idx_country_tenant_visa" json:"visa_type"`
	VisaProcessingDays     *int            `gorm:"column:visa_processing_days" json:"visa_processing_days,omitempty"`
	PassportValidityMonths int             `gorm:"column:passport_validity_months;not null;default:6" json:"passport_validity_months"`
	EntryPolicy            json.RawMessage `gorm:"column:entry_policy;type:jsonb" json:"entry_policy,omitempty"`
	CashRegulation         json.RawMessage `gorm:"column:cash_regulation;type:jsonb" json:"cash_regulation,omitempty"`
	ProhibitedItems        json.RawMessage `gorm:"column:prohibited_items;type:jsonb" json:"prohibited_items,omitempty"`
	EntryCardGuide         json.RawMessage `gorm:"column:entry_card_guide;type:jsonb" json:"entry_card_guide,omitempty"`
	CustomsGuide           json.RawMessage `gorm:"column:customs_guide;type:jsonb" json:"customs_guide,omitempty"`
	EmergencyContacts      json.RawMessage `gorm:"column:emergency_contacts;type:jsonb" json:"emergency_contacts,omitempty"`
	Status                 string          `gorm:"column:status;size:20;not null;default:active" json:"status"`
	CreatedAt              time.Time       `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt              time.Time       `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
}

// TableName overrides the table name.
func (Country) TableName() string {
	return "country"
}

// Country status constants.
const (
	CountryStatusActive   = "active"
	CountryStatusInactive = "inactive"
)

// VisaInfo contains visa-related information embedded in Product.
type VisaInfo struct {
	VisaType           string   `json:"visa_type"`
	ProcessingDays     int      `json:"processing_days,omitempty"`
	Fee                int      `json:"fee,omitempty"` // cents
	MaterialPreview    []string `json:"material_preview,omitempty"`
	ConsularDistrict   string   `json:"consular_district,omitempty"`
	ValidityPeriod     string   `json:"validity_period,omitempty"`
	StayPeriod         string   `json:"stay_period,omitempty"`
	LatestSubmitDate   string   `json:"latest_submit_date,omitempty"`
	RejectRefundPolicy string   `json:"reject_refund_policy,omitempty"`
}

// VisaMaterialTemplate defines the material requirements for a country+occupation combination.
type VisaMaterialTemplate struct {
	ID             int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID       int64  `gorm:"column:tenant_id;not null;index:idx_visa_template_country" json:"tenant_id"`
	CountryID      int64  `gorm:"column:country_id;not null;index:idx_visa_template_country" json:"country_id"`
	OccupationType string `gorm:"column:occupation_type;size:20;not null;index:idx_visa_template_country" json:"occupation_type"`
	MaterialType   string `gorm:"column:material_type;size:50;not null" json:"material_type"`
	MaterialName   string `gorm:"column:material_name;size:100;not null" json:"material_name"`
	IsRequired     bool   `gorm:"column:is_required;not null;default:true" json:"is_required"`
	Description    string `gorm:"column:description;type:text" json:"description,omitempty"`
	FileFormat     string `gorm:"column:file_format;size:50" json:"file_format,omitempty"`
	MaxSizeMB      int    `gorm:"column:max_size_mb;not null;default:10" json:"max_size_mb"`
	SortOrder      int    `gorm:"column:sort_order;not null;default:0" json:"sort_order"`
	Status         string `gorm:"column:status;size:20;not null;default:active" json:"status"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:now()" json:"created_at"`
}

// TableName overrides the table name.
func (VisaMaterialTemplate) TableName() string {
	return "visa_material_template"
}

// Occupation type constants for visa material templates.
const (
	OccupationEmployed   = "employed"
	OccupationFreelance  = "freelance"
	OccupationRetired    = "retired"
	OccupationStudent    = "student"
	OccupationChild      = "child"
)

// InsuranceRequirements defines insurance requirements for outbound products.
type InsuranceRequirements struct {
	Required       bool   `json:"required"`
	MinMedicalCost int    `json:"min_medical_cost,omitempty"` // cents, e.g. Schengen requires ≥30000 EUR
	CoveragePeriod string `json:"coverage_period,omitempty"`  // full_trip
	Schengen       bool   `json:"schengen,omitempty"`         // requires Schengen-compliant insurance
	Description    string `json:"description,omitempty"`
}

// PreTripServices defines pre-trip service configuration for outbound products.
type PreTripServices struct {
	EntryPolicy       bool `json:"entry_policy"`
	EntryMaterials    bool `json:"entry_materials"`
	CashRegulation    bool `json:"cash_regulation"`
	ProhibitedItems   bool `json:"prohibited_items"`
	EntryCardGuide    bool `json:"entry_card_guide"`
	CustomsGuide      bool `json:"customs_guide"`
	FlightTracking    bool `json:"flight_tracking"`
	WeatherForecast   bool `json:"weather_forecast"`
	TimeDifference    bool `json:"time_difference"`
	Checklist         bool `json:"checklist"`
	EmergencyContacts bool `json:"emergency_contacts"`
}
