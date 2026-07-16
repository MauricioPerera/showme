package cli

import (
	"github.com/MauricioPerera/showme/internal/domain"
	"github.com/MauricioPerera/showme/internal/storage"
)

// ArchiveProjectCommandInput is the raw data needed to run the
// "project archive"/"project unarchive" CLI commands.
type ArchiveProjectCommandInput struct {
	Path     string
	Archived bool
	OutDir   string
}

// ArchiveProjectCommandResult is the JSON-stable result of running the
// "project archive"/"project unarchive" CLI commands.
type ArchiveProjectCommandResult struct {
	OK       bool
	Path     string
	Errors   []string
	Warnings []string
}

// RunArchiveProjectCommand loads the Project at Path, sets its Archived
// field via domain.SetProjectArchived, and saves it back under OutDir
// (overwriting the same file when OutDir and Name match the original).
//
// A file-system error loading Path or saving under OutDir is returned via
// err. Setting Archived cannot itself fail (see set-project-archived-
// usecase), so a non-nil err always means an I/O problem, never a
// validation one.
func RunArchiveProjectCommand(input ArchiveProjectCommandInput) (ArchiveProjectCommandResult, error) {
	proj, err := storage.LoadProject(input.Path)
	if err != nil {
		return ArchiveProjectCommandResult{}, err
	}

	updated := domain.SetProjectArchived(domain.SetProjectArchivedInput{
		Project:  proj,
		Archived: input.Archived,
	})

	path, saveReport, err := storage.SaveProject(storage.SaveProjectRequest{
		Dir: input.OutDir,
		Input: domain.ProjectInput{
			Name:          updated.Name,
			Deck:          updated.Deck,
			DesignPath:    updated.DesignPath,
			KnowledgePath: updated.KnowledgePath,
			Version:       updated.Version,
			Archived:      updated.Archived,
			Runs:          updated.Runs,
		},
	})
	if err != nil {
		return ArchiveProjectCommandResult{}, err
	}

	return ArchiveProjectCommandResult{
		OK:       len(saveReport.Errors) == 0,
		Path:     path,
		Errors:   append([]string{}, saveReport.Errors...),
		Warnings: append([]string{}, saveReport.Warnings...),
	}, nil
}
