// Package translation provides background translation of LLM-generated option text.
package translation

import "encoding/json"

// TranslatableAttribute holds label and value for translation.
// Only attributes with type "text" have their value translated;
// other types (price, score, boolean) copy value verbatim.
type TranslatableAttribute struct {
	Label string `json:"label"`
	Value string `json:"value"` // already stringified for text attrs
}

// TranslationInput is the compact payload sent to the LLM for translation.
type TranslationInput struct {
	Name       string                  `json:"name"`
	Notes      string                  `json:"notes,omitempty"`
	Warnings   []string                `json:"warnings,omitempty"`
	Category   string                  `json:"category,omitempty"`
	Attributes []TranslatableAttribute `json:"attributes,omitempty"`
}

// TranslationBlob is the LLM response shape — identical structure to TranslationInput.
// Stored verbatim as the locale value inside options.translations JSONB.
type TranslationBlob struct {
	Name       string                  `json:"name"`
	Notes      string                  `json:"notes,omitempty"`
	Warnings   []string                `json:"warnings,omitempty"`
	Category   string                  `json:"category,omitempty"`
	Attributes []TranslatableAttribute `json:"attributes,omitempty"`
}

// RawBlob serializes a TranslationBlob to json.RawMessage.
func RawBlob(b TranslationBlob) (json.RawMessage, error) {
	return json.Marshal(b)
}
