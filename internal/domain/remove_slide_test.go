package domain

import "testing"

func TestRemoveSlide_Valid(t *testing.T) {
	deck := deckWithTwoSlides(t)

	updated, report := RemoveSlide(RemoveSlideInput{Deck: deck, SlideID: "plan"})

	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if len(updated.Slides) != 1 || updated.Slides[0].ID != "intro" {
		t.Fatalf("expected only 'intro' remaining, got %+v", updated.Slides)
	}
}

func TestRemoveSlide_NotFound(t *testing.T) {
	deck := deckWithTwoSlides(t)

	updated, report := RemoveSlide(RemoveSlideInput{Deck: deck, SlideID: "missing"})

	if !containsError(report.Errors, "slide not found: missing") {
		t.Fatalf("expected 'slide not found: missing' error, got %v", report.Errors)
	}
	if len(updated.Slides) != 2 {
		t.Fatalf("expected deck unchanged, got %+v", updated.Slides)
	}
}

func TestRemoveSlide_LastSlideRefused(t *testing.T) {
	deck, deckReport := NewDeck(DeckInput{
		Title:  "Solo una slide",
		Slides: []Slide{{ID: "intro", Title: "Introduccion"}},
	})
	if len(deckReport.Errors) != 0 {
		t.Fatalf("setup: unexpected deck errors: %v", deckReport.Errors)
	}

	updated, report := RemoveSlide(RemoveSlideInput{Deck: deck, SlideID: "intro"})

	if !containsError(report.Errors, "deck must have at least one slide") {
		t.Fatalf("expected 'deck must have at least one slide' error, got %v", report.Errors)
	}
	if len(updated.Slides) != 1 {
		t.Fatalf("expected deck unchanged, got %+v", updated.Slides)
	}
}

func TestRemoveSlide_DoesNotMutateOriginalDeck(t *testing.T) {
	deck := deckWithTwoSlides(t)

	_, _ = RemoveSlide(RemoveSlideInput{Deck: deck, SlideID: "plan"})

	if len(deck.Slides) != 2 {
		t.Fatalf("expected original deck untouched, got %+v", deck.Slides)
	}
}
