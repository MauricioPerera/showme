package cli

import (
	"github.com/MauricioPerera/showme/internal/domain"
	"github.com/MauricioPerera/showme/internal/storage"
)

// UpdateDeckInfoCommandInput is the raw data needed to run the
// "project update-info" CLI command.
type UpdateDeckInfoCommandInput struct {
	Path     string
	Title    string
	Audience string
	OutDir   string
}

// UpdateDeckInfoCommandResult is the JSON-stable result of running the
// "project update-info" CLI command.
type UpdateDeckInfoCommandResult struct {
	OK       bool
	Path     string
	Errors   []string
	Warnings []string
}

// RunUpdateDeckInfoCommand loads the Project at Path, replaces its Deck's
// Title/Audience via domain.UpdateDeckInfo, and, if valid, saves the
// updated Project under OutDir (overwriting the same file when OutDir and
// Name match the original). The project's own Name is left untouched.
//
// A file-system error loading Path or saving under OutDir is returned via
// err. Validation problems (empty Title) are returned in the result's
// Errors, with OK false and Path empty; the project is left untouched in
// that case.
func RunUpdateDeckInfoCommand(input UpdateDeckInfoCommandInput) (UpdateDeckInfoCommandResult, error) {
	proj, err := storage.LoadProject(input.Path)
	if err != nil {
		return UpdateDeckInfoCommandResult{}, err
	}

	updatedDeck, report := domain.UpdateDeckInfo(domain.UpdateDeckInfoInput{
		Deck:     proj.Deck,
		Title:    input.Title,
		Audience: input.Audience,
	})

	result := UpdateDeckInfoCommandResult{
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
		return UpdateDeckInfoCommandResult{}, err
	}

	result.Errors = append(result.Errors, saveReport.Errors...)
	result.Warnings = append(result.Warnings, saveReport.Warnings...)
	result.Path = path
	result.OK = len(result.Errors) == 0
	return result, nil
}
