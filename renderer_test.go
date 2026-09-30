package main

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestPrintPromptUsesTerminalSafeLineBreaks(t *testing.T) {
	var output bytes.Buffer
	printPrompt(&output, Prompt{Words: []string{"two", "three"}})
	if got, want := output.String(), "  two three\r\n> "; got != want {
		t.Fatalf("prompt output = %q, want %q", got, want)
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
	session := NewSession([]string{"one", "two"}, DefaultTimingConfig())
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
	if got, want := output.String(), "\x1b[31mnrhe\x1b[0m -> one\r\n\x1b[31mtwheh\x1b[0m -> two\r\n"; !strings.HasPrefix(got, want) {
		t.Fatalf("summary misses = %q, want prefix %q", got, want)
	}
}
