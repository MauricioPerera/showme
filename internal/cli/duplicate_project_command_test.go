package cli

import (
	"path/filepath"
	"testing"

	"github.com/MauricioPerera/showme/internal/storage"
)

func TestRunDuplicateProjectCommand_Valid(t *testing.T) {
	dir := t.TempDir()
	sourcePath := saveNamedProject(t, dir, "Original")

	result, err := RunDuplicateProjectCommand(DuplicateProjectCommandInput{
		SourcePath: sourcePath,
		NewName:    "Copia",
		OutDir:     dir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected OK, got errors: %v", result.Errors)
	}

	wantPath := filepath.Join(dir, "copia.json")
	if result.Path != wantPath {
		t.Fatalf("expected path %q, got %q", wantPath, result.Path)
	}

	dup, loadErr := storage.LoadProject(result.Path)
	if loadErr != nil {
		t.Fatalf("unexpected error loading duplicate: %v", loadErr)
	}
	if dup.Name != "Copia" || dup.Version != 1 {
		t.Fatalf("expected name/version reset, got %+v", dup)
	}
}

func TestRunDuplicateProjectCommand_EmptyNewName(t *testing.T) {
	dir := t.TempDir()
	sourcePath := saveNamedProject(t, dir, "Original")

	result, err := RunDuplicateProjectCommand(DuplicateProjectCommandInput{
		SourcePath: sourcePath,
		NewName:    "",
		OutDir:     dir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OK {
		t.Fatalf("expected not OK, got %+v", result)
	}
	if !containsError(result.Errors, "name is required") {
		t.Fatalf("expected 'name is required' error, got %v", result.Errors)
	}
}

func TestRunDuplicateProjectCommand_MissingSource(t *testing.T) {
	dir := t.TempDir()

	_, err := RunDuplicateProjectCommand(DuplicateProjectCommandInput{
		SourcePath: filepath.Join(dir, "does-not-exist.json"),
		NewName:    "Copia",
		OutDir:     dir,
	})

	if err == nil {
		t.Fatalf("expected an error for a missing source file")
	}
}
