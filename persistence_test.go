package main

import (
	"path/filepath"
	"testing"
	"time"
)

func TestFileSessionStoreAppendsCompleteSessionRecords(t *testing.T) {
	base := time.Unix(100, 0)
	config := DefaultInteractiveConfig()
	config.Selection = SessionSelection{
		Count:   2,
		Pool:    "1-200",
		Pattern: "^[a-z]+$",
		Seed:    42,
	}
	session := NewSession([]string{"one", "two"}, config.Timing)
	for _, word := range []string{"one", "two"} {
		for _, r := range word {
			if err := session.Apply(RuneInput(r), base.Add(time.Second)); err != nil {
				t.Fatal(err)
			}
		}
		if err := session.Apply(Input{Kind: InputSpace}, base.Add(2*time.Second)); err != nil {
			t.Fatal(err)
		}
	}

	path := filepath.Join(t.TempDir(), "nested", "sessions.jsonl")
	store := FileSessionStore{Path: path}
	if err := store.Save(session.Record(config, base.Add(3*time.Second))); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(session.Record(config, base.Add(4*time.Second))); err != nil {
		t.Fatal(err)
	}

	records, err := ReadSessionRecords(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("record count = %d, want 2", len(records))
	}
	record := records[0]
	if record.Version != sessionRecordVersion || record.Status != "completed" {
		t.Fatalf("record identity = %+v, want version %d completed", record, sessionRecordVersion)
	}
	if record.StartedAt == "" || record.EndedAt == "" {
		t.Fatalf("record timestamps = %q, %q, want both present", record.StartedAt, record.EndedAt)
	}
	if record.Selection.Pattern != "^[a-z]+$" || record.Selection.Seed != 42 {
		t.Fatalf("selection = %+v, want pattern and seed", record.Selection)
	}
	if len(record.Targets) != 2 || len(record.Attempts) != 2 || record.Metrics.Correct != 2 {
		t.Fatalf("record outcome = %+v, want two targets, attempts, and correct words", record)
	}
}

func TestDefaultSessionStorePathUsesXDGDataHome(t *testing.T) {
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)

	path, err := DefaultSessionStorePath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dataHome, "ttypist", "sessions.jsonl")
	if path != want {
		t.Fatalf("store path = %q, want %q", path, want)
	}
}
