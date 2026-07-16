package cli

import (
	"path/filepath"
	"testing"

	"github.com/MauricioPerera/showme/internal/storage"
)

func TestRunUpdateDeckInfoCommand_Valid(t *testing.T) {
	dir := t.TempDir()
	path := saveNamedProject(t, dir, "Roadmap Q3")

	result, err := RunUpdateDeckInfoCommand(UpdateDeckInfoCommandInput{
		Path:     path,
		Title:    "Roadmap Q4",
		Audience: "Equipo ejecutivo",
		OutDir:   dir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected OK, got errors: %v", result.Errors)
	}

	updated, loadErr := storage.LoadProject(path)
	if loadErr != nil {
		t.Fatalf("unexpected error reloading project: %v", loadErr)
	}
	if updated.Deck.Title != "Roadmap Q4" || updated.Deck.Audience != "Equipo ejecutivo" {
		t.Fatalf("expected title/audience updated, got %+v", updated.Deck)
	}
	if len(updated.Deck.Slides) != 1 || updated.Deck.Slides[0].ID != "intro" {
		t.Fatalf("expected slides preserved, got %+v", updated.Deck.Slides)
	}
	if updated.Name != "Roadmap Q3" {
		t.Fatalf("expected project Name untouched, got %q", updated.Name)
	}
}

func TestRunUpdateDeckInfoCommand_EmptyTitle(t *testing.T) {
	dir := t.TempDir()
	path := saveNamedProject(t, dir, "Roadmap Q3")

	result, err := RunUpdateDeckInfoCommand(UpdateDeckInfoCommandInput{
		Path:   path,
		Title:  "",
		OutDir: dir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OK {
		t.Fatalf("expected not OK, got %+v", result)
	}
	if !containsError(result.Errors, "title is required") {
		t.Fatalf("expected 'title is required' error, got %v", result.Errors)
	}
}

func TestRunUpdateDeckInfoCommand_MissingSource(t *testing.T) {
	dir := t.TempDir()

	_, err := RunUpdateDeckInfoCommand(UpdateDeckInfoCommandInput{
		Path:   filepath.Join(dir, "does-not-exist.json"),
		Title:  "Roadmap Q4",
		OutDir: dir,
	})

	if err == nil {
		t.Fatalf("expected an error for a missing source file")
	}
}
