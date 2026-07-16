package domain

import "fmt"

// ApplyReviewInput is the raw data used by ApplyReview.
type ApplyReviewInput struct {
	Deck   Deck
	Review ReviewInput
}

func statusForDecision(decision ReviewDecision) SlideStatus {
	switch decision {
	case ReviewDecisionAccepted:
		return SlideStatusAccepted
	case ReviewDecisionRejected:
		return SlideStatusRejected
	default: // ReviewDecisionEdited
		return SlideStatusDraft
	}
}

// ApplyReview validates a Review and, if valid, returns a copy of Deck with
// the reviewed slide's Status updated to reflect the decision: accepted ->
// SlideStatusAccepted, rejected -> SlideStatusRejected, edited ->
// SlideStatusDraft (an edited slide needs a fresh review). The original
// Deck is never mutated.
func ApplyReview(input ApplyReviewInput) (Deck, Report) {
	review, report := NewReview(input.Review)

	slides := make([]Slide, len(input.Deck.Slides))
	copy(slides, input.Deck.Slides)
	updated := Deck{
		Title:    input.Deck.Title,
		Audience: input.Deck.Audience,
		Slides:   slides,
	}

	if len(report.Errors) != 0 {
		return updated, report
	}

	for i, slide := range slides {
		if slide.ID == review.SlideID {
			slides[i].Status = statusForDecision(review.Decision)
			return updated, report
		}
	}

	report.Errors = append(report.Errors, fmt.Sprintf("slide not found: %s", review.SlideID))
	return updated, report
}
