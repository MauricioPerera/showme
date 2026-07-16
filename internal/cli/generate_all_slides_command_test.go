package cli

import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/MauricioPerera/showme/internal/domain"
	"github.com/MauricioPerera/showme/internal/storage"
)

func saveProjectWithTwoIntentSlides(t *testing.T, dir string, secondHasContent bool) string {
	t.Helper()
	second := domain.Slide{ID: "plan", Title: "Plan", Intent: "Explicar el plan"}
	if secondHasContent {
		second.Content = "Ya tiene contenido."
	}
	deck, deckReport := domain.NewDeck(domain.DeckInput{
		Title: "Roadmap",
		Slides: []domain.Slide{
			{ID: "intro", Title: "Introduccion", Intent: "Dar la bienvenida"},
			second,
		},
	})
	if len(deckReport.Errors) != 0 {
		t.Fatalf("setup: unexpected deck errors: %v", deckReport.Errors)
	}

	path, report, err := storage.SaveProject(storage.SaveProjectRequest{
		Dir: dir,
		Input: domain.ProjectInput{
			Name:          "Roadmap",
			Deck:          deck,
			DesignPath:    "DESIGN.md",
			KnowledgePath: dir,
		},
	})
	if err != nil || len(report.Errors) != 0 {
		t.Fatalf("setup: unexpected SaveProject failure: err=%v report=%v", err, report)
	}
	return path
}

func TestRunGenerateAllSlidesCommand_GeneratesOnlyEmptySlides(t *testing.T) {
	dir := t.TempDir()
	path := saveProjectWithTwoIntentSlides(t, dir, true)
	server := fakeChatServer(t, "Contenido generado.", http.StatusOK)
	defer server.Close()

	result, err := RunGenerateAllSlidesCommand(GenerateAllSlidesCommandInput{
		Path:    path,
		BaseURL: server.URL,
		Model:   "test-model",
		OutDir:  dir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected OK, got errors: %v", result.Errors)
	}
	if len(result.Generated) != 1 || result.Generated[0] != "intro" {
		t.Fatalf("expected only 'intro' generated, got %+v", result.Generated)
	}
	if len(result.Skipped) != 1 || result.Skipped[0] != "plan" {
		t.Fatalf("expected 'plan' skipped, got %+v", result.Skipped)
	}

	updated, loadErr := storage.LoadProject(path)
	if loadErr != nil {
		t.Fatalf("unexpected error reloading project: %v", loadErr)
	}
	if updated.Deck.Slides[0].Content != "Contenido generado." {
		t.Fatalf("expected intro content generated, got %+v", updated.Deck.Slides[0])
	}
	if updated.Deck.Slides[1].Content != "Ya tiene contenido." {
		t.Fatalf("expected plan content untouched, got %+v", updated.Deck.Slides[1])
	}
}

func TestRunGenerateAllSlidesCommand_AllAlreadyHaveContent(t *testing.T) {
	dir := t.TempDir()
	path := saveProjectWithTwoIntentSlides(t, dir, true)
	server := fakeChatServer(t, "no deberia usarse", http.StatusOK)
	defer server.Close()

	proj, loadErr := storage.LoadProject(path)
	if loadErr != nil {
		t.Fatalf("setup: %v", loadErr)
	}
	proj.Deck.Slides[0].Content = "Ya tiene contenido tambien."
	if _, report, err := storage.SaveProject(storage.SaveProjectRequest{
		Dir: dir,
		Input: domain.ProjectInput{
			Name: proj.Name, Deck: proj.Deck, DesignPath: proj.DesignPath, KnowledgePath: proj.KnowledgePath,
		},
	}); err != nil || len(report.Errors) != 0 {
		t.Fatalf("setup: unexpected SaveProject failure: err=%v report=%v", err, report)
	}

	result, err := RunGenerateAllSlidesCommand(GenerateAllSlidesCommandInput{
		Path:    path,
		BaseURL: server.URL,
		Model:   "test-model",
		OutDir:  dir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected OK, got errors: %v", result.Errors)
	}
	if len(result.Generated) != 0 {
		t.Fatalf("expected nothing generated, got %+v", result.Generated)
	}
	if len(result.Skipped) != 2 {
		t.Fatalf("expected both slides skipped, got %+v", result.Skipped)
	}
}

func TestRunGenerateAllSlidesCommand_ProviderErrorContinuesToNextSlide(t *testing.T) {
	dir := t.TempDir()
	path := saveProjectWithTwoIntentSlides(t, dir, false)
	server := fakeChatServer(t, "", http.StatusInternalServerError)
	defer server.Close()

	result, err := RunGenerateAllSlidesCommand(GenerateAllSlidesCommandInput{
		Path:    path,
		BaseURL: server.URL,
		Model:   "test-model",
		OutDir:  dir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OK {
		t.Fatalf("expected not OK, got %+v", result)
	}
	if len(result.Errors) != 2 {
		t.Fatalf("expected both slides to fail, got %+v", result.Errors)
	}
}

func TestRunGenerateAllSlidesCommand_MissingSource(t *testing.T) {
	dir := t.TempDir()
	server := fakeChatServer(t, "x", http.StatusOK)
	defer server.Close()

	_, err := RunGenerateAllSlidesCommand(GenerateAllSlidesCommandInput{
		Path:    filepath.Join(dir, "does-not-exist.json"),
		BaseURL: server.URL,
		Model:   "test-model",
		OutDir:  dir,
	})

	if err == nil {
		t.Fatalf("expected an error for a missing source file")
	}
}
