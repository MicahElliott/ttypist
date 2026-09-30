package main

import (
	"errors"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestParseCLIArgsSupportsRunConfiguration(t *testing.T) {
	options, positional, err := parseCLIArgs([]string{
		"run", "--nwords", "7", "--pool", "3-9", "--pattern", "^[a-z]+$",
		"--seed", "1234", "--lookahead", "4", "--slow-ms-per-rune", "300",
		"--target-wpm", "50", "--penalty-seconds", "2", "--min-wpm", "40",
		"--min-accuracy", "92", "one", "two",
	}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if !reflectCLIOptions(options, cliOptions{
		nwords:         7,
		pool:           "3-9",
		seed:           1234,
		lookahead:      4,
		slowPerRuneMS:  300,
		targetWPM:      50,
		penaltySeconds: 2,
		minWPM:         40,
		minAccuracy:    92,
	}) {
		t.Fatalf("options = %+v", options)
	}
	if options.pattern == nil || !options.pattern.MatchString("abc") || options.pattern.MatchString("ABC!") {
		t.Fatalf("pattern = %v, want lowercase word pattern", options.pattern)
	}
	if strings.Join(positional, " ") != "one two" {
		t.Fatalf("positional = %v, want [one two]", positional)
	}
}

func TestResolveTargetsLoadsPlainInputAndDictionaryFiles(t *testing.T) {
	directory := t.TempDir()
	inputPath := directory + "/input.txt"
	if err := os.WriteFile(inputPath, []byte("one two three four\n"), 0600); err != nil {
		t.Fatal(err)
	}
	options := defaultCLIOptions()
	options.inputPath = inputPath
	options.nwords = 3
	targets, err := resolveTargets(options, nil)
	if err != nil || strings.Join(targets, " ") != "one two three" {
		t.Fatalf("input targets = %v, error = %v", targets, err)
	}

	dictionaryPath := directory + "/words.txt"
	if err := os.WriteFile(dictionaryPath, []byte("alpha\nbeta\ngamma\n"), 0600); err != nil {
		t.Fatal(err)
	}
	options = defaultCLIOptions()
	options.dictionaryPath = dictionaryPath
	options.nwords = 2
	options.pool = "1-3"
	options.pattern = regexp.MustCompile("^[ab]")
	targets, err = resolveTargets(options, nil)
	if err != nil || len(targets) != 2 {
		t.Fatalf("dictionary targets = %v, error = %v", targets, err)
	}
	for _, target := range targets {
		if target != "alpha" && target != "beta" {
			t.Fatalf("dictionary target = %q", target)
		}
	}
}

func TestParsePoolAcceptsSingleRankAndRange(t *testing.T) {
	for value, want := range map[string][2]int{
		"7":   {7, 7},
		"3-9": {3, 9},
	} {
		start, end, err := parsePool(value)
		if err != nil || start != want[0] || end != want[1] {
			t.Fatalf("parsePool(%q) = %d-%d, %v; want %d-%d", value, start, end, err, want[0], want[1])
		}
	}
	if _, _, err := parsePool("  "); err == nil {
		t.Fatal(`parsePool("  ") succeeded, want error`)
	}
}

func TestCompletionThresholds(t *testing.T) {
	config := InteractiveConfig{MinWPM: 50, MinAccuracy: 92}
	if err := checkCompletionThresholds(Metrics{PenalizedWPM: 50, Accuracy: 92}, config); err != nil {
		t.Fatalf("thresholds rejected passing metrics: %v", err)
	}
	err := checkCompletionThresholds(Metrics{PenalizedWPM: 49.9, Accuracy: 91}, config)
	if err == nil || !errors.Is(err, ErrThresholdNotMet) {
		t.Fatalf("threshold error = %v, want ErrThresholdNotMet", err)
	}
}

func reflectCLIOptions(got, want cliOptions) bool {
	return got.nwords == want.nwords && got.pool == want.pool && got.seed == want.seed &&
		got.lookahead == want.lookahead && got.slowPerRuneMS == want.slowPerRuneMS &&
		got.targetWPM == want.targetWPM && got.penaltySeconds == want.penaltySeconds &&
		got.minWPM == want.minWPM && got.minAccuracy == want.minAccuracy
}
