package storage

import (
	"os"
	"path/filepath"
	"strings"
)

// ListDecks returns the sorted paths of the deck JSON files directly under dir.
func ListDecks(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		paths = append(paths, filepath.Join(dir, entry.Name()))
	}

	return paths, nil
}
