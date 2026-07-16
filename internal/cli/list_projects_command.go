package cli

import (
	"fmt"

	"github.com/MauricioPerera/showme/internal/storage"
)

// ListProjectsCommandInput is the raw data needed to run the "project list"
// CLI command.
type ListProjectsCommandInput struct {
	Dir string
}

// ProjectSummary is the name and path of one saved project.
type ProjectSummary struct {
	Name string
	Path string
}

// ListProjectsCommandResult is the JSON-stable result of running the
// "project list" CLI command.
type ListProjectsCommandResult struct {
	Projects []ProjectSummary
	Errors   []string
}

// RunListProjectsCommand lists the projects saved under Dir, loading each
// one to report its Name alongside its Path.
//
// A missing Dir is returned via err. A file that fails to load (e.g.
// corrupted JSON) is skipped from Projects and reported in Errors as
// "<path>: <underlying error>", so one broken file does not hide the rest.
func RunListProjectsCommand(input ListProjectsCommandInput) (ListProjectsCommandResult, error) {
	paths, err := storage.ListDecks(input.Dir)
	if err != nil {
		return ListProjectsCommandResult{}, err
	}

	result := ListProjectsCommandResult{}
	for _, path := range paths {
		proj, loadErr := storage.LoadProject(path)
		if loadErr != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", path, loadErr))
			continue
		}
		result.Projects = append(result.Projects, ProjectSummary{Name: proj.Name, Path: path})
	}

	return result, nil
}
