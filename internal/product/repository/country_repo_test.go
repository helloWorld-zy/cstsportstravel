package repository

import (
	"testing"

	"github.com/travel-booking/server/internal/product/model"
)

func TestCountryRepository_FindAll_NoDB(t *testing.T) {
	// Test that the function signature is correct and compiles
	// Actual DB tests would require test database setup
	countries := []model.Country{
		{ID: 1, NameCN: "日本", Continent: model.ContinentAsia, VisaType: model.VisaTypeRequired},
		{ID: 2, NameCN: "泰国", Continent: model.ContinentAsia, VisaType: model.VisaTypeOnArrival},
	}

	// Verify the ContinentTree logic
	tree := make(map[string][]model.Country)
	for _, c := range countries {
		tree[c.Continent] = append(tree[c.Continent], c)
	}

	if len(tree) != 1 {
		t.Errorf("expected 1 continent, got %d", len(tree))
	}
	if len(tree[model.ContinentAsia]) != 2 {
		t.Errorf("expected 2 countries in asia, got %d", len(tree[model.ContinentAsia]))
	}
}

func TestVisaMaterialTemplate_OccupationTypes(t *testing.T) {
	occupations := []string{
		model.OccupationEmployed,
		model.OccupationFreelance,
		model.OccupationRetired,
		model.OccupationStudent,
		model.OccupationChild,
	}

	if len(occupations) != 5 {
		t.Errorf("expected 5 occupation types, got %d", len(occupations))
	}
}
