package output

import "testing"

func TestParseFormat_Valid(t *testing.T) {
	cases := []struct {
		input    string
		expected Format
	}{
		{"text", FormatText},
		{"TEXT", FormatText},
		{"", FormatText},
		{"json", FormatJSON},
		{"JSON", FormatJSON},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := ParseFormat(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestParseFormat_Invalid(t *testing.T) {
	_, err := ParseFormat("yaml")
	if err == nil {
		t.Error("expected error for unknown format")
	}
}
