package domain

import "testing"

func TestReorderSlides_Valid(t *testing.T) {
	deck := deckWithTwoSlides(t) // intro, plan

	updated, report := ReorderSlides(ReorderSlidesInput{
		Deck:  deck,
		Order: []string{"plan", "intro"},
	})

	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if updated.Slides[0].ID != "plan" || updated.Slides[1].ID != "intro" {
		t.Fatalf("expected reordered slides, got %+v", updated.Slides)
	}
	if updated.Slides[0].Status != SlideStatusDraft || updated.Slides[1].Status != SlideStatusAccepted {
		t.Fatalf("expected slide contents preserved, got %+v", updated.Slides)
	}
}

func TestReorderSlides_MissingSlideID(t *testing.T) {
	deck := deckWithTwoSlides(t)

	updated, report := ReorderSlides(ReorderSlidesInput{
		Deck:  deck,
		Order: []string{"intro"},
	})

	if !containsError(report.Errors, "missing slide id in order: plan") {
		t.Fatalf("expected 'missing slide id in order: plan' error, got %v", report.Errors)
	}
	if updated.Slides[0].ID != "intro" || updated.Slides[1].ID != "plan" {
		t.Fatalf("expected deck unchanged, got %+v", updated.Slides)
	}
}

func TestReorderSlides_UnknownSlideID(t *testing.T) {
	deck := deckWithTwoSlides(t)

	_, report := ReorderSlides(ReorderSlidesInput{
		Deck:  deck,
		Order: []string{"intro", "plan", "closing"},
	})

	if !containsError(report.Errors, "unknown slide id: closing") {
		t.Fatalf("expected 'unknown slide id: closing' error, got %v", report.Errors)
	}
}

func TestReorderSlides_DuplicateSlideID(t *testing.T) {
	deck := deckWithTwoSlides(t)

	_, report := ReorderSlides(ReorderSlidesInput{
		Deck:  deck,
		Order: []string{"intro", "intro"},
	})

	if !containsError(report.Errors, "duplicate slide id in order: intro") {
		t.Fatalf("expected 'duplicate slide id in order: intro' error, got %v", report.Errors)
	}
}

func TestReorderSlides_DoesNotMutateOriginalDeck(t *testing.T) {
	deck := deckWithTwoSlides(t)

	_, _ = ReorderSlides(ReorderSlidesInput{
		Deck:  deck,
		Order: []string{"plan", "intro"},
	})

	if deck.Slides[0].ID != "intro" || deck.Slides[1].ID != "plan" {
		t.Fatalf("expected original deck untouched, got %+v", deck.Slides)
	}
}
