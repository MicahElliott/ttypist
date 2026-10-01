package main

import (
	"testing"
	"time"
)

func TestSessionCompletesAndRecordsIndependentWordTimes(t *testing.T) {
	base := time.Unix(0, 0)
	s := NewSession([]string{"one", "two"}, DefaultTimingConfig())

	apply := func(input Input, offset time.Duration) {
		t.Helper()
		if err := s.Apply(input, base.Add(offset)); err != nil {
			t.Fatal(err)
		}
	}
	for _, r := range "one" {
		apply(RuneInput(r), 100*time.Millisecond)
	}
	apply(Input{Kind: InputSpace}, 700*time.Millisecond)
	for _, r := range "two" {
		apply(RuneInput(r), 900*time.Millisecond)
	}
	apply(Input{Kind: InputSpace}, 1500*time.Millisecond)

	if s.Status() != SessionCompleted {
		t.Fatalf("status = %v, want completed", s.Status())
	}
	attempts := s.Attempts()
	if len(attempts) != 2 {
		t.Fatalf("attempt count = %d, want 2", len(attempts))
	}
	if attempts[0].Duration != 600*time.Millisecond {
		t.Fatalf("first duration = %s, want 600ms", attempts[0].Duration)
	}
	if attempts[1].Duration != 600*time.Millisecond {
		t.Fatalf("second duration = %s, want 600ms", attempts[1].Duration)
	}
	metrics := s.Metrics(base.Add(3 * time.Second))
	if metrics.Correct != 2 || metrics.Accuracy != 100 {
		t.Fatalf("metrics = %+v, want two correct attempts at 100%%", metrics)
	}
}

func TestEditingKeepsCorrectionTimeAndCanDeleteWord(t *testing.T) {
	base := time.Unix(0, 0)
	s := NewSession([]string{"same"}, DefaultTimingConfig())
	for i, r := range []rune("samx") {
		if err := s.Apply(RuneInput(r), base.Add(time.Duration(i+1)*100*time.Millisecond)); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Apply(Input{Kind: InputDeleteWord}, base.Add(2*time.Second)); err != nil {
		t.Fatal(err)
	}
	for i, r := range []rune("same") {
		if err := s.Apply(RuneInput(r), base.Add(time.Duration(21+i)*100*time.Millisecond)); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Apply(Input{Kind: InputSpace}, base.Add(3*time.Second)); err != nil {
		t.Fatal(err)
	}
	attempts := s.Attempts()
	if len(attempts) != 1 || !attempts[0].Correct || attempts[0].Entered != "same" {
		t.Fatalf("attempts = %+v, want one correct same attempt", attempts)
	}
	if attempts[0].Duration != 2900*time.Millisecond {
		t.Fatalf("duration = %s, want 2.9s including correction", attempts[0].Duration)
	}
}

func TestIncorrectWordAdvances(t *testing.T) {
	base := time.Unix(0, 0)
	s := NewSession([]string{"which", "day"}, DefaultTimingConfig())
	for i, r := range []rune("who") {
		if err := s.Apply(RuneInput(r), base.Add(time.Duration(i+1)*100*time.Millisecond)); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Apply(Input{Kind: InputSpace}, base.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if s.FocusIndex() != 1 || s.CurrentTarget() != "day" {
		t.Fatalf("focus = %d target = %q, want index 1 target day", s.FocusIndex(), s.CurrentTarget())
	}
	attempts := s.Attempts()
	if len(attempts) != 1 || attempts[0].Correct {
		t.Fatalf("attempts = %+v, want one incorrect attempt", attempts)
	}
}

func TestParagraphUsesWholeLinesAndKeepsFinalWordsTogether(t *testing.T) {
	targets := []string{"when", "this", "made", "while", "from", "however", "some", "between", "now", "world"}
	paragraph, err := BuildParagraph(targets, 35)
	if err != nil {
		t.Fatal(err)
	}
	if len(paragraph.Lines) != 2 {
		t.Fatalf("paragraph lines = %+v, want two lines", paragraph.Lines)
	}
	if got, want := paragraph.Lines[0], (ParagraphLine{Start: 0, End: 6}); got != want {
		t.Fatalf("first line = %+v, want %+v", got, want)
	}
	if got, want := paragraph.Lines[1], (ParagraphLine{Start: 6, End: 10}); got != want {
		t.Fatalf("last line = %+v, want %+v", got, want)
	}
}

func TestParagraphUsesWholeWideLine(t *testing.T) {
	targets := []string{"when", "this", "made", "while", "from", "however", "some", "between", "now", "world"}
	paragraph, err := BuildParagraph(targets, 118)
	if err != nil {
		t.Fatal(err)
	}
	if len(paragraph.Lines) != 1 || paragraph.Lines[0] != (ParagraphLine{Start: 0, End: len(targets)}) {
		t.Fatalf("wide paragraph = %+v, want all targets in one line", paragraph)
	}
}

func TestParagraphRejectsWordWiderThanTerminal(t *testing.T) {
	if _, err := BuildParagraph([]string{"this-word-is-too-wide"}, 8); err != ErrWordTooWide {
		t.Fatalf("error = %v, want ErrWordTooWide", err)
	}
}

func TestTargetWPMOverridesPerRuneThreshold(t *testing.T) {
	config := DefaultTimingConfig()
	config.TargetWPM = 60
	if got, want := config.slowThreshold("hello"), time.Second; got != want {
		t.Fatalf("threshold = %s, want %s", got, want)
	}
}
