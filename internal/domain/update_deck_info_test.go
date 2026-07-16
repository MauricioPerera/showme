package domain

import "testing"

func TestUpdateDeckInfo_Valid(t *testing.T) {
	deck := deckWithTwoSlides(t)

	updated, report := UpdateDeckInfo(UpdateDeckInfoInput{
		Deck:     deck,
		Title:    "Roadmap Q4",
		Audience: "Equipo ejecutivo",
	})

	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if updated.Title != "Roadmap Q4" || updated.Audience != "Equipo ejecutivo" {
		t.Fatalf("expected title/audience updated, got %+v", updated)
	}
	if len(updated.Slides) != 2 || updated.Slides[0].ID != "intro" || updated.Slides[1].ID != "plan" {
		t.Fatalf("expected slides preserved, got %+v", updated.Slides)
	}
}

func TestUpdateDeckInfo_EmptyTitle(t *testing.T) {
	deck := deckWithTwoSlides(t)

	updated, report := UpdateDeckInfo(UpdateDeckInfoInput{
		Deck:     deck,
		Title:    "",
		Audience: "Equipo ejecutivo",
	})

	if !containsError(report.Errors, "title is required") {
		t.Fatalf("expected 'title is required' error, got %v", report.Errors)
	}
	if updated.Title != deck.Title {
		t.Fatalf("expected deck unchanged, got %+v", updated)
	}
}

func TestUpdateDeckInfo_DoesNotMutateOriginalDeck(t *testing.T) {
	deck := deckWithTwoSlides(t)

	_, _ = UpdateDeckInfo(UpdateDeckInfoInput{
		Deck:     deck,
		Title:    "Roadmap Q4",
		Audience: "Equipo ejecutivo",
	})

	if deck.Title == "Roadmap Q4" {
		t.Fatalf("expected original deck untouched, got %+v", deck)
	}
}
