package report

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"
)

// Entry holds audit data for a single secret.
type Entry struct {
	Path     string
	Level    string
	Expiry   time.Time
	DaysLeft int
}

// Report holds a collection of secret audit entries.
type Report struct {
	entries []Entry
}

// New creates an empty Report.
func New() *Report {
	return &Report{}
}

// Render writes a formatted table of entries to w.
func (r *Report) Render(w io.Writer) {
	if len(r.entries) == 0 {
		fmt.Fprintln(w, "No secrets audited.")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PATH\tLEVEL\tEXPIRY\tDAYS LEFT")
	for _, e := range r.entries {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\n",
			e.Path, e.Level, e.Expiry.Format("2006-01-02"), e.DaysLeft)
	}
	tw.Flush()
	r.renderSummary(w)
}

// renderSummary prints counts by level.
func (r *Report) renderSummary(w io.Writer) {
	counts := r.Summary()
	fmt.Fprintf(w, "\nSummary: %d critical, %d warning, %d healthy\n",
		counts["critical"], counts["warning"], counts["healthy"])
}

// Summary returns a map of level -> count.
func (r *Report) Summary() map[string]int {
	m := map[string]int{"critical": 0, "warning": 0, "healthy": 0}
	for _, e := range r.entries {
		if _, ok := m[e.Level]; ok {
			m[e.Level]++
		}
	}
	return m
}

// Print renders the report to stdout.
func (r *Report) Print() {
	r.Render(os.Stdout)
}
