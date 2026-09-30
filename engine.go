package main

import (
	"errors"
	"time"
	"unicode/utf8"
)

const defaultSlowPerRune = 250 * time.Millisecond

var ErrSessionNotActive = errors.New("typing session is not active")
var ErrWordTooWide = errors.New("word is wider than the available prompt")

type SessionStatus uint8

const (
	SessionActive SessionStatus = iota
	SessionCompleted
	SessionAborted
)

type InputKind uint8

const (
	InputRune InputKind = iota
	InputSpace
	InputBackspace
	InputDeleteWord
	InputCtrlC
)

type Input struct {
	Kind InputKind
	Rune rune
}

func RuneInput(r rune) Input { return Input{Kind: InputRune, Rune: r} }

type TimingConfig struct {
	SlowPerRune time.Duration
	TargetWPM   float64
	Penalty     time.Duration
}

func DefaultTimingConfig() TimingConfig {
	return TimingConfig{SlowPerRune: defaultSlowPerRune, Penalty: time.Second}
}

func (c TimingConfig) slowThreshold(target string) time.Duration {
	runes := utf8.RuneCountInString(target)
	if c.TargetWPM > 0 {
		seconds := float64(runes) * 12 / c.TargetWPM
		return time.Duration(seconds * float64(time.Second))
	}
	perRune := c.SlowPerRune
	if perRune <= 0 {
		perRune = defaultSlowPerRune
	}
	return time.Duration(runes) * perRune
}

type Attempt struct {
	Target    string
	Entered   string
	Correct   bool
	Slow      bool
	Duration  time.Duration
	TargetIdx int
}

type Session struct {
	targets        []string
	config         TimingConfig
	status         SessionStatus
	focus          int
	current        []rune
	wordStarted    time.Time
	sessionStarted time.Time
	endedAt        time.Time
	attempts       []Attempt
}

func NewSession(targets []string, config TimingConfig) *Session {
	copyTargets := append([]string(nil), targets...)
	if config.SlowPerRune <= 0 && config.TargetWPM <= 0 {
		config.SlowPerRune = defaultSlowPerRune
	}
	return &Session{
		targets: copyTargets,
		config:  config,
		status:  SessionActive,
	}
}

func (s *Session) Apply(input Input, now time.Time) error {
	if s.status != SessionActive {
		return ErrSessionNotActive
	}

	switch input.Kind {
	case InputCtrlC:
		s.abort(now)
	case InputBackspace:
		if len(s.current) > 0 {
			s.current = s.current[:len(s.current)-1]
		}
	case InputDeleteWord:
		s.current = nil
	case InputRune:
		if input.Rune == ' ' || input.Rune == '\t' || input.Rune == '\r' || input.Rune == '\n' {
			return nil
		}
		if !utf8.ValidRune(input.Rune) {
			return nil
		}
		if s.wordStarted.IsZero() {
			s.wordStarted = now
			if s.sessionStarted.IsZero() {
				s.sessionStarted = now
			}
		}
		s.current = append(s.current, input.Rune)
	case InputSpace:
		if len(s.current) == 0 {
			return nil
		}
		s.commit(now)
	}
	return nil
}

func (s *Session) commit(now time.Time) {
	target := s.targets[s.focus]
	entered := string(s.current)
	duration := now.Sub(s.wordStarted)
	s.attempts = append(s.attempts, Attempt{
		Target:    target,
		Entered:   entered,
		Correct:   entered == target,
		Slow:      duration > s.config.slowThreshold(target),
		Duration:  duration,
		TargetIdx: s.focus,
	})
	s.current = nil
	s.wordStarted = time.Time{}
	s.focus++
	if s.focus == len(s.targets) {
		s.status = SessionCompleted
		s.endedAt = now
	}
}

func (s *Session) abort(now time.Time) {
	s.status = SessionAborted
	s.endedAt = now
}

