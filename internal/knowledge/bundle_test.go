package knowledge

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsConceptsAndSkipsIndex(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "index.md"), "# Bundle\n")
	writeFile(t, filepath.Join(root, "audience.md"), "---\ntype: Audience\ntitle: Technical audience\ndescription: People who build systems.\ntags: [technical, systems]\n---\n\n# Context\n\nThey value evidence.\n")
	writeFile(t, filepath.Join(root, "slides", "slide-01.md"), "---\ntype: Slide\ntitle: Opening\n---\n\n# Objective\n\nFrame the problem.\n")

	bundle, report := Load(root)
	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if len(bundle.Concepts) != 2 {
		t.Fatalf("expected 2 concepts, got %d", len(bundle.Concepts))
	}
	if bundle.Concepts[0].ID != "audience" || bundle.Concepts[1].ID != "slides/slide-01" {
		t.Fatalf("unexpected concept ids: %#v", bundle.Concepts)
	}
	if bundle.Concepts[0].Tags[0] != "technical" {
		t.Fatalf("expected parsed tags, got %#v", bundle.Concepts[0].Tags)
	}
}

func TestLoadReportsMissingType(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "missing-type.md"), "---\ntitle: Incomplete\n---\n\nBody\n")

	_, report := Load(root)
	if !contains(report.Errors, "missing-type.md: type is required") {
		t.Fatalf("expected missing type error, got %v", report.Errors)
	}
}

func TestLoadPreservesUnknownTypeAndBody(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "metric.md"), "---\ntype: Metric\ntitle: Conversion\n---\n\n# Definition\n\nUnknown types remain readable.\n")

	bundle, report := Load(root)
	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if bundle.Concepts[0].Type != "Metric" || bundle.Concepts[0].Body == "" {
		t.Fatalf("expected type and body to be preserved, got %#v", bundle.Concepts[0])
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
