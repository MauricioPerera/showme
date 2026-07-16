package project

import (
	"github.com/MauricioPerera/showme/internal/domain"
	"github.com/MauricioPerera/showme/internal/storage"
)

// DuplicateProjectInput is the raw data used to duplicate a Project.
type DuplicateProjectInput struct {
	SourcePath string
	NewName    string
	Dir        string
}

// DuplicateProject loads the Project at SourcePath and saves a copy under
// Dir with NewName and its Version reset to the default (1). The source
// file is never modified.
func DuplicateProject(input DuplicateProjectInput) (string, domain.Report, error) {
	source, err := storage.LoadProject(input.SourcePath)
	if err != nil {
		return "", domain.Report{}, err
	}

	return storage.SaveProject(storage.SaveProjectRequest{
		Dir: input.Dir,
		Input: domain.ProjectInput{
			Name:          input.NewName,
			Deck:          source.Deck,
			DesignPath:    source.DesignPath,
			KnowledgePath: source.KnowledgePath,
		},
	})
}
