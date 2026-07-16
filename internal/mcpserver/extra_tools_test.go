package mcpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcptest"
)

func fakeChatServer(t *testing.T, content string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encoded, err := json.Marshal(content)
		if err != nil {
			t.Fatalf("setup: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":` + string(encoded) + `}}]}`))
	}))
}

func createdProjectPath(t *testing.T, srv *mcptest.Server, root string) string {
	t.Helper()
	designPath := filepath.Join(root, "DESIGN.md")
	writeFile(t, designPath, mcpValidDesign)
	knowledgeRoot := filepath.Join(root, "knowledge")
	if err := os.Mkdir(knowledgeRoot, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	writeFile(t, filepath.Join(knowledgeRoot, "brand.md"), "---\ntype: Brand\n---\n\nBody.\n")
	deckPath := filepath.Join(root, "deck.json")
	writeFile(t, deckPath, `{"title":"Roadmap Q3","audience":"Equipo","slides":[{"id":"intro","title":"Introduccion","intent":"Dar la bienvenida"}]}`)
	dataDir := filepath.Join(root, "data")
	if err := os.Mkdir(dataDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	out, isError := callTool(t, srv, "create_project", map[string]any{
		"name":           "Roadmap Q3",
		"design_path":    designPath,
		"knowledge_root": knowledgeRoot,
		"deck_path":      deckPath,
		"out_dir":        dataDir,
	})
	if isError {
		t.Fatalf("create_project setup failed: %s", out)
	}
	var created struct {
		OK       bool
		Path     string
		Errors   []string
		Warnings []string
	}
	if err := json.Unmarshal([]byte(out), &created); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if !created.OK || created.Path == "" {
		t.Fatalf("create_project setup did not succeed: %+v", created)
	}
	return created.Path
}

func TestSlideTools_AddUpdateReorderRemove(t *testing.T) {
	root := t.TempDir()
	srv, err := mcptest.NewServer(t, Tools()...)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer srv.Close()

	path := createdProjectPath(t, srv, root)
	dataDir := filepath.Dir(path)

	out, isError := callTool(t, srv, "add_slide", map[string]any{
		"path":     path,
		"slide_id": "closing",
		"title":    "Cierre",
		"out_dir":  dataDir,
	})
	if isError {
		t.Fatalf("add_slide failed: %s", out)
	}

	out, isError = callTool(t, srv, "update_slide", map[string]any{
		"path":     path,
		"slide_id": "closing",
		"title":    "Cierre y proximos pasos",
		"out_dir":  dataDir,
	})
	if isError {
		t.Fatalf("update_slide failed: %s", out)
	}

	out, isError = callTool(t, srv, "reorder_slides", map[string]any{
		"path":    path,
		"order":   []any{"closing", "intro"},
		"out_dir": dataDir,
	})
	if isError {
		t.Fatalf("reorder_slides failed: %s", out)
	}

	out, isError = callTool(t, srv, "update_deck_info", map[string]any{
		"path":     path,
		"title":    "Roadmap Q4",
		"audience": "Directorio",
		"out_dir":  dataDir,
	})
	if isError {
		t.Fatalf("update_deck_info failed: %s", out)
	}

	out, isError = callTool(t, srv, "remove_slide", map[string]any{
		"path":     path,
		"slide_id": "closing",
		"out_dir":  dataDir,
	})
	if isError {
		t.Fatalf("remove_slide failed: %s", out)
	}
}

func TestSlideTools_AddSlide_MissingRequired(t *testing.T) {
	srv, err := mcptest.NewServer(t, Tools()...)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer srv.Close()

	_, isError := callTool(t, srv, "add_slide", map[string]any{
		"path":    filepath.Join(t.TempDir(), "does-not-exist.json"),
		"out_dir": t.TempDir(),
	})
	if !isError {
		t.Fatalf("expected a tool error for missing required arguments")
	}
}

func TestProjectManagementTools(t *testing.T) {
	root := t.TempDir()
	srv, err := mcptest.NewServer(t, Tools()...)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer srv.Close()

	path := createdProjectPath(t, srv, root)
	dataDir := filepath.Dir(path)

	out, isError := callTool(t, srv, "review_project", map[string]any{
		"path":     path,
		"slide_id": "intro",
		"decision": "accept",
		"out_dir":  dataDir,
	})
	if isError {
		t.Fatalf("review_project failed: %s", out)
	}

	out, isError = callTool(t, srv, "archive_project", map[string]any{
		"path":     path,
		"archived": true,
		"out_dir":  dataDir,
	})
	if isError {
		t.Fatalf("archive_project failed: %s", out)
	}

	htmlPath := filepath.Join(root, "out.html")
	out, isError = callTool(t, srv, "export_project", map[string]any{
		"path":     path,
		"out_path": htmlPath,
	})
	if isError {
		t.Fatalf("export_project failed: %s", out)
	}
	if _, statErr := os.Stat(htmlPath); statErr != nil {
		t.Fatalf("expected exported file to exist: %v", statErr)
	}

	out, isError = callTool(t, srv, "duplicate_project", map[string]any{
		"source_path": path,
		"new_name":    "Roadmap Q3 copia",
		"out_dir":     dataDir,
	})
	if isError {
		t.Fatalf("duplicate_project failed: %s", out)
	}

	out, isError = callTool(t, srv, "rename_project", map[string]any{
		"source_path": path,
		"new_name":    "Roadmap Q3 renombrado",
		"out_dir":     dataDir,
	})
	if isError {
		t.Fatalf("rename_project failed: %s", out)
	}
}

func TestProjectManagementTools_ReviewProject_InvalidDecision(t *testing.T) {
	root := t.TempDir()
	srv, err := mcptest.NewServer(t, Tools()...)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer srv.Close()

	path := createdProjectPath(t, srv, root)
	dataDir := filepath.Dir(path)

	out, isError := callTool(t, srv, "review_project", map[string]any{
		"path":     path,
		"slide_id": "intro",
		"decision": "not-a-real-decision",
		"out_dir":  dataDir,
	})
	if isError {
		t.Fatalf("expected a validation-error result, not a tool error: %s", out)
	}
	var reviewResult struct {
		OK     bool
		Errors []string
	}
	if err := json.Unmarshal([]byte(out), &reviewResult); err != nil {
		t.Fatalf("expected JSON result, got %q: %v", out, err)
	}
	if reviewResult.OK || len(reviewResult.Errors) == 0 {
		t.Fatalf("expected a validation error for an invalid review decision, got %+v", reviewResult)
	}
}

func TestAITools_GenerateSlideContentAndAll(t *testing.T) {
	root := t.TempDir()
	server := fakeChatServer(t, "Contenido generado.")
	defer server.Close()

	srv, err := mcptest.NewServer(t, Tools()...)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer srv.Close()

	path := createdProjectPath(t, srv, root)
	dataDir := filepath.Dir(path)

	out, isError := callTool(t, srv, "generate_slide_content", map[string]any{
		"path":     path,
		"slide_id": "intro",
		"base_url": server.URL,
		"model":    "test-model",
		"out_dir":  dataDir,
	})
	if isError {
		t.Fatalf("generate_slide_content failed: %s", out)
	}

	out, isError = callTool(t, srv, "add_slide", map[string]any{
		"path":     path,
		"slide_id": "closing",
		"title":    "Cierre",
		"intent":   "Resumir los proximos pasos",
		"out_dir":  dataDir,
	})
	if isError {
		t.Fatalf("add_slide setup failed: %s", out)
	}

	out, isError = callTool(t, srv, "generate_all_slides", map[string]any{
		"path":     path,
		"base_url": server.URL,
		"model":    "test-model",
		"out_dir":  dataDir,
	})
	if isError {
		t.Fatalf("generate_all_slides failed: %s", out)
	}
	var allResult struct {
		OK        bool
		Generated []string
		Skipped   []string
		Errors    []string
	}
	if err := json.Unmarshal([]byte(out), &allResult); err != nil {
		t.Fatalf("expected JSON result, got %q: %v", out, err)
	}
	if !allResult.OK || len(allResult.Generated) != 1 || allResult.Generated[0] != "closing" {
		t.Fatalf("expected only 'closing' generated, got %+v", allResult)
	}
	if len(allResult.Skipped) != 1 || allResult.Skipped[0] != "intro" {
		t.Fatalf("expected 'intro' skipped, got %+v", allResult)
	}
}

func TestAITools_GenerateStoryboard(t *testing.T) {
	server := fakeChatServer(t, `[{"title":"Introduccion","intent":"Dar la bienvenida"},{"title":"Plan","intent":"Explicar los proximos pasos"}]`)
	defer server.Close()

	srv, err := mcptest.NewServer(t, Tools()...)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer srv.Close()

	root := t.TempDir()
	outPath := filepath.Join(root, "deck.json")

	out, isError := callTool(t, srv, "generate_storyboard", map[string]any{
		"objective":  "Presentar el roadmap",
		"audience":   "Equipo",
		"base_url":   server.URL,
		"model":      "test-model",
		"deck_title": "Roadmap Q3",
		"count":      2,
		"out_path":   outPath,
	})
	if isError {
		t.Fatalf("generate_storyboard failed: %s", out)
	}
	if _, statErr := os.Stat(outPath); statErr != nil {
		t.Fatalf("expected deck file to exist: %v", statErr)
	}
}

func TestAITools_GenerateSlideContent_ProviderError(t *testing.T) {
	root := t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	srv, err := mcptest.NewServer(t, Tools()...)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer srv.Close()

	path := createdProjectPath(t, srv, root)
	dataDir := filepath.Dir(path)

	out, isError := callTool(t, srv, "generate_slide_content", map[string]any{
		"path":     path,
		"slide_id": "intro",
		"base_url": server.URL,
		"model":    "test-model",
		"out_dir":  dataDir,
	})
	if isError {
		t.Fatalf("expected a validation-error result, not a tool error: %s", out)
	}
	var genResult struct {
		OK     bool
		Errors []string
	}
	if err := json.Unmarshal([]byte(out), &genResult); err != nil {
		t.Fatalf("expected JSON result, got %q: %v", out, err)
	}
	if genResult.OK || len(genResult.Errors) == 0 {
		t.Fatalf("expected an error when the AI provider fails, got %+v", genResult)
	}
}
