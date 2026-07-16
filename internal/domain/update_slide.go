package domain

import "fmt"

// UpdateSlideInput is the raw data used by UpdateSlide. Slide.ID identifies
// which existing slide to update; the rest of Slide's fields become its new
// values.
type UpdateSlideInput struct {
	Deck  Deck
	Slide Slide
}

// UpdateSlide validates new field values for an existing slide and, if
// valid, returns a copy of Deck with that slide replaced in place. If
// Slide.Status is empty, the slide's previous status is preserved instead
// of resetting to draft (unlike AddSlide, which defaults a fresh slide to
// draft). The original Deck is never mutated.
func UpdateSlide(input UpdateSlideInput) (Deck, Report) {
	report := Report{}
	slide := input.Slide

	index := -1
	for i, existing := range input.Deck.Slides {
		if existing.ID == slide.ID {
			index = i
			break
		}
	}

	if slide.ID == "" {
		report.Errors = append(report.Errors, "slide id is required")
	} else if index == -1 {
		report.Errors = append(report.Errors, fmt.Sprintf("slide not found: %s", slide.ID))
	}

	if slide.Title == "" {
		report.Errors = append(report.Errors, "slide title is required")
	}

	if slide.Status == "" {
		if index != -1 {
			slide.Status = input.Deck.Slides[index].Status
		}
	} else if !isValidSlideStatus(slide.Status) {
		report.Errors = append(report.Errors, fmt.Sprintf("invalid status: %s", slide.Status))
	}

	slides := make([]Slide, len(input.Deck.Slides))
	copy(slides, input.Deck.Slides)
	if len(report.Errors) == 0 {
		slides[index] = slide
	}

	updated := Deck{
		Title:    input.Deck.Title,
		Audience: input.Deck.Audience,
		Slides:   slides,
	}
	return updated, report
}
