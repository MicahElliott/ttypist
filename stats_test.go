package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildStatsReportShowsRecentSessionsAndHardestWords(t *testing.T) {
	records := []SessionRecord{
		{
			StartedAt: "2026-10-01T10:00:00Z",
			Status:    "completed",
			Attempts: []Attempt{
				{Target: "zebra", Correct: false},
				{Target: "apple", Correct: false},
			},
		},
		{
			StartedAt: "2026-10-02T10:00:00Z",
			Status:    "completed",
			Attempts: []Attempt{
				{Target: "zebra", Correct: false},
				{Target: "apple", Correct: true},
			},
		},
		{
			StartedAt: "2026-10-03T10:00:00Z",
			Status:    "aborted",
			Attempts: []Attempt{
				{Target: "apple", Correct: false},
			},
		},
	}

	report := BuildStatsReport(records, 2, 2)
	if len(report.Sessions) != 2 || report.Sessions[0].StartedAt != records[2].StartedAt || report.Sessions[1].StartedAt != records[1].StartedAt {
		t.Fatalf("recent sessions = %+v, want newest two first", report.Sessions)
	}
	if len(report.Hardest) != 2 || report.Hardest[0] != (HardestWord{Target: "apple", Misses: 2}) || report.Hardest[1] != (HardestWord{Target: "zebra", Misses: 2}) {
		t.Fatalf("hardest words = %+v, want alphabetical tie order", report.Hardest)
	}
}

func TestPrintStatsIncludesAccuracyGraphAndEmptyHardestSection(t *testing.T) {
	var output bytes.Buffer
	printStats(&output, StatsReport{
		Sessions: []SessionRecord{{
			StartedAt: "2026-10-03T10:00:00Z",
			Status:    "completed",
			Metrics:   Metrics{Attempted: 2, PenalizedWPM: 42.5, Accuracy: 50},
		}},
	})
	text := output.String()
	for _, want := range []string{"Recent sessions (1):", "completed", "[##########----------] 50%", "Hardest words: none"} {
		if !strings.Contains(text, want) {
			t.Fatalf("stats output = %q, want %q", text, want)
		}
	}
	if strings.Contains(strings.ReplaceAll(text, "\r\n", ""), "\n") {
		t.Fatalf("stats output = %q, want CRLF line endings", text)
	}
}

func TestStatsCommandReportsMissingStore(t *testing.T) {
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)
	var output bytes.Buffer
	command := newCLI(nil, &output, &output)
	if err := command.Run(t.Context(), []string{"ttypist", "stats"}); err != nil {
		t.Fatal(err)
	}
	if got := output.String(); got != "No sessions recorded.\r\n" {
		t.Fatalf("stats output = %q, want empty-store message", got)
	}
}

func TestStatsCommandReadsStoreAndAppliesLimits(t *testing.T) {
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)
	path := filepath.Join(dataHome, "ttypist", "sessions.jsonl")
	store := FileSessionStore{Path: path}
	for _, record := range []SessionRecord{
		{
			StartedAt: "2026-10-01T10:00:00Z",
			Status:    "completed",
			Metrics:   Metrics{Attempted: 2, PenalizedWPM: 30, Accuracy: 50},
			Attempts: []Attempt{
				{Target: "old", Correct: false},
				{Target: "old", Correct: false},
			},
		},
		{
			StartedAt: "2026-10-03T10:00:00Z",
			Status:    "completed",
			Metrics:   Metrics{Attempted: 1, PenalizedWPM: 60, Accuracy: 100},
			Attempts:  []Attempt{{Target: "new", Correct: true}},
		},
	} {
		if err := store.Save(record); err != nil {
			t.Fatal(err)
		}
	}

	var output bytes.Buffer
	command := newCLI(nil, &output, &output)
	if err := command.Run(t.Context(), []string{"ttypist", "stats", "--limit", "1", "--hardest", "1"}); err != nil {
		t.Fatal(err)
	}
	text := output.String()
	for _, want := range []string{"Recent sessions (1):", "2026-10-03", "old                  2 misses"} {
		if !strings.Contains(text, want) {
			t.Fatalf("stats output = %q, want %q", text, want)
		}
	}
	if strings.Contains(text, "2026-10-01") {
		t.Fatalf("stats output = %q, want session limit to hide old session", text)
	}
}
