package domain

import "testing"

func TestNewReview_Valid(t *testing.T) {
	input := ReviewInput{
		SlideID:  "intro",
		Decision: ReviewDecisionAccepted,
	}

	review, report := NewReview(input)

	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if review.SlideID != "intro" {
		t.Fatalf("expected slide id preserved, got %q", review.SlideID)
	}
	if review.Decision != ReviewDecisionAccepted {
		t.Fatalf("expected decision preserved, got %q", review.Decision)
	}
}

func TestNewReview_EmptySlideID(t *testing.T) {
	input := ReviewInput{Decision: ReviewDecisionAccepted}

	_, report := NewReview(input)

	if !containsError(report.Errors, "slide id is required") {
		t.Fatalf("expected 'slide id is required' error, got %v", report.Errors)
	}
}

func TestNewReview_EmptyDecision(t *testing.T) {
	input := ReviewInput{SlideID: "intro"}

	_, report := NewReview(input)

	if !containsError(report.Errors, "decision is required") {
		t.Fatalf("expected 'decision is required' error, got %v", report.Errors)
	}
}

func TestNewReview_InvalidDecision(t *testing.T) {
	input := ReviewInput{SlideID: "intro", Decision: ReviewDecision("archived")}

	_, report := NewReview(input)

	if !containsError(report.Errors, "invalid decision: archived") {
		t.Fatalf("expected 'invalid decision: archived' error, got %v", report.Errors)
	}
}

func TestNewReview_RejectedWithNotes(t *testing.T) {
	input := ReviewInput{
		SlideID:  "intro",
		Decision: ReviewDecisionRejected,
		Notes:    "Claim sin fuente",
	}

	review, report := NewReview(input)

	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if review.Notes != "Claim sin fuente" {
		t.Fatalf("expected notes preserved, got %q", review.Notes)
	}
}

func TestNewReview_EditedDecision(t *testing.T) {
	input := ReviewInput{SlideID: "intro", Decision: ReviewDecisionEdited}

	review, report := NewReview(input)

	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if review.Decision != ReviewDecisionEdited {
		t.Fatalf("expected decision preserved, got %q", review.Decision)
	}
}
