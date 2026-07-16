package web

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/MauricioPerera/showme/internal/cli"
	"github.com/MauricioPerera/showme/internal/storage"
)

// SlideInput is one slide submitted by the multi-slide "create project"
// web flow (see propose-storyboard-usecase).
type SlideInput struct {
	Title  string
	Intent string
}

// CreateProjectWithSlidesInput is the raw data submitted by the multi-slide
// "create project" web form.
type CreateProjectWithSlidesInput struct {
	Name          string
	DesignPath    string
	KnowledgeRoot string
	DeckTitle     string
	DeckAudience  string
	Slides        []SlideInput
	Dir           string
}

func uniqueSlideID(seen map[string]bool, title string) string {
	base := storage.Slugify(title)
	if base == "" {
		base = "slide"
	}
	id := base
	for suffix := 2; seen[id]; suffix++ {
		id = base + "-" + strconv.Itoa(suffix)
	}
	seen[id] = true
	return id
}

// HandleCreateProjectWithSlides turns the submitted form fields (including
// an arbitrary number of slides, e.g. from a reviewed AI-proposed
// storyboard) into a deck JSON file and delegates to
// cli.RunCreateProjectCommand, same pattern as HandleCreateProjectForm but
// for N slides instead of exactly one. Each slide gets a deterministic id
// via storage.Slugify, deduplicated with a numeric suffix on collision. The
// temporary deck file is always removed before returning.
//
// A file-system error is returned via err. Validation problems (including
// an empty Slides list, which domain.NewDeck rejects) are returned in the
// result's Errors, with OK false and Path empty.
func HandleCreateProjectWithSlides(input CreateProjectWithSlidesInput) (CreateProjectFormResult, error) {
	seen := map[string]bool{}
	slides := make([]formDeckSlide, len(input.Slides))
	for i, s := range input.Slides {
		slides[i] = formDeckSlide{ID: uniqueSlideID(seen, s.Title), Title: s.Title, Intent: s.Intent}
	}

	deck := formDeckInput{
		Title:    input.DeckTitle,
		Audience: input.DeckAudience,
		Slides:   slides,
	}
	encoded, err := json.Marshal(deck)
	if err != nil {
		return CreateProjectFormResult{}, err
	}

	tempDeckFile, err := os.CreateTemp("", "showme-web-deck-*.json")
	if err != nil {
		return CreateProjectFormResult{}, err
	}
	deckPath := tempDeckFile.Name()
	defer os.Remove(deckPath)

	if _, err := tempDeckFile.Write(encoded); err != nil {
		_ = tempDeckFile.Close()
		return CreateProjectFormResult{}, err
	}
	if err := tempDeckFile.Close(); err != nil {
		return CreateProjectFormResult{}, err
	}

	result, err := cli.RunCreateProjectCommand(cli.CreateProjectCommandInput{
		Name:          input.Name,
		DesignPath:    input.DesignPath,
		KnowledgeRoot: input.KnowledgeRoot,
		DeckPath:      deckPath,
		OutDir:        input.Dir,
	})
	if err != nil {
		return CreateProjectFormResult{}, err
	}

	return CreateProjectFormResult{
		OK:       result.OK,
		Path:     result.Path,
		Errors:   result.Errors,
		Warnings: result.Warnings,
	}, nil
}
