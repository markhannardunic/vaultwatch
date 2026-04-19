package filter_test

import (
	"testing"

	"github.com/vaultwatch/internal/filter"
)

func TestFilter_Allow_NoRules(t *testing.T) {
	f := filter.New(filter.Rule{})
	if !f.Allow("secret/myapp/db") {
		t.Error("expected path to be allowed with no rules")
	}
}

func TestFilter_Allow_MatchingPrefix(t *testing.T) {
	f := filter.New(filter.Rule{Prefixes: []string{"secret/myapp"}})
	if !f.Allow("secret/myapp/db") {
		t.Error("expected path to be allowed")
	}
}

func TestFilter_Allow_NonMatchingPrefix(t *testing.T) {
	f := filter.New(filter.Rule{Prefixes: []string{"secret/myapp"}})
	if f.Allow("secret/other/db") {
		t.Error("expected path to be denied")
	}
}

func TestFilter_Allow_ExcludeOverridesPrefix(t *testing.T) {
	f := filter.New(filter.Rule{
		Prefixes: []string{"secret/"},
		Exclude:  []string{"secret/internal"},
	})
	if f.Allow("secret/internal/key") {
		t.Error("expected excluded path to be denied")
	}
	if !f.Allow("secret/myapp/key") {
		t.Error("expected non-excluded path to be allowed")
	}
}

func TestFilter_Apply(t *testing.T) {
	f := filter.New(filter.Rule{
		Prefixes: []string{"secret/prod"},
		Exclude:  []string{"secret/prod/legacy"},
	})
	paths := []string{
		"secret/prod/api",
		"secret/prod/legacy/old",
		"secret/dev/api",
	}
	got := f.Apply(paths)
	if len(got) != 1 || got[0] != "secret/prod/api" {
		t.Errorf("unexpected result: %v", got)
	}
}
