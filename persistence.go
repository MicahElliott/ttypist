package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const sessionRecordVersion = 1

type SessionSelection struct {
	Count      int    `json:"count"`
	Pool       string `json:"pool,omitempty"`
	Pattern    string `json:"pattern,omitempty"`
	Dictionary string `json:"dictionary,omitempty"`
	Input      string `json:"input,omitempty"`
	Seed       int64  `json:"seed"`
}

type SessionRecord struct {
	Version   int              `json:"version"`
	StartedAt string           `json:"started_at,omitempty"`
	EndedAt   string           `json:"ended_at,omitempty"`
	Status    string           `json:"status"`
	Selection SessionSelection `json:"selection"`
	Timing    TimingConfig     `json:"timing"`
	MinWPM    float64          `json:"min_wpm,omitempty"`
	MinAcc    float64          `json:"min_accuracy,omitempty"`
	Targets   []string         `json:"targets"`
	Attempts  []Attempt        `json:"attempts"`
	Metrics   Metrics          `json:"metrics"`
}

type SessionStore interface {
	Save(SessionRecord) error
}

type FileSessionStore struct {
	Path string
}

func (s *Session) Record(config InteractiveConfig, now time.Time) SessionRecord {
	record := SessionRecord{
		Version:   sessionRecordVersion,
		Status:    sessionStatusName(s.status),
		Selection: config.Selection,
		Timing:    s.config,
		MinWPM:    config.MinWPM,
		MinAcc:    config.MinAccuracy,
		Targets:   s.Targets(),
		Attempts:  s.Attempts(),
		Metrics:   s.Metrics(now),
	}
	if !s.StartedAt().IsZero() {
		record.StartedAt = s.StartedAt().UTC().Format(time.RFC3339Nano)
	}
	if !s.EndedAt().IsZero() {
		record.EndedAt = s.EndedAt().UTC().Format(time.RFC3339Nano)
	}
	return record
}

func sessionStatusName(status SessionStatus) string {
	switch status {
	case SessionCompleted:
		return "completed"
	case SessionAborted:
		return "aborted"
	default:
		return "active"
	}
}

func (s FileSessionStore) Save(record SessionRecord) error {
	if s.Path == "" {
		return errors.New("session store path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(s.Path), 0700); err != nil {
		return fmt.Errorf("create session data directory: %w", err)
	}
	file, err := os.OpenFile(s.Path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("open session store: %w", err)
	}
	defer file.Close()

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("encode session record: %w", err)
	}
	data = append(data, '\n')
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("write session record: %w", err)
	}
	return nil
}

func DefaultSessionStorePath() (string, error) {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("find home directory: %w", err)
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, "ttypist", "sessions.jsonl"), nil
}

func ReadSessionRecords(path string) ([]SessionRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var records []SessionRecord
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for line := 1; scanner.Scan(); line++ {
		var record SessionRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, fmt.Errorf("decode session record at line %d: %w", line, err)
		}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read session store: %w", err)
	}
	return records, nil
}
