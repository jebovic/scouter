package translation

import (
	"encoding/json"
	"fmt"

	"github.com/jibei/scouter/internal/option"
)

// TranslatableFields extracts only the content that needs LLM translation.
// Only "text" type attributes are included; others are omitted.
func TranslatableFields(o option.Option) TranslationInput {
	input := TranslationInput{
		Name:     o.Name,
		Notes:    o.Notes,
		Warnings: append([]string(nil), o.Warnings...),
		Category: o.Category,
	}
	for _, a := range o.Attributes {
		if a.Type != "text" {
			continue
		}
		val := fmt.Sprintf("%v", a.Value)
		input.Attributes = append(input.Attributes, TranslatableAttribute{
			Label: a.Label,
			Value: val,
		})
	}
	return input
}

// MergeTranslation overlays translated text fields on top of the English base option.
// Returns a new Option value — never mutates the input.
// If blob is nil or unmarshalling fails, the original option is returned unchanged.
func MergeTranslation(base option.Option, locale string, blob json.RawMessage) option.Option {
	if blob == nil {
		return base
	}
	var tb TranslationBlob
	if err := json.Unmarshal(blob, &tb); err != nil {
		return base
	}

	out := base // shallow copy
	out.Name = tb.Name
	if tb.Notes != "" {
		out.Notes = tb.Notes
	}
	if len(tb.Warnings) > 0 {
		out.Warnings = append([]string(nil), tb.Warnings...)
	}
	if tb.Category != "" {
		out.Category = tb.Category
	}

	// Merge text-typed attributes: match by position within the text-attr subsequence.
	newAttrs := make([]option.Attribute, len(base.Attributes))
	copy(newAttrs, base.Attributes)
	textIdx := 0
	for i, a := range newAttrs {
		if a.Type != "text" {
			continue
		}
		if textIdx < len(tb.Attributes) {
			newAttrs[i].Label = tb.Attributes[textIdx].Label
			newAttrs[i].Value = tb.Attributes[textIdx].Value
		}
		textIdx++
	}
	out.Attributes = newAttrs
	return out
}
