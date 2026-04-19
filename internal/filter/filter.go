package filter

import "strings"

// Rule defines inclusion/exclusion criteria for secrets.
type Rule struct {
	Prefixes []string
	Exclude  []string
}

// Filter decides which secret paths should be audited.
type Filter struct {
	rule Rule
}

// New creates a Filter from the given Rule.
func New(r Rule) *Filter {
	return &Filter{rule: r}
}

// Allow returns true if the given path should be audited.
func (f *Filter) Allow(path string) bool {
	for _, ex := range f.rule.Exclude {
		if strings.HasPrefix(path, ex) {
			return false
		}
	}
	if len(f.rule.Prefixes) == 0 {
		return true
	}
	for _, p := range f.rule.Prefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// Apply filters a slice of paths, returning only allowed ones.
func (f *Filter) Apply(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if f.Allow(p) {
			out = append(out, p)
		}
	}
	return out
}
