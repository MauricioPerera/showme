package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/MauricioPerera/showme/internal/domain"
)

func validProjectDeck(t *testing.T) domain.Deck {
	t.Helper()
	deck, report := domain.NewDeck(domain.DeckInput{
		Title:  "Roadmap Q3",
		Slides: []domain.Slide{{ID: "intro", Title: "Introduccion"}},
	})
	if len(report.Errors) != 0 {
		t.Fatalf("setup: unexpected deck errors: %v", report.Errors)
	}
	return deck
}

func TestSaveProject_ValidWritesJSONFile(t *testing.T) {
	dir := t.TempDir()
	input := domain.ProjectInput{
		Name:          "Presentacion Q3",
		Deck:          validProjectDeck(t),
		DesignPath:    "DESIGN.md",
		KnowledgePath: "knowledge/showme",
	}

	path, report, err := SaveProject(SaveProjectRequest{Dir: dir, Input: input})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}

	wantPath := filepath.Join(dir, "presentacion-q3.json")
	if path != wantPath {
		t.Fatalf("expected path %q, got %q", wantPath, path)
	}

	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("expected file to exist at %q: %v", path, readErr)
	}

	var proj domain.Project
	if unmarshalErr := json.Unmarshal(data, &proj); unmarshalErr != nil {
		t.Fatalf("expected valid JSON, got error: %v", unmarshalErr)
	}
	if proj.Name != "Presentacion Q3" || proj.DesignPath != "DESIGN.md" || proj.KnowledgePath != "knowledge/showme" {
		t.Fatalf("expected fields round-trip, got %+v", proj)
	}
	if proj.Version != 1 || len(proj.Deck.Slides) != 1 {
		t.Fatalf("expected version/deck round-trip, got %+v", proj)
	}
}

func TestSaveProject_InvalidProjectWritesNothing(t *testing.T) {
	dir := t.TempDir()
	input := domain.ProjectInput{Name: ""}

	path, report, err := SaveProject(SaveProjectRequest{Dir: dir, Input: input})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "" {
		t.Fatalf("expected empty path when project is invalid, got %q", path)
	}
	if len(report.Errors) == 0 {
		t.Fatalf("expected validation errors from an empty name")
	}

	entries, readErr := os.ReadDir(dir)
	if readErr != nil {
		t.Fatalf("unexpected error reading dir: %v", readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no files written for an invalid project, found %v", entries)
	}
}

func TestSaveProject_NameProducesEmptySlug(t *testing.T) {
	dir := t.TempDir()
	input := domain.ProjectInput{
		Name:          "!!!",
		Deck:          validProjectDeck(t),
		DesignPath:    "DESIGN.md",
		KnowledgePath: "knowledge/showme",
	}

	path, report, err := SaveProject(SaveProjectRequest{Dir: dir, Input: input})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "" {
		t.Fatalf("expected empty path, got %q", path)
	}
	if !containsError(report.Errors, "name produces an empty slug") {
		t.Fatalf("expected 'name produces an empty slug' error, got %v", report.Errors)
	}
}

func TestSaveProject_MissingDirReturnsError(t *testing.T) {
	input := domain.ProjectInput{
		Name:          "Presentacion Q3",
		Deck:          validProjectDeck(t),
		DesignPath:    "DESIGN.md",
		KnowledgePath: "knowledge/showme",
	}

	_, report, err := SaveProject(SaveProjectRequest{Dir: "/no/such/dir/for-showme-tests", Input: input})

	if err == nil {
		t.Fatalf("expected an I/O error for a missing directory")
	}
	if len(report.Errors) != 0 {
		t.Fatalf("expected no validation errors, got %v", report.Errors)
	}
}
