package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MauricioPerera/showme/internal/domain"
	"github.com/MauricioPerera/showme/internal/storage"
)

func saveProjectForRename(t *testing.T, dir, name string) string {
	t.Helper()
	deck, deckReport := domain.NewDeck(domain.DeckInput{
		Title:  "Roadmap",
		Slides: []domain.Slide{{ID: "intro", Title: "Introduccion"}},
	})
	if len(deckReport.Errors) != 0 {
		t.Fatalf("setup: unexpected deck errors: %v", deckReport.Errors)
	}

	path, report, err := storage.SaveProject(storage.SaveProjectRequest{
		Dir: dir,
		Input: domain.ProjectInput{
			Name:          name,
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

func TestRenameProject_Valid(t *testing.T) {
	dir := t.TempDir()
	sourcePath := saveProjectForRename(t, dir, "Original")

	newPath, report, err := RenameProject(RenameProjectInput{
		SourcePath: sourcePath,
		NewName:    "Renombrado",
		Dir:        dir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}

	wantPath := filepath.Join(dir, "renombrado.json")
	if newPath != wantPath {
		t.Fatalf("expected path %q, got %q", wantPath, newPath)
	}

	if _, statErr := os.Stat(sourcePath); !os.IsNotExist(statErr) {
		t.Fatalf("expected the old file to be removed, stat err: %v", statErr)
	}

	renamed, loadErr := storage.LoadProject(newPath)
	if loadErr != nil {
		t.Fatalf("unexpected error loading renamed project: %v", loadErr)
	}
	if renamed.Name != "Renombrado" || renamed.Version != 3 {
		t.Fatalf("expected name updated and version preserved, got %+v", renamed)
	}
}

func TestRenameProject_SameNameIsANoOp(t *testing.T) {
	dir := t.TempDir()
	sourcePath := saveProjectForRename(t, dir, "Original")

	newPath, report, err := RenameProject(RenameProjectInput{
		SourcePath: sourcePath,
		NewName:    "Original",
		Dir:        dir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if newPath != sourcePath {
		t.Fatalf("expected same path, got %q vs %q", newPath, sourcePath)
	}
	if _, statErr := os.Stat(sourcePath); statErr != nil {
		t.Fatalf("expected the file to still exist, got: %v", statErr)
	}
}

func TestRenameProject_CollisionRefused(t *testing.T) {
	dir := t.TempDir()
	sourcePath := saveProjectForRename(t, dir, "Original")
	otherPath := saveProjectForRename(t, dir, "Ya Existente")

	newPath, report, err := RenameProject(RenameProjectInput{
		SourcePath: sourcePath,
		NewName:    "Ya Existente",
		Dir:        dir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newPath != "" {
		t.Fatalf("expected empty path, got %q", newPath)
	}
	if !containsError(report.Errors, "a project already exists at that name") {
		t.Fatalf("expected collision error, got %v", report.Errors)
	}
	if _, statErr := os.Stat(sourcePath); statErr != nil {
		t.Fatalf("expected source file untouched, got: %v", statErr)
	}
	if _, statErr := os.Stat(otherPath); statErr != nil {
		t.Fatalf("expected other project untouched, got: %v", statErr)
	}
}

func TestRenameProject_EmptyNewName(t *testing.T) {
	dir := t.TempDir()
	sourcePath := saveProjectForRename(t, dir, "Original")

	newPath, report, err := RenameProject(RenameProjectInput{
		SourcePath: sourcePath,
		NewName:    "",
		Dir:        dir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newPath != "" {
		t.Fatalf("expected empty path, got %q", newPath)
	}
	if !containsError(report.Errors, "name is required") {
		t.Fatalf("expected 'name is required' error, got %v", report.Errors)
	}
}

func TestRenameProject_MissingSource(t *testing.T) {
	dir := t.TempDir()

	_, _, err := RenameProject(RenameProjectInput{
		SourcePath: filepath.Join(dir, "does-not-exist.json"),
		NewName:    "Nuevo",
		Dir:        dir,
	})

	if err == nil {
		t.Fatalf("expected an error for a missing source file")
	}
}
