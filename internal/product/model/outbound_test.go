package model

import (
	"encoding/json"
	"testing"
)

func TestProduct_OutboundFields(t *testing.T) {
	visaInfo := VisaInfo{
		VisaType:       VisaTypeRequired,
		ProcessingDays: 7,
		Fee:            50000,
	}
	visaData, _ := json.Marshal(visaInfo)

	insurance := InsuranceRequirements{
		Required: true,
		Schengen: true,
	}
	insuranceData, _ := json.Marshal(insurance)

	flightInfo := InternationalFlightInfo{
		Airline:     "中国国航",
		FlightNo:    "CA925",
		DepartCity:  "北京",
		ArriveCity:  "东京",
	}
	flightData, _ := json.Marshal(flightInfo)

	services := PreTripServices{
		EntryPolicy:    true,
		EntryMaterials: true,
		FlightTracking: true,
	}
	servicesData, _ := json.Marshal(services)

	p := Product{
		ProductType:             ProductTypeOutbound,
		DestinationCountryID:    int64Ptr(1),
		VisaInfo:                visaData,
		InsuranceRequirements:   insuranceData,
		InternationalFlightInfo: flightData,
		PreTripServices:         servicesData,
	}

	if p.ProductType != ProductTypeOutbound {
		t.Errorf("expected product_type '%s', got '%s'", ProductTypeOutbound, p.ProductType)
	}

	if p.DestinationCountryID == nil || *p.DestinationCountryID != 1 {
		t.Error("expected destination_country_id to be 1")
	}

	var decodedVisa VisaInfo
	if err := json.Unmarshal(p.VisaInfo, &decodedVisa); err != nil {
		t.Fatalf("failed to unmarshal visa_info: %v", err)
	}
	if decodedVisa.VisaType != VisaTypeRequired {
		t.Errorf("expected visa_type '%s', got '%s'", VisaTypeRequired, decodedVisa.VisaType)
	}

	var decodedInsurance InsuranceRequirements
	if err := json.Unmarshal(p.InsuranceRequirements, &decodedInsurance); err != nil {
		t.Fatalf("failed to unmarshal insurance_requirements: %v", err)
	}
	if !decodedInsurance.Schengen {
		t.Error("expected schengen to be true")
	}

	var decodedFlight InternationalFlightInfo
	if err := json.Unmarshal(p.InternationalFlightInfo, &decodedFlight); err != nil {
		t.Fatalf("failed to unmarshal international_flight_info: %v", err)
	}
	if decodedFlight.FlightNo != "CA925" {
		t.Errorf("expected flight_no 'CA925', got '%s'", decodedFlight.FlightNo)
	}

	var decodedServices PreTripServices
	if err := json.Unmarshal(p.PreTripServices, &decodedServices); err != nil {
		t.Fatalf("failed to unmarshal pre_trip_services: %v", err)
	}
	if !decodedServices.EntryPolicy {
		t.Error("expected entry_policy to be true")
	}
}

func TestProductType_Outbound(t *testing.T) {
	if ProductTypeOutbound != "outbound_group" {
		t.Errorf("expected 'outbound_group', got '%s'", ProductTypeOutbound)
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}
