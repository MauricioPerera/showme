package web

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MauricioPerera/showme/internal/storage"
)

const webValidDesign = `---
name: "Test Brand"
colors:
  primary: "#111827"
---

## Overview

Brand overview.
`

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
}

func setupWebDirs(t *testing.T) (designPath, knowledgeRoot, dataDir string) {
	t.Helper()
	root := t.TempDir()
	designPath = filepath.Join(root, "DESIGN.md")
	writeFile(t, designPath, webValidDesign)

	knowledgeRoot = filepath.Join(root, "knowledge")
	if err := os.Mkdir(knowledgeRoot, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	writeFile(t, filepath.Join(knowledgeRoot, "brand.md"), "---\ntype: Brand\n---\n\nBody.\n")

	dataDir = filepath.Join(root, "data")
	if err := os.Mkdir(dataDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	return designPath, knowledgeRoot, dataDir
}

func TestHandleCreateProjectForm_Valid(t *testing.T) {
	designPath, knowledgeRoot, dataDir := setupWebDirs(t)

	result, err := HandleCreateProjectForm(CreateProjectFormInput{
		Name:          "Presentacion Q3",
		DesignPath:    designPath,
		KnowledgeRoot: knowledgeRoot,
		DeckTitle:     "Roadmap Q3",
		DeckAudience:  "Equipo de producto",
		SlideTitle:    "Introduccion",
		SlideIntent:   "Dar la bienvenida",
		Dir:           dataDir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected OK, got errors: %v", result.Errors)
	}

	proj, loadErr := storage.LoadProject(result.Path)
	if loadErr != nil {
		t.Fatalf("expected saved project to load, got error: %v", loadErr)
	}
	if proj.Name != "Presentacion Q3" || len(proj.Deck.Slides) != 1 {
		t.Fatalf("expected project persisted, got %+v", proj)
	}
	if proj.Deck.Slides[0].Title != "Introduccion" || proj.Deck.Slides[0].Intent != "Dar la bienvenida" {
		t.Fatalf("expected slide fields persisted, got %+v", proj.Deck.Slides[0])
	}
}

func TestHandleCreateProjectForm_MissingDesignFile(t *testing.T) {
	_, knowledgeRoot, dataDir := setupWebDirs(t)

	_, err := HandleCreateProjectForm(CreateProjectFormInput{
		Name:          "Presentacion Q3",
		DesignPath:    filepath.Join(dataDir, "does-not-exist.md"),
		KnowledgeRoot: knowledgeRoot,
		DeckTitle:     "Roadmap Q3",
		SlideTitle:    "Introduccion",
		SlideIntent:   "Dar la bienvenida",
		Dir:           dataDir,
	})

	if err == nil {
		t.Fatalf("expected an error for a missing design file")
	}
}

func TestHandleCreateProjectForm_ValidationErrorsDoNotSave(t *testing.T) {
	designPath, knowledgeRoot, dataDir := setupWebDirs(t)
	writeFile(t, designPath, "no frontmatter here")

	result, err := HandleCreateProjectForm(CreateProjectFormInput{
		Name:          "Presentacion Q3",
		DesignPath:    designPath,
		KnowledgeRoot: knowledgeRoot,
		DeckTitle:     "Roadmap Q3",
		SlideTitle:    "Introduccion",
		SlideIntent:   "Dar la bienvenida",
		Dir:           dataDir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OK {
		t.Fatalf("expected not OK, got %+v", result)
	}
	if !containsError(result.Errors, "frontmatter is required") {
		t.Fatalf("expected 'frontmatter is required' error, got %v", result.Errors)
	}

	entries, readErr := os.ReadDir(dataDir)
	if readErr != nil {
		t.Fatalf("unexpected error reading dataDir: %v", readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no files written on validation error, found %v", entries)
	}
}

func TestHandleCreateProjectForm_MissingDataDir(t *testing.T) {
	designPath, knowledgeRoot, dataDir := setupWebDirs(t)

	_, err := HandleCreateProjectForm(CreateProjectFormInput{
		Name:          "Presentacion Q3",
		DesignPath:    designPath,
		KnowledgeRoot: knowledgeRoot,
		DeckTitle:     "Roadmap Q3",
		SlideTitle:    "Introduccion",
		SlideIntent:   "Dar la bienvenida",
		Dir:           filepath.Join(dataDir, "no-such-subdir"),
	})

	if err == nil {
		t.Fatalf("expected an error for a missing data directory")
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
