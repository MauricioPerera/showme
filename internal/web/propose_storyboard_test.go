package web

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

func TestHandleProposeStoryboard_ValidWithoutKnowledge(t *testing.T) {
	server := fakeStoryboardServer(t, `[{"title":"Introduccion","intent":"Dar la bienvenida"},{"title":"Plan","intent":"Explicar los proximos pasos"}]`, http.StatusOK)
	defer server.Close()

	result, err := HandleProposeStoryboard(ProposeStoryboardInput{
		Objective: "Presentar el roadmap",
		Audience:  "Equipo",
		BaseURL:   server.URL,
		Model:     "test-model",
		Count:     2,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", result.Errors)
	}
	if len(result.Slides) != 2 || result.Slides[0].Title != "Introduccion" || result.Slides[1].Title != "Plan" {
		t.Fatalf("expected 2 slides in order, got %+v", result.Slides)
	}
}

func TestHandleProposeStoryboard_WithKnowledgeContext(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"[{\"title\":\"Intro\",\"intent\":\"a\"}]"}}]}`))
	}))
	defer server.Close()

	dir := t.TempDir()
	knowledgeRoot := filepath.Join(dir, "knowledge")
	if err := os.Mkdir(knowledgeRoot, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	writeFile(t, filepath.Join(knowledgeRoot, "brand.md"), "---\ntype: Brand\ntitle: Bienvenida\n---\n\nContenido de marca sobre la bienvenida.\n")

	result, err := HandleProposeStoryboard(ProposeStoryboardInput{
		Objective:     "Presentar la marca",
		KnowledgeRoot: knowledgeRoot,
		BaseURL:       server.URL,
		Model:         "test-model",
		Count:         1,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", result.Errors)
	}
	messages, _ := gotBody["messages"].([]any)
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %v", gotBody["messages"])
	}
}

func TestHandleProposeStoryboard_ProviderError(t *testing.T) {
	server := fakeStoryboardServer(t, "", http.StatusInternalServerError)
	defer server.Close()

	result, err := HandleProposeStoryboard(ProposeStoryboardInput{
		Objective: "x",
		BaseURL:   server.URL,
		Model:     "test-model",
		Count:     3,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Errors) == 0 {
		t.Fatalf("expected an error from the provider, got none")
	}
	if len(result.Slides) != 0 {
		t.Fatalf("expected no slides, got %+v", result.Slides)
	}
}

func TestHandleProposeStoryboard_MissingKnowledgeRoot(t *testing.T) {
	server := fakeStoryboardServer(t, `[{"title":"x","intent":"y"}]`, http.StatusOK)
	defer server.Close()

	_, err := HandleProposeStoryboard(ProposeStoryboardInput{
		Objective:     "x",
		KnowledgeRoot: filepath.Join(t.TempDir(), "does-not-exist"),
		BaseURL:       server.URL,
		Model:         "test-model",
		Count:         1,
	})

	if err != nil {
		t.Fatalf("expected knowledge.Load to tolerate a missing dir gracefully, got error: %v", err)
	}
}
