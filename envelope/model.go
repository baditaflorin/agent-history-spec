package envelope

import (
	"regexp"
	"strings"
)

var (
	// modelPattern is the form of Event.Model: a lower-case id of at most 128 bytes with no
	// whitespace and no control characters. It allows what real ids use: letters, digits and
	// . _ : / @ + - [ ] (claude-opus-5-5, gpt-5.6-terra, claude-opus-5-5[1m], llama3:8b, vendor/name).
	modelPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._:/@+\[\]-]{0,127}$`)
	// effortPattern is the form of Event.Effort: a lower-case word of at most 32 bytes.
	effortPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,31}$`)
)

// CleanModel returns s trimmed and lower-cased when the result is a valid Event.Model, and ""
// otherwise. A producer calls it on whatever the tool wrote, so that an odd value costs the field
// and never the event. "<synthetic>", the placeholder some tools write for records they made
// themselves, is not a model and comes back empty.
func CleanModel(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if !modelPattern.MatchString(s) {
		return ""
	}
	return s
}

// CleanEffort is CleanModel for Event.Effort.
func CleanEffort(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if !effortPattern.MatchString(s) {
		return ""
	}
	return s
}
