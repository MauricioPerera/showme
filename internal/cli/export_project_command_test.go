package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunExportProjectCommand_Valid(t *testing.T) {
	dir := t.TempDir()
	path := saveNamedProject(t, dir, "Roadmap Q3")
	outPath := filepath.Join(dir, "roadmap-q3.html")

	result, err := RunExportProjectCommand(ExportProjectCommandInput{
		Path:    path,
		OutPath: outPath,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected OK, got %+v", result)
	}
	if result.Path != outPath {
		t.Fatalf("expected path %q, got %q", outPath, result.Path)
	}

	data, readErr := os.ReadFile(outPath)
	if readErr != nil {
		t.Fatalf("expected exported file to exist: %v", readErr)
	}
	if !strings.HasPrefix(string(data), "<!doctype html>") {
		t.Fatalf("expected exported file to be HTML, got: %s", data[:40])
	}
	if !strings.Contains(string(data), "Introduccion") {
		t.Fatalf("expected slide content in export, got: %s", data)
	}
}

func TestRunExportProjectCommand_MissingSource(t *testing.T) {
	dir := t.TempDir()

	_, err := RunExportProjectCommand(ExportProjectCommandInput{
		Path:    filepath.Join(dir, "does-not-exist.json"),
		OutPath: filepath.Join(dir, "out.html"),
	})

	if err == nil {
		t.Fatalf("expected an error for a missing source file")
	}
}

func TestRunExportProjectCommand_MissingOutDir(t *testing.T) {
	dir := t.TempDir()
	path := saveNamedProject(t, dir, "Roadmap Q3")

	_, err := RunExportProjectCommand(ExportProjectCommandInput{
		Path:    path,
		OutPath: filepath.Join(dir, "no-such-subdir", "out.html"),
	})

	if err == nil {
		t.Fatalf("expected an error for a missing output directory")
	}
}
