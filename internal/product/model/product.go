// Package model defines GORM models for the Product domain.
package model

import (
	"encoding/json"
	"time"
)

// Category represents a product category with tree structure.
type Category struct {
	ID        int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string     `gorm:"column:name;size:100;not null" json:"name"`
	ParentID  *int64     `gorm:"column:parent_id;index" json:"parent_id,omitempty"`
	IconURL   string     `gorm:"column:icon_url;size:500" json:"icon_url,omitempty"`
	SortOrder int        `gorm:"column:sort_order;not null;default:0" json:"sort_order"`
	Status    string     `gorm:"column:status;size:20;not null;default:active" json:"status"`
	CreatedAt time.Time  `gorm:"column:created_at;not null;default:now()" json:"created_at"`
}

// TableName overrides the table name.
func (Category) TableName() string {
	return "category"
}

// Product represents a travel product listing.
type Product struct {
	ID                 int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductNo          string          `gorm:"column:product_no;size:30;uniqueIndex;not null" json:"product_no"`
	ProductName        string          `gorm:"column:product_name;size:200;not null" json:"product_name"`
	CategoryID         int64           `gorm:"column:category_id;not null;index" json:"category_id"`
	ProductType        string          `gorm:"column:product_type;size:30;not null;default:group_tour" json:"product_type"`
	// Outbound travel fields (Phase 2)
	DestinationCountryID    *int64          `gorm:"column:destination_country_id;index" json:"destination_country_id,omitempty"`
	VisaInfo                json.RawMessage `gorm:"column:visa_info;type:jsonb" json:"visa_info,omitempty"`
	InternationalFlightInfo json.RawMessage `gorm:"column:international_flight_info;type:jsonb" json:"international_flight_info,omitempty"`
	InsuranceRequirements   json.RawMessage `gorm:"column:insurance_requirements;type:jsonb" json:"insurance_requirements,omitempty"`
	PreTripServices         json.RawMessage `gorm:"column:pre_trip_services;type:jsonb" json:"pre_trip_services,omitempty"`
	OriginCity         string          `gorm:"column:origin_city;size:50;not null" json:"origin_city"`
	DestinationCities  json.RawMessage `gorm:"column:destination_cities;type:jsonb;not null" json:"destination_cities"`
	DestinationTags    json.RawMessage `gorm:"column:destination_tags;type:jsonb" json:"destination_tags,omitempty"`
	Days               int             `gorm:"column:days;not null" json:"days"`
	Nights             int             `gorm:"column:nights;not null" json:"nights"`
	TransportMode      string          `gorm:"column:transport_mode;size:50" json:"transport_mode,omitempty"`
	MinGroupSize       int             `gorm:"column:min_group_size;not null;default:2" json:"min_group_size"`
	MaxGroupSize       int             `gorm:"column:max_group_size;not null;default:50" json:"max_group_size"`
	ProductGrade       string          `gorm:"column:product_grade;size:20" json:"product_grade,omitempty"`
	CoverImage         string          `gorm:"column:cover_image;size:500" json:"cover_image,omitempty"`
	Images             json.RawMessage `gorm:"column:images;type:jsonb" json:"images,omitempty"`
	Summary            string          `gorm:"column:summary;size:500" json:"summary,omitempty"`
	Description        string          `gorm:"column:description;type:text" json:"description,omitempty"`
	FeeIncluded        string          `gorm:"column:fee_included;type:text" json:"fee_included,omitempty"`
	FeeExcluded        string          `gorm:"column:fee_excluded;type:text" json:"fee_excluded,omitempty"`
	BookingNotes       string          `gorm:"column:booking_notes;type:text" json:"booking_notes,omitempty"`
	Status             string          `gorm:"column:status;size:30;not null;default:draft" json:"status"`
	RejectReason       string          `gorm:"column:reject_reason;size:500" json:"reject_reason,omitempty"`
	SupplierID         *int64          `gorm:"column:supplier_id;index" json:"supplier_id,omitempty"`
	CommissionRate     float64         `gorm:"column:commission_rate;type:decimal(5,4);default:0" json:"commission_rate"`
	ViewCount          int             `gorm:"column:view_count;not null;default:0" json:"view_count"`
	OrderCount         int             `gorm:"column:order_count;not null;default:0" json:"order_count"`
	SatisfactionRate   *float64        `gorm:"column:satisfaction_rate;type:decimal(5,2)" json:"satisfaction_rate,omitempty"`
	CreatedAt          time.Time       `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt          time.Time       `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`

	// Relations
	Itineraries    []Itinerary    `gorm:"foreignKey:ProductID" json:"itineraries,omitempty"`
	DepartureDates []DepartureDate `gorm:"foreignKey:ProductID" json:"departure_dates,omitempty"`
	PriceRules     []PriceRule    `gorm:"foreignKey:ProductID" json:"price_rules,omitempty"`
	RefundRules    []RefundRule   `gorm:"foreignKey:ProductID" json:"refund_rules,omitempty"`
	Reviews        []ProductReview `gorm:"foreignKey:ProductID" json:"reviews,omitempty"`
	Category       *Category      `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	DestinationCountry *Country   `gorm:"foreignKey:DestinationCountryID" json:"destination_country,omitempty"`
}