func (s *Session) Status() SessionStatus { return s.status }

func (s *Session) FocusIndex() int { return s.focus }

func (s *Session) CurrentTarget() string {
	if s.focus >= len(s.targets) {
		return ""
	}
	return s.targets[s.focus]
}

func (s *Session) CurrentText() string { return string(s.current) }

func (s *Session) Targets() []string {
	return append([]string(nil), s.targets...)
}

func (s *Session) Attempts() []Attempt {
	return append([]Attempt(nil), s.attempts...)
}

type Metrics struct {
	Elapsed      time.Duration
	RawWPM       float64
	PenalizedWPM float64
	Accuracy     float64
	Correct      int
	Incorrect    int
	Unattempted  int
	Attempted    int
}

func (s *Session) Metrics(now time.Time) Metrics {
	m := Metrics{
		Attempted:   len(s.attempts),
		Unattempted: len(s.targets) - len(s.attempts),
	}
	if m.Unattempted < 0 {
		m.Unattempted = 0
	}
	for _, attempt := range s.attempts {
		if attempt.Correct {
			m.Correct++
		} else {
			m.Incorrect++
		}
	}
	if m.Attempted > 0 {
		m.Accuracy = float64(m.Correct) / float64(m.Attempted) * 100
	}
	if s.sessionStarted.IsZero() {
		return m
	}
	end := now
	if !s.endedAt.IsZero() {
		end = s.endedAt
	}
	m.Elapsed = end.Sub(s.sessionStarted)
	if m.Elapsed <= 0 {
		return m
	}
	var targetRunes int
	for _, attempt := range s.attempts {
		targetRunes += utf8.RuneCountInString(attempt.Target)
	}
	minutes := m.Elapsed.Minutes()
	m.RawWPM = float64(targetRunes) / 5 / minutes
	penalized := m.Elapsed + time.Duration(m.Incorrect)*s.config.Penalty
	if penalized > 0 {
		m.PenalizedWPM = float64(targetRunes) / 5 / penalized.Minutes()
	}
	return m
}

type Prompt struct {
	Start     int
	BodyEnd   int
	End       int
	Words     []string
	Body      []string
	Lookahead []string
}

// BuildPrompt returns a terminal-width prompt window. BodyEnd is the logical
// index at which the next window starts; the words between BodyEnd and End are
// lookahead and will be repeated visually in that next window.
func BuildPrompt(targets []string, start, width, lookahead int) (Prompt, error) {
	if start < 0 {
		start = 0
	}
	if start > len(targets) {
		start = len(targets)
	}
	if width < 1 {
		width = 1
	}
	if lookahead < 0 {
		lookahead = 0
	}
	if start == len(targets) {
		return Prompt{Start: start, BodyEnd: start, End: start}, nil
	}
	for _, target := range targets[start:] {
		if utf8.RuneCountInString(target) > width {
			return Prompt{}, ErrWordTooWide
		}
	}

	end := start
	lineWidth := 0
	for end < len(targets) {
		wordWidth := utf8.RuneCountInString(targets[end])
		nextWidth := wordWidth
		if end > start {
			nextWidth++ // separating space
		}
		if end > start && lineWidth+nextWidth > width {
			break
		}
		lineWidth += nextWidth
		end++
	}
	if end == start {
		return Prompt{}, ErrWordTooWide
	}

	overlap := lookahead
	if overlap >= end-start {
		overlap = 0
	}
	bodyEnd := end - overlap
	if bodyEnd <= start {
		bodyEnd = end
	}

	words := append([]string(nil), targets[start:end]...)
	body := append([]string(nil), targets[start:bodyEnd]...)
	trailing := append([]string(nil), targets[bodyEnd:end]...)
	return Prompt{
		Start:     start,
		BodyEnd:   bodyEnd,
		End:       end,
		Words:     words,
		Body:      body,
		Lookahead: trailing,
	}, nil
}
