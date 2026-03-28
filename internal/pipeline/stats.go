package pipeline

import (
	"fmt"
	"strings"
	"time"
)

// ItemError records a single item failure.
type ItemError struct {
	URL   string
	Title string
	Err   error
}

// Stats tracks aggregate results of a ProcessBatch run.
type Stats struct {
	Success int
	Skipped int
	Failed  int
	Errors  []ItemError
	Start   time.Time
	End     time.Time
}

// Duration returns the wall-clock duration of the batch.
func (s *Stats) Duration() time.Duration { return s.End.Sub(s.Start) }

// Report returns a human-readable summary of the batch results.
func (s *Stats) Report() string {
	var b strings.Builder
	total := s.Success + s.Skipped + s.Failed

	b.WriteString(fmt.Sprintf("Pipeline completed in %s\n", s.Duration().Round(time.Second)))
	b.WriteString(fmt.Sprintf("  Total:   %d\n", total))
	b.WriteString(fmt.Sprintf("  Success: %d\n", s.Success))
	b.WriteString(fmt.Sprintf("  Skipped: %d\n", s.Skipped))
	b.WriteString(fmt.Sprintf("  Failed:  %d\n", s.Failed))

	if len(s.Errors) > 0 {
		b.WriteString("\nErrors:\n")
		for _, e := range s.Errors {
			title := e.Title
			if title == "" {
				title = e.URL
			}
			b.WriteString(fmt.Sprintf("  - %s: %v\n", title, e.Err))
		}
	}

	return b.String()
}