// TableName overrides the table name.
func (Product) TableName() string {
	return "product"
}

// Itinerary represents a day-by-day itinerary for a product.
type Itinerary struct {
	ID          int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID   int64           `gorm:"column:product_id;not null;uniqueIndex:idx_itinerary_product_day" json:"product_id"`
	DayNo       int             `gorm:"column:day_no;not null;uniqueIndex:idx_itinerary_product_day" json:"day_no"`
	Title       string          `gorm:"column:title;size:200;not null" json:"title"`
	Description string          `gorm:"column:description;type:text" json:"description,omitempty"`
	Meals       json.RawMessage `gorm:"column:meals;type:jsonb" json:"meals,omitempty"`
	Hotel       string          `gorm:"column:hotel;size:200" json:"hotel,omitempty"`
	Transport   string          `gorm:"column:transport;size:100" json:"transport,omitempty"`
	Spots       json.RawMessage `gorm:"column:spots;type:jsonb" json:"spots,omitempty"`
	Images      json.RawMessage `gorm:"column:images;type:jsonb" json:"images,omitempty"`
	CreatedAt   time.Time       `gorm:"column:created_at;not null;default:now()" json:"created_at"`
}

// TableName overrides the table name.
func (Itinerary) TableName() string {
	return "itinerary"
}

// DepartureDate represents a specific departure date with pricing and stock.
type DepartureDate struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID       int64     `gorm:"column:product_id;not null;uniqueIndex:idx_departure_product_date" json:"product_id"`
	DepartureDate   time.Time `gorm:"column:departure_date;not null;uniqueIndex:idx_departure_product_date" json:"departure_date"`
	ReturnDate      time.Time `gorm:"column:return_date;not null" json:"return_date"`
	AdultPrice      int       `gorm:"column:adult_price;not null" json:"adult_price"`              // cents
	ChildPrice      int       `gorm:"column:child_price;not null" json:"child_price"`              // cents
	InfantPrice     int       `gorm:"column:infant_price;not null;default:0" json:"infant_price"`  // cents
	SingleSupplement int      `gorm:"column:single_supplement;not null;default:0" json:"single_supplement"` // cents
	TotalStock      int       `gorm:"column:total_stock;not null" json:"total_stock"`
	SoldCount       int       `gorm:"column:sold_count;not null;default:0" json:"sold_count"`
	LockedCount     int       `gorm:"column:locked_count;not null;default:0" json:"locked_count"`
	CutoffDays      int       `gorm:"column:cutoff_days;not null;default:1" json:"cutoff_days"`
	Status          string    `gorm:"column:status;size:20;not null;default:open" json:"status"`
	CreatedAt       time.Time `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
}

// TableName overrides the table name.
func (DepartureDate) TableName() string {
	return "departure_date"
}

// AvailableStock returns the available stock count.
func (d *DepartureDate) AvailableStock() int {
	return d.TotalStock - d.SoldCount - d.LockedCount
}

// PriceRule represents a pricing rule for specific date ranges.
type PriceRule struct {
	ID               int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID        int64     `gorm:"column:product_id;not null;index" json:"product_id"`
	DateFrom         time.Time `gorm:"column:date_from;not null" json:"date_from"`
	DateTo           time.Time `gorm:"column:date_to;not null" json:"date_to"`
	AdultPrice       *int      `gorm:"column:adult_price" json:"adult_price,omitempty"`             // cents
	ChildPrice       *int      `gorm:"column:child_price" json:"child_price,omitempty"`             // cents
	InfantPrice      *int      `gorm:"column:infant_price" json:"infant_price,omitempty"`           // cents
	SingleSupplement *int      `gorm:"column:single_supplement" json:"single_supplement,omitempty"` // cents
	PriceType        string    `gorm:"column:price_type;size:20;not null;default:standard" json:"price_type"`
	Priority         int       `gorm:"column:priority;not null;default:0" json:"priority"`
	CreatedAt        time.Time `gorm:"column:created_at;not null;default:now()" json:"created_at"`
}

// TableName overrides the table name.
func (PriceRule) TableName() string {
	return "price_rule"
}

// RefundRule represents a cancellation/refund rule.
type RefundRule struct {
	ID               int64   `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID        *int64  `gorm:"column:product_id;index" json:"product_id,omitempty"` // null for global template
	RuleName         string  `gorm:"column:rule_name;size:100;not null" json:"rule_name"`
	DaysBeforeMin    int     `gorm:"column:days_before_min;not null" json:"days_before_min"`
	DaysBeforeMax    *int    `gorm:"column:days_before_max" json:"days_before_max,omitempty"`
	RefundPercentage float64 `gorm:"column:refund_percentage;type:decimal(5,2);not null" json:"refund_percentage"`
	Description      string  `gorm:"column:description;size:500" json:"description,omitempty"`
	IsTemplate       bool    `gorm:"column:is_template;not null;default:false" json:"is_template"`
	CreatedAt        time.Time `gorm:"column:created_at;not null;default:now()" json:"created_at"`
}

