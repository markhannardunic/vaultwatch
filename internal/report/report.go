package report

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"
)

// Entry represents a single secret audit result.
type Entry struct {
	Path      string
	Level     string
	ExpiresAt time.Time
	Message   string
}

// Report holds a collection of audit entries.
type Report struct {
	entries []Entry
	writer  io.Writer
}

// New creates a new Report writing to w. If w is nil, os.Stdout is used.
func New(w io.Writer) *Report {
	if w == nil {
		w = os.Stdout
	}
	return &Report{writer: w}
}

// Add appends an entry to the report.
func (r *Report) Add(e Entry) {
	r.entries = append(r.entries, e)
}

// Summary returns counts by level.
func (r *Report) Summary() map[string]int {
	counts := map[string]int{}
	for _, e := range r.entries {
		counts[e.Level]++
	}
	return counts
}

// Render writes a formatted table of entries to the report's writer.
func (r *Report) Render() error {
	tw := tabwriter.NewWriter(r.writer, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PATH\tLEVEL\tEXPIRES\tMESSAGE")
	fmt.Fprintln(tw, "----\t-----\t-------\t-------")
	for _, e := range r.entries {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			e.Path,
			e.Level,
			e.ExpiresAt.Format(time.RFC3339),
			e.Message,
		)
	}
	if err := tw.Flush(); err != nil {
		return fmt.Errorf("report render: %w", err)
	}
	s := r.Summary()
	fmt.Fprintf(r.writer, "\nSummary: critical=%d warning=%d healthy=%d\n",
		s["critical"], s["warning"], s["healthy"])
	return nil
}
