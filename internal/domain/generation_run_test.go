package domain

import "testing"

func TestNewGenerationRun_Valid(t *testing.T) {
	run, report := NewGenerationRun(GenerationRunInput{
		SlideID:   "intro",
		Model:     "Ternary-Bonsai-27B-Q2_0.gguf",
		Provider:  "http://127.0.0.1:8080/v1",
		Intent:    "Dar la bienvenida",
		Context:   "Contexto seleccionado.",
		Output:    "Contenido generado.",
		CreatedAt: "2026-07-16T12:00:00Z",
	})

	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if run.SlideID != "intro" || run.Output != "Contenido generado." {
		t.Fatalf("expected fields preserved, got %+v", run)
	}
}

func TestNewGenerationRun_EmptySlideID(t *testing.T) {
	_, report := NewGenerationRun(GenerationRunInput{
		Model:     "m",
		Output:    "o",
		CreatedAt: "2026-07-16T12:00:00Z",
	})

	if !containsError(report.Errors, "slide id is required") {
		t.Fatalf("expected 'slide id is required' error, got %v", report.Errors)
	}
}

func TestNewGenerationRun_EmptyModel(t *testing.T) {
	_, report := NewGenerationRun(GenerationRunInput{
		SlideID:   "intro",
		Output:    "o",
		CreatedAt: "2026-07-16T12:00:00Z",
	})

	if !containsError(report.Errors, "model is required") {
		t.Fatalf("expected 'model is required' error, got %v", report.Errors)
	}
}

func TestNewGenerationRun_EmptyOutput(t *testing.T) {
	_, report := NewGenerationRun(GenerationRunInput{
		SlideID:   "intro",
		Model:     "m",
		CreatedAt: "2026-07-16T12:00:00Z",
	})

	if !containsError(report.Errors, "output is required") {
		t.Fatalf("expected 'output is required' error, got %v", report.Errors)
	}
}

func TestNewGenerationRun_InvalidCreatedAt(t *testing.T) {
	_, report := NewGenerationRun(GenerationRunInput{
		SlideID:   "intro",
		Model:     "m",
		Output:    "o",
		CreatedAt: "not-a-timestamp",
	})

	if !containsError(report.Errors, "created at must be a valid RFC3339 timestamp") {
		t.Fatalf("expected timestamp format error, got %v", report.Errors)
	}
}

func TestNewGenerationRun_EmptyCreatedAt(t *testing.T) {
	_, report := NewGenerationRun(GenerationRunInput{
		SlideID: "intro",
		Model:   "m",
		Output:  "o",
	})

	if !containsError(report.Errors, "created at is required") {
		t.Fatalf("expected 'created at is required' error, got %v", report.Errors)
	}
}
