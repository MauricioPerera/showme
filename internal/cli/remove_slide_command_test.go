package cli

import (
	"path/filepath"
	"testing"

	"github.com/MauricioPerera/showme/internal/storage"
)

func TestRunRemoveSlideCommand_Valid(t *testing.T) {
	dir := t.TempDir()
	path := saveNamedProject(t, dir, "Roadmap Q3")

	addResult, err := RunAddSlideCommand(AddSlideCommandInput{
		Path: path, SlideID: "closing", Title: "Cierre", OutDir: dir,
	})
	if err != nil || !addResult.OK {
		t.Fatalf("setup: unexpected AddSlide failure: err=%v result=%+v", err, addResult)
	}

	result, err := RunRemoveSlideCommand(RemoveSlideCommandInput{
		Path: path, SlideID: "closing", OutDir: dir,
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
	if len(updated.Deck.Slides) != 1 || updated.Deck.Slides[0].ID != "intro" {
		t.Fatalf("expected only 'intro' remaining, got %+v", updated.Deck.Slides)
	}
}

func TestRunRemoveSlideCommand_LastSlideRefused(t *testing.T) {
	dir := t.TempDir()
	path := saveNamedProject(t, dir, "Roadmap Q3")

	result, err := RunRemoveSlideCommand(RemoveSlideCommandInput{
		Path: path, SlideID: "intro", OutDir: dir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OK {
		t.Fatalf("expected not OK, got %+v", result)
	}
	if !containsError(result.Errors, "deck must have at least one slide") {
		t.Fatalf("expected 'deck must have at least one slide' error, got %v", result.Errors)
	}
}

func TestRunRemoveSlideCommand_MissingSource(t *testing.T) {
	dir := t.TempDir()

	_, err := RunRemoveSlideCommand(RemoveSlideCommandInput{
		Path: filepath.Join(dir, "does-not-exist.json"), SlideID: "intro", OutDir: dir,
	})

	if err == nil {
		t.Fatalf("expected an error for a missing source file")
	}
}