// TableName overrides the table name.
func (RefundRule) TableName() string {
	return "refund_rule"
}

// ProductReview represents a user review for a product.
type ProductReview struct {
	ID          int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID   int64           `gorm:"column:product_id;not null;index" json:"product_id"`
	UserID      int64           `gorm:"column:user_id;not null" json:"user_id"`
	OrderID     int64           `gorm:"column:order_id;not null;uniqueIndex" json:"order_id"`
	Rating      int             `gorm:"column:rating;not null" json:"rating"`
	Content     string          `gorm:"column:content;type:text" json:"content,omitempty"`
	Images      json.RawMessage `gorm:"column:images;type:jsonb" json:"images,omitempty"`
	IsAnonymous bool            `gorm:"column:is_anonymous;not null;default:false" json:"is_anonymous"`
	CreatedAt   time.Time       `gorm:"column:created_at;not null;default:now()" json:"created_at"`
}

// TableName overrides the table name.
func (ProductReview) TableName() string {
	return "product_review"
}

// Product status constants.
const (
	ProductStatusDraft               = "draft"
	ProductStatusPendingReview       = "pending_review"
	ProductStatusApproved            = "approved"
	ProductStatusSuspended           = "suspended"
	ProductStatusChangePendingReview = "change_pending_review"
)

// Product type constants.
const (
	ProductTypeGroupTour = "group_tour"
	ProductTypeOutbound  = "outbound_group"
)

// InternationalFlightInfo contains international flight details for outbound products.
type InternationalFlightInfo struct {
	Airline        string `json:"airline"`
	FlightNo       string `json:"flight_no"`
	DepartCity     string `json:"depart_city"`
	DepartAirport  string `json:"depart_airport,omitempty"`
	ArriveCity     string `json:"arrive_city"`
	ArriveAirport  string `json:"arrive_airport,omitempty"`
	DepartTime     string `json:"depart_time,omitempty"`
	ArriveTime     string `json:"arrive_time,omitempty"`
	Stops          int    `json:"stops,omitempty"` // 0 = direct
	Aircraft       string `json:"aircraft,omitempty"`
	BaggageAllowance string `json:"baggage_allowance,omitempty"`
}

// Departure status constants.
const (
	DepartureStatusOpen      = "open"
	DepartureStatusFull      = "full"
	DepartureStatusClosed    = "closed"
	DepartureStatusCancelled = "cancelled"
)

// Price type constants.
const (
	PriceTypeStandard  = "standard"
	PriceTypeEarlyBird = "early_bird"
	PriceTypePromotion = "promotion"
)

// Destination represents a travel destination.
type Destination struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"column:name;size:100;not null" json:"name"`
	Province    string    `gorm:"column:province;size:50" json:"province,omitempty"`
	City        string    `gorm:"column:city;size:50" json:"city,omitempty"`
	CoverImage  string    `gorm:"column:cover_image;size:500" json:"cover_image,omitempty"`
	Description string    `gorm:"column:description;type:text" json:"description,omitempty"`
	SortOrder   int       `gorm:"column:sort_order;not null;default:0" json:"sort_order"`
	Status      string    `gorm:"column:status;size:20;not null;default:active" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:now()" json:"created_at"`
}

// TableName overrides the table name.
func (Destination) TableName() string {
	return "destination"
}
