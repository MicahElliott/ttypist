package main

import (
	"bufio"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

//go:embed data/10k-3.num
var embeddedDefaultDictionary []byte

var (
	ErrInvalidDictionaryLine = errors.New("invalid dictionary line")
	ErrInvalidPool           = errors.New("invalid dictionary pool")
	ErrNotEnoughWords        = errors.New("dictionary does not contain enough matching words")
)

type DictionaryEntry struct {
	Rank      int
	Frequency int
	Word      string
}

type Dictionary struct {
	entries []DictionaryEntry
}

type SelectionConfig struct {
	Count     int
	PoolStart int
	PoolEnd   int
	Pattern   *regexp.Regexp
	Seed      int64
}

func DefaultSelectionConfig() SelectionConfig {
	return SelectionConfig{
		Count:     50,
		PoolStart: 1,
		PoolEnd:   200,
	}
}

func ParseDictionary(r io.Reader) (Dictionary, error) {
	scanner := bufio.NewScanner(r)
	entries := make([]DictionaryEntry, 0)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		fields := strings.SplitN(scanner.Text(), "\t", 3)
		if len(fields) != 3 || fields[2] == "" {
			return Dictionary{}, fmt.Errorf("%w at line %d", ErrInvalidDictionaryLine, lineNumber)
		}
		rank, err := strconv.Atoi(fields[0])
		if err != nil || rank < 1 {
			return Dictionary{}, fmt.Errorf("%w at line %d: rank %q", ErrInvalidDictionaryLine, lineNumber, fields[0])
		}
		frequency, err := strconv.Atoi(fields[1])
		if err != nil || frequency < 0 {
			return Dictionary{}, fmt.Errorf("%w at line %d: frequency %q", ErrInvalidDictionaryLine, lineNumber, fields[1])
		}
		entries = append(entries, DictionaryEntry{
			Rank:      rank,
			Frequency: frequency,
			Word:      fields[2],
		})
	}
	if err := scanner.Err(); err != nil {
		return Dictionary{}, err
	}
	return Dictionary{entries: entries}, nil
}

func DefaultDictionary() Dictionary {
	return defaultDictionary
}

func (d Dictionary) Entries() []DictionaryEntry {
	return append([]DictionaryEntry(nil), d.entries...)
}

func (d Dictionary) Select(config SelectionConfig) ([]string, error) {
	if config.Count < 0 {
		return nil, fmt.Errorf("%w: word count cannot be negative", ErrInvalidPool)
	}
	if config.PoolStart <= 0 {
		config.PoolStart = 1
	}
	if config.PoolEnd <= 0 {
		config.PoolEnd = maxRank(d.entries)
	}
	if config.PoolStart > config.PoolEnd {
		return nil, fmt.Errorf("%w: %d-%d", ErrInvalidPool, config.PoolStart, config.PoolEnd)
	}
	if config.Count == 0 {
		return []string{}, nil
	}

	candidates := make([]string, 0)
	for _, entry := range d.entries {
		if entry.Rank < config.PoolStart || entry.Rank > config.PoolEnd || !usableDictionaryWord(entry.Word) {
			continue
		}
		if config.Pattern != nil && !config.Pattern.MatchString(entry.Word) {
			continue
		}
		candidates = append(candidates, entry.Word)
	}
	if config.Count > len(candidates) {
		return nil, fmt.Errorf("%w: requested %d, found %d", ErrNotEnoughWords, config.Count, len(candidates))
	}

	rand.New(rand.NewSource(config.Seed)).Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})
	return append([]string(nil), candidates[:config.Count]...), nil
}

func usableDictionaryWord(word string) bool {
	return utf8.RuneCountInString(word) >= 3 && !strings.ContainsAny(word, "'_&")
}

func maxRank(entries []DictionaryEntry) int {
	max := 0
	for _, entry := range entries {
		if entry.Rank > max {
			max = entry.Rank
		}
	}
	return max
}

var defaultDictionary = mustParseEmbeddedDictionary()

func mustParseEmbeddedDictionary() Dictionary {
	dictionary, err := ParseDictionary(strings.NewReader(string(embeddedDefaultDictionary)))
	if err != nil {
		panic(err)
	}
	return dictionary
}
