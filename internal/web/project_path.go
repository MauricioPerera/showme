package web

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ProjectFilePath resolves a URL-supplied project slug to the file path
// under dir where that project would be saved (same convention as
// storage.Slugify: "<dir>/<slug>.json").
//
// slug is validated to be a plain filename component: empty, containing a
// path separator ("/" or "\"), or containing ".." is rejected. This is the
// only thing standing between an HTTP request path and the filesystem, so
// it never trusts the caller to have sanitized it already.
func ProjectFilePath(dir, slug string) (string, error) {
	if slug == "" {
		return "", fmt.Errorf("slug is required")
	}
	if strings.ContainsAny(slug, "/\\") || strings.Contains(slug, "..") {
		return "", fmt.Errorf("invalid slug: %s", slug)
	}
	return filepath.Join(dir, slug+".json"), nil
}
