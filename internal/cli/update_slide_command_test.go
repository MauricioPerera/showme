package cli

import (
	"path/filepath"
	"testing"

	"github.com/MauricioPerera/showme/internal/domain"
	"github.com/MauricioPerera/showme/internal/storage"
)

func TestRunUpdateSlideCommand_PreservesStatusWhenNotGiven(t *testing.T) {
	dir := t.TempDir()
	path := saveNamedProject(t, dir, "Roadmap Q3")

	reviewResult, err := RunReviewProjectCommand(ReviewProjectCommandInput{
		Path: path, SlideID: "intro", Decision: "accepted", OutDir: dir,
	})
	if err != nil || !reviewResult.OK {
		t.Fatalf("setup: unexpected review failure: err=%v result=%+v", err, reviewResult)
	}

	result, err := RunUpdateSlideCommand(UpdateSlideCommandInput{
		Path: path, SlideID: "intro", Title: "Introduccion revisada", OutDir: dir,
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
	if updated.Deck.Slides[0].Title != "Introduccion revisada" {
		t.Fatalf("expected title updated, got %+v", updated.Deck.Slides[0])
	}
	if updated.Deck.Slides[0].Status != domain.SlideStatusAccepted {
		t.Fatalf("expected status preserved as accepted, got %q", updated.Deck.Slides[0].Status)
	}
}

func TestRunUpdateSlideCommand_NotFound(t *testing.T) {
	dir := t.TempDir()
	path := saveNamedProject(t, dir, "Roadmap Q3")

	result, err := RunUpdateSlideCommand(UpdateSlideCommandInput{
		Path: path, SlideID: "missing", Title: "Titulo", OutDir: dir,
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

func TestRunUpdateSlideCommand_MissingSource(t *testing.T) {
	dir := t.TempDir()

	_, err := RunUpdateSlideCommand(UpdateSlideCommandInput{
		Path: filepath.Join(dir, "does-not-exist.json"), SlideID: "intro", Title: "Titulo", OutDir: dir,
	})

	if err == nil {
		t.Fatalf("expected an error for a missing source file")
	}
}
