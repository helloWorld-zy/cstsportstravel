package service

import (
	"testing"

	"github.com/travel-booking/server/internal/product/model"
)

func TestVisaFormService_GenerateForm(t *testing.T) {
	svc := NewVisaFormService()

	country := &model.Country{
		ID:       1,
		NameCN:   "日本",
		NameEN:   "Japan",
		VisaType: model.VisaTypeRequired,
	}

	form := svc.GenerateForm(country)

	if form.CountryID != 1 {
		t.Errorf("expected country_id 1, got %d", form.CountryID)
	}
	if form.CountryName != "日本" {
		t.Errorf("expected country_name '日本', got '%s'", form.CountryName)
	}
	if form.Title != "日本签证申请表" {
		t.Errorf("expected title '日本签证申请表', got '%s'", form.Title)
	}
	if len(form.Fields) == 0 {
		t.Error("expected fields to be non-empty")
	}
	if len(form.Groups) == 0 {
		t.Error("expected groups to be non-empty")
	}
}

func TestVisaFormService_FieldGroups(t *testing.T) {
	svc := NewVisaFormService()

	country := &model.Country{
		ID:       1,
		NameCN:   "泰国",
		NameEN:   "Thailand",
		VisaType: model.VisaTypeOnArrival,
	}

	form := svc.GenerateForm(country)

	groupNames := make(map[string]bool)
	for _, g := range form.Groups {
		groupNames[g.Name] = true
	}

	expectedGroups := []string{"personal", "passport", "travel", "employment", "contact"}
	for _, g := range expectedGroups {
		if !groupNames[g] {
			t.Errorf("expected group '%s' to exist", g)
		}
	}
}

func TestVisaFormService_RequiredFields(t *testing.T) {
	svc := NewVisaFormService()

	country := &model.Country{
		ID:       1,
		NameCN:   "日本",
		NameEN:   "Japan",
		VisaType: model.VisaTypeRequired,
	}

	form := svc.GenerateForm(country)

	requiredFields := []string{
		"surname_cn", "given_name_cn", "surname_en", "given_name_en",
		"gender", "birth_date", "passport_number", "passport_expiry_date",
		"purpose", "entry_date", "exit_date",
		"emergency_contact_name", "emergency_contact_phone",
	}

	for _, fieldName := range requiredFields {
		found := false
		for _, f := range form.Fields {
			if f.Name == fieldName {
				found = true
				if !f.Required {
					t.Errorf("field '%s' should be required", fieldName)
				}
				break
			}
		}
		if !found {
			t.Errorf("expected required field '%s' not found", fieldName)
		}
	}
}

func TestVisaFormService_SchengenCountry(t *testing.T) {
	svc := NewVisaFormService()

	// Test Schengen country (France)
	country := &model.Country{
		ID:       2,
		NameCN:   "法国",
		NameEN:   "France",
		VisaType: model.VisaTypeRequired,
	}

	form := svc.GenerateForm(country)

	// Check for Schengen-specific fields
	schengenFields := []string{"schengen_main_destination", "schengen_first_entry", "insurance_policy_number"}
	for _, fieldName := range schengenFields {
		found := false
		for _, f := range form.Fields {
			if f.Name == fieldName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected Schengen field '%s' for France", fieldName)
		}
	}
}

func TestVisaFormService_NonSchengenCountry(t *testing.T) {
	svc := NewVisaFormService()

	// Test non-Schengen country (Japan)
	country := &model.Country{
		ID:       1,
		NameCN:   "日本",
		NameEN:   "Japan",
		VisaType: model.VisaTypeRequired,
	}

	form := svc.GenerateForm(country)

	// Check that Schengen-specific fields are NOT present
	schengenFields := []string{"schengen_main_destination", "schengen_first_entry", "insurance_policy_number"}
	for _, fieldName := range schengenFields {
		for _, f := range form.Fields {
			if f.Name == fieldName {
				t.Errorf("Schengen field '%s' should not exist for Japan", fieldName)
			}
		}
	}
}

func TestVisaFormService_FieldTypes(t *testing.T) {
	svc := NewVisaFormService()

	country := &model.Country{
		ID:       1,
		NameCN:   "日本",
		NameEN:   "Japan",
		VisaType: model.VisaTypeRequired,
	}

	form := svc.GenerateForm(country)

	// Check that select fields have options
	for _, f := range form.Fields {
		if f.Type == "select" && len(f.Options) == 0 {
			t.Errorf("select field '%s' should have options", f.Name)
		}
	}
}

func TestIsSchengenCountry(t *testing.T) {
	svc := NewVisaFormService()

	tests := []struct {
		country string
		want    bool
	}{
		{"France", true},
		{"Germany", true},
		{"Italy", true},
		{"Spain", true},
		{"Japan", false},
		{"USA", false},
		{"Thailand", false},
	}

	for _, tt := range tests {
		t.Run(tt.country, func(t *testing.T) {
			got := svc.isSchengenCountry(tt.country)
			if got != tt.want {
				t.Errorf("isSchengenCountry(%s) = %v, want %v", tt.country, got, tt.want)
			}
		})
	}
}
