package metrics

import (
	"bytes"
	"strings"
	"testing"
)

func TestCounter_Record_And_Snapshot(t *testing.T) {
	c := New()
	c.Record(LevelHealthy)
	c.Record(LevelWarning)
	c.Record(LevelWarning)
	c.Record(LevelCritical)

	snap := c.Snapshot()
	if snap[LevelHealthy] != 1 {
		t.Errorf("expected 1 healthy, got %d", snap[LevelHealthy])
	}
	if snap[LevelWarning] != 2 {
		t.Errorf("expected 2 warning, got %d", snap[LevelWarning])
	}
	if snap[LevelCritical] != 1 {
		t.Errorf("expected 1 critical, got %d", snap[LevelCritical])
	}
	if c.TotalRuns != 1 {
		t.Errorf("expected TotalRuns=1, got %d", c.TotalRuns)
	}
}

func TestCounter_Reset(t *testing.T) {
	c := New()
	c.Record(LevelCritical)
	c.Reset()
	snap := c.Snapshot()
	if Total(snap) != 0 {
		t.Errorf("expected 0 after reset, got %d", Total(snap))
	}
}

func TestTotal(t *testing.T) {
	counts := map[Level]int{
		LevelHealthy:  3,
		LevelWarning:  2,
		LevelCritical: 1,
	}
	if got := Total(counts); got != 6 {
		t.Errorf("expected 6, got %d", got)
	}
}

func TestSummary_ContainsLevels(t *testing.T) {
	counts := map[Level]int{
		LevelHealthy:  5,
		LevelWarning:  2,
		LevelCritical: 1,
	}
	var buf bytes.Buffer
	if err := Summary(&buf, counts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"critical", "warning", "healthy", "LEVEL", "COUNT"} {
		if !strings.Contains(out, want) {
			t.Errorf("summary missing %q", want)
		}
	}
}
