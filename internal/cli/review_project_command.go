package cli

import (
	"github.com/MauricioPerera/showme/internal/domain"
	"github.com/MauricioPerera/showme/internal/storage"
)

// ReviewProjectCommandInput is the raw data needed to run the
// "project review" CLI command.
type ReviewProjectCommandInput struct {
	Path     string
	SlideID  string
	Decision string
	Notes    string
	OutDir   string
}

// ReviewProjectCommandResult is the JSON-stable result of running the
// "project review" CLI command.
type ReviewProjectCommandResult struct {
	OK       bool
	Path     string
	Errors   []string
	Warnings []string
}

// RunReviewProjectCommand loads the Project at Path, applies a review to
// one of its slides via domain.ApplyReview, and, if valid, saves the
// updated Project under OutDir (overwriting the same file when OutDir and
// Name match the original).
//
// A file-system error loading Path or saving under OutDir is returned via
// err. Review validation problems (invalid decision, slide not found) are
// returned in the result's Errors, with OK false and Path empty; the
// project is left untouched in that case.
func RunReviewProjectCommand(input ReviewProjectCommandInput) (ReviewProjectCommandResult, error) {
	proj, err := storage.LoadProject(input.Path)
	if err != nil {
		return ReviewProjectCommandResult{}, err
	}

	updatedDeck, report := domain.ApplyReview(domain.ApplyReviewInput{
		Deck: proj.Deck,
		Review: domain.ReviewInput{
			SlideID:  input.SlideID,
			Decision: domain.ReviewDecision(input.Decision),
			Notes:    input.Notes,
		},
	})

	result := ReviewProjectCommandResult{
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
		return ReviewProjectCommandResult{}, err
	}

	result.Errors = append(result.Errors, saveReport.Errors...)
	result.Warnings = append(result.Warnings, saveReport.Warnings...)
	result.Path = path
	result.OK = len(result.Errors) == 0
	return result, nil
}
