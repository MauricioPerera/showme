package cli

import (
	"fmt"

	"github.com/MauricioPerera/showme/internal/storage"
)

// GenerateAllSlidesCommandInput is the raw data needed to run the
// "project generate-all" CLI command.
type GenerateAllSlidesCommandInput struct {
	Path    string
	BaseURL string
	Model   string
	OutDir  string
}

// GenerateAllSlidesCommandResult is the JSON-stable result of running the
// "project generate-all" CLI command.
type GenerateAllSlidesCommandResult struct {
	OK        bool
	Generated []string
	Skipped   []string
	Errors    []string
}

// RunGenerateAllSlidesCommand loads the Project at Path and, for every
// slide whose Content is empty, calls RunGenerateSlideContentCommand to
// generate it. Slides that already have Content are left untouched and
// reported in Skipped. A failure generating one slide is recorded in
// Errors as "<slideID>: <message>" and does not stop the remaining
// slides from being attempted.
//
// A file-system error loading Path is returned via err. OK is true only
// when every attempted slide generated successfully (Errors is empty);
// having Skipped entries does not affect OK.
func RunGenerateAllSlidesCommand(input GenerateAllSlidesCommandInput) (GenerateAllSlidesCommandResult, error) {
	proj, err := storage.LoadProject(input.Path)
	if err != nil {
		return GenerateAllSlidesCommandResult{}, err
	}

	result := GenerateAllSlidesCommandResult{}
	for _, slide := range proj.Deck.Slides {
		if slide.Content != "" {
			result.Skipped = append(result.Skipped, slide.ID)
			continue
		}

		genResult, genErr := RunGenerateSlideContentCommand(GenerateSlideContentCommandInput{
			Path:    input.Path,
			SlideID: slide.ID,
			BaseURL: input.BaseURL,
			Model:   input.Model,
			OutDir:  input.OutDir,
		})
		if genErr != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", slide.ID, genErr))
			continue
		}
		if !genResult.OK {
			for _, e := range genResult.Errors {
				result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", slide.ID, e))
			}
			continue
		}

		result.Generated = append(result.Generated, slide.ID)
	}

	result.OK = len(result.Errors) == 0
	return result, nil
}
