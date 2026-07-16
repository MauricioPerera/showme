package domain

import "testing"

func TestUpdateSlide_PreservesStatusWhenNotGiven(t *testing.T) {
	deck := deckWithTwoSlides(t) // intro=accepted, plan=draft

	updated, report := UpdateSlide(UpdateSlideInput{
		Deck:  deck,
		Slide: Slide{ID: "intro", Title: "Introduccion revisada", Content: "Nuevo contenido"},
	})

	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if updated.Slides[0].Title != "Introduccion revisada" || updated.Slides[0].Content != "Nuevo contenido" {
		t.Fatalf("expected title/content updated, got %+v", updated.Slides[0])
	}
	if updated.Slides[0].Status != SlideStatusAccepted {
		t.Fatalf("expected status preserved as accepted, got %q", updated.Slides[0].Status)
	}
	if updated.Slides[1].ID != "plan" || updated.Slides[1].Status != SlideStatusDraft {
		t.Fatalf("expected other slide untouched, got %+v", updated.Slides[1])
	}
}

func TestUpdateSlide_ExplicitStatusOverrides(t *testing.T) {
	deck := deckWithTwoSlides(t)

	updated, report := UpdateSlide(UpdateSlideInput{
		Deck:  deck,
		Slide: Slide{ID: "intro", Title: "Introduccion", Status: SlideStatusRejected},
	})

	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if updated.Slides[0].Status != SlideStatusRejected {
		t.Fatalf("expected status overridden to rejected, got %q", updated.Slides[0].Status)
	}
}

func TestUpdateSlide_NotFound(t *testing.T) {
	deck := deckWithTwoSlides(t)

	updated, report := UpdateSlide(UpdateSlideInput{
		Deck:  deck,
		Slide: Slide{ID: "missing", Title: "Titulo"},
	})

	if !containsError(report.Errors, "slide not found: missing") {
		t.Fatalf("expected 'slide not found: missing' error, got %v", report.Errors)
	}
	if updated.Slides[0].Title != "Introduccion" {
		t.Fatalf("expected deck unchanged, got %+v", updated.Slides)
	}
}

func TestUpdateSlide_EmptyID(t *testing.T) {
	deck := deckWithTwoSlides(t)

	_, report := UpdateSlide(UpdateSlideInput{
		Deck:  deck,
		Slide: Slide{ID: "", Title: "Titulo"},
	})

	if !containsError(report.Errors, "slide id is required") {
		t.Fatalf("expected 'slide id is required' error, got %v", report.Errors)
	}
}

func TestUpdateSlide_EmptyTitle(t *testing.T) {
	deck := deckWithTwoSlides(t)

	_, report := UpdateSlide(UpdateSlideInput{
		Deck:  deck,
		Slide: Slide{ID: "intro", Title: ""},
	})

	if !containsError(report.Errors, "slide title is required") {
		t.Fatalf("expected 'slide title is required' error, got %v", report.Errors)
	}
}

func TestUpdateSlide_InvalidStatus(t *testing.T) {
	deck := deckWithTwoSlides(t)

	_, report := UpdateSlide(UpdateSlideInput{
		Deck:  deck,
		Slide: Slide{ID: "intro", Title: "Introduccion", Status: SlideStatus("archived")},
	})

	if !containsError(report.Errors, "invalid status: archived") {
		t.Fatalf("expected 'invalid status: archived' error, got %v", report.Errors)
	}
}

func TestUpdateSlide_DoesNotMutateOriginalDeck(t *testing.T) {
	deck := deckWithTwoSlides(t)

	_, _ = UpdateSlide(UpdateSlideInput{
		Deck:  deck,
		Slide: Slide{ID: "intro", Title: "Cambiado"},
	})

	if deck.Slides[0].Title != "Introduccion" {
		t.Fatalf("expected original deck untouched, got %+v", deck.Slides[0])
	}
}
