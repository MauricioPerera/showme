package domain

import "testing"

func TestAddSlide_Valid(t *testing.T) {
	deck := deckWithTwoSlides(t)

	updated, report := AddSlide(AddSlideInput{
		Deck:  deck,
		Slide: Slide{ID: "closing", Title: "Cierre"},
	})

	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if len(updated.Slides) != 3 {
		t.Fatalf("expected 3 slides, got %d", len(updated.Slides))
	}
	if updated.Slides[2].ID != "closing" || updated.Slides[2].Status != SlideStatusDraft {
		t.Fatalf("expected new slide appended with draft status, got %+v", updated.Slides[2])
	}
	if updated.Slides[0].ID != "intro" || updated.Slides[1].ID != "plan" {
		t.Fatalf("expected existing slides preserved in order, got %+v", updated.Slides)
	}
}

func TestAddSlide_EmptyID(t *testing.T) {
	deck := deckWithTwoSlides(t)

	updated, report := AddSlide(AddSlideInput{
		Deck:  deck,
		Slide: Slide{ID: "", Title: "Cierre"},
	})

	if !containsError(report.Errors, "slide id is required") {
		t.Fatalf("expected 'slide id is required' error, got %v", report.Errors)
	}
	if len(updated.Slides) != 2 {
		t.Fatalf("expected deck unchanged, got %+v", updated.Slides)
	}
}

func TestAddSlide_EmptyTitle(t *testing.T) {
	deck := deckWithTwoSlides(t)

	updated, report := AddSlide(AddSlideInput{
		Deck:  deck,
		Slide: Slide{ID: "closing", Title: ""},
	})

	if !containsError(report.Errors, "slide title is required") {
		t.Fatalf("expected 'slide title is required' error, got %v", report.Errors)
	}
	if len(updated.Slides) != 2 {
		t.Fatalf("expected deck unchanged, got %+v", updated.Slides)
	}
}

func TestAddSlide_DuplicateID(t *testing.T) {
	deck := deckWithTwoSlides(t)

	updated, report := AddSlide(AddSlideInput{
		Deck:  deck,
		Slide: Slide{ID: "intro", Title: "Otra intro"},
	})

	if !containsError(report.Errors, "duplicate slide id: intro") {
		t.Fatalf("expected 'duplicate slide id: intro' error, got %v", report.Errors)
	}
	if len(updated.Slides) != 2 {
		t.Fatalf("expected deck unchanged, got %+v", updated.Slides)
	}
}

func TestAddSlide_InvalidStatus(t *testing.T) {
	deck := deckWithTwoSlides(t)

	_, report := AddSlide(AddSlideInput{
		Deck:  deck,
		Slide: Slide{ID: "closing", Title: "Cierre", Status: SlideStatus("archived")},
	})

	if !containsError(report.Errors, "invalid status: archived") {
		t.Fatalf("expected 'invalid status: archived' error, got %v", report.Errors)
	}
}

func TestAddSlide_DoesNotMutateOriginalDeck(t *testing.T) {
	deck := deckWithTwoSlides(t)

	_, _ = AddSlide(AddSlideInput{
		Deck:  deck,
		Slide: Slide{ID: "closing", Title: "Cierre"},
	})

	if len(deck.Slides) != 2 {
		t.Fatalf("expected original deck untouched, got %+v", deck.Slides)
	}
}
