package project

import (
	"os"
	"path/filepath"

	"github.com/MauricioPerera/showme/internal/domain"
	"github.com/MauricioPerera/showme/internal/storage"
)

// RenameProjectInput is the raw data used to rename a Project in place.
type RenameProjectInput struct {
	SourcePath string
	NewName    string
	Dir        string
}

// RenameProject loads the Project at SourcePath and saves it under Dir with
// NewName, preserving its Deck, DesignPath, KnowledgePath, Version and
// Archived state (unlike DuplicateProject, which resets Version). If the
// new name resolves to the same file as SourcePath, this is a no-op that
// still succeeds. If it resolves to a different, already-existing file, the
// rename is refused to avoid silently overwriting another project; neither
// file is touched in that case. Otherwise, SourcePath is removed once the
// renamed file has been written successfully.
func RenameProject(input RenameProjectInput) (string, domain.Report, error) {
	source, err := storage.LoadProject(input.SourcePath)
	if err != nil {
		return "", domain.Report{}, err
	}

	targetPath := filepath.Join(input.Dir, storage.Slugify(input.NewName)+".json")
	if storage.Slugify(input.NewName) != "" && targetPath != input.SourcePath {
		if _, statErr := os.Stat(targetPath); statErr == nil {
			return "", domain.Report{Errors: []string{"a project already exists at that name"}}, nil
		}
	}

	newPath, report, err := storage.SaveProject(storage.SaveProjectRequest{
		Dir: input.Dir,
		Input: domain.ProjectInput{
			Name:          input.NewName,
			Deck:          source.Deck,
			DesignPath:    source.DesignPath,
			KnowledgePath: source.KnowledgePath,
			Version:       source.Version,
			Archived:      source.Archived,
		},
	})
	if err != nil {
		return "", report, err
	}
	if len(report.Errors) != 0 {
		return "", report, nil
	}

	if newPath != input.SourcePath {
		if err := os.Remove(input.SourcePath); err != nil {
			return "", report, err
		}
	}

	return newPath, report, nil
}
