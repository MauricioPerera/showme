package cli

import (
	"github.com/MauricioPerera/showme/internal/project"
)

// DuplicateProjectCommandInput is the raw data needed to run the
// "project duplicate" CLI command.
type DuplicateProjectCommandInput struct {
	SourcePath string
	NewName    string
	OutDir     string
}

// DuplicateProjectCommandResult is the JSON-stable result of running the
// "project duplicate" CLI command.
type DuplicateProjectCommandResult struct {
	OK       bool
	Path     string
	Errors   []string
	Warnings []string
}

// RunDuplicateProjectCommand wraps project.DuplicateProject for the CLI.
//
// A file-system error reading SourcePath or writing under OutDir is
// returned via err. Validation problems (e.g. an empty NewName) are
// returned in the result's Errors, with OK false and Path empty.
func RunDuplicateProjectCommand(input DuplicateProjectCommandInput) (DuplicateProjectCommandResult, error) {
	path, report, err := project.DuplicateProject(project.DuplicateProjectInput{
		SourcePath: input.SourcePath,
		NewName:    input.NewName,
		Dir:        input.OutDir,
	})
	if err != nil {
		return DuplicateProjectCommandResult{}, err
	}

	return DuplicateProjectCommandResult{
		OK:       len(report.Errors) == 0,
		Path:     path,
		Errors:   append([]string{}, report.Errors...),
		Warnings: append([]string{}, report.Warnings...),
	}, nil
}
