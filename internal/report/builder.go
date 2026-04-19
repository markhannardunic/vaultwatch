package report

import "time"

// Builder accumulates report entries from checker results.
type Builder struct {
	report *Report
}

// NewBuilder creates a new Builder.
func NewBuilder() *Builder {
	return &Builder{report: New()}
}

// Add appends an entry to the report.
func (b *Builder) Add(path, level string, expiry time.Time, daysLeft int) {
	b.report.entries = append(b.report.entries, Entry{
		Path:     path,
		Level:    level,
		Expiry:   expiry,
		DaysLeft: daysLeft,
	})
}

// Build returns the completed Report.
func (b *Builder) Build() *Report {
	return b.report
}

// Len returns the number of entries added so far.
func (b *Builder) Len() int {
	return len(b.report.entries)
}
