package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/MauricioPerera/showme/internal/domain"
	"github.com/MauricioPerera/showme/internal/storage"
)

func saveNamedProject(t *testing.T, dir, name string) string {
	t.Helper()
	deck, deckReport := domain.NewDeck(domain.DeckInput{
		Title:  "Roadmap",
		Slides: []domain.Slide{{ID: "intro", Title: "Introduccion"}},
	})
	if len(deckReport.Errors) != 0 {
		t.Fatalf("setup: unexpected deck errors: %v", deckReport.Errors)
	}

	path, report, err := storage.SaveProject(storage.SaveProjectRequest{
		Dir: dir,
		Input: domain.ProjectInput{
			Name:          name,
			Deck:          deck,
			DesignPath:    "DESIGN.md",
			KnowledgePath: "knowledge/showme",
		},
	})
	if err != nil || len(report.Errors) != 0 {
		t.Fatalf("setup: unexpected SaveProject failure: err=%v report=%v", err, report)
	}
	return path
}

func TestRunListProjectsCommand_ReturnsNamesAndPaths(t *testing.T) {
	dir := t.TempDir()
	pathA := saveNamedProject(t, dir, "Roadmap Q3")
	pathB := saveNamedProject(t, dir, "Onboarding")

	result, err := RunListProjectsCommand(ListProjectsCommandInput{Dir: dir})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Projects) != 2 {
		t.Fatalf("expected 2 projects, got %+v", result.Projects)
	}

	byPath := map[string]string{}
	for _, p := range result.Projects {
		byPath[p.Path] = p.Name
	}
	if byPath[pathA] != "Roadmap Q3" {
		t.Fatalf("expected %q for %q, got %+v", "Roadmap Q3", pathA, result.Projects)
	}
	if byPath[pathB] != "Onboarding" {
		t.Fatalf("expected %q for %q, got %+v", "Onboarding", pathB, result.Projects)
	}
}

func TestRunListProjectsCommand_ReflectsArchivedState(t *testing.T) {
	dir := t.TempDir()
	path := saveNamedProject(t, dir, "Roadmap Q3")

	archiveResult, err := RunArchiveProjectCommand(ArchiveProjectCommandInput{
		Path: path, Archived: true, OutDir: dir,
	})
	if err != nil || !archiveResult.OK {
		t.Fatalf("setup: unexpected archive failure: err=%v result=%+v", err, archiveResult)
	}

	result, err := RunListProjectsCommand(ListProjectsCommandInput{Dir: dir})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Projects) != 1 || !result.Projects[0].Archived {
		t.Fatalf("expected the listed project to be archived, got %+v", result.Projects)
	}
}

func TestRunListProjectsCommand_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	result, err := RunListProjectsCommand(ListProjectsCommandInput{Dir: dir})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Projects) != 0 {
		t.Fatalf("expected no projects, got %+v", result.Projects)
	}
}

func TestRunListProjectsCommand_MissingDir(t *testing.T) {
	dir := t.TempDir()

	_, err := RunListProjectsCommand(ListProjectsCommandInput{Dir: filepath.Join(dir, "does-not-exist")})

	if err == nil {
		t.Fatalf("expected an error for a missing directory")
	}
}

func TestRunListProjectsCommand_SkipsUnreadableFile(t *testing.T) {
	dir := t.TempDir()
	pathA := saveNamedProject(t, dir, "Roadmap Q3")
	writeFile(t, filepath.Join(dir, "broken.json"), "{not json")

	result, err := RunListProjectsCommand(ListProjectsCommandInput{Dir: dir})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Projects) != 1 || result.Projects[0].Path != pathA {
		t.Fatalf("expected only the readable project, got %+v", result.Projects)
	}
	wantPrefix := filepath.Join(dir, "broken.json") + ": "
	found := false
	for _, e := range result.Errors {
		if strings.HasPrefix(e, wantPrefix) {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a parse error naming %q, got %v", wantPrefix, result.Errors)
	}
}
