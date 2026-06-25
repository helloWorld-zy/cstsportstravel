// Package service provides business logic for the Order domain.
//
// This file implements pricing calculation rules per PRD §4.2.5:
//   - Single room supplement: auto-added when adult count is odd
//   - Child pricing: 2-12 years old, no bed, uses child_price
//   - Infant pricing: <2 years old, uses infant_price
//   - Child must link to adult; max 1 infant per adult
package service

import (
	"fmt"
	"time"
)

// PriceBreakdown holds the detailed pricing calculation result.
type PriceBreakdown struct {
	AdultPrice         int64 `json:"adult_price"`          // per-adult price in cents
	ChildPrice         int64 `json:"child_price"`          // per-child price in cents
	InfantPrice        int64 `json:"infant_price"`         // per-infant price in cents
	SingleSupplement   int64 `json:"single_supplement"`    // per-unit supplement in cents
	AdultCount         int   `json:"adult_count"`
	ChildCount         int   `json:"child_count"`
	InfantCount        int   `json:"infant_count"`
	AdultSubtotal      int64 `json:"adult_subtotal"`       // adults * adult_price
	ChildSubtotal      int64 `json:"child_subtotal"`       // children * child_price
	InfantSubtotal     int64 `json:"infant_subtotal"`      // infants * infant_price
	SupplementCount    int   `json:"supplement_count"`      // number of supplement units
	SupplementSubtotal int64 `json:"supplement_subtotal"`  // supplement_count * single_supplement
	AddonSubtotal      int64 `json:"addon_subtotal"`       // addon services total
	TotalAmount        int64 `json:"total_amount"`         // sum of all subtotals
	DiscountAmount     int64 `json:"discount_amount"`      // discount (MVP: 0)
	PayableAmount      int64 `json:"payable_amount"`       // total - discount
}

// CalculatePrice computes the order price breakdown.
// Single room supplement rule (PRD §4.2.5): when adult count is odd, auto-add 1 supplement.
func CalculatePrice(
	adultPrice, childPrice, infantPrice, singleSupplement int64,
	adultCount, childCount, infantCount int,
	addonAmount int64,
) *PriceBreakdown {
	bd := &PriceBreakdown{
		AdultPrice:       adultPrice,
		ChildPrice:       childPrice,
		InfantPrice:      infantPrice,
		SingleSupplement: singleSupplement,
		AdultCount:       adultCount,
		ChildCount:       childCount,
		InfantCount:      infantCount,
		AddonSubtotal:    addonAmount,
	}

	// Calculate subtotals
	bd.AdultSubtotal = int64(adultCount) * adultPrice
	bd.ChildSubtotal = int64(childCount) * childPrice
	bd.InfantSubtotal = int64(infantCount) * infantPrice

	// Single room supplement: auto-add when adult count is odd
	// Per PRD §4.2.5: "成人数为奇数时自动附加1份单房差费用"
	if adultCount > 0 && adultCount%2 != 0 {
		bd.SupplementCount = 1
		bd.SupplementSubtotal = singleSupplement
	}

	// Total
	bd.TotalAmount = bd.AdultSubtotal + bd.ChildSubtotal + bd.InfantSubtotal +
		bd.SupplementSubtotal + bd.AddonSubtotal

	// MVP: no discount
	bd.DiscountAmount = 0
	bd.PayableAmount = bd.TotalAmount - bd.DiscountAmount

	return bd
}

// ValidateTravellerConstraints validates traveller business rules per PRD §4.2.5.
// Rules:
//   - Total travellers must equal adultCount + childCount + infantCount
//   - Children (is_child=true) must have linked_adult_traveller_index set
//   - Infants (is_infant=true) must have linked_adult_traveller_index set
//   - Max 1 infant per adult
//   - Child age: 2-12 years old (validated by birth_date)
//   - Infant age: <2 years old (validated by birth_date)
type TravellerInput struct {
	RealName                string  `json:"real_name"`
	IDCardNo                string  `json:"id_card_no"`
	Phone                   string  `json:"phone"`
	BirthDate               string  `json:"birth_date"`
	Gender                  string  `json:"gender"`
	IsChild                 bool    `json:"is_child"`
	IsInfant                bool    `json:"is_infant"`
	LinkedAdultTravellerIndex *int  `json:"linked_adult_traveller_index,omitempty"`
}

