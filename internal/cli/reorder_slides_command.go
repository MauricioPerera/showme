package cli

import (
	"github.com/MauricioPerera/showme/internal/domain"
	"github.com/MauricioPerera/showme/internal/storage"
)

// ReorderSlidesCommandInput is the raw data needed to run the
// "project reorder-slides" CLI command.
type ReorderSlidesCommandInput struct {
	Path   string
	Order  []string
	OutDir string
}

// ReorderSlidesCommandResult is the JSON-stable result of running the
// "project reorder-slides" CLI command.
type ReorderSlidesCommandResult struct {
	OK       bool
	Path     string
	Errors   []string
	Warnings []string
}

// RunReorderSlidesCommand loads the Project at Path, reorders its Deck's
// slides via domain.ReorderSlides, and, if valid, saves the updated
// Project under OutDir (overwriting the same file when OutDir and Name
// match the original).
//
// A file-system error loading Path or saving under OutDir is returned via
// err. Validation problems (incomplete, unknown or duplicate slide ids in
// Order) are returned in the result's Errors, with OK false and Path
// empty; the project is left untouched in that case.
func RunReorderSlidesCommand(input ReorderSlidesCommandInput) (ReorderSlidesCommandResult, error) {
	proj, err := storage.LoadProject(input.Path)
	if err != nil {
		return ReorderSlidesCommandResult{}, err
	}

	updatedDeck, report := domain.ReorderSlides(domain.ReorderSlidesInput{
		Deck:  proj.Deck,
		Order: input.Order,
	})

	result := ReorderSlidesCommandResult{
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
			Runs:          proj.Runs,
		},
	})
	if err != nil {
		return ReorderSlidesCommandResult{}, err
	}

	result.Errors = append(result.Errors, saveReport.Errors...)
	result.Warnings = append(result.Warnings, saveReport.Warnings...)
	result.Path = path
	result.OK = len(result.Errors) == 0
	return result, nil
}
