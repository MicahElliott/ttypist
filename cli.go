package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/urfave/cli-altsrc/v3"
	docs "github.com/urfave/cli-docs/v3"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
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
	configPath     string
}

func defaultCLIOptions() cliOptions {
	return cliOptions{
		nwords:         50,
		pool:           "1-200",
		seed:           time.Now().UnixNano(),
		slowPerRuneMS:  150,
		penaltySeconds: 1,
	}
}

func parseCLIArgs(args []string, stderr io.Writer) (cliOptions, []string, error) {
	if len(args) > 0 && args[0] == "run" {
		args = args[1:]
	}

	configPath := ""
	command := &cli.Command{
		Name:      "ttypist",
		Flags:     cliFlags(&configPath),
		Action:    func(context.Context, *cli.Command) error { return nil },
		Writer:    io.Discard,
		ErrWriter: stderr,
	}
	if err := command.Run(context.Background(), append([]string{"ttypist"}, args...)); err != nil {
		return cliOptions{}, nil, err
	}
	options, err := cliOptionsFromCommand(command)
	if err != nil {
		return cliOptions{}, nil, err
	}
	return options, command.Args().Slice(), nil
}

func cliFlags(configPath *string) []cli.Flag {
	defaults := defaultCLIOptions()
	return []cli.Flag{
		&cli.StringFlag{Name: "config", Aliases: []string{"c"}, Usage: "optional YAML, JSON, or TOML configuration file", TakesFile: true, Destination: configPath, Sources: cli.EnvVars("TTYP_CONFIG")},
		&cli.IntFlag{Name: "nwords", Aliases: []string{"n"}, Value: defaults.nwords, Usage: "number of words to select", Sources: cliSources(configPath, "nwords", "TTYP_NWORDS")},
		&cli.StringFlag{Name: "pool", Aliases: []string{"p"}, Value: defaults.pool, Usage: "inclusive dictionary rank range, such as 1-200", Sources: cliSources(configPath, "pool", "TTYP_POOL")},
		&cli.StringFlag{Name: "pattern", Aliases: []string{"e"}, Value: ".", Usage: "regular expression applied to dictionary words", Sources: cliSources(configPath, "pattern", "TTYP_PATTERN")},
		&cli.StringFlag{Name: "dict", Aliases: []string{"d"}, Usage: "custom dictionary file", TakesFile: true, Sources: cliSources(configPath, "dict", "TTYP_DICT")},
		&cli.StringFlag{Name: "input", Aliases: []string{"i"}, Usage: "file containing the target text", TakesFile: true, Sources: cliSources(configPath, "input", "TTYP_INPUT")},
		&cli.Int64Flag{Name: "seed", Aliases: []string{"s"}, Value: defaults.seed, Usage: "random selection seed (default: random)", Sources: cliSources(configPath, "seed", "TTYP_SEED")},
		&cli.IntFlag{Name: "slow-ms-per-rune", Value: defaults.slowPerRuneMS, Usage: "slow-word threshold in milliseconds per target rune", Sources: cliSources(configPath, "slow-ms-per-rune", "TTYP_SLOW_MS_PER_RUNE")},
		&cli.Float64Flag{Name: "target-wpm", Value: defaults.targetWPM, Usage: "target WPM used to derive the slow-word threshold", Sources: cliSources(configPath, "target-wpm", "TTYP_TARGET_WPM")},
		&cli.Float64Flag{Name: "penalty-seconds", Value: defaults.penaltySeconds, Usage: "penalty added to elapsed time per incorrect word", Sources: cliSources(configPath, "penalty-seconds", "TTYP_PENALTY_SECONDS")},
		&cli.Float64Flag{Name: "min-wpm", Value: defaults.minWPM, Usage: "minimum penalized WPM for a successful completion", Sources: cliSources(configPath, "min-wpm", "TTYP_MIN_WPM")},
		&cli.Float64Flag{Name: "min-accuracy", Value: defaults.minAccuracy, Usage: "minimum accuracy percentage for a successful completion", Sources: cliSources(configPath, "min-accuracy", "TTYP_MIN_ACCURACY")},
	}
}

func cliSources(configPath *string, key, envName string) cli.ValueSourceChain {
	return cli.NewValueSourceChain(cli.EnvVar(envName), configFileSource(configPath, key))
}

func configFileSource(configPath *string, key string) cli.ValueSource {
	return altsrc.NewValueSource(func(data []byte, destination any) error {
		return unmarshalConfig(*configPath, data, destination)
	}, "config", key, altsrc.NewStringPtrSourcer(configPath))
}

