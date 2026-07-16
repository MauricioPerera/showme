package domain

import "fmt"

// RemoveSlideInput is the raw data used by RemoveSlide.
type RemoveSlideInput struct {
	Deck    Deck
	SlideID string
}

// RemoveSlide validates that SlideID exists in Deck and that removing it
// would not leave the deck empty, then returns a copy of Deck without that
// slide. The original Deck is never mutated.
func RemoveSlide(input RemoveSlideInput) (Deck, Report) {
	report := Report{}

	index := -1
	for i, slide := range input.Deck.Slides {
		if slide.ID == input.SlideID {
			index = i
			break
		}
	}

	if index == -1 {
		report.Errors = append(report.Errors, fmt.Sprintf("slide not found: %s", input.SlideID))
	} else if len(input.Deck.Slides) == 1 {
		report.Errors = append(report.Errors, "deck must have at least one slide")
	}

	slides := make([]Slide, len(input.Deck.Slides))
	copy(slides, input.Deck.Slides)
	if len(report.Errors) == 0 {
		slides = append(slides[:index], slides[index+1:]...)
	}

	updated := Deck{
		Title:    input.Deck.Title,
		Audience: input.Deck.Audience,
		Slides:   slides,
	}
	return updated, report
}
