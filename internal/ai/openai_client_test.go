package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAIClient_GenerateContent_SendsExpectedRequest(t *testing.T) {
	var gotPath string
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if decodeErr := json.NewDecoder(r.Body).Decode(&gotBody); decodeErr != nil {
			t.Fatalf("failed to decode request body: %v", decodeErr)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"finish_reason":"stop","index":0,"message":{"role":"assistant","content":"Contenido generado."}}]}`))
	}))
	defer server.Close()

	client := NewOpenAIClient(server.URL, "Ternary-Bonsai-27B-Q2_0.gguf")

	content, err := client.GenerateContent(GenerateContentRequest{
		Intent:  "Explicar el roadmap",
		Context: "El roadmap tiene 3 fases.",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "Contenido generado." {
		t.Fatalf("expected generated content, got %q", content)
	}
	if gotPath != "/chat/completions" {
		t.Fatalf("expected path '/chat/completions', got %q", gotPath)
	}
	if gotBody["model"] != "Ternary-Bonsai-27B-Q2_0.gguf" {
		t.Fatalf("expected model field set, got %v", gotBody["model"])
	}
	messages, ok := gotBody["messages"].([]any)
	if !ok || len(messages) == 0 {
		t.Fatalf("expected non-empty messages array, got %v", gotBody["messages"])
	}
	kwargs, ok := gotBody["chat_template_kwargs"].(map[string]any)
	if !ok || kwargs["enable_thinking"] != false {
		t.Fatalf("expected chat_template_kwargs.enable_thinking=false, got %v", gotBody["chat_template_kwargs"])
	}
}

func TestOpenAIClient_GenerateContent_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"boom"}`))
	}))
	defer server.Close()

	client := NewOpenAIClient(server.URL, "some-model")

	_, err := client.GenerateContent(GenerateContentRequest{Intent: "x"})

	if err == nil {
		t.Fatalf("expected an error for a non-200 response")
	}
}

func TestOpenAIClient_GenerateContent_NoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[]}`))
	}))
	defer server.Close()

	client := NewOpenAIClient(server.URL, "some-model")

	_, err := client.GenerateContent(GenerateContentRequest{Intent: "x"})

	if err == nil {
		t.Fatalf("expected an error when the response has no choices")
	}
}

func TestOpenAIClient_GenerateContent_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{not json`))
	}))
	defer server.Close()

	client := NewOpenAIClient(server.URL, "some-model")

	_, err := client.GenerateContent(GenerateContentRequest{Intent: "x"})

	if err == nil {
		t.Fatalf("expected an error for invalid JSON in the response")
	}
}

func TestOpenAIClient_GenerateStoryboard_SendsExpectedRequest(t *testing.T) {
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if decodeErr := json.NewDecoder(r.Body).Decode(&gotBody); decodeErr != nil {
			t.Fatalf("failed to decode request body: %v", decodeErr)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"[{\"title\":\"Intro\",\"intent\":\"Bienvenida\"}]"}}]}`))
	}))
	defer server.Close()

	client := NewOpenAIClient(server.URL, "test-model")

	raw, err := client.GenerateStoryboard(GenerateStoryboardRequest{
		Objective: "Presentar el roadmap",
		Audience:  "Equipo",
		Count:     3,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if raw != `[{"title":"Intro","intent":"Bienvenida"}]` {
		t.Fatalf("expected raw JSON passthrough, got %q", raw)
	}
	messages, ok := gotBody["messages"].([]any)
	if !ok || len(messages) != 2 {
		t.Fatalf("expected 2 messages (system+user), got %v", gotBody["messages"])
	}
	userMsg := messages[1].(map[string]any)
	if userMsg["content"] == nil || !strings.Contains(userMsg["content"].(string), "Presentar el roadmap") {
		t.Fatalf("expected objective in the user message, got %v", userMsg["content"])
	}
}

func TestOpenAIClient_GenerateStoryboard_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewOpenAIClient(server.URL, "some-model")

	_, err := client.GenerateStoryboard(GenerateStoryboardRequest{Objective: "x", Count: 3})

	if err == nil {
		t.Fatalf("expected an error for a non-200 response")
	}
}
