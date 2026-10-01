package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/term"
)

const (
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
	ansiBold   = "\x1b[1m"
	ansiReset  = "\x1b[0m"
	ansiClear  = "\x1b[2K"
)

var errInputClosed = errors.New("input closed")
var ErrThresholdNotMet = errors.New("completion threshold not met")

type InteractiveConfig struct {
	Timing      TimingConfig
	MinWPM      float64
	MinAccuracy float64
	Selection   SessionSelection
	Store       SessionStore
}

func DefaultInteractiveConfig() InteractiveConfig {
	return InteractiveConfig{Timing: DefaultTimingConfig()}
}

// runInteractive renders a static target paragraph followed by append-only
// input lines that use the same word breaks.
func runInteractive(targets []string, in *os.File, out io.Writer) error {
	return runInteractiveWithConfig(targets, DefaultInteractiveConfig(), in, out)
}

func runInteractiveWithConfig(targets []string, config InteractiveConfig, in *os.File, out io.Writer) error {
	if len(targets) == 0 {
		return errors.New("input must contain at least one word")
	}
	width := terminalPromptWidth(out)
	paragraph, err := BuildParagraph(targets, width)
	if err != nil {
		return err
	}
	store, err := sessionStore(config.Store)
	if err != nil {
		return err
	}

	session := NewSession(targets, config.Timing)
	oldState, err := term.MakeRaw(int(in.Fd()))
	if err != nil {
		return fmt.Errorf("could not switch terminal to raw mode: %w", err)
	}
	defer term.Restore(int(in.Fd()), oldState)

	fmt.Fprint(out, "Start typing to begin test, Ctrl-C to exit.\r\n\r\n")
	printTargetParagraph(out, targets, paragraph)
	fmt.Fprint(out, "\r\n> ")
	lineIndex := 0

	reader := bufio.NewReader(in)
	for session.Status() == SessionActive {
		input, ok, err := readInput(reader)
		if err != nil {
			if errors.Is(err, errInputClosed) {
				session.abort(time.Now())
				if err := persistSession(store, session, config); err != nil {
					return err
				}
				return nil
			}
			return err
		}
		if !ok {
			continue
		}

		if err := session.Apply(input, time.Now()); err != nil {
			return err
		}
		if input.Kind == InputCtrlC {
			fmt.Fprint(out, "\r\n")
			printSessionSummary(out, session)
			if err := persistSession(store, session, config); err != nil {
				return err
			}
			return nil
		}

		line := paragraph.Lines[lineIndex]
		printInputLine(out, session, line)
		if session.Status() == SessionCompleted {
			fmt.Fprint(out, "\r\n")
			printSessionSummary(out, session)
			if err := persistSession(store, session, config); err != nil {
				return err
			}
			if err := checkCompletionThresholds(session.Metrics(time.Now()), config); err != nil {
				return err
			}
			return nil
		}

		if session.FocusIndex() >= line.End {
			fmt.Fprint(out, "\r\n")
			lineIndex++
			fmt.Fprint(out, "> ")
		}
	}
	return nil
}

func sessionStore(store SessionStore) (SessionStore, error) {
	if store != nil {
		return store, nil
	}
	path, err := DefaultSessionStorePath()
	if err != nil {
		return nil, err
	}
	return FileSessionStore{Path: path}, nil
}

func persistSession(store SessionStore, session *Session, config InteractiveConfig) error {
	if session.Status() != SessionCompleted && len(session.Attempts()) == 0 {
		return nil
	}
	return store.Save(session.Record(config, time.Now()))
}

func checkCompletionThresholds(metrics Metrics, config InteractiveConfig) error {
	failures := make([]string, 0, 2)
	if config.MinWPM > 0 && metrics.PenalizedWPM < config.MinWPM {
		failures = append(failures, fmt.Sprintf("WPM %.1f is below %.1f", metrics.PenalizedWPM, config.MinWPM))
	}
	if config.MinAccuracy > 0 && metrics.Accuracy < config.MinAccuracy {
		failures = append(failures, fmt.Sprintf("accuracy %.1f%% is below %.1f%%", metrics.Accuracy, config.MinAccuracy))
	}
	if len(failures) == 0 {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrThresholdNotMet, strings.Join(failures, "; "))
}

func terminalPromptWidth(out io.Writer) int {
	if file, ok := out.(*os.File); ok {
		if columns, _, err := term.GetSize(int(file.Fd())); err == nil {
			return promptWidth(columns)
		}
	}
	return 78
}

