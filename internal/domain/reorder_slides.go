package domain

import "fmt"

// ReorderSlidesInput is the raw data used by ReorderSlides. Order is the
// desired final sequence of slide IDs; it must contain every slide of Deck
// exactly once.
type ReorderSlidesInput struct {
	Deck  Deck
	Order []string
}

// ReorderSlides validates that Order is a permutation of Deck.Slides' IDs
// and, if valid, returns a copy of Deck with its slides in that order. The
// original Deck is never mutated.
func ReorderSlides(input ReorderSlidesInput) (Deck, Report) {
	report := Report{}

	byID := make(map[string]Slide, len(input.Deck.Slides))
	for _, slide := range input.Deck.Slides {
		byID[slide.ID] = slide
	}

	seen := make(map[string]bool, len(input.Order))
	reordered := make([]Slide, 0, len(input.Order))
	for _, id := range input.Order {
		slide, known := byID[id]
		if !known {
			report.Errors = append(report.Errors, fmt.Sprintf("unknown slide id: %s", id))
			continue
		}
		if seen[id] {
			report.Errors = append(report.Errors, fmt.Sprintf("duplicate slide id in order: %s", id))
			continue
		}
		seen[id] = true
		reordered = append(reordered, slide)
	}

	for _, slide := range input.Deck.Slides {
		if !seen[slide.ID] {
			report.Errors = append(report.Errors, fmt.Sprintf("missing slide id in order: %s", slide.ID))
		}
	}

	slides := make([]Slide, len(input.Deck.Slides))
	copy(slides, input.Deck.Slides)
	if len(report.Errors) == 0 {
		slides = reordered
	}

	updated := Deck{
		Title:    input.Deck.Title,
		Audience: input.Deck.Audience,
		Slides:   slides,
	}
	return updated, report
}
