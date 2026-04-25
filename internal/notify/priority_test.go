package notify

import (
	"fmt"
	"testing"
)

// collectSender records all messages sent to it.
type collectSender struct {
	msgs []string
	fail bool
}

func (c *collectSender) Send(msg string) error {
	if c.fail {
		return fmt.Errorf("send error")
	}
	c.msgs = append(c.msgs, msg)
	return nil
}

func TestNewPriorityQueue_NilSender(t *testing.T) {
	_, err := NewPriorityQueue(nil)
	if err == nil {
		t.Fatal("expected error for nil sender")
	}
}

func TestNewPriorityQueue_ValidSender(t *testing.T) {
	s := &collectSender{}
	pq, err := NewPriorityQueue(s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pq == nil {
		t.Fatal("expected non-nil PriorityQueue")
	}
}

func TestPriorityQueue_Flush_OrderedByCriticality(t *testing.T) {
	s := &collectSender{}
	pq, _ := NewPriorityQueue(s)

	pq.Enqueue(PriorityInfo, "info-msg")
	pq.Enqueue(PriorityWarning, "warning-msg")
	pq.Enqueue(PriorityCritical, "critical-msg")

	if err := pq.Flush(); err != nil {
		t.Fatalf("unexpected flush error: %v", err)
	}

	if len(s.msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(s.msgs))
	}
	if s.msgs[0] != "critical-msg" {
		t.Errorf("expected critical first, got %q", s.msgs[0])
	}
	if s.msgs[1] != "warning-msg" {
		t.Errorf("expected warning second, got %q", s.msgs[1])
	}
	if s.msgs[2] != "info-msg" {
		t.Errorf("expected info third, got %q", s.msgs[2])
	}
}

func TestPriorityQueue_Flush_EmptyBuckets(t *testing.T) {
	s := &collectSender{}
	pq, _ := NewPriorityQueue(s)
	if err := pq.Flush(); err != nil {
		t.Fatalf("unexpected error on empty flush: %v", err)
	}
	if len(s.msgs) != 0 {
		t.Errorf("expected no messages, got %d", len(s.msgs))
	}
}

func TestPriorityQueue_Flush_ReturnsErrorOnSendFailure(t *testing.T) {
	s := &collectSender{fail: true}
	pq, _ := NewPriorityQueue(s)
	pq.Enqueue(PriorityCritical, "will-fail")
	if err := pq.Flush(); err == nil {
		t.Fatal("expected error from failed send")
	}
}

func TestLevelToPriority(t *testing.T) {
	cases := []struct {
		level    string
		want     Priority
	}{
		{"critical", PriorityCritical},
		{"CRITICAL", PriorityCritical},
		{"warning", PriorityWarning},
		{"Warning", PriorityWarning},
		{"info", PriorityInfo},
		{"unknown", PriorityInfo},
	}
	for _, tc := range cases {
		got := LevelToPriority(tc.level)
		if got != tc.want {
			t.Errorf("LevelToPriority(%q) = %d, want %d", tc.level, got, tc.want)
		}
	}
}
