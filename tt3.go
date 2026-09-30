package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

const (
	colorRed    = "\x1b[31m"
	colorGreen  = "\x1b[32m"
	colorYellow = "\x1b[33m"
	colorReset  = "\x1b[0m"

	millisPerCharacter = 250
)

type attempt struct {
	target  string
	entered string
	correct bool
	elapsed time.Duration
}

func eraseInput(n int) {
	if n == 0 {
		return
	}
	fmt.Print(strings.Repeat("\b", n))
	fmt.Print(strings.Repeat(" ", n))
	fmt.Print(strings.Repeat("\b", n))
}

// conductTest displays one target line and reads one character at a time.
// The bool reports whether Ctrl-C interrupted the test.
func conductTest(targets []string) ([]attempt, bool) {
	fmt.Println("\n  " + strings.Join(targets, " "))
	fmt.Print("> ")

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not switch terminal to raw mode:", err)
		return nil, false
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	attempts := make([]attempt, 0, len(targets))
	for _, target := range targets {
		var typed []byte
		var started time.Time

	wordInput:
		for {
			var buf [1]byte
			n, err := os.Stdin.Read(buf[:])
			if err != nil {
				if err == io.EOF {
					fmt.Print(colorReset + "\n")
					return attempts, false
				}
				fmt.Fprintln(os.Stderr, "could not read terminal input:", err)
				return attempts, false
			}
			if n == 0 {
				continue
			}

			key := buf[0]
			switch key {
			case 3: // Ctrl-C
				fmt.Print(colorReset + "\n")
				return attempts, true

			case ' ':
				if len(typed) == 0 {
					continue
				}
				entered := string(typed)
				if entered == "exit" {
					eraseInput(len(typed))
					fmt.Print(colorReset + "\n")
					return attempts, false
				}

				elapsed := time.Since(started)
				isCorrect := entered == target
				isSlow := elapsed > time.Duration((len(typed)+1)*millisPerCharacter)*time.Millisecond
				attempts = append(attempts, attempt{
					target:  target,
					entered: entered,
					correct: isCorrect,
					elapsed: elapsed,
				})

				color := colorGreen
				if !isCorrect {
					color = colorRed
				} else if isSlow {
					color = colorYellow
				}
				eraseInput(len(typed))
				fmt.Print(color + target + colorReset + " ")
				break wordInput

			case 127: // Backspace
				if len(typed) > 0 {
					eraseInput(1)
					typed = typed[:len(typed)-1]
				}

			case 8, 23: // Ctrl-Backspace or Ctrl-W
				eraseInput(len(typed))
				typed = nil
				started = time.Time{}

			case '\t':
				// Keep tab from becoming part of the submitted word.

			default:
				if started.IsZero() {
					started = time.Now()
				}
				typed = append(typed, key)
				fmt.Printf("%c", key)
			}
		}
	}

	fmt.Print(colorReset + "\n")
	return attempts, false
}

func printSummary(attempts []attempt) {
	fmt.Println("\nTypos:")
	for _, a := range attempts {
		if !a.correct {
			fmt.Printf("<%s,%s>", a.target, a.entered)
		}
		fmt.Println("elapsed:", a.elapsed)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Must pass an input string of words")
		os.Exit(1)
	}

	targets := strings.Fields(os.Args[1])
	if len(targets) == 0 {
		fmt.Fprintln(os.Stderr, "Input must contain at least one word")
		os.Exit(1)
	}

	if err := runInteractive(targets, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
