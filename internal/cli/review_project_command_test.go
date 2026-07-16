package cli

import (
	"path/filepath"
	"testing"

	"github.com/MauricioPerera/showme/internal/domain"
	"github.com/MauricioPerera/showme/internal/storage"
)

func TestRunReviewProjectCommand_Valid(t *testing.T) {
	dir := t.TempDir()
	path := saveNamedProject(t, dir, "Roadmap Q3")

	result, err := RunReviewProjectCommand(ReviewProjectCommandInput{
		Path:     path,
		SlideID:  "intro",
		Decision: "accepted",
		OutDir:   dir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected OK, got errors: %v", result.Errors)
	}
	if result.Path != path {
		t.Fatalf("expected path %q (same file, same name), got %q", path, result.Path)
	}

	updated, loadErr := storage.LoadProject(path)
	if loadErr != nil {
		t.Fatalf("unexpected error reloading project: %v", loadErr)
	}
	if updated.Deck.Slides[0].Status != domain.SlideStatusAccepted {
		t.Fatalf("expected slide accepted, got %q", updated.Deck.Slides[0].Status)
	}
}

func TestRunReviewProjectCommand_InvalidDecision(t *testing.T) {
	dir := t.TempDir()
	path := saveNamedProject(t, dir, "Roadmap Q3")

	result, err := RunReviewProjectCommand(ReviewProjectCommandInput{
		Path:     path,
		SlideID:  "intro",
		Decision: "archived",
		OutDir:   dir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OK {
		t.Fatalf("expected not OK, got %+v", result)
	}
	if !containsError(result.Errors, "invalid decision: archived") {
		t.Fatalf("expected 'invalid decision: archived' error, got %v", result.Errors)
	}
}

func TestRunReviewProjectCommand_SlideNotFound(t *testing.T) {
	dir := t.TempDir()
	path := saveNamedProject(t, dir, "Roadmap Q3")

	result, err := RunReviewProjectCommand(ReviewProjectCommandInput{
		Path:     path,
		SlideID:  "missing",
		Decision: "accepted",
		OutDir:   dir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OK {
		t.Fatalf("expected not OK, got %+v", result)
	}
	if !containsError(result.Errors, "slide not found: missing") {
		t.Fatalf("expected 'slide not found: missing' error, got %v", result.Errors)
	}
}

func TestRunReviewProjectCommand_MissingSource(t *testing.T) {
	dir := t.TempDir()

	_, err := RunReviewProjectCommand(ReviewProjectCommandInput{
		Path:     filepath.Join(dir, "does-not-exist.json"),
		SlideID:  "intro",
		Decision: "accepted",
		OutDir:   dir,
	})

	if err == nil {
		t.Fatalf("expected an error for a missing source file")
	}
}
