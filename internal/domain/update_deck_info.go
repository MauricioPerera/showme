package domain

// UpdateDeckInfoInput is the raw data used by UpdateDeckInfo.
type UpdateDeckInfoInput struct {
	Deck     Deck
	Title    string
	Audience string
}

// UpdateDeckInfo validates a new Title/Audience for an existing Deck and,
// if valid, returns a copy of Deck with those fields replaced. Slides are
// always preserved untouched. The original Deck is never mutated.
func UpdateDeckInfo(input UpdateDeckInfoInput) (Deck, Report) {
	report := Report{}

	if input.Title == "" {
		report.Errors = append(report.Errors, "title is required")
	}

	slides := make([]Slide, len(input.Deck.Slides))
	copy(slides, input.Deck.Slides)

	if len(report.Errors) != 0 {
		return Deck{
			Title:    input.Deck.Title,
			Audience: input.Deck.Audience,
			Slides:   slides,
		}, report
	}

	updated := Deck{
		Title:    input.Title,
		Audience: input.Audience,
		Slides:   slides,
	}
	return updated, report
}
