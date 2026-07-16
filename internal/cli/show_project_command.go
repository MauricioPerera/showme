package cli

import (
	"github.com/MauricioPerera/showme/internal/domain"
	"github.com/MauricioPerera/showme/internal/storage"
)

// ShowProjectCommandInput is the raw data needed to run the "project show"
// CLI command.
type ShowProjectCommandInput struct {
	Path string
}

// ShowProjectCommandResult is the JSON-stable result of running the
// "project show" CLI command.
type ShowProjectCommandResult struct {
	Project domain.Project
}

// RunShowProjectCommand loads the Project at Path and returns it in full.
//
// A missing file or invalid JSON content is returned via err.
func RunShowProjectCommand(input ShowProjectCommandInput) (ShowProjectCommandResult, error) {
	proj, err := storage.LoadProject(input.Path)
	if err != nil {
		return ShowProjectCommandResult{}, err
	}

	return ShowProjectCommandResult{Project: proj}, nil
}
