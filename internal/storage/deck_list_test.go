package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MauricioPerera/showme/internal/domain"
)

func TestListDecks_ReturnsSortedJSONPaths(t *testing.T) {
	dir := t.TempDir()

	titles := []string{"Roadmap Q3", "Onboarding", "Kickoff"}
	for _, title := range titles {
		_, report, err := SaveDeck(SaveDeckRequest{
			Dir: dir,
			Input: domain.DeckInput{
				Title:  title,
				Slides: []domain.Slide{{ID: "s1", Title: "Uno"}},
			},
		})
		if err != nil || len(report.Errors) != 0 {
			t.Fatalf("setup: unexpected SaveDeck failure: err=%v report=%v", err, report)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("ignore me"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "subdir.json"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	paths, err := ListDecks(dir)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{
		filepath.Join(dir, "kickoff.json"),
		filepath.Join(dir, "onboarding.json"),
		filepath.Join(dir, "roadmap-q3.json"),
	}
	if len(paths) != len(want) {
		t.Fatalf("expected %v, got %v", want, paths)
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, paths)
		}
	}
}

func TestListDecks_EmptyDirReturnsEmptySlice(t *testing.T) {
	dir := t.TempDir()

	paths, err := ListDecks(dir)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 0 {
		t.Fatalf("expected no paths, got %v", paths)
	}
}

func TestListDecks_MissingDirReturnsError(t *testing.T) {
	dir := t.TempDir()

	_, err := ListDecks(filepath.Join(dir, "does-not-exist"))

	if err == nil {
		t.Fatalf("expected an error for a missing directory")
	}
}
