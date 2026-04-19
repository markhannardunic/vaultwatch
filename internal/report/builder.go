package report

import (
	"context"
	"fmt"

	"github.com/your-org/vaultwatch/internal/vault"
)

// Builder constructs a Report by checking a list of secret paths.
type Builder struct {
	checker *vault.Checker
	report  *Report
}

// NewBuilder creates a Builder using the given checker and report.
func NewBuilder(checker *vault.Checker, r *Report) *Builder {
	return &Builder{checker: checker, report: r}
}

// Build iterates over paths, checks each secret, and adds entries to the report.
func (b *Builder) Build(ctx context.Context, paths []string) error {
	for _, path := range paths {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		result, err := b.checker.Check(ctx, path)
		if err != nil {
			return fmt.Errorf("build report: check %q: %w", path, err)
		}

		b.report.Add(Entry{
			Path:      path,
			Level:     result.Level,
			ExpiresAt: result.ExpiresAt,
			Message:   vault.FormatAlert(result),
		})
	}
	return nil
}