func promptWidth(columns int) int {
	if columns > 4 {
		return columns - 2 // reserve the two-character prompt prefix
	}
	return 78
}

func printTargetParagraph(out io.Writer, targets []string, paragraph Paragraph) {
	for _, line := range paragraph.Lines {
		fmt.Fprintf(out, "  %s%s%s\r\n", ansiBold, strings.Join(targets[line.Start:line.End], " "), ansiReset)
	}
}

func printInputLine(out io.Writer, session *Session, line ParagraphLine) {
	fmt.Fprintf(out, "\r%s> ", ansiClear)
	for _, attempt := range session.Attempts() {
		if attempt.TargetIdx < line.Start || attempt.TargetIdx >= line.End {
			continue
		}
		color := ansiGreen
		if !attempt.Correct {
			color = ansiRed
		} else if attempt.Slow {
			color = ansiYellow
		}
		fmt.Fprintf(out, "%s%s%s ", color, attempt.Target, ansiReset)
	}
	fmt.Fprint(out, session.CurrentText())
}

func readInput(reader *bufio.Reader) (Input, bool, error) {
	r, _, err := reader.ReadRune()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return Input{}, false, errInputClosed
		}
		return Input{}, false, err
	}

	switch r {
	case 3:
		return Input{Kind: InputCtrlC}, true, nil
	case 23:
		return Input{Kind: InputDeleteWord}, true, nil
	case '\b', 127:
		return Input{Kind: InputBackspace}, true, nil
	case ' ':
		return Input{Kind: InputSpace}, true, nil
	case '\t', '\r', '\n':
		return Input{}, false, nil
	case 27:
		return Input{}, false, discardEscapeSequence(reader)
	default:
		return RuneInput(r), true, nil
	}
}

func discardEscapeSequence(reader *bufio.Reader) error {
	r, _, err := reader.ReadRune()
	if err != nil {
		return err
	}
	if r != '[' {
		return nil
	}
	for {
		r, _, err = reader.ReadRune()
		if err != nil {
			return err
		}
		if r >= '@' && r <= '~' {
			return nil
		}
	}
}

func printSessionSummary(out io.Writer, session *Session) {
	fmt.Fprint(out, "\r\n")
	attempts := session.Attempts()
	misses := make([]Attempt, 0)
	slow := make([]Attempt, 0)
	maxMissedWidth := 0
	for _, attempt := range attempts {
		if !attempt.Correct {
			misses = append(misses, attempt)
			if width := utf8.RuneCountInString(attempt.Entered); width > maxMissedWidth {
				maxMissedWidth = width
			}
		}
		if attempt.Slow {
			slow = append(slow, attempt)
		}
	}
	if len(misses) > 0 {
		fmt.Fprint(out, "Missed:\r\n")
		for _, attempt := range misses {
			entered := padRight(attempt.Entered, maxMissedWidth)
			fmt.Fprintf(out, "%s%s%s -> %s\r\n", ansiRed, entered, ansiReset, attempt.Target)
		}
	}
	if len(slow) > 0 {
		if len(misses) > 0 {
			fmt.Fprint(out, "\r\n")
		}
		fmt.Fprint(out, "Slow:\r\n")
		words := make([]string, 0, len(slow))
		for _, attempt := range slow {
			word := attempt.Target
			if !attempt.Correct {
				word = attempt.Entered
			}
			words = append(words, fmt.Sprintf("%s(%d)", word, attempt.Duration.Milliseconds()))
		}
		fmt.Fprintf(out, "%s\r\n", strings.Join(words, " "))
	}
	fmt.Fprint(out, "\r\n")
	metrics := session.Metrics(time.Now())
	status := "Test"
	if session.Status() == SessionAborted {
		status = "Interrupted test"
	}
	fmt.Fprintf(out, "%s of %d words took %d seconds.\r\n", status, metrics.Attempted, int(metrics.Elapsed.Round(time.Second).Seconds()))
	fmt.Fprintf(out, "WPM: %.1f (raw: %.1f)\r\n", metrics.PenalizedWPM, metrics.RawWPM)
	fmt.Fprintf(out, "Acc: %.0f%% (%d/%d)\r\n", metrics.Accuracy, metrics.Correct, metrics.Attempted)
}

func padRight(value string, width int) string {
	padding := width - utf8.RuneCountInString(value)
	if padding <= 0 {
		return value
	}
	return value + strings.Repeat(" ", padding)
}
