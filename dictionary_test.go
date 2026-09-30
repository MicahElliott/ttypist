package main

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestDefaultDictionaryIsEmbedded(t *testing.T) {
	dictionary := DefaultDictionary()
	entries := dictionary.Entries()
	if len(entries) != 10000 {
		t.Fatalf("default dictionary entries = %d, want 10000", len(entries))
	}
	if got := entries[0]; got.Rank != 1 || got.Word != "!!WHOLE_CORPUS" {
		t.Fatalf("first dictionary entry = %+v, want rank 1 corpus marker", got)
	}
	if got := entries[1]; got.Rank != 2 || got.Word != "the" {
		t.Fatalf("second dictionary entry = %+v, want rank 2 the", got)
	}
}

func TestDictionarySelectionIsDeterministicAndFiltered(t *testing.T) {
	dictionary, err := ParseDictionary(strings.NewReader(strings.Join([]string{
		"1\t100\tone",
		"2\t90\ttwo",
		"3\t80\ttoo",
		"4\t70\ta_b",
		"5\t60\tno",
		"6\t50\tthree",
		"7\t40\tother",
	}, "\n")))
	if err != nil {
		t.Fatal(err)
	}
	config := SelectionConfig{
		Count:     3,
		PoolStart: 1,
		PoolEnd:   6,
		Pattern:   regexp.MustCompile(`^t`),
		Seed:      42,
	}
	first, err := dictionary.Select(config)
	if err != nil {
		t.Fatal(err)
	}
	second, err := dictionary.Select(config)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("same selection config produced %v and %v", first, second)
	}
	wantSet := map[string]bool{"one": true, "two": true, "too": true, "three": true}
	for _, word := range first {
		if !wantSet[word] || !strings.HasPrefix(word, "t") {
			t.Fatalf("selected word %q does not satisfy pool and pattern", word)
		}
	}
}

func TestDictionarySelectionRejectsInsufficientMatches(t *testing.T) {
	dictionary, err := ParseDictionary(strings.NewReader("1\t1\tone\n"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = dictionary.Select(SelectionConfig{Count: 2, PoolStart: 1, PoolEnd: 1})
	if err == nil || !strings.Contains(err.Error(), ErrNotEnoughWords.Error()) {
		t.Fatalf("error = %v, want insufficient-word error", err)
	}
}
