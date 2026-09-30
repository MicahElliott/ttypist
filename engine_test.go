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

func TestPromptRepeatsLookaheadAcrossWindows(t *testing.T) {
	targets := []string{"always", "man", "good", "same", "from", "going", "most", "after", "made", "again", "small", "which", "day", "first", "next"}
	first, err := BuildPrompt(targets, 0, 75, 2)
	if err != nil {
		t.Fatal(err)
	}
	if first.BodyEnd != 12 || first.End != 14 {
		t.Fatalf("first prompt = %+v, want body end 12 and end 14", first)
	}
	if got := first.Lookahead; len(got) != 2 || got[0] != "day" || got[1] != "first" {
		t.Fatalf("first lookahead = %v, want [day first]", got)
	}
	second, err := BuildPrompt(targets, first.BodyEnd, 75, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Words) == 0 || second.Words[0] != "day" || second.Words[1] != "first" {
		t.Fatalf("second words = %v, want to begin with [day first]", second.Words)
	}
}

func TestPromptRejectsWordWiderThanTerminal(t *testing.T) {
	if _, err := BuildPrompt([]string{"this-word-is-too-wide"}, 0, 8, 2); err != ErrWordTooWide {
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
