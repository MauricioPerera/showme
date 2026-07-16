package web

import (
	"path/filepath"
	"testing"
)

func TestProjectFilePath_Valid(t *testing.T) {
	path, err := ProjectFilePath("/data", "roadmap-q3")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join("/data", "roadmap-q3.json")
	if path != want {
		t.Fatalf("expected %q, got %q", want, path)
	}
}

func TestProjectFilePath_EmptySlug(t *testing.T) {
	_, err := ProjectFilePath("/data", "")

	if err == nil {
		t.Fatalf("expected an error for an empty slug")
	}
}

func TestProjectFilePath_RejectsPathSeparators(t *testing.T) {
	for _, slug := range []string{"a/b", "a\\b", "../secret", "..", "a/../b"} {
		if _, err := ProjectFilePath("/data", slug); err == nil {
			t.Fatalf("expected an error for slug %q", slug)
		}
	}
}
