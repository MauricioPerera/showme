package cli

import (
	"os"

	"github.com/MauricioPerera/showme/internal/export"
	"github.com/MauricioPerera/showme/internal/storage"
)

// ExportProjectCommandInput is the raw data needed to run the
// "project export" CLI command.
type ExportProjectCommandInput struct {
	Path    string
	OutPath string
}

// ExportProjectCommandResult is the JSON-stable result of running the
// "project export" CLI command.
type ExportProjectCommandResult struct {
	OK   bool
	Path string
}

// RunExportProjectCommand loads the Project at Path, renders it with
// export.ExportProjectHTML, and writes the result to OutPath.
//
// A file-system error loading Path or writing OutPath is returned via err.
// Rendering itself cannot fail (see export-project-html), so any error
// here is always an I/O problem.
func RunExportProjectCommand(input ExportProjectCommandInput) (ExportProjectCommandResult, error) {
	proj, err := storage.LoadProject(input.Path)
	if err != nil {
		return ExportProjectCommandResult{}, err
	}

	rendered := export.ExportProjectHTML(proj)

	if err := os.WriteFile(input.OutPath, []byte(rendered), 0o644); err != nil {
		return ExportProjectCommandResult{}, err
	}

	return ExportProjectCommandResult{OK: true, Path: input.OutPath}, nil
}