func unmarshalConfig(path string, data []byte, destination any) error {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		var values map[string]any
		if err := json.Unmarshal(data, &values); err != nil {
			return err
		}
		mapped, ok := destination.(*map[any]any)
		if !ok {
			return errors.New("config JSON destination has an unexpected type")
		}
		*mapped = make(map[any]any, len(values))
		for key, value := range values {
			(*mapped)[key] = value
		}
		return nil
	case ".toml":
		return toml.Unmarshal(data, destination)
	case ".yaml", ".yml":
		return yaml.Unmarshal(data, destination)
	default:
		return fmt.Errorf("unsupported config format %q: use .json, .toml, .yaml, or .yml", filepath.Ext(path))
	}
}

func validateConfigFile(path string) error {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file %q: %w", path, err)
	}
	var values map[any]any
	if err := unmarshalConfig(path, data, &values); err != nil {
		return fmt.Errorf("parse config file %q: %w", path, err)
	}
	return nil
}

func cliOptionsFromCommand(command *cli.Command) (cliOptions, error) {
	options := cliOptions{
		nwords:         command.Int("nwords"),
		pool:           command.String("pool"),
		dictionaryPath: command.String("dict"),
		inputPath:      command.String("input"),
		seed:           command.Int64("seed"),
		slowPerRuneMS:  command.Int("slow-ms-per-rune"),
		targetWPM:      command.Float64("target-wpm"),
		penaltySeconds: command.Float64("penalty-seconds"),
		minWPM:         command.Float64("min-wpm"),
		minAccuracy:    command.Float64("min-accuracy"),
		configPath:     command.String("config"),
	}
	if err := validateConfigFile(options.configPath); err != nil {
		return cliOptions{}, err
	}
	pattern := command.String("pattern")
	if options.nwords < 1 {
		return cliOptions{}, errors.New("--nwords must be positive")
	}
	if options.slowPerRuneMS < 1 {
		return cliOptions{}, errors.New("--slow-ms-per-rune must be positive")
	}
	if options.targetWPM < 0 || options.penaltySeconds < 0 || options.minWPM < 0 || options.minAccuracy < 0 || options.minAccuracy > 100 {
		return cliOptions{}, errors.New("timing and completion thresholds cannot be negative, and --min-accuracy must be at most 100")
	}
	compiledPattern, err := regexp.Compile(pattern)
	if err != nil {
		return cliOptions{}, fmt.Errorf("invalid --pattern: %w", err)
	}
	options.pattern = compiledPattern
	options.patternText = pattern
	if _, _, err := parsePool(options.pool); err != nil {
		return cliOptions{}, err
	}
	return options, nil
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

func newCLI(in *os.File, out, stderr io.Writer) *cli.Command {
	runAction := runCommandAction(in, out)
	configPath := ""
	return &cli.Command{
		Name:                   "ttypist",
		Usage:                  "terminal typing tutor",
		Description:            "Configuration precedence is CLI flags, environment variables, config file, then built-in defaults. Use TTYP_* variables or --config for non-CLI configuration.",
		ArgsUsage:              "[words...]",
		Flags:                  cliFlags(&configPath),
		Commands:               []*cli.Command{{Name: "run", Aliases: []string{"r"}, Usage: "run a typing test", ArgsUsage: "[words...]", Action: runAction}, manCommand()},
		Action:                 runAction,
		EnableShellCompletion:  true,
		Suggest:                true,
		UseShortOptionHandling: true,
		Writer:                 out,
		ErrWriter:              stderr,
	}
}

func runCommandAction(in *os.File, out io.Writer) cli.ActionFunc {
	return func(_ context.Context, command *cli.Command) error {
		options, err := cliOptionsFromCommand(command)
		if err != nil {
			return cli.Exit(err, 2)
		}
		targets, err := resolveTargets(options, command.Args().Slice())
		if err != nil {
			return cli.Exit(err, 1)
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
			return cli.Exit(err, 1)
		}
		return nil
	}
}

func manCommand() *cli.Command {
	return &cli.Command{
		Name:   "man",
		Usage:  "write the generated man page",
		Hidden: true,
		Action: func(_ context.Context, command *cli.Command) error {
			man, err := docs.ToMan(command.Root())
			if err != nil {
				return cli.Exit(err, 1)
			}
			_, err = io.WriteString(command.Root().Writer, man)
			if err != nil {
				return cli.Exit(err, 1)
			}
			return nil
		},
	}
}

func runCLI(args []string, in *os.File, out, stderr io.Writer) int {
	app := newCLI(in, out, stderr)
	exitCode := 0
	app.ExitErrHandler = func(_ context.Context, _ *cli.Command, err error) {
		exitCode = 1
		var exitErr cli.ExitCoder
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
		fmt.Fprintln(stderr, err)
	}
	if err := app.Run(context.Background(), append([]string{"ttypist"}, args...)); err != nil {
		if exitCode != 0 {
			return exitCode
		}
		return 2
	}
	return exitCode
}
