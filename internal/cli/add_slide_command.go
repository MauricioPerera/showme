package cli

import (
	"github.com/MauricioPerera/showme/internal/domain"
	"github.com/MauricioPerera/showme/internal/storage"
)

// AddSlideCommandInput is the raw data needed to run the
// "project add-slide" CLI command.
type AddSlideCommandInput struct {
	Path    string
	SlideID string
	Title   string
	Intent  string
	Content string
	Status  string
	OutDir  string
}

// AddSlideCommandResult is the JSON-stable result of running the
// "project add-slide" CLI command.
type AddSlideCommandResult struct {
	OK       bool
	Path     string
	Errors   []string
	Warnings []string
}

// RunAddSlideCommand loads the Project at Path, appends a new slide to its
// Deck via domain.AddSlide, and, if valid, saves the updated Project under
// OutDir (overwriting the same file when OutDir and Name match the
// original).
//
// A file-system error loading Path or saving under OutDir is returned via
// err. Validation problems (missing id/title, duplicate id, invalid
// status) are returned in the result's Errors, with OK false and Path
// empty; the project is left untouched in that case.
func RunAddSlideCommand(input AddSlideCommandInput) (AddSlideCommandResult, error) {
	proj, err := storage.LoadProject(input.Path)
	if err != nil {
		return AddSlideCommandResult{}, err
	}

	updatedDeck, report := domain.AddSlide(domain.AddSlideInput{
		Deck: proj.Deck,
		Slide: domain.Slide{
			ID:      input.SlideID,
			Title:   input.Title,
			Intent:  input.Intent,
			Content: input.Content,
			Status:  domain.SlideStatus(input.Status),
		},
	})

	result := AddSlideCommandResult{
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
			Archived:      proj.Archived,
		},
	})
	if err != nil {
		return AddSlideCommandResult{}, err
	}

	result.Errors = append(result.Errors, saveReport.Errors...)
	result.Warnings = append(result.Warnings, saveReport.Warnings...)
	result.Path = path
	result.OK = len(result.Errors) == 0
	return result, nil
}
