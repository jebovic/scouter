package translation_test

import (
	"encoding/json"
	"testing"

	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/translation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleOption() option.Option {
	return option.Option{
		Name:     "Sony WH-1000XM5",
		Notes:    "Best-in-class ANC",
		Warnings: []string{"Expensive"},
		Category: "Headphones",
		Attributes: []option.Attribute{
			{Key: "anc", Label: "ANC Quality", Value: "Excellent", Type: "text"},
			{Key: "price", Label: "Price", Value: 350.0, Type: "price"},
			{Key: "score", Label: "Overall", Value: 9, Type: "score"},
			{Key: "foldable", Label: "Foldable", Value: true, Type: "boolean"},
		},
	}
}

func TestTranslatableFields_ExtractsTextOnly(t *testing.T) {
	input := translation.TranslatableFields(sampleOption())

	assert.Equal(t, "Sony WH-1000XM5", input.Name)
	assert.Equal(t, "Best-in-class ANC", input.Notes)
	assert.Equal(t, []string{"Expensive"}, input.Warnings)
	assert.Equal(t, "Headphones", input.Category)
	require.Len(t, input.Attributes, 1, "only text attributes should be included")
	assert.Equal(t, "ANC Quality", input.Attributes[0].Label)
	assert.Equal(t, "Excellent", input.Attributes[0].Value)
}

func TestMergeTranslation_OverlaysTextFields(t *testing.T) {
	base := sampleOption()
	blob := translation.TranslationBlob{
		Name:     "Sony WH-1000XM5 (FR)",
		Notes:    "ANC de première classe",
		Warnings: []string{"Cher"},
		Category: "Casques",
		Attributes: []translation.TranslatableAttribute{
			{Label: "Qualité ANC", Value: "Excellente"},
		},
	}
	raw, err := json.Marshal(blob)
	require.NoError(t, err)

	merged := translation.MergeTranslation(base, "fr", raw)

	assert.Equal(t, "Sony WH-1000XM5 (FR)", merged.Name)
	assert.Equal(t, "ANC de première classe", merged.Notes)
	assert.Equal(t, []string{"Cher"}, merged.Warnings)
	assert.Equal(t, "Casques", merged.Category)

	// Text attribute should be translated; others preserved from base
	require.Len(t, merged.Attributes, 4)
	var textAttr option.Attribute
	for _, a := range merged.Attributes {
		if a.Type == "text" {
			textAttr = a
		}
	}
	assert.Equal(t, "Qualité ANC", textAttr.Label)
	assert.Equal(t, "Excellente", textAttr.Value)

	// price attribute label unchanged
	for _, a := range merged.Attributes {
		if a.Type == "price" {
			assert.Equal(t, "Price", a.Label, "non-text label unchanged")
		}
	}
}

func TestMergeTranslation_NoopWhenNilBlob(t *testing.T) {
	base := sampleOption()
	merged := translation.MergeTranslation(base, "fr", nil)
	assert.Equal(t, base.Name, merged.Name)
}
