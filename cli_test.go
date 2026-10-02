package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestParseCLIArgsSupportsRunConfiguration(t *testing.T) {
	options, positional, err := parseCLIArgs([]string{
		"run", "--nwords", "7", "--pool", "3-9", "--pattern", "^[a-z]+$",
		"--seed", "1234", "--slow-ms-per-rune", "300",
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

func TestCLIGeneratesZshCompletionAndManPage(t *testing.T) {
	var completion bytes.Buffer
	completionCLI := newCLI(nil, &completion, io.Discard)
	if err := completionCLI.Run(context.Background(), []string{"ttypist", "completion", "zsh"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(completion.String(), "#compdef ttypist") {
		t.Fatalf("completion = %q, want zsh compdef", completion.String())
	}

	var help bytes.Buffer
	helpCLI := newCLI(nil, &help, io.Discard)
	if err := helpCLI.Run(context.Background(), []string{"ttypist", "--help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(help.String(), "TTYP_NWORDS") {
		t.Fatalf("help = %q, want environment variable hint", help.String())
	}

	var man bytes.Buffer
	manCLI := newCLI(nil, &man, io.Discard)
	if err := manCLI.Run(context.Background(), []string{"ttypist", "man"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(man.String(), ".TH ttypist 8") || !strings.Contains(man.String(), "--nwords") {
		t.Fatalf("man page does not contain the command name and flags: %q", man.String())
	}
}

func TestParseCLIArgsUsesEnvironmentConfigAndCLIPrecedence(t *testing.T) {
	clearCLIEnv(t)
	configPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configPath, []byte("nwords = 12\npool = \"2-4\"\nseed = 21\ntarget-wpm = 45\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TTYP_CONFIG", configPath)
	t.Setenv("TTYP_NWORDS", "9")
	t.Setenv("TTYP_SEED", "22")

	options, positional, err := parseCLIArgs([]string{"--nwords", "7", "one"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if options.nwords != 7 || options.pool != "2-4" || options.seed != 22 || options.targetWPM != 45 {
		t.Fatalf("options = %+v, want CLI over env over config", options)
	}
	if strings.Join(positional, " ") != "one" {
		t.Fatalf("positional = %v, want [one]", positional)
	}
}

func TestParseCLIArgsSupportsJSONYAMLAndTOMLConfig(t *testing.T) {
	clearCLIEnv(t)
	tests := []struct {
		extension string
		contents  string
	}{
		{extension: ".json", contents: `{"nwords":12,"pool":"2-4","seed":21}`},
		{extension: ".toml", contents: "nwords = 12\npool = \"2-4\"\nseed = 21\n"},
		{extension: ".yaml", contents: "nwords: 12\npool: 2-4\nseed: 21\n"},
	}
	for _, test := range tests {
		t.Run(test.extension, func(t *testing.T) {
			configPath := filepath.Join(t.TempDir(), "config"+test.extension)
			if err := os.WriteFile(configPath, []byte(test.contents), 0600); err != nil {
				t.Fatal(err)
			}
			options, _, err := parseCLIArgs([]string{"--config", configPath}, io.Discard)
			if err != nil {
				t.Fatal(err)
			}
			if options.nwords != 12 || options.pool != "2-4" || options.seed != 21 {
				t.Fatalf("options = %+v, want values from %s", options, test.extension)
			}
		})
	}
}

func TestParseCLIArgsRejectsInvalidConfigFile(t *testing.T) {
	clearCLIEnv(t)
	_, _, err := parseCLIArgs([]string{"--config", filepath.Join(t.TempDir(), "missing.toml")}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "read config file") {
		t.Fatalf("error = %v, want missing config error", err)
	}

	configPath := filepath.Join(t.TempDir(), "config.ini")
	if err := os.WriteFile(configPath, []byte("nwords=12\n"), 0600); err != nil {
		t.Fatal(err)
	}
	_, _, err = parseCLIArgs([]string{"--config", configPath}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "unsupported config format") {
		t.Fatalf("error = %v, want unsupported format error", err)
	}
}

func TestDefaultCLISeedIsRandomized(t *testing.T) {
	if got := defaultCLIOptions().seed; got == 0 {
		t.Fatal("default CLI seed = 0, want a generated seed")
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
		got.slowPerRuneMS == want.slowPerRuneMS &&
		got.targetWPM == want.targetWPM && got.penaltySeconds == want.penaltySeconds &&
		got.minWPM == want.minWPM && got.minAccuracy == want.minAccuracy
}

func clearCLIEnv(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"TTYP_CONFIG", "TTYP_NWORDS", "TTYP_POOL", "TTYP_PATTERN", "TTYP_DICT", "TTYP_INPUT",
		"TTYP_SEED", "TTYP_SLOW_MS_PER_RUNE", "TTYP_TARGET_WPM", "TTYP_PENALTY_SECONDS",
		"TTYP_MIN_WPM", "TTYP_MIN_ACCURACY",
	} {
		old, wasSet := os.LookupEnv(name)
		if err := os.Unsetenv(name); err != nil {
			t.Fatal(err)
		}
		envName, previous, hadValue := name, old, wasSet
		t.Cleanup(func() {
			if hadValue {
				_ = os.Setenv(envName, previous)
			} else {
				_ = os.Unsetenv(envName)
			}
		})
	}
}
