package template

import (
	"testing"
)

func TestRegistry_All_ReturnsExpectedCount(t *testing.T) {
	r := NewRegistry()
	all := r.All()
	if len(all) < 10 {
		t.Errorf("All() returned %d templates; want at least 10", len(all))
	}
}

func TestRegistry_All_RequiredFieldsNonEmpty(t *testing.T) {
	r := NewRegistry()
	for _, tmpl := range r.All() {
		if tmpl.Slug == "" {
			t.Errorf("template %q has empty slug", tmpl.Name)
		}
		if tmpl.Name == "" {
			t.Errorf("template slug=%q has empty name", tmpl.Slug)
		}
		if tmpl.Icon == "" {
			t.Errorf("template slug=%q has empty icon", tmpl.Slug)
		}
		if tmpl.Category == "" {
			t.Errorf("template slug=%q has empty category", tmpl.Slug)
		}
		if tmpl.Description == "" {
			t.Errorf("template slug=%q has empty description", tmpl.Slug)
		}
		if len(tmpl.Constraints) == 0 {
			t.Errorf("template slug=%q has no constraints", tmpl.Slug)
		}
		if len(tmpl.CostCategories) == 0 {
			t.Errorf("template slug=%q has no cost categories", tmpl.Slug)
		}
	}
}

func TestRegistry_All_SlugsAreUnique(t *testing.T) {
	r := NewRegistry()
	seen := make(map[string]bool)
	for _, tmpl := range r.All() {
		if seen[tmpl.Slug] {
			t.Errorf("duplicate slug: %q", tmpl.Slug)
		}
		seen[tmpl.Slug] = true
	}
}

func TestRegistry_All_CategoriesAreValid(t *testing.T) {
	valid := map[string]bool{
		"travel":      true,
		"renovation":  true,
		"electronics": true,
		"computing":   true,
		"custom":      true,
	}
	r := NewRegistry()
	for _, tmpl := range r.All() {
		if !valid[tmpl.Category] {
			t.Errorf("template slug=%q has invalid category %q", tmpl.Slug, tmpl.Category)
		}
	}
}

func TestRegistry_GetBySlug_Found(t *testing.T) {
	r := NewRegistry()
	tmpl, err := r.GetBySlug("laptop")
	if err != nil {
		t.Fatalf("GetBySlug(%q) error: %v", "laptop", err)
	}
	if tmpl.Slug != "laptop" {
		t.Errorf("Slug = %q; want %q", tmpl.Slug, "laptop")
	}
}

func TestRegistry_GetBySlug_NotFound(t *testing.T) {
	r := NewRegistry()
	_, err := r.GetBySlug("nonexistent-template-slug")
	if err == nil {
		t.Error("GetBySlug(nonexistent) should return error")
	}
}

func TestRegistry_GetBySlug_AllTemplatesRetrievable(t *testing.T) {
	r := NewRegistry()
	for _, tmpl := range r.All() {
		got, err := r.GetBySlug(tmpl.Slug)
		if err != nil {
			t.Errorf("GetBySlug(%q) error: %v", tmpl.Slug, err)
			continue
		}
		if got.Name != tmpl.Name {
			t.Errorf("GetBySlug(%q).Name = %q; want %q", tmpl.Slug, got.Name, tmpl.Name)
		}
	}
}

func TestRegistry_WeightProfileSumsPositive(t *testing.T) {
	r := NewRegistry()
	for _, tmpl := range r.All() {
		sum := tmpl.WeightProfile.Price + tmpl.WeightProfile.Quality + tmpl.WeightProfile.Feature
		if sum <= 0 {
			t.Errorf("template slug=%q weight profile sums to %v; want > 0", tmpl.Slug, sum)
		}
	}
}
