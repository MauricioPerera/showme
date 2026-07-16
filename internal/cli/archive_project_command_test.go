package cli

import (
	"path/filepath"
	"testing"

	"github.com/MauricioPerera/showme/internal/storage"
)

func TestRunArchiveProjectCommand_Archive(t *testing.T) {
	dir := t.TempDir()
	path := saveNamedProject(t, dir, "Roadmap Q3")

	result, err := RunArchiveProjectCommand(ArchiveProjectCommandInput{
		Path: path, Archived: true, OutDir: dir,
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
	if !updated.Archived {
		t.Fatalf("expected project archived, got %+v", updated)
	}
}

func TestRunArchiveProjectCommand_Unarchive(t *testing.T) {
	dir := t.TempDir()
	path := saveNamedProject(t, dir, "Roadmap Q3")

	_, err := RunArchiveProjectCommand(ArchiveProjectCommandInput{Path: path, Archived: true, OutDir: dir})
	if err != nil {
		t.Fatalf("setup: unexpected error: %v", err)
	}

	result, err := RunArchiveProjectCommand(ArchiveProjectCommandInput{
		Path: path, Archived: false, OutDir: dir,
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
	if updated.Archived {
		t.Fatalf("expected project unarchived, got %+v", updated)
	}
}

func TestRunArchiveProjectCommand_MissingSource(t *testing.T) {
	dir := t.TempDir()

	_, err := RunArchiveProjectCommand(ArchiveProjectCommandInput{
		Path: filepath.Join(dir, "does-not-exist.json"), Archived: true, OutDir: dir,
	})

	if err == nil {
		t.Fatalf("expected an error for a missing source file")
	}
}
