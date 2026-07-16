package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func fakeStoryboardServer(t *testing.T, rawSlidesJSON string, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		encoded, err := json.Marshal(rawSlidesJSON)
		if err != nil {
			t.Fatalf("setup: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":` + string(encoded) + `}}]}`))
	}))
}

func TestRunGenerateStoryboardCommand_Valid(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "deck.json")
	server := fakeStoryboardServer(t, `[{"title":"Introduccion","intent":"Dar la bienvenida"},{"title":"Plan","intent":"Explicar los proximos pasos"}]`, http.StatusOK)
	defer server.Close()

	result, err := RunGenerateStoryboardCommand(GenerateStoryboardCommandInput{
		Objective: "Presentar el roadmap",
		Audience:  "Equipo",
		DeckTitle: "Roadmap Q3",
		BaseURL:   server.URL,
		Model:     "test-model",
		Count:     2,
		OutPath:   outPath,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected OK, got errors: %v", result.Errors)
	}
	if result.Path != outPath {
		t.Fatalf("expected path %q, got %q", outPath, result.Path)
	}

	data, readErr := os.ReadFile(outPath)
	if readErr != nil {
		t.Fatalf("expected deck file to exist: %v", readErr)
	}
	deckInput, parseErr := parseDeckInput(data)
	if parseErr != nil {
		t.Fatalf("expected the written file to be a valid deck input: %v", parseErr)
	}
	if deckInput.Title != "Roadmap Q3" || len(deckInput.Slides) != 2 {
		t.Fatalf("expected 2 slides under the given deck title, got %+v", deckInput)
	}
	if deckInput.Slides[0].ID == "" || deckInput.Slides[1].ID == "" || deckInput.Slides[0].ID == deckInput.Slides[1].ID {
		t.Fatalf("expected unique, non-empty slide ids, got %+v", deckInput.Slides)
	}
}

func TestRunGenerateStoryboardCommand_DedupesCollidingSlugs(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "deck.json")
	server := fakeStoryboardServer(t, `[{"title":"Introduccion","intent":"a"},{"title":"Introduccion","intent":"b"}]`, http.StatusOK)
	defer server.Close()

	result, err := RunGenerateStoryboardCommand(GenerateStoryboardCommandInput{
		Objective: "x",
		DeckTitle: "Deck",
		BaseURL:   server.URL,
		Model:     "test-model",
		Count:     2,
		OutPath:   outPath,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected OK, got errors: %v", result.Errors)
	}

	data, _ := os.ReadFile(outPath)
	deckInput, _ := parseDeckInput(data)
	if deckInput.Slides[0].ID == deckInput.Slides[1].ID {
		t.Fatalf("expected deduped ids, got %+v", deckInput.Slides)
	}
}

func TestRunGenerateStoryboardCommand_ProviderError(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "deck.json")
	server := fakeStoryboardServer(t, "", http.StatusInternalServerError)
	defer server.Close()

	result, err := RunGenerateStoryboardCommand(GenerateStoryboardCommandInput{
		Objective: "x",
		DeckTitle: "Deck",
		BaseURL:   server.URL,
		Model:     "test-model",
		Count:     3,
		OutPath:   outPath,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OK {
		t.Fatalf("expected not OK, got %+v", result)
	}
	if _, statErr := os.Stat(outPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected no file written on provider error")
	}
}

func TestRunGenerateStoryboardCommand_MissingOutDir(t *testing.T) {
	dir := t.TempDir()
	server := fakeStoryboardServer(t, `[{"title":"Intro","intent":"a"}]`, http.StatusOK)
	defer server.Close()

	_, err := RunGenerateStoryboardCommand(GenerateStoryboardCommandInput{
		Objective: "x",
		DeckTitle: "Deck",
		BaseURL:   server.URL,
		Model:     "test-model",
		Count:     1,
		OutPath:   filepath.Join(dir, "no-such-subdir", "deck.json"),
	})

	if err == nil {
		t.Fatalf("expected an error for a missing output directory")
	}
}
