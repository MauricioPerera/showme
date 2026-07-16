package domain

import "fmt"

// ReviewDecision is the human verdict over a slide or a proposal.
type ReviewDecision string

const (
	ReviewDecisionAccepted ReviewDecision = "accepted"
	ReviewDecisionEdited   ReviewDecision = "edited"
	ReviewDecisionRejected ReviewDecision = "rejected"
)

// ReviewInput is the raw data used to build a Review.
type ReviewInput struct {
	SlideID  string
	Decision ReviewDecision
	Notes    string
}

// Review is a human decision over a slide or a proposal.
type Review struct {
	SlideID  string
	Decision ReviewDecision
	Notes    string
}

func isValidReviewDecision(decision ReviewDecision) bool {
	switch decision {
	case ReviewDecisionAccepted, ReviewDecisionEdited, ReviewDecisionRejected:
		return true
	}
	return false
}

// NewReview builds a Review from input, enforcing its structural invariants.
func NewReview(input ReviewInput) (Review, Report) {
	report := Report{}

	if input.SlideID == "" {
		report.Errors = append(report.Errors, "slide id is required")
	}

	if input.Decision == "" {
		report.Errors = append(report.Errors, "decision is required")
	} else if !isValidReviewDecision(input.Decision) {
		report.Errors = append(report.Errors, fmt.Sprintf("invalid decision: %s", input.Decision))
	}

	review := Review{
		SlideID:  input.SlideID,
		Decision: input.Decision,
		Notes:    input.Notes,
	}
	return review, report
}
