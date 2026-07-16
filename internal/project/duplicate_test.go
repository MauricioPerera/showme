package project

import (
	"path/filepath"
	"testing"

	"github.com/MauricioPerera/showme/internal/domain"
	"github.com/MauricioPerera/showme/internal/storage"
)

func saveSourceProject(t *testing.T, dir string) string {
	t.Helper()
	deck, deckReport := domain.NewDeck(domain.DeckInput{
		Title:  "Roadmap Q3",
		Slides: []domain.Slide{{ID: "intro", Title: "Introduccion"}},
	})
	if len(deckReport.Errors) != 0 {
		t.Fatalf("setup: unexpected deck errors: %v", deckReport.Errors)
	}

	path, report, err := storage.SaveProject(storage.SaveProjectRequest{
		Dir: dir,
		Input: domain.ProjectInput{
			Name:          "Original",
			Deck:          deck,
			DesignPath:    "DESIGN.md",
			KnowledgePath: "knowledge/showme",
			Version:       3,
		},
	})
	if err != nil || len(report.Errors) != 0 {
		t.Fatalf("setup: unexpected SaveProject failure: err=%v report=%v", err, report)
	}
	return path
}

func TestDuplicateProject_Valid(t *testing.T) {
	dir := t.TempDir()
	sourcePath := saveSourceProject(t, dir)

	newPath, report, err := DuplicateProject(DuplicateProjectInput{
		SourcePath: sourcePath,
		NewName:    "Copia",
		Dir:        dir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}

	wantPath := filepath.Join(dir, "copia.json")
	if newPath != wantPath {
		t.Fatalf("expected path %q, got %q", wantPath, newPath)
	}

	duplicated, loadErr := storage.LoadProject(newPath)
	if loadErr != nil {
		t.Fatalf("unexpected error loading duplicate: %v", loadErr)
	}
	if duplicated.Name != "Copia" {
		t.Fatalf("expected name 'Copia', got %q", duplicated.Name)
	}
	if duplicated.Version != 1 {
		t.Fatalf("expected version reset to 1, got %d", duplicated.Version)
	}
	if duplicated.DesignPath != "DESIGN.md" || duplicated.KnowledgePath != "knowledge/showme" {
		t.Fatalf("expected design/knowledge paths preserved, got %+v", duplicated)
	}
	if len(duplicated.Deck.Slides) != 1 || duplicated.Deck.Slides[0].ID != "intro" {
		t.Fatalf("expected deck preserved, got %+v", duplicated.Deck)
	}

	original, origErr := storage.LoadProject(sourcePath)
	if origErr != nil {
		t.Fatalf("unexpected error loading original: %v", origErr)
	}
	if original.Name != "Original" || original.Version != 3 {
		t.Fatalf("expected original untouched, got %+v", original)
	}
}

func TestDuplicateProject_EmptyNewName(t *testing.T) {
	dir := t.TempDir()
	sourcePath := saveSourceProject(t, dir)

	path, report, err := DuplicateProject(DuplicateProjectInput{
		SourcePath: sourcePath,
		NewName:    "",
		Dir:        dir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "" {
		t.Fatalf("expected empty path, got %q", path)
	}
	if !containsError(report.Errors, "name is required") {
		t.Fatalf("expected 'name is required' error, got %v", report.Errors)
	}
}

func TestDuplicateProject_MissingSource(t *testing.T) {
	dir := t.TempDir()

	_, _, err := DuplicateProject(DuplicateProjectInput{
		SourcePath: filepath.Join(dir, "does-not-exist.json"),
		NewName:    "Copia",
		Dir:        dir,
	})

	if err == nil {
		t.Fatalf("expected an error for a missing source file")
	}
}

func TestDuplicateProject_MissingTargetDir(t *testing.T) {
	dir := t.TempDir()
	sourcePath := saveSourceProject(t, dir)

	_, _, err := DuplicateProject(DuplicateProjectInput{
		SourcePath: sourcePath,
		NewName:    "Copia",
		Dir:        filepath.Join(dir, "no-such-subdir"),
	})

	if err == nil {
		t.Fatalf("expected an error for a missing target directory")
	}
}
