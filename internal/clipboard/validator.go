package clipboard

import (
	"strings"

	"github.com/atotto/clipboard"
	"github.com/surge-downloader/surge/internal/source"
)

// Validator checks and extracts valid downloadable URLs from text
type Validator struct {
}

// NewValidator creates a new URL validator
func NewValidator() *Validator {
	return &Validator{}
}

// ExtractURL validates and returns a clean URL, or empty string if invalid
func (v *Validator) ExtractURL(text string) string {
	text = strings.TrimSpace(text)

	// Quick reject: too long, contains newlines, or obviously not a URL
	if len(text) > 2048 || strings.ContainsAny(text, "\n\r") {
		return ""
	}

	if !source.IsSupported(text) {
		return ""
	}

	return source.Normalize(text)
}

// ReadURL reads the clipboard and returns a valid URL if found, or empty string otherwise
func ReadURL() string {
	text, err := clipboard.ReadAll()
	if err != nil {
		return ""
	}
	validator := NewValidator()
	return validator.ExtractURL(text)
}
