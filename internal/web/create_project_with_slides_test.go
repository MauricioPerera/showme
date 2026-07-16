package web

import (
	"path/filepath"
	"testing"

	"github.com/MauricioPerera/showme/internal/storage"
)

func TestHandleCreateProjectWithSlides_Valid(t *testing.T) {
	designPath, knowledgeRoot, dataDir := setupWebDirs(t)

	result, err := HandleCreateProjectWithSlides(CreateProjectWithSlidesInput{
		Name:          "Presentacion Q3",
		DesignPath:    designPath,
		KnowledgeRoot: knowledgeRoot,
		DeckTitle:     "Roadmap Q3",
		DeckAudience:  "Equipo de producto",
		Slides: []SlideInput{
			{Title: "Introduccion", Intent: "Dar la bienvenida"},
			{Title: "Plan", Intent: "Explicar los proximos pasos"},
		},
		Dir: dataDir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected OK, got errors: %v", result.Errors)
	}

	proj, loadErr := storage.LoadProject(result.Path)
	if loadErr != nil {
		t.Fatalf("expected saved project to load, got error: %v", loadErr)
	}
	if len(proj.Deck.Slides) != 2 {
		t.Fatalf("expected 2 slides, got %+v", proj.Deck.Slides)
	}
	if proj.Deck.Slides[0].ID == "" || proj.Deck.Slides[1].ID == "" || proj.Deck.Slides[0].ID == proj.Deck.Slides[1].ID {
		t.Fatalf("expected unique, non-empty slide ids, got %+v", proj.Deck.Slides)
	}
}

func TestHandleCreateProjectWithSlides_DedupesCollidingSlugs(t *testing.T) {
	designPath, knowledgeRoot, dataDir := setupWebDirs(t)

	result, err := HandleCreateProjectWithSlides(CreateProjectWithSlidesInput{
		Name:          "Presentacion Q3",
		DesignPath:    designPath,
		KnowledgeRoot: knowledgeRoot,
		DeckTitle:     "Roadmap Q3",
		Slides: []SlideInput{
			{Title: "Introduccion", Intent: "a"},
			{Title: "Introduccion", Intent: "b"},
		},
		Dir: dataDir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected OK, got errors: %v", result.Errors)
	}

	proj, loadErr := storage.LoadProject(result.Path)
	if loadErr != nil {
		t.Fatalf("unexpected error: %v", loadErr)
	}
	if proj.Deck.Slides[0].ID == proj.Deck.Slides[1].ID {
		t.Fatalf("expected deduped ids, got %+v", proj.Deck.Slides)
	}
}

func TestHandleCreateProjectWithSlides_NoSlides(t *testing.T) {
	designPath, knowledgeRoot, dataDir := setupWebDirs(t)

	result, err := HandleCreateProjectWithSlides(CreateProjectWithSlidesInput{
		Name:          "Presentacion Q3",
		DesignPath:    designPath,
		KnowledgeRoot: knowledgeRoot,
		DeckTitle:     "Roadmap Q3",
		Slides:        nil,
		Dir:           dataDir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OK {
		t.Fatalf("expected not OK, got %+v", result)
	}
	if !containsError(result.Errors, "at least one slide is required") {
		t.Fatalf("expected 'at least one slide is required' error, got %v", result.Errors)
	}
}

func TestHandleCreateProjectWithSlides_MissingDataDir(t *testing.T) {
	designPath, knowledgeRoot, dataDir := setupWebDirs(t)

	_, err := HandleCreateProjectWithSlides(CreateProjectWithSlidesInput{
		Name:          "Presentacion Q3",
		DesignPath:    designPath,
		KnowledgeRoot: knowledgeRoot,
		DeckTitle:     "Roadmap Q3",
		Slides:        []SlideInput{{Title: "Introduccion", Intent: "x"}},
		Dir:           filepath.Join(dataDir, "no-such-subdir"),
	})

	if err == nil {
		t.Fatalf("expected an error for a missing data directory")
	}
}
