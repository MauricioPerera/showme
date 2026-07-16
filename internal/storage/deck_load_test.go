package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MauricioPerera/showme/internal/domain"
)

func TestLoadDeck_RoundTripsWhatSaveDeckWrote(t *testing.T) {
	dir := t.TempDir()
	input := domain.DeckInput{
		Title:    "Roadmap Q3",
		Audience: "Equipo de producto",
		Slides: []domain.Slide{
			{ID: "intro", Title: "Introduccion"},
			{ID: "plan", Title: "Plan", Status: domain.SlideStatusAccepted},
		},
	}

	path, report, err := SaveDeck(SaveDeckRequest{Dir: dir, Input: input})
	if err != nil || len(report.Errors) != 0 {
		t.Fatalf("setup: unexpected SaveDeck failure: err=%v report=%v", err, report)
	}

	deck, loadErr := LoadDeck(path)

	if loadErr != nil {
		t.Fatalf("unexpected error: %v", loadErr)
	}
	if deck.Title != "Roadmap Q3" || deck.Audience != "Equipo de producto" {
		t.Fatalf("expected title/audience round-trip, got %+v", deck)
	}
	if len(deck.Slides) != 2 || deck.Slides[0].ID != "intro" || deck.Slides[1].ID != "plan" {
		t.Fatalf("expected slide order preserved, got %+v", deck.Slides)
	}
	if deck.Slides[1].Status != domain.SlideStatusAccepted {
		t.Fatalf("expected status round-trip, got %q", deck.Slides[1].Status)
	}
}

func TestLoadDeck_MissingFileReturnsError(t *testing.T) {
	dir := t.TempDir()

	_, err := LoadDeck(filepath.Join(dir, "does-not-exist.json"))

	if err == nil {
		t.Fatalf("expected an error for a missing file")
	}
}

func TestLoadDeck_InvalidJSONReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "broken.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, err := LoadDeck(path)

	if err == nil {
		t.Fatalf("expected an error for invalid JSON content")
	}
}
