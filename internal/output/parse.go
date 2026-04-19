package output

import (
	"fmt"
	"strings"
)

// ParseFormat converts a string to a Format, returning an error for unknown values.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "text", "":
		return FormatText, nil
	case "json":
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("unknown output format %q: must be \"text\" or \"json\"", s)
	}
}
