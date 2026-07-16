package domain

import "testing"

func validDeck(t *testing.T) Deck {
	t.Helper()
	deck, report := NewDeck(DeckInput{
		Title: "Roadmap Q3",
		Slides: []Slide{
			{ID: "intro", Title: "Introduccion"},
		},
	})
	if len(report.Errors) != 0 {
		t.Fatalf("setup: unexpected deck errors: %v", report.Errors)
	}
	return deck
}

func TestNewProject_Valid(t *testing.T) {
	input := ProjectInput{
		Name:          "Presentacion Q3",
		Deck:          validDeck(t),
		DesignPath:    "DESIGN.md",
		KnowledgePath: "knowledge/showme",
	}

	project, report := NewProject(input)

	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if project.Name != "Presentacion Q3" {
		t.Fatalf("expected name preserved, got %q", project.Name)
	}
	if project.Version != 1 {
		t.Fatalf("expected version to default to 1, got %d", project.Version)
	}
	if len(project.Deck.Slides) != 1 {
		t.Fatalf("expected deck preserved, got %+v", project.Deck)
	}
}

func TestNewProject_EmptyName(t *testing.T) {
	input := ProjectInput{
		Deck:          validDeck(t),
		DesignPath:    "DESIGN.md",
		KnowledgePath: "knowledge/showme",
	}

	_, report := NewProject(input)

	if !containsError(report.Errors, "name is required") {
		t.Fatalf("expected 'name is required' error, got %v", report.Errors)
	}
}

func TestNewProject_DeckWithoutSlides(t *testing.T) {
	input := ProjectInput{
		Name:          "Presentacion Q3",
		Deck:          Deck{Title: "Vacio"},
		DesignPath:    "DESIGN.md",
		KnowledgePath: "knowledge/showme",
	}

	_, report := NewProject(input)

	if !containsError(report.Errors, "deck must have at least one slide") {
		t.Fatalf("expected 'deck must have at least one slide' error, got %v", report.Errors)
	}
}

func TestNewProject_MissingDesignPath(t *testing.T) {
	input := ProjectInput{
		Name:          "Presentacion Q3",
		Deck:          validDeck(t),
		KnowledgePath: "knowledge/showme",
	}

	_, report := NewProject(input)

	if !containsError(report.Errors, "design path is required") {
		t.Fatalf("expected 'design path is required' error, got %v", report.Errors)
	}
}

func TestNewProject_MissingKnowledgePath(t *testing.T) {
	input := ProjectInput{
		Name:       "Presentacion Q3",
		Deck:       validDeck(t),
		DesignPath: "DESIGN.md",
	}

	_, report := NewProject(input)

	if !containsError(report.Errors, "knowledge path is required") {
		t.Fatalf("expected 'knowledge path is required' error, got %v", report.Errors)
	}
}

func TestNewProject_NegativeVersion(t *testing.T) {
	input := ProjectInput{
		Name:          "Presentacion Q3",
		Deck:          validDeck(t),
		DesignPath:    "DESIGN.md",
		KnowledgePath: "knowledge/showme",
		Version:       -1,
	}

	_, report := NewProject(input)

	if !containsError(report.Errors, "version must be positive") {
		t.Fatalf("expected 'version must be positive' error, got %v", report.Errors)
	}
}

func TestNewProject_ExplicitVersionPreserved(t *testing.T) {
	input := ProjectInput{
		Name:          "Presentacion Q3",
		Deck:          validDeck(t),
		DesignPath:    "DESIGN.md",
		KnowledgePath: "knowledge/showme",
		Version:       3,
	}

	project, report := NewProject(input)

	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if project.Version != 3 {
		t.Fatalf("expected version 3 preserved, got %d", project.Version)
	}
}
