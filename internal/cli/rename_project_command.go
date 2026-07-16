package cli

import (
	"github.com/MauricioPerera/showme/internal/project"
)

// RenameProjectCommandInput is the raw data needed to run the
// "project rename" CLI command.
type RenameProjectCommandInput struct {
	SourcePath string
	NewName    string
	OutDir     string
}

// RenameProjectCommandResult is the JSON-stable result of running the
// "project rename" CLI command.
type RenameProjectCommandResult struct {
	OK       bool
	Path     string
	Errors   []string
	Warnings []string
}

// RunRenameProjectCommand wraps project.RenameProject for the CLI.
//
// A file-system error (missing source, I/O failure while saving or
// removing the old file) is returned via err. Validation problems (empty
// NewName, a name collision with another project) are returned in the
// result's Errors, with OK false and Path empty.
func RunRenameProjectCommand(input RenameProjectCommandInput) (RenameProjectCommandResult, error) {
	path, report, err := project.RenameProject(project.RenameProjectInput{
		SourcePath: input.SourcePath,
		NewName:    input.NewName,
		Dir:        input.OutDir,
	})
	if err != nil {
		return RenameProjectCommandResult{}, err
	}

	return RenameProjectCommandResult{
		OK:       len(report.Errors) == 0,
		Path:     path,
		Errors:   append([]string{}, report.Errors...),
		Warnings: append([]string{}, report.Warnings...),
	}, nil
}
