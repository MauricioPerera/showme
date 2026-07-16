package domain

import "time"

// GenerationRunInput is the raw data used to build a GenerationRun.
type GenerationRunInput struct {
	SlideID   string
	Model     string
	Provider  string
	Intent    string
	Context   string
	Output    string
	Warnings  []string
	CreatedAt string
}

// GenerationRun records one AI generation: the inputs and configuration
// that produced Output, for traceability (DEFINITION.md: "Cada
// GenerationRun debe conservar suficiente informacion para explicar que
// contexto, tokens y configuracion produjo la salida").
type GenerationRun struct {
	SlideID   string
	Model     string
	Provider  string
	Intent    string
	Context   string
	Output    string
	Warnings  []string
	CreatedAt string
}

// NewGenerationRun builds a GenerationRun from input, enforcing its
// structural invariants. CreatedAt is supplied by the caller (never
// generated here) so this constructor stays pure and deterministic; the
// CLI layer that calls it is expected to pass a real timestamp.
func NewGenerationRun(input GenerationRunInput) (GenerationRun, Report) {
	report := Report{}

	if input.SlideID == "" {
		report.Errors = append(report.Errors, "slide id is required")
	}
	if input.Model == "" {
		report.Errors = append(report.Errors, "model is required")
	}
	if input.Output == "" {
		report.Errors = append(report.Errors, "output is required")
	}
	if input.CreatedAt == "" {
		report.Errors = append(report.Errors, "created at is required")
	} else if _, err := time.Parse(time.RFC3339, input.CreatedAt); err != nil {
		report.Errors = append(report.Errors, "created at must be a valid RFC3339 timestamp")
	}

	run := GenerationRun{
		SlideID:   input.SlideID,
		Model:     input.Model,
		Provider:  input.Provider,
		Intent:    input.Intent,
		Context:   input.Context,
		Output:    input.Output,
		Warnings:  input.Warnings,
		CreatedAt: input.CreatedAt,
	}
	return run, report
}
