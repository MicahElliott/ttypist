package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

const (
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
	ansiReset  = "\x1b[0m"
	ansiClear  = "\x1b[2K"
)

var errInputClosed = errors.New("input closed")

// runInteractive renders one prompt at a time and leaves completed prompt
// pairs in terminal scrollback.
func runInteractive(targets []string, in *os.File, out io.Writer) error {
	if len(targets) == 0 {
		return errors.New("input must contain at least one word")
	}

	width := terminalPromptWidth(out)
	prompt, err := BuildPrompt(targets, 0, width, 2)
	if err != nil {
		return err
	}

	session := NewSession(targets, DefaultTimingConfig())
	oldState, err := term.MakeRaw(int(in.Fd()))
	if err != nil {
		return fmt.Errorf("could not switch terminal to raw mode: %w", err)
	}
	defer term.Restore(int(in.Fd()), oldState)

	fmt.Fprint(out, "Start typing to begin test, Ctrl-C to exit.\r\n")
	printPrompt(out, prompt)

	reader := bufio.NewReader(in)
	for session.Status() == SessionActive {
		input, ok, err := readInput(reader)
		if err != nil {
			if errors.Is(err, errInputClosed) {
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
			return nil
		}

		printInputLine(out, session, prompt)
		if session.Status() == SessionCompleted {
			fmt.Fprint(out, "\r\n")
			printSessionSummary(out, session)
			return nil
		}

		if session.FocusIndex() >= prompt.BodyEnd {
			fmt.Fprint(out, "\r\n")
			prompt, err = BuildPrompt(targets, prompt.BodyEnd, width, 2)
			if err != nil {
				return err
			}
			printPrompt(out, prompt)
		}
	}
	return nil
}

func terminalPromptWidth(out io.Writer) int {
	if file, ok := out.(*os.File); ok {
		if _, columns, err := term.GetSize(int(file.Fd())); err == nil && columns > 4 {
			return columns - 2 // reserve the two-character prompt prefix
		}
	}
	return 78
}

func printPrompt(out io.Writer, prompt Prompt) {
	fmt.Fprintf(out, "  %s\r\n> ", strings.Join(prompt.Words, " "))
}

func printInputLine(out io.Writer, session *Session, prompt Prompt) {
	fmt.Fprintf(out, "\r%s> ", ansiClear)
	for _, attempt := range session.Attempts() {
		if attempt.TargetIdx < prompt.Start || attempt.TargetIdx >= prompt.BodyEnd {
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
	for _, attempt := range session.Attempts() {
		if !attempt.Correct {
			fmt.Fprintf(out, "%s%s%s -> %s\r\n", ansiRed, attempt.Entered, ansiReset, attempt.Target)
		}
	}
	metrics := session.Metrics(time.Now())
	status := "Test"
	if session.Status() == SessionAborted {
		status = "Interrupted test"
	}
	fmt.Fprintf(out, "\r\n%s of %d words took %d seconds.\r\n", status, metrics.Attempted, int(metrics.Elapsed.Round(time.Second).Seconds()))
	fmt.Fprintf(out, "WPM: %.1f (raw: %.1f)\r\n", metrics.PenalizedWPM, metrics.RawWPM)
	fmt.Fprintf(out, "Acc: %.0f%% (%d/%d)\r\n", metrics.Accuracy, metrics.Correct, metrics.Attempted)
}
