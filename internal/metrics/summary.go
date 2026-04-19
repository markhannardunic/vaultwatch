package metrics

import (
	"fmt"
	"io"
	"text/tabwriter"
)

// Summary prints a formatted summary of the counter snapshot to w.
func Summary(w io.Writer, counts map[Level]int) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "LEVEL\tCOUNT")
	fmt.Fprintln(tw, "-----\t-----")
	for _, level := range []Level{LevelCritical, LevelWarning, LevelHealthy} {
		fmt.Fprintf(tw, "%s\t%d\n", level, counts[level])
	}
	return tw.Flush()
}

// Total returns the sum of all counts.
func Total(counts map[Level]int) int {
	n := 0
	for _, v := range counts {
		n += v
	}
	return n
}
