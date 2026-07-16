package cli

import (
	"path/filepath"
	"testing"

	"github.com/MauricioPerera/showme/internal/storage"
)

func TestRunRenameProjectCommand_Valid(t *testing.T) {
	dir := t.TempDir()
	sourcePath := saveNamedProject(t, dir, "Original")

	result, err := RunRenameProjectCommand(RenameProjectCommandInput{
		SourcePath: sourcePath,
		NewName:    "Renombrado",
		OutDir:     dir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected OK, got errors: %v", result.Errors)
	}

	wantPath := filepath.Join(dir, "renombrado.json")
	if result.Path != wantPath {
		t.Fatalf("expected path %q, got %q", wantPath, result.Path)
	}

	renamed, loadErr := storage.LoadProject(result.Path)
	if loadErr != nil {
		t.Fatalf("unexpected error loading renamed project: %v", loadErr)
	}
	if renamed.Name != "Renombrado" {
		t.Fatalf("expected name updated, got %q", renamed.Name)
	}
}

func TestRunRenameProjectCommand_Collision(t *testing.T) {
	dir := t.TempDir()
	sourcePath := saveNamedProject(t, dir, "Original")
	saveNamedProject(t, dir, "Ya Existente")

	result, err := RunRenameProjectCommand(RenameProjectCommandInput{
		SourcePath: sourcePath,
		NewName:    "Ya Existente",
		OutDir:     dir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OK {
		t.Fatalf("expected not OK, got %+v", result)
	}
	if !containsError(result.Errors, "a project already exists at that name") {
		t.Fatalf("expected collision error, got %v", result.Errors)
	}
}

func TestRunRenameProjectCommand_MissingSource(t *testing.T) {
	dir := t.TempDir()

	_, err := RunRenameProjectCommand(RenameProjectCommandInput{
		SourcePath: filepath.Join(dir, "does-not-exist.json"),
		NewName:    "Nuevo",
		OutDir:     dir,
	})

	if err == nil {
		t.Fatalf("expected an error for a missing source file")
	}
}
