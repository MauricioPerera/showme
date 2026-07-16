package mcpserver

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/mcptest"

	"github.com/MauricioPerera/showme/internal/domain"
)

const mcpValidDesign = `---
name: "Test Brand"
colors:
  primary: "#111827"
---

## Overview

Brand overview.
`

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
}

func callTool(t *testing.T, srv *mcptest.Server, name string, args map[string]any) (string, bool) {
	t.Helper()
	var req mcp.CallToolRequest
	req.Params.Name = name
	req.Params.Arguments = args

	result, err := srv.Client().CallTool(context.Background(), req)
	if err != nil {
		t.Fatalf("CallTool(%s): %v", name, err)
	}

	var b strings.Builder
	for _, content := range result.Content {
		text, ok := content.(mcp.TextContent)
		if !ok {
			t.Fatalf("unsupported content type: %T", content)
		}
		b.WriteString(text.Text)
	}
	return b.String(), result.IsError
}

func TestTools_CreateListShowProject(t *testing.T) {
	root := t.TempDir()
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

	srv, err := mcptest.NewServer(t, Tools()...)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer srv.Close()

	// create_project
	out, isError := callTool(t, srv, "create_project", map[string]any{
		"name":           "Presentacion Q3",
		"design_path":    designPath,
		"knowledge_root": knowledgeRoot,
		"deck_path":      deckPath,
		"out_dir":        dataDir,
	})
	if isError {
		t.Fatalf("create_project returned an error: %s", out)
	}
	var createResult struct {
		OK   bool
		Path string
	}
	if err := json.Unmarshal([]byte(out), &createResult); err != nil {
		t.Fatalf("expected JSON result, got %q: %v", out, err)
	}
	if !createResult.OK || createResult.Path == "" {
		t.Fatalf("expected OK with a path, got %+v", createResult)
	}

	// list_projects
	out, isError = callTool(t, srv, "list_projects", map[string]any{"dir": dataDir})
	if isError {
		t.Fatalf("list_projects returned an error: %s", out)
	}
	var listResult struct {
		Projects []struct {
			Name string
			Path string
		}
	}
	if err := json.Unmarshal([]byte(out), &listResult); err != nil {
		t.Fatalf("expected JSON result, got %q: %v", out, err)
	}
	if len(listResult.Projects) != 1 || listResult.Projects[0].Name != "Presentacion Q3" {
		t.Fatalf("expected one listed project, got %+v", listResult.Projects)
	}

	// show_project
	out, isError = callTool(t, srv, "show_project", map[string]any{"path": listResult.Projects[0].Path})
	if isError {
		t.Fatalf("show_project returned an error: %s", out)
	}
	var showResult struct {
		Project domain.Project
	}
	if err := json.Unmarshal([]byte(out), &showResult); err != nil {
		t.Fatalf("expected JSON result, got %q: %v", out, err)
	}
	if showResult.Project.Name != "Presentacion Q3" || len(showResult.Project.Deck.Slides) != 1 {
		t.Fatalf("expected the created project, got %+v", showResult.Project)
	}
}

func TestTools_CreateProject_ValidationError(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "data")
	if err := os.Mkdir(dataDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	deckPath := filepath.Join(root, "deck.json")
	writeFile(t, deckPath, `{"title":"x","slides":[{"id":"s","title":"s"}]}`)

	srv, err := mcptest.NewServer(t, Tools()...)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer srv.Close()

	out, isError := callTool(t, srv, "create_project", map[string]any{
		"name":           "x",
		"design_path":    filepath.Join(root, "does-not-exist.md"),
		"knowledge_root": root,
		"deck_path":      deckPath,
		"out_dir":        dataDir,
	})

	if !isError {
		t.Fatalf("expected a tool error for a missing design file, got %s", out)
	}
}

func TestTools_ListProjects_MissingDir(t *testing.T) {
	srv, err := mcptest.NewServer(t, Tools()...)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer srv.Close()

	_, isError := callTool(t, srv, "list_projects", map[string]any{
		"dir": filepath.Join(t.TempDir(), "does-not-exist"),
	})

	if !isError {
		t.Fatalf("expected a tool error for a missing directory")
	}
}

func TestTools_ShowProject_MissingFile(t *testing.T) {
	srv, err := mcptest.NewServer(t, Tools()...)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer srv.Close()

	_, isError := callTool(t, srv, "show_project", map[string]any{
		"path": filepath.Join(t.TempDir(), "does-not-exist.json"),
	})

	if !isError {
		t.Fatalf("expected a tool error for a missing project file")
	}
}
