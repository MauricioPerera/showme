package domain

import "testing"

func TestAppendGenerationRun_Valid(t *testing.T) {
	proj := validProjectForArchiving(t)
	run, runReport := NewGenerationRun(GenerationRunInput{
		SlideID:   "intro",
		Model:     "test-model",
		Output:    "Contenido generado.",
		CreatedAt: "2026-07-16T12:00:00Z",
	})
	if len(runReport.Errors) != 0 {
		t.Fatalf("setup: unexpected run errors: %v", runReport.Errors)
	}

	updated := AppendGenerationRun(AppendGenerationRunInput{Project: proj, Run: run})

	if len(updated.Runs) != 1 || updated.Runs[0].SlideID != "intro" {
		t.Fatalf("expected run appended, got %+v", updated.Runs)
	}
}

func TestAppendGenerationRun_AppendsToExisting(t *testing.T) {
	proj := validProjectForArchiving(t)
	first, _ := NewGenerationRun(GenerationRunInput{SlideID: "intro", Model: "m1", Output: "o1", CreatedAt: "2026-07-16T12:00:00Z"})
	proj = AppendGenerationRun(AppendGenerationRunInput{Project: proj, Run: first})

	second, _ := NewGenerationRun(GenerationRunInput{SlideID: "intro", Model: "m2", Output: "o2", CreatedAt: "2026-07-16T13:00:00Z"})
	updated := AppendGenerationRun(AppendGenerationRunInput{Project: proj, Run: second})

	if len(updated.Runs) != 2 || updated.Runs[0].Model != "m1" || updated.Runs[1].Model != "m2" {
		t.Fatalf("expected both runs in order, got %+v", updated.Runs)
	}
}

func TestAppendGenerationRun_DoesNotMutateOriginal(t *testing.T) {
	proj := validProjectForArchiving(t)
	run, _ := NewGenerationRun(GenerationRunInput{SlideID: "intro", Model: "m", Output: "o", CreatedAt: "2026-07-16T12:00:00Z"})

	_ = AppendGenerationRun(AppendGenerationRunInput{Project: proj, Run: run})

	if len(proj.Runs) != 0 {
		t.Fatalf("expected original project untouched, got %+v", proj.Runs)
	}
}
