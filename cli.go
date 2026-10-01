package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type cliOptions struct {
	nwords         int
	pool           string
	pattern        *regexp.Regexp
	dictionaryPath string
	inputPath      string
	seed           int64
	patternText    string
	slowPerRuneMS  int
	targetWPM      float64
	penaltySeconds float64
	minWPM         float64
	minAccuracy    float64
}

func defaultCLIOptions() cliOptions {
	return cliOptions{
		nwords:         50,
		pool:           "1-200",
		seed:           0,
		slowPerRuneMS:  250,
		penaltySeconds: 1,
	}
}

func parseCLIArgs(args []string, stderr io.Writer) (cliOptions, []string, error) {
	options := defaultCLIOptions()
	if len(args) > 0 && args[0] == "run" {
		args = args[1:]
	}

	flags := flag.NewFlagSet("ttypist", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprintln(stderr, "Usage: ttypist [run] [flags] [words]")
		fmt.Fprintln(stderr, "       ttypist [run] [flags]  (select words from the dictionary)")
		flags.PrintDefaults()
	}
	flags.IntVar(&options.nwords, "nwords", options.nwords, "number of words to select")
	flags.StringVar(&options.pool, "pool", options.pool, "inclusive dictionary rank range, such as 1-200")
	pattern := "."
	flags.StringVar(&pattern, "pattern", pattern, "regular expression applied to dictionary words")
	flags.StringVar(&options.dictionaryPath, "dict", "", "custom dictionary file")
	flags.StringVar(&options.inputPath, "input", "", "file containing the target text")
	flags.Int64Var(&options.seed, "seed", options.seed, "random selection seed")
	flags.IntVar(&options.slowPerRuneMS, "slow-ms-per-rune", options.slowPerRuneMS, "slow-word threshold in milliseconds per target rune")
	flags.Float64Var(&options.targetWPM, "target-wpm", 0, "target WPM used to derive the slow-word threshold")
	flags.Float64Var(&options.penaltySeconds, "penalty-seconds", options.penaltySeconds, "penalty added to elapsed time per incorrect word")
	flags.Float64Var(&options.minWPM, "min-wpm", 0, "minimum penalized WPM for a successful completion")
	flags.Float64Var(&options.minAccuracy, "min-accuracy", 0, "minimum accuracy percentage for a successful completion")

	if err := flags.Parse(args); err != nil {
		return cliOptions{}, nil, err
	}
	if options.nwords < 1 {
		return cliOptions{}, nil, errors.New("--nwords must be positive")
	}
	if options.slowPerRuneMS < 1 {
		return cliOptions{}, nil, errors.New("--slow-ms-per-rune must be positive")
	}
	if options.targetWPM < 0 || options.penaltySeconds < 0 || options.minWPM < 0 || options.minAccuracy < 0 || options.minAccuracy > 100 {
		return cliOptions{}, nil, errors.New("timing and completion thresholds cannot be negative, and --min-accuracy must be at most 100")
	}
	compiledPattern, err := regexp.Compile(pattern)
	if err != nil {
		return cliOptions{}, nil, fmt.Errorf("invalid --pattern: %w", err)
	}
	options.pattern = compiledPattern
	options.patternText = pattern
	if _, _, err := parsePool(options.pool); err != nil {
		return cliOptions{}, nil, err
	}
	return options, flags.Args(), nil
}

func parsePool(value string) (int, int, error) {
	parts := strings.Split(value, "-")
	if len(parts) < 1 || len(parts) > 2 || parts[0] == "" || (len(parts) == 2 && parts[1] == "") {
		return 0, 0, fmt.Errorf("invalid --pool %q: use N or N-M", value)
	}
	start, err := strconv.Atoi(parts[0])
	if err != nil || start < 1 {
		return 0, 0, fmt.Errorf("invalid --pool %q: use positive ranks", value)
	}
	end := start
	if len(parts) == 2 {
		end, err = strconv.Atoi(parts[1])
		if err != nil || end < 1 {
			return 0, 0, fmt.Errorf("invalid --pool %q: use positive ranks", value)
		}
	}
	if start > end {
		return 0, 0, fmt.Errorf("invalid --pool %q: start exceeds end", value)
	}
	return start, end, nil
}

func resolveTargets(options cliOptions, positional []string) ([]string, error) {
	if options.inputPath != "" && len(positional) > 0 {
		return nil, errors.New("cannot combine --input with positional words")
	}
	if options.inputPath != "" && options.dictionaryPath != "" {
		return nil, errors.New("cannot combine --input with --dict")
	}
	if options.dictionaryPath != "" && len(positional) > 0 {
		return nil, errors.New("cannot combine --dict with positional words")
	}
	if len(positional) > 0 {
		targets := strings.Fields(strings.Join(positional, " "))
		if len(targets) == 0 {
			return nil, errors.New("input must contain at least one word")
		}
		return targets, nil
	}
	if options.inputPath != "" {
		data, err := os.ReadFile(options.inputPath)
		if err != nil {
			return nil, fmt.Errorf("read input file: %w", err)
		}
		targets := strings.Fields(string(data))
		if len(targets) > options.nwords {
			targets = targets[:options.nwords]
		}
		if len(targets) == 0 {
			return nil, errors.New("input must contain at least one word")
		}
		return targets, nil
	}

	dictionary := DefaultDictionary()
	if options.dictionaryPath != "" {
		data, err := os.ReadFile(options.dictionaryPath)
		if err != nil {
			return nil, fmt.Errorf("read dictionary file: %w", err)
		}
		dictionary, err = ParseDictionaryData(data)
		if err != nil {
			return nil, fmt.Errorf("parse dictionary file: %w", err)
		}
	}
	poolStart, poolEnd, err := parsePool(options.pool)
	if err != nil {
		return nil, err
	}
	return dictionary.Select(SelectionConfig{
		Count:     options.nwords,
		PoolStart: poolStart,
		PoolEnd:   poolEnd,
		Pattern:   options.pattern,
		Seed:      options.seed,
	})
}

func runCLI(args []string, in *os.File, out, stderr io.Writer) int {
	options, positional, err := parseCLIArgs(args, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			fmt.Fprintln(stderr, err)
		}
		return 2
	}
	targets, err := resolveTargets(options, positional)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	config := InteractiveConfig{
		Selection: SessionSelection{
			Count:      options.nwords,
			Pool:       options.pool,
			Pattern:    options.patternText,
			Dictionary: options.dictionaryPath,
			Input:      options.inputPath,
			Seed:       options.seed,
		},
		Timing: TimingConfig{
			SlowPerRune: time.Duration(options.slowPerRuneMS) * time.Millisecond,
			TargetWPM:   options.targetWPM,
			Penalty:     time.Duration(options.penaltySeconds * float64(time.Second)),
		},
		MinWPM:      options.minWPM,
		MinAccuracy: options.minAccuracy,
	}
	if err := runInteractiveWithConfig(targets, config, in, out); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
