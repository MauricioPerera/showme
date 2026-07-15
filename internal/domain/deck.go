package domain

import "fmt"

// SlideStatus is the review state of a single slide.
type SlideStatus string

const (
	SlideStatusDraft    SlideStatus = "draft"
	SlideStatusAccepted SlideStatus = "accepted"
	SlideStatusRejected SlideStatus = "rejected"
)

// Slide is a single editable unit of a Deck.
type Slide struct {
	ID      string
	Title   string
	Intent  string
	Content string
	Status  SlideStatus
}

// DeckInput is the raw data used to build a Deck.
type DeckInput struct {
	Title    string
	Audience string
	Slides   []Slide
}

// Deck is an ordered collection of slides with a shared title and audience.
type Deck struct {
	Title    string
	Audience string
	Slides   []Slide
}

// Report collects the problems found while building a Deck.
type Report struct {
	Errors   []string
	Warnings []string
}

func isValidSlideStatus(status SlideStatus) bool {
	switch status {
	case SlideStatusDraft, SlideStatusAccepted, SlideStatusRejected:
		return true
	}
	return false
}

func validateSlide(index int, slide Slide, seen map[string]bool, report *Report) Slide {
	if slide.ID == "" {
		report.Errors = append(report.Errors, fmt.Sprintf("slide[%d]: id is required", index))
	} else if seen[slide.ID] {
		report.Errors = append(report.Errors, fmt.Sprintf("duplicate slide id: %s", slide.ID))
	} else {
		seen[slide.ID] = true
	}

	if slide.Title == "" {
		report.Errors = append(report.Errors, fmt.Sprintf("slide[%d]: title is required", index))
	}

	if slide.Status == "" {
		slide.Status = SlideStatusDraft
	} else if !isValidSlideStatus(slide.Status) {
		report.Errors = append(report.Errors, fmt.Sprintf("slide[%d]: invalid status: %s", index, slide.Status))
	}

	return slide
}

// NewDeck builds a Deck from input, enforcing its structural invariants.
func NewDeck(input DeckInput) (Deck, Report) {
	report := Report{}

	if input.Title == "" {
		report.Errors = append(report.Errors, "title is required")
	}

	if len(input.Slides) == 0 {
		report.Errors = append(report.Errors, "at least one slide is required")
	}

	seen := make(map[string]bool, len(input.Slides))
	slides := make([]Slide, len(input.Slides))
	for i, slide := range input.Slides {
		slides[i] = validateSlide(i, slide, seen, &report)
	}

	deck := Deck{
		Title:    input.Title,
		Audience: input.Audience,
		Slides:   slides,
	}
	return deck, report
}
