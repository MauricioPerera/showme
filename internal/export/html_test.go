package export

import (
	"strings"
	"testing"

	"github.com/MauricioPerera/showme/internal/domain"
)

func sampleProject(t *testing.T) domain.Project {
	t.Helper()
	deck, deckReport := domain.NewDeck(domain.DeckInput{
		Title:    "Roadmap Q3",
		Audience: "Equipo de producto",
		Slides: []domain.Slide{
			{ID: "intro", Title: "Introduccion", Content: "Bienvenida al roadmap.", Status: domain.SlideStatusAccepted},
			{ID: "plan", Title: "Plan", Content: "Los proximos pasos.", Status: domain.SlideStatusDraft},
		},
	})
	if len(deckReport.Errors) != 0 {
		t.Fatalf("setup: unexpected deck errors: %v", deckReport.Errors)
	}
	proj, projReport := domain.NewProject(domain.ProjectInput{
		Name:          "Presentacion Q3",
		Deck:          deck,
		DesignPath:    "DESIGN.md",
		KnowledgePath: "knowledge/showme",
	})
	if len(projReport.Errors) != 0 {
		t.Fatalf("setup: unexpected project errors: %v", projReport.Errors)
	}
	return proj
}

func TestExportProjectHTML_ContainsDoctypeTitleAndSections(t *testing.T) {
	proj := sampleProject(t)

	out := ExportProjectHTML(proj)

	if !strings.HasPrefix(out, "<!doctype html>") {
		t.Fatalf("expected output to start with '<!doctype html>', got: %s", out[:min(40, len(out))])
	}
	if !strings.Contains(out, "<title>Roadmap Q3</title>") {
		t.Fatalf("expected a <title> with the deck title, got: %s", out)
	}
	if strings.Count(out, "<section") != 2 {
		t.Fatalf("expected 2 <section> elements, got: %s", out)
	}
	if !strings.Contains(out, "Introduccion") || !strings.Contains(out, "Plan") {
		t.Fatalf("expected both slide titles present, got: %s", out)
	}
}

func TestExportProjectHTML_PreservesSlideOrder(t *testing.T) {
	proj := sampleProject(t)

	out := ExportProjectHTML(proj)

	introIndex := strings.Index(out, "Introduccion")
	planIndex := strings.Index(out, "Plan")
	if introIndex == -1 || planIndex == -1 || introIndex > planIndex {
		t.Fatalf("expected 'Introduccion' before 'Plan', got: %s", out)
	}
}

func TestExportProjectHTML_EscapesSlideContent(t *testing.T) {
	proj := sampleProject(t)
	proj.Deck.Slides[0].Title = "<script>alert(1)</script>"
	proj.Deck.Slides[0].Content = "Tom & Jerry"

	out := ExportProjectHTML(proj)

	if strings.Contains(out, "<script>alert(1)</script>") {
		t.Fatalf("expected slide title to be HTML-escaped, got: %s", out)
	}
	if !strings.Contains(out, "&lt;script&gt;") {
		t.Fatalf("expected escaped script tag, got: %s", out)
	}
	if !strings.Contains(out, "Tom &amp; Jerry") {
		t.Fatalf("expected escaped ampersand, got: %s", out)
	}
}

func TestExportProjectHTML_Deterministic(t *testing.T) {
	proj := sampleProject(t)

	first := ExportProjectHTML(proj)
	second := ExportProjectHTML(proj)

	if first != second {
		t.Fatalf("expected deterministic output, got two different renders")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
