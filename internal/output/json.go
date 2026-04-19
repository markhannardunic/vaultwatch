package output

import (
	"encoding/json"
	"io"
	"time"
)

type jsonEntry struct {
	Path      string `json:"path"`
	Status    string `json:"status"`
	ExpiresAt string `json:"expires_at"`
	DaysLeft  int    `json:"days_left"`
}

type jsonReport struct {
	GeneratedAt string      `json:"generated_at"`
	Entries     []jsonEntry `json:"entries"`
	Total       int         `json:"total"`
}

func writeJSON(w io.Writer, entries []Entry) error {
	je := make([]jsonEntry, len(entries))
	for i, e := range entries {
		je[i] = jsonEntry{
			Path:      e.Path,
			Status:    e.Status,
			ExpiresAt: e.ExpiresAt.Format(time.RFC3339),
			DaysLeft:  e.DaysLeft,
		}
	}
	report := jsonReport{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Entries:     je,
		Total:       len(je),
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}
