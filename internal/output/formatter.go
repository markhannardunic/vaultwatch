package output

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"
)

// Format controls the output format.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Entry represents a single secret status line.
type Entry struct {
	Path      string
	Status    string
	ExpiresAt time.Time
	DaysLeft  int
}

// Formatter writes entries to an io.Writer in a given format.
type Formatter struct {
	w      io.Writer
	format Format
}

// NewFormatter creates a Formatter writing to w.
func NewFormatter(w io.Writer, format Format) *Formatter {
	return &Formatter{w: w, format: format}
}

// Write renders all entries.
func (f *Formatter) Write(entries []Entry) error {
	switch f.format {
	case FormatJSON:
		return writeJSON(f.w, entries)
	default:
		return writeText(f.w, entries)
	}
}

func writeText(w io.Writer, entries []Entry) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PATH\tSTATUS\tDAYS LEFT\tEXPIRES AT")
	for _, e := range entries {
		fmt.Fprintf(tw, "%s\t%s\t%d\t%s\n",
			e.Path, e.Status, e.DaysLeft, e.ExpiresAt.Format(time.RFC3339))
	}
	return tw.Flush()
}
