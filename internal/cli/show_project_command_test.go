package cli

import (
	"path/filepath"
	"testing"
)

func TestRunShowProjectCommand_Valid(t *testing.T) {
	dir := t.TempDir()
	path := saveNamedProject(t, dir, "Roadmap Q3")

	result, err := RunShowProjectCommand(ShowProjectCommandInput{Path: path})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Project.Name != "Roadmap Q3" {
		t.Fatalf("expected name 'Roadmap Q3', got %q", result.Project.Name)
	}
	if len(result.Project.Deck.Slides) != 1 || result.Project.Deck.Slides[0].ID != "intro" {
		t.Fatalf("expected deck preserved, got %+v", result.Project.Deck)
	}
}

func TestRunShowProjectCommand_MissingFile(t *testing.T) {
	dir := t.TempDir()

	_, err := RunShowProjectCommand(ShowProjectCommandInput{Path: filepath.Join(dir, "does-not-exist.json")})

	if err == nil {
		t.Fatalf("expected an error for a missing file")
	}
}

func TestRunShowProjectCommand_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "broken.json")
	writeFile(t, path, "{not json")

	_, err := RunShowProjectCommand(ShowProjectCommandInput{Path: path})

	if err == nil {
		t.Fatalf("expected an error for invalid JSON content")
	}
}
