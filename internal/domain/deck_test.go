package domain

import "testing"

func TestNewDeck_Valid(t *testing.T) {
	input := DeckInput{
		Title:    "Roadmap Q3",
		Audience: "Equipo de producto",
		Slides: []Slide{
			{ID: "intro", Title: "Introduccion", Status: SlideStatusDraft},
			{ID: "plan", Title: "Plan", Status: SlideStatusAccepted},
		},
	}

	deck, report := NewDeck(input)

	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if deck.Title != "Roadmap Q3" {
		t.Fatalf("expected title preserved, got %q", deck.Title)
	}
	if len(deck.Slides) != 2 {
		t.Fatalf("expected 2 slides, got %d", len(deck.Slides))
	}
	if deck.Slides[0].ID != "intro" || deck.Slides[1].ID != "plan" {
		t.Fatalf("expected slide order preserved, got %+v", deck.Slides)
	}
}

func TestNewDeck_EmptyTitle(t *testing.T) {
	input := DeckInput{
		Title: "",
		Slides: []Slide{
			{ID: "intro", Title: "Introduccion"},
		},
	}

	_, report := NewDeck(input)

	if !containsError(report.Errors, "title is required") {
		t.Fatalf("expected 'title is required' error, got %v", report.Errors)
	}
}

func TestNewDeck_NoSlides(t *testing.T) {
	input := DeckInput{Title: "Deck sin slides"}

	_, report := NewDeck(input)

	if !containsError(report.Errors, "at least one slide is required") {
		t.Fatalf("expected 'at least one slide is required' error, got %v", report.Errors)
	}
}

func TestNewDeck_SlideMissingID(t *testing.T) {
	input := DeckInput{
		Title: "Deck",
		Slides: []Slide{
			{ID: "", Title: "Sin id"},
		},
	}

	_, report := NewDeck(input)

	if !containsError(report.Errors, "slide[0]: id is required") {
		t.Fatalf("expected 'slide[0]: id is required' error, got %v", report.Errors)
	}
}

func TestNewDeck_SlideMissingTitle(t *testing.T) {
	input := DeckInput{
		Title: "Deck",
		Slides: []Slide{
			{ID: "s1", Title: ""},
		},
	}

	_, report := NewDeck(input)

	if !containsError(report.Errors, "slide[0]: title is required") {
		t.Fatalf("expected 'slide[0]: title is required' error, got %v", report.Errors)
	}
}

func TestNewDeck_DuplicateSlideID(t *testing.T) {
	input := DeckInput{
		Title: "Deck",
		Slides: []Slide{
			{ID: "s1", Title: "Uno"},
			{ID: "s1", Title: "Dos"},
		},
	}

	_, report := NewDeck(input)

	if !containsError(report.Errors, "duplicate slide id: s1") {
		t.Fatalf("expected 'duplicate slide id: s1' error, got %v", report.Errors)
	}
}

func TestNewDeck_InvalidStatus(t *testing.T) {
	input := DeckInput{
		Title: "Deck",
		Slides: []Slide{
			{ID: "s1", Title: "Uno", Status: SlideStatus("archived")},
		},
	}

	_, report := NewDeck(input)

	if !containsError(report.Errors, "slide[0]: invalid status: archived") {
		t.Fatalf("expected 'slide[0]: invalid status: archived' error, got %v", report.Errors)
	}
}

func TestNewDeck_EmptyStatusDefaultsToDraft(t *testing.T) {
	input := DeckInput{
		Title: "Deck",
		Slides: []Slide{
			{ID: "s1", Title: "Uno"},
		},
	}

	deck, report := NewDeck(input)

	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if deck.Slides[0].Status != SlideStatusDraft {
		t.Fatalf("expected status to default to draft, got %q", deck.Slides[0].Status)
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
