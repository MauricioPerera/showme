package cli

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/MauricioPerera/showme/internal/domain"
	"github.com/MauricioPerera/showme/internal/storage"
)

func saveProjectWithIntentAndKnowledge(t *testing.T, dir string) (projectPath string) {
	t.Helper()
	deck, deckReport := domain.NewDeck(domain.DeckInput{
		Title: "Roadmap Q3",
		Slides: []domain.Slide{
			{ID: "intro", Title: "Introduccion", Intent: "Dar la bienvenida y presentar el objetivo", Status: domain.SlideStatusAccepted},
		},
	})
	if len(deckReport.Errors) != 0 {
		t.Fatalf("setup: unexpected deck errors: %v", deckReport.Errors)
	}

	knowledgeRoot := filepath.Join(dir, "knowledge")
	if err := os.Mkdir(knowledgeRoot, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	writeFile(t, filepath.Join(knowledgeRoot, "brand.md"), "---\ntype: Brand\ntitle: Bienvenida\n---\n\nContenido de marca sobre la bienvenida.\n")

	path, report, err := storage.SaveProject(storage.SaveProjectRequest{
		Dir: dir,
		Input: domain.ProjectInput{
			Name:          "Roadmap Q3",
			Deck:          deck,
			DesignPath:    "DESIGN.md",
			KnowledgePath: knowledgeRoot,
		},
	})
	if err != nil || len(report.Errors) != 0 {
		t.Fatalf("setup: unexpected SaveProject failure: err=%v report=%v", err, report)
	}
	return path
}

func fakeChatServer(t *testing.T, content string, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"` + content + `"}}]}`))
	}))
}

func TestRunGenerateSlideContentCommand_Valid(t *testing.T) {
	dir := t.TempDir()
	path := saveProjectWithIntentAndKnowledge(t, dir)
	server := fakeChatServer(t, "Contenido generado.", http.StatusOK)
	defer server.Close()

	result, err := RunGenerateSlideContentCommand(GenerateSlideContentCommandInput{
		Path:    path,
		SlideID: "intro",
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
	if result.Content != "Contenido generado." {
		t.Fatalf("expected generated content, got %q", result.Content)
	}

	updated, loadErr := storage.LoadProject(path)
	if loadErr != nil {
		t.Fatalf("unexpected error reloading project: %v", loadErr)
	}
	if updated.Deck.Slides[0].Content != "Contenido generado." {
		t.Fatalf("expected slide content updated, got %+v", updated.Deck.Slides[0])
	}
	if updated.Deck.Slides[0].Status != domain.SlideStatusAccepted {
		t.Fatalf("expected status preserved, got %q", updated.Deck.Slides[0].Status)
	}
}

func TestRunGenerateSlideContentCommand_SlideNotFound(t *testing.T) {
	dir := t.TempDir()
	path := saveProjectWithIntentAndKnowledge(t, dir)
	server := fakeChatServer(t, "no deberia usarse", http.StatusOK)
	defer server.Close()

	result, err := RunGenerateSlideContentCommand(GenerateSlideContentCommandInput{
		Path:    path,
		SlideID: "missing",
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
	if !containsError(result.Errors, "slide not found: missing") {
		t.Fatalf("expected 'slide not found: missing' error, got %v", result.Errors)
	}
}

func TestRunGenerateSlideContentCommand_ProviderError(t *testing.T) {
	dir := t.TempDir()
	path := saveProjectWithIntentAndKnowledge(t, dir)
	server := fakeChatServer(t, "", http.StatusInternalServerError)
	defer server.Close()

	result, err := RunGenerateSlideContentCommand(GenerateSlideContentCommandInput{
		Path:    path,
		SlideID: "intro",
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
	if len(result.Errors) == 0 {
		t.Fatalf("expected an error from the provider, got none")
	}

	unchanged, loadErr := storage.LoadProject(path)
	if loadErr != nil {
		t.Fatalf("unexpected error reloading project: %v", loadErr)
	}
	if unchanged.Deck.Slides[0].Content != "" {
		t.Fatalf("expected slide content untouched, got %+v", unchanged.Deck.Slides[0])
	}
}

func TestRunGenerateSlideContentCommand_MissingSource(t *testing.T) {
	dir := t.TempDir()
	server := fakeChatServer(t, "x", http.StatusOK)
	defer server.Close()

	_, err := RunGenerateSlideContentCommand(GenerateSlideContentCommandInput{
		Path:    filepath.Join(dir, "does-not-exist.json"),
		SlideID: "intro",
		BaseURL: server.URL,
		Model:   "test-model",
		OutDir:  dir,
	})

	if err == nil {
		t.Fatalf("expected an error for a missing source file")
	}
}
