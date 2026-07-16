package ai

import "testing"

type fakeStoryboardGenerator struct {
	raw   string
	err   error
	calls []GenerateStoryboardRequest
}

func (f *fakeStoryboardGenerator) GenerateStoryboard(request GenerateStoryboardRequest) (string, error) {
	f.calls = append(f.calls, request)
	return f.raw, f.err
}

func TestGenerateStoryboard_Valid(t *testing.T) {
	gen := &fakeStoryboardGenerator{raw: `[{"title":"Introduccion","intent":"Dar la bienvenida"},{"title":"Plan","intent":"Explicar los proximos pasos"}]`}

	result, report := GenerateStoryboard(GenerateStoryboardInput{
		Generator: gen,
		Objective: "Presentar el roadmap",
		Audience:  "Equipo de producto",
		Count:     2,
	})

	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if len(result.Slides) != 2 {
		t.Fatalf("expected 2 slides, got %+v", result.Slides)
	}
	if result.Slides[0].Title != "Introduccion" || result.Slides[1].Title != "Plan" {
		t.Fatalf("expected slides in order, got %+v", result.Slides)
	}
	if len(gen.calls) != 1 || gen.calls[0].Objective != "Presentar el roadmap" || gen.calls[0].Count != 2 {
		t.Fatalf("expected request forwarded as-is, got %+v", gen.calls)
	}
}

func TestGenerateStoryboard_EmptyObjective(t *testing.T) {
	gen := &fakeStoryboardGenerator{raw: "no deberia usarse"}

	_, report := GenerateStoryboard(GenerateStoryboardInput{Generator: gen, Count: 3})

	if !containsError(report.Errors, "objective is required") {
		t.Fatalf("expected 'objective is required' error, got %v", report.Errors)
	}
	if len(gen.calls) != 0 {
		t.Fatalf("expected the generator not to be called, got %d calls", len(gen.calls))
	}
}

func TestGenerateStoryboard_NonPositiveCount(t *testing.T) {
	gen := &fakeStoryboardGenerator{raw: "no deberia usarse"}

	_, report := GenerateStoryboard(GenerateStoryboardInput{Generator: gen, Objective: "x", Count: 0})

	if !containsError(report.Errors, "count must be positive") {
		t.Fatalf("expected 'count must be positive' error, got %v", report.Errors)
	}
}

func TestGenerateStoryboard_GeneratorError(t *testing.T) {
	gen := &fakeStoryboardGenerator{err: errUnavailable}

	_, report := GenerateStoryboard(GenerateStoryboardInput{Generator: gen, Objective: "x", Count: 3})

	if !containsError(report.Errors, errUnavailable.Error()) {
		t.Fatalf("expected the generator's error message, got %v", report.Errors)
	}
}

func TestGenerateStoryboard_InvalidJSON(t *testing.T) {
	gen := &fakeStoryboardGenerator{raw: "```json\n[{\"title\":\"x\"}]\n```"}

	_, report := GenerateStoryboard(GenerateStoryboardInput{Generator: gen, Objective: "x", Count: 3})

	if len(report.Errors) == 0 {
		t.Fatalf("expected a JSON parse error, got none")
	}
}

func TestGenerateStoryboard_EmptySlideList(t *testing.T) {
	gen := &fakeStoryboardGenerator{raw: "[]"}

	_, report := GenerateStoryboard(GenerateStoryboardInput{Generator: gen, Objective: "x", Count: 3})

	if !containsError(report.Errors, "generator returned no slides") {
		t.Fatalf("expected 'generator returned no slides' error, got %v", report.Errors)
	}
}

func TestGenerateStoryboard_SlideMissingFields(t *testing.T) {
	gen := &fakeStoryboardGenerator{raw: `[{"title":"","intent":""}]`}

	_, report := GenerateStoryboard(GenerateStoryboardInput{Generator: gen, Objective: "x", Count: 1})

	if !containsError(report.Errors, "slide[0]: title is required") {
		t.Fatalf("expected 'slide[0]: title is required' error, got %v", report.Errors)
	}
	if !containsError(report.Errors, "slide[0]: intent is required") {
		t.Fatalf("expected 'slide[0]: intent is required' error, got %v", report.Errors)
	}
}
