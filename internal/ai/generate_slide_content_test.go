package ai

import (
	"errors"
	"testing"
)

var errUnavailable = errors.New("provider unavailable")

type fakeGenerator struct {
	content string
	err     error
	calls   []GenerateContentRequest
}

func (f *fakeGenerator) GenerateContent(request GenerateContentRequest) (string, error) {
	f.calls = append(f.calls, request)
	return f.content, f.err
}

func TestGenerateSlideContent_Valid(t *testing.T) {
	gen := &fakeGenerator{content: "Contenido generado."}

	result, report := GenerateSlideContent(GenerateSlideContentInput{
		Generator: gen,
		Intent:    "Explicar el roadmap",
		Context:   "El roadmap tiene 3 fases.",
	})

	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if result.Content != "Contenido generado." {
		t.Fatalf("expected generated content, got %q", result.Content)
	}
	if len(gen.calls) != 1 {
		t.Fatalf("expected exactly one call to the generator, got %d", len(gen.calls))
	}
	if gen.calls[0].Intent != "Explicar el roadmap" || gen.calls[0].Context != "El roadmap tiene 3 fases." {
		t.Fatalf("expected intent/context forwarded as-is, got %+v", gen.calls[0])
	}
}

func TestGenerateSlideContent_EmptyIntent(t *testing.T) {
	gen := &fakeGenerator{content: "no deberia llegar aca"}

	result, report := GenerateSlideContent(GenerateSlideContentInput{
		Generator: gen,
		Intent:    "",
	})

	if !containsError(report.Errors, "intent is required") {
		t.Fatalf("expected 'intent is required' error, got %v", report.Errors)
	}
	if result.Content != "" {
		t.Fatalf("expected empty content, got %q", result.Content)
	}
	if len(gen.calls) != 0 {
		t.Fatalf("expected the generator not to be called, got %d calls", len(gen.calls))
	}
}

func TestGenerateSlideContent_GeneratorError(t *testing.T) {
	gen := &fakeGenerator{err: errUnavailable}

	_, report := GenerateSlideContent(GenerateSlideContentInput{
		Generator: gen,
		Intent:    "Explicar el roadmap",
	})

	if !containsError(report.Errors, errUnavailable.Error()) {
		t.Fatalf("expected the generator's error message, got %v", report.Errors)
	}
}

func TestGenerateSlideContent_EmptyGeneratedContent(t *testing.T) {
	gen := &fakeGenerator{content: ""}

	_, report := GenerateSlideContent(GenerateSlideContentInput{
		Generator: gen,
		Intent:    "Explicar el roadmap",
	})

	if !containsError(report.Errors, "generator returned empty content") {
		t.Fatalf("expected 'generator returned empty content' error, got %v", report.Errors)
	}
}

func containsError(errors []string, want string) bool {
	for _, e := range errors {
		if e == want {
			return true
		}
	}
	return false
}
