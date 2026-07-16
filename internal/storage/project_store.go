package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/MauricioPerera/showme/internal/domain"
)

// SaveProjectRequest is the input to SaveProject.
type SaveProjectRequest struct {
	Dir   string
	Input domain.ProjectInput
}

// SaveProject builds a Project from the request and persists it as JSON
// under Dir.
//
// An invalid project (per domain.NewProject) is never written to disk; its
// problems are returned in Report instead. A path error while writing to
// disk is returned via err, not Report.
func SaveProject(request SaveProjectRequest) (string, domain.Report, error) {
	proj, report := domain.NewProject(request.Input)
	if len(report.Errors) != 0 {
		return "", report, nil
	}

	slug := slugify(proj.Name)
	if slug == "" {
		report.Errors = append(report.Errors, "name produces an empty slug")
		return "", report, nil
	}

	data, err := json.MarshalIndent(proj, "", "  ")
	if err != nil {
		return "", report, err
	}

	path := filepath.Join(request.Dir, slug+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", report, err
	}

	return path, report, nil
}
