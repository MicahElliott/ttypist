package main

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

const (
	defaultStatsLimit   = 10
	defaultHardestLimit = 10
	statsBarWidth       = 20
)

type HardestWord struct {
	Target string
	Misses int
}

type StatsReport struct {
	Sessions []SessionRecord
	Hardest  []HardestWord
}

func BuildStatsReport(records []SessionRecord, sessionLimit, hardestLimit int) StatsReport {
	if sessionLimit < 1 {
		sessionLimit = defaultStatsLimit
	}
	if hardestLimit < 1 {
		hardestLimit = defaultHardestLimit
	}

	recent := make([]SessionRecord, 0, min(sessionLimit, len(records)))
	for index := len(records) - 1; index >= 0 && len(recent) < sessionLimit; index-- {
		recent = append(recent, records[index])
	}

	misses := make(map[string]int)
	for _, record := range records {
		for _, attempt := range record.Attempts {
			if !attempt.Correct {
				misses[attempt.Target]++
			}
		}
	}
	hardest := make([]HardestWord, 0, len(misses))
	for target, count := range misses {
		hardest = append(hardest, HardestWord{Target: target, Misses: count})
	}
	sort.Slice(hardest, func(i, j int) bool {
		if hardest[i].Misses != hardest[j].Misses {
			return hardest[i].Misses > hardest[j].Misses
		}
		return hardest[i].Target < hardest[j].Target
	})
	if len(hardest) > hardestLimit {
		hardest = hardest[:hardestLimit]
	}

	return StatsReport{Sessions: recent, Hardest: hardest}
}

func printStats(out io.Writer, report StatsReport) {
	if len(report.Sessions) == 0 {
		fmt.Fprint(out, "No sessions recorded.\r\n")
		return
	}

	fmt.Fprintf(out, "Recent sessions (%d):\r\n", len(report.Sessions))
	fmt.Fprint(out, "When                 Status      Words  WPM   Acc\r\n")
	for _, session := range report.Sessions {
		metrics := session.Metrics
		fmt.Fprintf(out, "%-20s %-11s %5d %5.1f %3.0f%%\r\n",
			statsWhen(session), statsStatus(session), metrics.Attempted,
			metrics.PenalizedWPM, metrics.Accuracy)
	}

	fmt.Fprint(out, "\r\nAccuracy:\r\n")
	for _, session := range report.Sessions {
		filled := int(session.Metrics.Accuracy * statsBarWidth / 100)
		if filled < 0 {
			filled = 0
		}
		if filled > statsBarWidth {
			filled = statsBarWidth
		}
		bar := strings.Repeat("#", filled) + strings.Repeat("-", statsBarWidth-filled)
		fmt.Fprintf(out, "%s [%s] %.0f%%\r\n", statsWhen(session), bar, session.Metrics.Accuracy)
	}

	if len(report.Hardest) == 0 {
		fmt.Fprint(out, "\r\nHardest words: none\r\n")
		return
	}
	fmt.Fprint(out, "\r\nHardest words:\r\n")
	for _, word := range report.Hardest {
		fmt.Fprintf(out, "  %-20s %d misses\r\n", word.Target, word.Misses)
	}
}

func statsWhen(record SessionRecord) string {
	if record.StartedAt == "" {
		return "-"
	}
	started, err := time.Parse(time.RFC3339Nano, record.StartedAt)
	if err != nil {
		return record.StartedAt
	}
	return started.Local().Format("2006-01-02 15:04")
}

func statsStatus(record SessionRecord) string {
	if record.Status == "" {
		return "unknown"
	}
	return record.Status
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
