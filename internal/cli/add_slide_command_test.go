package cli

import (
	"path/filepath"
	"testing"

	"github.com/MauricioPerera/showme/internal/storage"
)

func TestRunAddSlideCommand_Valid(t *testing.T) {
	dir := t.TempDir()
	path := saveNamedProject(t, dir, "Roadmap Q3")

	result, err := RunAddSlideCommand(AddSlideCommandInput{
		Path:    path,
		SlideID: "closing",
		Title:   "Cierre",
		OutDir:  dir,
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
	if len(updated.Deck.Slides) != 2 || updated.Deck.Slides[1].ID != "closing" {
		t.Fatalf("expected new slide appended, got %+v", updated.Deck.Slides)
	}
}

func TestRunAddSlideCommand_DuplicateID(t *testing.T) {
	dir := t.TempDir()
	path := saveNamedProject(t, dir, "Roadmap Q3")

	result, err := RunAddSlideCommand(AddSlideCommandInput{
		Path:    path,
		SlideID: "intro",
		Title:   "Otra intro",
		OutDir:  dir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OK {
		t.Fatalf("expected not OK, got %+v", result)
	}
	if !containsError(result.Errors, "duplicate slide id: intro") {
		t.Fatalf("expected 'duplicate slide id: intro' error, got %v", result.Errors)
	}
}

func TestRunAddSlideCommand_PreservesArchivedFlag(t *testing.T) {
	dir := t.TempDir()
	path := saveNamedProject(t, dir, "Roadmap Q3")

	archiveResult, err := RunArchiveProjectCommand(ArchiveProjectCommandInput{
		Path: path, Archived: true, OutDir: dir,
	})
	if err != nil || !archiveResult.OK {
		t.Fatalf("setup: unexpected archive failure: err=%v result=%+v", err, archiveResult)
	}

	result, err := RunAddSlideCommand(AddSlideCommandInput{
		Path: path, SlideID: "closing", Title: "Cierre", OutDir: dir,
	})
	if err != nil || !result.OK {
		t.Fatalf("unexpected AddSlide failure: err=%v result=%+v", err, result)
	}

	updated, loadErr := storage.LoadProject(path)
	if loadErr != nil {
		t.Fatalf("unexpected error reloading project: %v", loadErr)
	}
	if !updated.Archived {
		t.Fatalf("expected Archived preserved through add-slide, got %+v", updated)
	}
	if len(updated.Deck.Slides) != 2 {
		t.Fatalf("expected new slide still appended, got %+v", updated.Deck.Slides)
	}
}

func TestRunAddSlideCommand_MissingSource(t *testing.T) {
	dir := t.TempDir()

	_, err := RunAddSlideCommand(AddSlideCommandInput{
		Path:    filepath.Join(dir, "does-not-exist.json"),
		SlideID: "closing",
		Title:   "Cierre",
		OutDir:  dir,
	})

	if err == nil {
		t.Fatalf("expected an error for a missing source file")
	}
}
