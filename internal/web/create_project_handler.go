package web

import (
	"encoding/json"
	"os"

	"github.com/MauricioPerera/showme/internal/cli"
)

// CreateProjectFormInput is the raw data submitted by the "create project"
// web form.
type CreateProjectFormInput struct {
	Name          string
	DesignPath    string
	KnowledgeRoot string
	DeckTitle     string
	DeckAudience  string
	SlideTitle    string
	SlideIntent   string
	Dir           string
}

// CreateProjectFormResult is the outcome of handling the "create project"
// web form, used to render either a success or an error page.
type CreateProjectFormResult struct {
	OK       bool
	Path     string
	Errors   []string
	Warnings []string
}

type formDeckSlide struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Intent string `json:"intent"`
}

type formDeckInput struct {
	Title    string          `json:"title"`
	Audience string          `json:"audience"`
	Slides   []formDeckSlide `json:"slides"`
}

// HandleCreateProjectForm turns the submitted form fields into a deck JSON
// file (the same shape cli.RunCreateProjectCommand's --deck flag expects),
// writes it to a temporary file, and delegates to
// cli.RunCreateProjectCommand to validate, assemble and persist the
// Project under Dir. The temporary deck file is always removed before
// returning.
//
// A file-system error (reading DesignPath, writing the temp deck file,
// saving under Dir) is returned via err. Validation problems are returned
// in the result's Errors, with OK false and Path empty; nothing is
// persisted in that case.
func HandleCreateProjectForm(input CreateProjectFormInput) (CreateProjectFormResult, error) {
	deck := formDeckInput{
		Title:    input.DeckTitle,
		Audience: input.DeckAudience,
		Slides: []formDeckSlide{
			{ID: "slide-1", Title: input.SlideTitle, Intent: input.SlideIntent},
		},
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
