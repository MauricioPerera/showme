package cli

import (
	"path/filepath"
	"testing"

	"github.com/MauricioPerera/showme/internal/domain"
	"github.com/MauricioPerera/showme/internal/storage"
)

func saveProjectWithTwoSlides(t *testing.T, dir, name string) string {
	t.Helper()
	deck, deckReport := domain.NewDeck(domain.DeckInput{
		Title: "Roadmap",
		Slides: []domain.Slide{
			{ID: "intro", Title: "Introduccion"},
			{ID: "plan", Title: "Plan"},
		},
	})
	if len(deckReport.Errors) != 0 {
		t.Fatalf("setup: unexpected deck errors: %v", deckReport.Errors)
	}

	path, report, err := storage.SaveProject(storage.SaveProjectRequest{
		Dir: dir,
		Input: domain.ProjectInput{
			Name:          name,
			Deck:          deck,
			DesignPath:    "DESIGN.md",
			KnowledgePath: "knowledge/showme",
		},
	})
	if err != nil || len(report.Errors) != 0 {
		t.Fatalf("setup: unexpected SaveProject failure: err=%v report=%v", err, report)
	}
	return path
}

func TestRunReorderSlidesCommand_Valid(t *testing.T) {
	dir := t.TempDir()
	path := saveProjectWithTwoSlides(t, dir, "Roadmap Q3")

	result, err := RunReorderSlidesCommand(ReorderSlidesCommandInput{
		Path:   path,
		Order:  []string{"plan", "intro"},
		OutDir: dir,
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
	if updated.Deck.Slides[0].ID != "plan" || updated.Deck.Slides[1].ID != "intro" {
		t.Fatalf("expected reordered slides, got %+v", updated.Deck.Slides)
	}
}

func TestRunReorderSlidesCommand_MissingSlideID(t *testing.T) {
	dir := t.TempDir()
	path := saveProjectWithTwoSlides(t, dir, "Roadmap Q3")

	result, err := RunReorderSlidesCommand(ReorderSlidesCommandInput{
		Path:   path,
		Order:  []string{"intro"},
		OutDir: dir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OK {
		t.Fatalf("expected not OK, got %+v", result)
	}
	if !containsError(result.Errors, "missing slide id in order: plan") {
		t.Fatalf("expected 'missing slide id in order: plan' error, got %v", result.Errors)
	}
}

func TestRunReorderSlidesCommand_MissingSource(t *testing.T) {
	dir := t.TempDir()

	_, err := RunReorderSlidesCommand(ReorderSlidesCommandInput{
		Path:   filepath.Join(dir, "does-not-exist.json"),
		Order:  []string{"intro", "plan"},
		OutDir: dir,
	})

	if err == nil {
		t.Fatalf("expected an error for a missing source file")
	}
}
