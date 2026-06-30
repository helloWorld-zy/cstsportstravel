package model

import (
	"encoding/json"
	"testing"
)

func TestCountry_TableName(t *testing.T) {
	c := Country{}
	if c.TableName() != "country" {
		t.Errorf("expected 'country', got '%s'", c.TableName())
	}
}

func TestVisaMaterialTemplate_TableName(t *testing.T) {
	v := VisaMaterialTemplate{}
	if v.TableName() != "visa_material_template" {
		t.Errorf("expected 'visa_material_template', got '%s'", v.TableName())
	}
}

func TestContinentConstants(t *testing.T) {
	continents := []string{
		ContinentAsia, ContinentEurope, ContinentNorthAmerica,
		ContinentSouthAmerica, ContinentOceania, ContinentAfrica,
	}
	for _, c := range continents {
		if c == "" {
			t.Error("continent constant should not be empty")
		}
	}
}

func TestVisaTypeConstants(t *testing.T) {
	types := []string{
		VisaTypeFreeOnArrival, VisaTypeOnArrival,
		VisaTypeEVisa, VisaTypeRequired,
	}
	for _, vt := range types {
		if vt == "" {
			t.Error("visa type constant should not be empty")
		}
	}
}

func TestOccupationTypeConstants(t *testing.T) {
	occupations := []string{
		OccupationEmployed, OccupationFreelance,
		OccupationRetired, OccupationStudent, OccupationChild,
	}
	for _, o := range occupations {
		if o == "" {
			t.Error("occupation type constant should not be empty")
		}
	}
}

func TestVisaInfo_JSON(t *testing.T) {
	info := VisaInfo{
		VisaType:        VisaTypeRequired,
		ProcessingDays:  7,
		Fee:             50000, // 500 yuan in cents
		MaterialPreview: []string{"护照原件", "照片", "在职证明"},
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("failed to marshal VisaInfo: %v", err)
	}

	var decoded VisaInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal VisaInfo: %v", err)
	}

	if decoded.VisaType != VisaTypeRequired {
		t.Errorf("expected visa_type '%s', got '%s'", VisaTypeRequired, decoded.VisaType)
	}
	if decoded.ProcessingDays != 7 {
		t.Errorf("expected processing_days 7, got %d", decoded.ProcessingDays)
	}
	if len(decoded.MaterialPreview) != 3 {
		t.Errorf("expected 3 material preview items, got %d", len(decoded.MaterialPreview))
	}
}

func TestInsuranceRequirements_JSON(t *testing.T) {
	ins := InsuranceRequirements{
		Required:       true,
		MinMedicalCost: 3000000, // 30000 EUR in cents
		Schengen:       true,
		Description:    "申根签证要求旅行保险医疗保额≥3万欧元",
	}

	data, err := json.Marshal(ins)
	if err != nil {
		t.Fatalf("failed to marshal InsuranceRequirements: %v", err)
	}

	var decoded InsuranceRequirements
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal InsuranceRequirements: %v", err)
	}

	if !decoded.Required {
		t.Error("expected required to be true")
	}
	if !decoded.Schengen {
		t.Error("expected schengen to be true")
	}
}

func TestCountry_PassportValidityDefault(t *testing.T) {
	c := Country{
		PassportValidityMonths: 6,
	}
	if c.PassportValidityMonths != 6 {
		t.Errorf("expected default passport validity 6 months, got %d", c.PassportValidityMonths)
	}
}
