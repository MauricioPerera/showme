package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/MauricioPerera/showme/internal/domain"
)

var nonSlugChars = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify converts an arbitrary title/name into the deterministic
// lowercase-hyphenated slug used for a saved file's basename.
func Slugify(title string) string {
	lower := strings.ToLower(title)
	slug := nonSlugChars.ReplaceAllString(lower, "-")
	return strings.Trim(slug, "-")
}

// SaveDeckRequest is the input to SaveDeck.
type SaveDeckRequest struct {
	Dir   string
	Input domain.DeckInput
}

// SaveDeck builds a Deck from the request and persists it as JSON under Dir.
//
// An invalid deck (per domain.NewDeck) is never written to disk; its
// problems are returned in Report instead. A path error while writing to
// disk is returned via err, not Report.
func SaveDeck(request SaveDeckRequest) (string, domain.Report, error) {
	deck, report := domain.NewDeck(request.Input)
	if len(report.Errors) != 0 {
		return "", report, nil
	}

	slug := Slugify(deck.Title)
	if slug == "" {
		report.Errors = append(report.Errors, "title produces an empty slug")
		return "", report, nil
	}

	data, err := json.MarshalIndent(deck, "", "  ")
	if err != nil {
		return "", report, err
	}

	path := filepath.Join(request.Dir, slug+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", report, err
	}

	return path, report, nil
}
