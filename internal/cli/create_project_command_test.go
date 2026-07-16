package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MauricioPerera/showme/internal/storage"
)

const commandValidDesign = `---
name: "Test Brand"
colors:
  primary: "#111827"
---

## Overview

Brand overview.
`

const commandValidDeckJSON = `{
  "title": "Roadmap Q3",
  "audience": "Equipo de producto",
  "slides": [
    {"id": "intro", "title": "Introduccion"}
  ]
}`

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
}

func setupCommandDirs(t *testing.T) (designPath, knowledgeRoot, deckPath, outDir string) {
	t.Helper()
	root := t.TempDir()
	designPath = filepath.Join(root, "DESIGN.md")
	writeFile(t, designPath, commandValidDesign)

	knowledgeRoot = filepath.Join(root, "knowledge")
	if err := os.Mkdir(knowledgeRoot, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	writeFile(t, filepath.Join(knowledgeRoot, "brand.md"), "---\ntype: Brand\n---\n\nBody.\n")

	deckPath = filepath.Join(root, "deck.json")
	writeFile(t, deckPath, commandValidDeckJSON)

	outDir = filepath.Join(root, "out")
	if err := os.Mkdir(outDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	return designPath, knowledgeRoot, deckPath, outDir
}

func TestRunCreateProjectCommand_Valid(t *testing.T) {
	designPath, knowledgeRoot, deckPath, outDir := setupCommandDirs(t)

	result, err := RunCreateProjectCommand(CreateProjectCommandInput{
		Name:          "Presentacion Q3",
		DesignPath:    designPath,
		KnowledgeRoot: knowledgeRoot,
		DeckPath:      deckPath,
		OutDir:        outDir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected OK, got errors: %v", result.Errors)
	}

	wantPath := filepath.Join(outDir, "presentacion-q3.json")
	if result.Path != wantPath {
		t.Fatalf("expected path %q, got %q", wantPath, result.Path)
	}

	proj, loadErr := storage.LoadProject(result.Path)
	if loadErr != nil {
		t.Fatalf("expected saved project to load, got error: %v", loadErr)
	}
	if proj.Name != "Presentacion Q3" || len(proj.Deck.Slides) != 1 {
		t.Fatalf("expected project persisted, got %+v", proj)
	}
}

func TestRunCreateProjectCommand_MissingDesignFile(t *testing.T) {
	_, knowledgeRoot, deckPath, outDir := setupCommandDirs(t)

	_, err := RunCreateProjectCommand(CreateProjectCommandInput{
		Name:          "Presentacion Q3",
		DesignPath:    filepath.Join(outDir, "does-not-exist.md"),
		KnowledgeRoot: knowledgeRoot,
		DeckPath:      deckPath,
		OutDir:        outDir,
	})

	if err == nil {
		t.Fatalf("expected an error for a missing design file")
	}
}

func TestRunCreateProjectCommand_InvalidDeckJSON(t *testing.T) {
	designPath, knowledgeRoot, _, outDir := setupCommandDirs(t)
	badDeckPath := filepath.Join(outDir, "bad-deck.json")
	writeFile(t, badDeckPath, "{not json")

	_, err := RunCreateProjectCommand(CreateProjectCommandInput{
		Name:          "Presentacion Q3",
		DesignPath:    designPath,
		KnowledgeRoot: knowledgeRoot,
		DeckPath:      badDeckPath,
		OutDir:        outDir,
	})

	if err == nil {
		t.Fatalf("expected an error for invalid deck JSON")
	}
}

func TestRunCreateProjectCommand_ValidationErrorsDoNotSave(t *testing.T) {
	designPath, knowledgeRoot, deckPath, outDir := setupCommandDirs(t)
	writeFile(t, designPath, "no frontmatter here")

	result, err := RunCreateProjectCommand(CreateProjectCommandInput{
		Name:          "Presentacion Q3",
		DesignPath:    designPath,
		KnowledgeRoot: knowledgeRoot,
		DeckPath:      deckPath,
		OutDir:        outDir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OK {
		t.Fatalf("expected not OK, got %+v", result)
	}
	if result.Path != "" {
		t.Fatalf("expected empty path, got %q", result.Path)
	}
	if !containsError(result.Errors, "frontmatter is required") {
		t.Fatalf("expected 'frontmatter is required' error, got %v", result.Errors)
	}

	entries, readErr := os.ReadDir(outDir)
	if readErr != nil {
		t.Fatalf("unexpected error reading outDir: %v", readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no files written on validation error, found %v", entries)
	}
}

func TestRunCreateProjectCommand_NameProducesEmptySlug(t *testing.T) {
	designPath, knowledgeRoot, deckPath, outDir := setupCommandDirs(t)

	result, err := RunCreateProjectCommand(CreateProjectCommandInput{
		Name:          "!!!",
		DesignPath:    designPath,
		KnowledgeRoot: knowledgeRoot,
		DeckPath:      deckPath,
		OutDir:        outDir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OK {
		t.Fatalf("expected not OK, got %+v", result)
	}
	if !containsError(result.Errors, "name produces an empty slug") {
		t.Fatalf("expected 'name produces an empty slug' error, got %v", result.Errors)
	}
}

func containsError(errors []string, want string) bool {
	for _, e := range errors {
		if e == want {
			return true
		}
	}
	return false
}
