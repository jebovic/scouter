package template

import "fmt"

// Registry holds all built-in templates indexed by slug for O(1) lookup.
type Registry struct {
	all    []Template
	bySlug map[string]Template
}

// NewRegistry constructs a Registry from the compiled-in catalog.
func NewRegistry() *Registry {
	bySlug := make(map[string]Template, len(catalog))
	for _, t := range catalog {
		bySlug[t.Slug] = t
	}
	return &Registry{all: catalog, bySlug: bySlug}
}

// All returns a copy of all templates in definition order.
// Callers may not mutate the returned slice or its elements' slice fields.
func (r *Registry) All() []Template {
	out := make([]Template, len(r.all))
	copy(out, r.all)
	return out
}

// GetBySlug returns the template with the given slug, or an error if not found.
func (r *Registry) GetBySlug(slug string) (Template, error) {
	t, ok := r.bySlug[slug]
	if !ok {
		return Template{}, fmt.Errorf("template %q not found", slug)
	}
	return t, nil
}