// TravellerValidationError holds a validation error for a specific traveller.
type TravellerValidationError struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
}

func (e *TravellerValidationError) Error() string {
	return fmt.Sprintf("traveller[%d]: %s", e.Index, e.Message)
}

// ValidateTravellers validates traveller inputs against business rules.
func ValidateTravellers(travellers []TravellerInput, adultCount, childCount, infantCount int, departureDate time.Time) []TravellerValidationError {
	var errs []TravellerValidationError

	// Check total count
	expectedTotal := adultCount + childCount + infantCount
	if len(travellers) != expectedTotal {
		errs = append(errs, TravellerValidationError{
			Index:   -1,
			Message: fmt.Sprintf("traveller count mismatch: expected %d, got %d", expectedTotal, len(travellers)),
		})
		return errs
	}

	// Count actual adults/children/infants and track infant-per-adult
	actualAdults := 0
	actualChildren := 0
	actualInfants := 0
	infantPerAdult := make(map[int]int) // adult index -> infant count

	for i, t := range travellers {
		// Validate ID card format
		if t.IDCardNo == "" {
			errs = append(errs, TravellerValidationError{
				Index:   i,
				Message: "id_card_no is required",
			})
			continue
		}
		if !ValidateIDCard(t.IDCardNo) {
			errs = append(errs, TravellerValidationError{
				Index:   i,
				Message: "invalid id_card_no format",
			})
			continue
		}

		// Determine age category from birth date
		if t.IsChild {
			actualChildren++
			// Child must link to an adult
			if t.LinkedAdultTravellerIndex == nil {
				errs = append(errs, TravellerValidationError{
					Index:   i,
					Message: "child must have linked_adult_traveller_index",
				})
			} else {
				adultIdx := *t.LinkedAdultTravellerIndex
				if adultIdx < 0 || adultIdx >= len(travellers) || travellers[adultIdx].IsChild || travellers[adultIdx].IsInfant {
					errs = append(errs, TravellerValidationError{
						Index:   i,
						Message: "linked_adult_traveller_index is invalid",
					})
				}
			}
		} else if t.IsInfant {
			actualInfants++
			// Infant must link to an adult
			if t.LinkedAdultTravellerIndex == nil {
				errs = append(errs, TravellerValidationError{
					Index:   i,
					Message: "infant must have linked_adult_traveller_index",
				})
			} else {
				adultIdx := *t.LinkedAdultTravellerIndex
				if adultIdx < 0 || adultIdx >= len(travellers) || travellers[adultIdx].IsChild || travellers[adultIdx].IsInfant {
					errs = append(errs, TravellerValidationError{
						Index:   i,
						Message: "linked_adult_traveller_index is invalid",
					})
				} else {
					infantPerAdult[adultIdx]++
				}
			}
		} else {
			actualAdults++
		}

		// Validate real_name
		if t.RealName == "" {
			errs = append(errs, TravellerValidationError{
				Index:   i,
				Message: "real_name is required",
			})
		}
	}

	// Check infant limit: max 1 per adult
	for adultIdx, count := range infantPerAdult {
		if count > 1 {
			errs = append(errs, TravellerValidationError{
				Index:   adultIdx,
				Message: fmt.Sprintf("adult can have at most 1 infant, got %d", count),
			})
		}
	}

	// Verify counts match
	if actualAdults != adultCount {
		errs = append(errs, TravellerValidationError{
			Index:   -1,
			Message: fmt.Sprintf("adult count mismatch: expected %d, got %d", adultCount, actualAdults),
		})
	}
	if actualChildren != childCount {
		errs = append(errs, TravellerValidationError{
			Index:   -1,
			Message: fmt.Sprintf("child count mismatch: expected %d, got %d", childCount, actualChildren),
		})
	}
	if actualInfants != infantCount {
		errs = append(errs, TravellerValidationError{
			Index:   -1,
			Message: fmt.Sprintf("infant count mismatch: expected %d, got %d", infantCount, actualInfants),
		})
	}

	return errs
}
