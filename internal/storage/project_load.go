package storage

import (
	"encoding/json"
	"os"

	"github.com/MauricioPerera/showme/internal/domain"
)

// LoadProject reads the JSON file at path and deserializes it into a domain.Project.
func LoadProject(path string) (domain.Project, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.Project{}, err
	}

	var proj domain.Project
	if err := json.Unmarshal(data, &proj); err != nil {
		return domain.Project{}, err
	}

	return proj, nil
}
