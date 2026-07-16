package cli

import (
	"github.com/MauricioPerera/showme/internal/domain"
	"github.com/MauricioPerera/showme/internal/storage"
)

// RemoveSlideCommandInput is the raw data needed to run the
// "project remove-slide" CLI command.
type RemoveSlideCommandInput struct {
	Path    string
	SlideID string
	OutDir  string
}

// RemoveSlideCommandResult is the JSON-stable result of running the
// "project remove-slide" CLI command.
type RemoveSlideCommandResult struct {
	OK       bool
	Path     string
	Errors   []string
	Warnings []string
}

// RunRemoveSlideCommand loads the Project at Path, removes a slide from its
// Deck via domain.RemoveSlide, and, if valid, saves the updated Project
// under OutDir (overwriting the same file when OutDir and Name match the
// original).
//
// A file-system error loading Path or saving under OutDir is returned via
// err. Validation problems (slide not found, removing the only slide) are
// returned in the result's Errors, with OK false and Path empty; the
// project is left untouched in that case.
func RunRemoveSlideCommand(input RemoveSlideCommandInput) (RemoveSlideCommandResult, error) {
	proj, err := storage.LoadProject(input.Path)
	if err != nil {
		return RemoveSlideCommandResult{}, err
	}

	updatedDeck, report := domain.RemoveSlide(domain.RemoveSlideInput{
		Deck:    proj.Deck,
		SlideID: input.SlideID,
	})

	result := RemoveSlideCommandResult{
		Errors:   append([]string{}, report.Errors...),
		Warnings: append([]string{}, report.Warnings...),
	}
	if len(report.Errors) != 0 {
		return result, nil
	}

	path, saveReport, err := storage.SaveProject(storage.SaveProjectRequest{
		Dir: input.OutDir,
		Input: domain.ProjectInput{
			Name:          proj.Name,
			Deck:          updatedDeck,
			DesignPath:    proj.DesignPath,
			KnowledgePath: proj.KnowledgePath,
			Version:       proj.Version,
		},
	})
	if err != nil {
		return RemoveSlideCommandResult{}, err
	}

	result.Errors = append(result.Errors, saveReport.Errors...)
	result.Warnings = append(result.Warnings, saveReport.Warnings...)
	result.Path = path
	result.OK = len(result.Errors) == 0
	return result, nil
}
