package notify

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

type captureSender struct {
	msgs []string
	fail bool
}

func (c *captureSender) Send(path, level, message string) error {
	if c.fail {
		return errors.New("send failed")
	}
	c.msgs = append(c.msgs, fmt.Sprintf("%s|%s|%s", path, level, message))
	return nil
}

func TestNewReplayStore_InvalidMaxSize(t *testing.T) {
	_, err := NewReplayStore(0)
	if err == nil {
		t.Fatal("expected error for maxSize=0")
	}
}

func TestNewReplayStore_Valid(t *testing.T) {
	s, err := NewReplayStore(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Len() != 0 {
		t.Fatalf("expected empty store, got %d", s.Len())
	}
}

func TestReplayStore_RecordAndLen(t *testing.T) {
	s, _ := NewReplayStore(5)
	s.Record("secret/a", "warning", "expires soon")
	s.Record("secret/b", "critical", "expires now")
	if s.Len() != 2 {
		t.Fatalf("expected 2 entries, got %d", s.Len())
	}
}

func TestReplayStore_EvictsOldestWhenFull(t *testing.T) {
	s, _ := NewReplayStore(2)
	s.Record("a", "warning", "first")
	s.Record("b", "warning", "second")
	s.Record("c", "critical", "third")
	if s.Len() != 2 {
		t.Fatalf("expected 2 entries after eviction, got %d", s.Len())
	}
	sender := &captureSender{}
	_ = s.Replay(sender)
	for _, m := range sender.msgs {
		if strings.Contains(m, "first") {
			t.Error("oldest entry should have been evicted")
		}
	}
}

func TestReplayStore_Replay_SendsAll(t *testing.T) {
	s, _ := NewReplayStore(10)
	s.Record("secret/x", "warning", "msg1")
	s.Record("secret/y", "critical", "msg2")
	sender := &captureSender{}
	if err := s.Replay(sender); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sender.msgs) != 2 {
		t.Fatalf("expected 2 replayed messages, got %d", len(sender.msgs))
	}
}

func TestReplayStore_Replay_CollectsErrors(t *testing.T) {
	s, _ := NewReplayStore(5)
	s.Record("secret/a", "warning", "fail me")
	sender := &captureSender{fail: true}
	err := s.Replay(sender)
	if err == nil {
		t.Fatal("expected error when sender fails")
	}
}

func TestReplayStore_Clear(t *testing.T) {
	s, _ := NewReplayStore(5)
	s.Record("secret/a", "warning", "x")
	s.Clear()
	if s.Len() != 0 {
		t.Fatalf("expected empty store after Clear, got %d", s.Len())
	}
}
