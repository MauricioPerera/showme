package domain

import "testing"

func deckWithTwoSlides(t *testing.T) Deck {
	t.Helper()
	deck, report := NewDeck(DeckInput{
		Title: "Roadmap Q3",
		Slides: []Slide{
			{ID: "intro", Title: "Introduccion", Status: SlideStatusAccepted},
			{ID: "plan", Title: "Plan", Status: SlideStatusDraft},
		},
	})
	if len(report.Errors) != 0 {
		t.Fatalf("setup: unexpected deck errors: %v", report.Errors)
	}
	return deck
}

func TestApplyReview_Accepted(t *testing.T) {
	deck := deckWithTwoSlides(t)

	updated, report := ApplyReview(ApplyReviewInput{
		Deck:   deck,
		Review: ReviewInput{SlideID: "plan", Decision: ReviewDecisionAccepted},
	})

	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if updated.Slides[1].Status != SlideStatusAccepted {
		t.Fatalf("expected plan slide accepted, got %q", updated.Slides[1].Status)
	}
	if updated.Slides[0].Status != SlideStatusAccepted {
		t.Fatalf("expected intro slide untouched, got %q", updated.Slides[0].Status)
	}
}

func TestApplyReview_Rejected(t *testing.T) {
	deck := deckWithTwoSlides(t)

	updated, report := ApplyReview(ApplyReviewInput{
		Deck:   deck,
		Review: ReviewInput{SlideID: "intro", Decision: ReviewDecisionRejected},
	})

	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if updated.Slides[0].Status != SlideStatusRejected {
		t.Fatalf("expected intro slide rejected, got %q", updated.Slides[0].Status)
	}
}

func TestApplyReview_EditedResetsToDraft(t *testing.T) {
	deck := deckWithTwoSlides(t)

	updated, report := ApplyReview(ApplyReviewInput{
		Deck:   deck,
		Review: ReviewInput{SlideID: "intro", Decision: ReviewDecisionEdited},
	})

	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if updated.Slides[0].Status != SlideStatusDraft {
		t.Fatalf("expected intro slide reset to draft, got %q", updated.Slides[0].Status)
	}
}

func TestApplyReview_SlideNotFound(t *testing.T) {
	deck := deckWithTwoSlides(t)

	updated, report := ApplyReview(ApplyReviewInput{
		Deck:   deck,
		Review: ReviewInput{SlideID: "missing", Decision: ReviewDecisionAccepted},
	})

	if !containsError(report.Errors, "slide not found: missing") {
		t.Fatalf("expected 'slide not found: missing' error, got %v", report.Errors)
	}
	if updated.Slides[0].Status != SlideStatusAccepted || updated.Slides[1].Status != SlideStatusDraft {
		t.Fatalf("expected deck unchanged, got %+v", updated.Slides)
	}
}

func TestApplyReview_InvalidReviewLeavesDeckUnchanged(t *testing.T) {
	deck := deckWithTwoSlides(t)

	updated, report := ApplyReview(ApplyReviewInput{
		Deck:   deck,
		Review: ReviewInput{SlideID: "", Decision: ReviewDecisionAccepted},
	})

	if !containsError(report.Errors, "slide id is required") {
		t.Fatalf("expected 'slide id is required' error, got %v", report.Errors)
	}
	if updated.Slides[0].Status != SlideStatusAccepted || updated.Slides[1].Status != SlideStatusDraft {
		t.Fatalf("expected deck unchanged, got %+v", updated.Slides)
	}
}

func TestApplyReview_DoesNotMutateOriginalDeck(t *testing.T) {
	deck := deckWithTwoSlides(t)

	_, _ = ApplyReview(ApplyReviewInput{
		Deck:   deck,
		Review: ReviewInput{SlideID: "plan", Decision: ReviewDecisionAccepted},
	})

	if deck.Slides[1].Status != SlideStatusDraft {
		t.Fatalf("expected original deck untouched, got %q", deck.Slides[1].Status)
	}
}
