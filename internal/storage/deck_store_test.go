package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/MauricioPerera/showme/internal/domain"
)

func TestSaveDeck_ValidWritesJSONFile(t *testing.T) {
	dir := t.TempDir()
	input := domain.DeckInput{
		Title:    "Roadmap Q3",
		Audience: "Equipo de producto",
		Slides: []domain.Slide{
			{ID: "intro", Title: "Introduccion"},
		},
	}

	path, report, err := SaveDeck(SaveDeckRequest{Dir: dir, Input: input})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}

	wantPath := filepath.Join(dir, "roadmap-q3.json")
	if path != wantPath {
		t.Fatalf("expected path %q, got %q", wantPath, path)
	}

	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("expected file to exist at %q: %v", path, readErr)
	}

	var deck domain.Deck
	if unmarshalErr := json.Unmarshal(data, &deck); unmarshalErr != nil {
		t.Fatalf("expected valid JSON, got error: %v", unmarshalErr)
	}
	if deck.Title != "Roadmap Q3" {
		t.Fatalf("expected title round-trip, got %q", deck.Title)
	}
	if len(deck.Slides) != 1 || deck.Slides[0].ID != "intro" {
		t.Fatalf("expected slides round-trip, got %+v", deck.Slides)
	}
}

func TestSaveDeck_InvalidDeckWritesNothing(t *testing.T) {
	dir := t.TempDir()
	input := domain.DeckInput{Title: ""}

	path, report, err := SaveDeck(SaveDeckRequest{Dir: dir, Input: input})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "" {
		t.Fatalf("expected empty path when deck is invalid, got %q", path)
	}
	if len(report.Errors) == 0 {
		t.Fatalf("expected validation errors from an empty title")
	}

	entries, readErr := os.ReadDir(dir)
	if readErr != nil {
		t.Fatalf("unexpected error reading dir: %v", readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no files written for an invalid deck, found %v", entries)
	}
}

func TestSaveDeck_TitleProducesEmptySlug(t *testing.T) {
	dir := t.TempDir()
	input := domain.DeckInput{
		Title:  "!!!",
		Slides: []domain.Slide{{ID: "s1", Title: "Uno"}},
	}

	path, report, err := SaveDeck(SaveDeckRequest{Dir: dir, Input: input})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "" {
		t.Fatalf("expected empty path, got %q", path)
	}
	if !containsError(report.Errors, "title produces an empty slug") {
		t.Fatalf("expected 'title produces an empty slug' error, got %v", report.Errors)
	}
}

func TestSaveDeck_MissingDirReturnsError(t *testing.T) {
	input := domain.DeckInput{
		Title:  "Deck",
		Slides: []domain.Slide{{ID: "s1", Title: "Uno"}},
	}

	_, report, err := SaveDeck(SaveDeckRequest{Dir: "/no/such/dir/for-showme-tests", Input: input})

	if err == nil {
		t.Fatalf("expected an I/O error for a missing directory")
	}
	if len(report.Errors) != 0 {
		t.Fatalf("expected no validation errors, got %v", report.Errors)
	}
}

func containsError(errors []string, want string) bool {
	for _, e := range errors {
		if e == want {
			return true
		}
	}
	return false
}
