package main

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestPrintTargetParagraphUsesTerminalSafeLineBreaks(t *testing.T) {
	var output bytes.Buffer
	printTargetParagraph(&output, []string{"two", "three", "four"}, Paragraph{Lines: []ParagraphLine{{Start: 0, End: 2}, {Start: 2, End: 3}}})
	if got, want := output.String(), "  \x1b[1mtwo three\x1b[0m\r\n  \x1b[1mfour\x1b[0m\r\n"; got != want {
		t.Fatalf("prompt output = %q, want %q", got, want)
	}
}

func TestPromptWidthUsesTerminalColumns(t *testing.T) {
	if got, want := promptWidth(120), 118; got != want {
		t.Fatalf("prompt width = %d, want %d", got, want)
	}
	if got, want := promptWidth(3), 78; got != want {
		t.Fatalf("narrow-terminal fallback width = %d, want %d", got, want)
	}
}

func TestReadInputMapsEditingKeysAndUTF8(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("é\b\x17 \r\x03"))

	input, ok, err := readInput(reader)
	if err != nil || !ok || input.Kind != InputRune || input.Rune != 'é' {
		t.Fatalf("first input = %+v, %v, %v", input, ok, err)
	}
	input, ok, err = readInput(reader)
	if err != nil || !ok || input.Kind != InputBackspace {
		t.Fatalf("backspace input = %+v, %v, %v", input, ok, err)
	}
	input, ok, err = readInput(reader)
	if err != nil || !ok || input.Kind != InputDeleteWord {
		t.Fatalf("delete-word input = %+v, %v, %v", input, ok, err)
	}
	input, ok, err = readInput(reader)
	if err != nil || !ok || input.Kind != InputSpace {
		t.Fatalf("space input = %+v, %v, %v", input, ok, err)
	}
	_, ok, err = readInput(reader)
	if err != nil || ok {
		t.Fatalf("return input = %v, %v, want ignored", ok, err)
	}
	input, ok, err = readInput(reader)
	if err != nil || !ok || input.Kind != InputCtrlC {
		t.Fatalf("ctrl-c input = %+v, %v, %v", input, ok, err)
	}
}

func TestReadInputDiscardsCSIArrowSequence(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("\x1b[Ax"))
	_, ok, err := readInput(reader)
	if err != nil || ok {
		t.Fatalf("escape input = %v, %v, want ignored", ok, err)
	}
	input, ok, err := readInput(reader)
	if err != nil || !ok || input.Kind != InputRune || input.Rune != 'x' {
		t.Fatalf("post-escape input = %+v, %v, %v", input, ok, err)
	}
}

func TestPrintSessionSummaryKeepsMissesAtColumnZero(t *testing.T) {
	base := time.Unix(0, 0)
	config := DefaultTimingConfig()
	config.SlowPerRune = 10 * time.Second
	session := NewSession([]string{"one", "two"}, config)
	for _, word := range []string{"nrhe", "twheh"} {
		for _, r := range word {
			if err := session.Apply(RuneInput(r), base.Add(time.Second)); err != nil {
				t.Fatal(err)
			}
		}
		if err := session.Apply(Input{Kind: InputSpace}, base.Add(2*time.Second)); err != nil {
			t.Fatal(err)
		}
	}

	var output bytes.Buffer
	printSessionSummary(&output, session)
	if got, want := output.String(), "\r\nMissed:\r\n\x1b[31mnrhe \x1b[0m -> one\r\n\x1b[31mtwheh\x1b[0m -> two\r\n\r\n"; !strings.HasPrefix(got, want) {
		t.Fatalf("summary misses = %q, want prefix %q", got, want)
	}
}

func TestPrintSessionSummaryShowsSlowWordsInMilliseconds(t *testing.T) {
	base := time.Unix(0, 0)
	config := DefaultTimingConfig()
	config.SlowPerRune = time.Millisecond
	session := NewSession([]string{"same", "which"}, config)
	for _, r := range "same" {
		if err := session.Apply(RuneInput(r), base.Add(100*time.Millisecond)); err != nil {
			t.Fatal(err)
		}
	}
	if err := session.Apply(Input{Kind: InputSpace}, base.Add(355*time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	for _, r := range "who" {
		if err := session.Apply(RuneInput(r), base.Add(400*time.Millisecond)); err != nil {
			t.Fatal(err)
		}
	}
	if err := session.Apply(Input{Kind: InputSpace}, base.Add(620*time.Millisecond)); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	printSessionSummary(&output, session)
	if got, want := output.String(), "Slow:\r\nsame(255) who(220)\r\n"; !strings.Contains(got, want) {
		t.Fatalf("summary slow words = %q, want substring %q", got, want)
	}
	if !strings.Contains(output.String(), "Missed:\r\n\x1b[31mwho\x1b[0m -> which\r\n\r\nSlow:\r\n") {
		t.Fatalf("summary sections = %q, want blank line between sections", output.String())
	}
	if strings.Contains(output.String(), "who(220)\r\n\r\n\r\nTest") {
		t.Fatalf("summary has too many blank lines before test: %q", output.String())
	}
}
