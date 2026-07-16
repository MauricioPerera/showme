package domain

import "fmt"

// AddSlideInput is the raw data used by AddSlide.
type AddSlideInput struct {
	Deck  Deck
	Slide Slide
}

// AddSlide validates a new Slide against an existing Deck and, if valid,
// returns a copy of Deck with the slide appended at the end. The original
// Deck is never mutated.
func AddSlide(input AddSlideInput) (Deck, Report) {
	report := Report{}
	slide := input.Slide

	if slide.ID == "" {
		report.Errors = append(report.Errors, "slide id is required")
	} else {
		for _, existing := range input.Deck.Slides {
			if existing.ID == slide.ID {
				report.Errors = append(report.Errors, fmt.Sprintf("duplicate slide id: %s", slide.ID))
				break
			}
		}
	}

	if slide.Title == "" {
		report.Errors = append(report.Errors, "slide title is required")
	}

	if slide.Status == "" {
		slide.Status = SlideStatusDraft
	} else if !isValidSlideStatus(slide.Status) {
		report.Errors = append(report.Errors, fmt.Sprintf("invalid status: %s", slide.Status))
	}

	slides := make([]Slide, len(input.Deck.Slides), len(input.Deck.Slides)+1)
	copy(slides, input.Deck.Slides)
	if len(report.Errors) == 0 {
		slides = append(slides, slide)
	}

	updated := Deck{
		Title:    input.Deck.Title,
		Audience: input.Deck.Audience,
		Slides:   slides,
	}
	return updated, report
}
