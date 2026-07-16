package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MauricioPerera/showme/internal/domain"
)

func TestLoadProject_RoundTripsWhatSaveProjectWrote(t *testing.T) {
	dir := t.TempDir()
	input := domain.ProjectInput{
		Name:          "Presentacion Q3",
		Deck:          validProjectDeck(t),
		DesignPath:    "DESIGN.md",
		KnowledgePath: "knowledge/showme",
		Version:       2,
	}

	path, report, err := SaveProject(SaveProjectRequest{Dir: dir, Input: input})
	if err != nil || len(report.Errors) != 0 {
		t.Fatalf("setup: unexpected SaveProject failure: err=%v report=%v", err, report)
	}

	proj, loadErr := LoadProject(path)

	if loadErr != nil {
		t.Fatalf("unexpected error: %v", loadErr)
	}
	if proj.Name != "Presentacion Q3" || proj.DesignPath != "DESIGN.md" || proj.KnowledgePath != "knowledge/showme" {
		t.Fatalf("expected fields round-trip, got %+v", proj)
	}
	if proj.Version != 2 {
		t.Fatalf("expected version round-trip, got %d", proj.Version)
	}
	if len(proj.Deck.Slides) != 1 || proj.Deck.Slides[0].ID != "intro" {
		t.Fatalf("expected deck round-trip, got %+v", proj.Deck)
	}
}

func TestLoadProject_MissingFileReturnsError(t *testing.T) {
	dir := t.TempDir()

	_, err := LoadProject(filepath.Join(dir, "does-not-exist.json"))

	if err == nil {
		t.Fatalf("expected an error for a missing file")
	}
}

func TestLoadProject_InvalidJSONReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "broken.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, err := LoadProject(path)

	if err == nil {
		t.Fatalf("expected an error for invalid JSON content")
	}
}
