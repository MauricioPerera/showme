package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MauricioPerera/showme/internal/domain"
)

const validDesign = `---
name: "Test Brand"
colors:
  primary: "#111827"
---

## Overview

Brand overview.
`

func writeConcept(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
}

func validDeckInput() domain.DeckInput {
	return domain.DeckInput{
		Title:  "Roadmap Q3",
		Slides: []domain.Slide{{ID: "intro", Title: "Introduccion"}},
	}
}

func TestCreateProject_Valid(t *testing.T) {
	dir := t.TempDir()
	writeConcept(t, dir, "brand.md", "---\ntype: Brand\n---\n\nBody.\n")

	input := CreateProjectInput{
		Name:          "Presentacion Q3",
		DesignContent: validDesign,
		DesignPath:    "DESIGN.md",
		KnowledgeRoot: dir,
		DeckInput:     validDeckInput(),
	}

	proj, report := CreateProject(input)

	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if proj.Name != "Presentacion Q3" {
		t.Fatalf("expected name preserved, got %q", proj.Name)
	}
	if proj.Version != 1 {
		t.Fatalf("expected version to default to 1, got %d", proj.Version)
	}
	if len(proj.Deck.Slides) != 1 {
		t.Fatalf("expected deck slides preserved, got %+v", proj.Deck)
	}
}

func TestCreateProject_InvalidDesignContent(t *testing.T) {
	dir := t.TempDir()
	writeConcept(t, dir, "brand.md", "---\ntype: Brand\n---\n\nBody.\n")

	input := CreateProjectInput{
		Name:          "Presentacion Q3",
		DesignContent: "no frontmatter here",
		DesignPath:    "DESIGN.md",
		KnowledgeRoot: dir,
		DeckInput:     validDeckInput(),
	}

	_, report := CreateProject(input)

	if !containsError(report.Errors, "frontmatter is required") {
		t.Fatalf("expected 'frontmatter is required' error, got %v", report.Errors)
	}
}

func TestCreateProject_InvalidKnowledgeConcept(t *testing.T) {
	dir := t.TempDir()
	writeConcept(t, dir, "broken.md", "---\ntitle: Sin tipo\n---\n\nBody.\n")

	input := CreateProjectInput{
		Name:          "Presentacion Q3",
		DesignContent: validDesign,
		DesignPath:    "DESIGN.md",
		KnowledgeRoot: dir,
		DeckInput:     validDeckInput(),
	}

	_, report := CreateProject(input)

	if !containsError(report.Errors, "broken.md: type is required") {
		t.Fatalf("expected 'broken.md: type is required' error, got %v", report.Errors)
	}
}

func TestCreateProject_InvalidDeckCascadesToProject(t *testing.T) {
	dir := t.TempDir()
	writeConcept(t, dir, "brand.md", "---\ntype: Brand\n---\n\nBody.\n")

	input := CreateProjectInput{
		Name:          "Presentacion Q3",
		DesignContent: validDesign,
		DesignPath:    "DESIGN.md",
		KnowledgeRoot: dir,
		DeckInput:     domain.DeckInput{Title: "Vacio"},
	}

	_, report := CreateProject(input)

	if !containsError(report.Errors, "at least one slide is required") {
		t.Fatalf("expected deck error, got %v", report.Errors)
	}
	if !containsError(report.Errors, "deck must have at least one slide") {
		t.Fatalf("expected cascaded project error, got %v", report.Errors)
	}
}

func TestCreateProject_EmptyName(t *testing.T) {
	dir := t.TempDir()
	writeConcept(t, dir, "brand.md", "---\ntype: Brand\n---\n\nBody.\n")

	input := CreateProjectInput{
		DesignContent: validDesign,
		DesignPath:    "DESIGN.md",
		KnowledgeRoot: dir,
		DeckInput:     validDeckInput(),
	}

	_, report := CreateProject(input)

	if !containsError(report.Errors, "name is required") {
		t.Fatalf("expected 'name is required' error, got %v", report.Errors)
	}
}

func containsError(errors []string, want string) bool {
	for _, e := range errors {
		if e == want {
			return true
		}
	}
	return false
}
