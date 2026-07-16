package ai

import (
	"encoding/json"
	"fmt"
)

// GenerateStoryboardRequest is the input to a StoryboardGenerator.
type GenerateStoryboardRequest struct {
	Objective string
	Audience  string
	Context   string
	Count     int
}

// StoryboardGenerator produces a raw JSON array of proposed slides (each
// with "title" and "intent") from a presentation's objective, audience and
// supporting context. Implementations may call an external AI provider.
type StoryboardGenerator interface {
	GenerateStoryboard(request GenerateStoryboardRequest) (string, error)
}

// StoryboardSlide is one proposed slide in a generated storyboard.
type StoryboardSlide struct {
	Title  string
	Intent string
}

// GenerateStoryboardInput is the raw data used by GenerateStoryboard.
type GenerateStoryboardInput struct {
	Generator StoryboardGenerator
	Objective string
	Audience  string
	Context   string
	Count     int
}

// GenerateStoryboardResult is the outcome of GenerateStoryboard.
type GenerateStoryboardResult struct {
	Slides []StoryboardSlide
}

type storyboardSlideDTO struct {
	Title  string `json:"title"`
	Intent string `json:"intent"`
}

// GenerateStoryboard validates Objective/Count and delegates to Generator
// to propose a storyboard, parsing its raw JSON response into
// StoryboardSlide values. The JSON is parsed as-is: a response wrapped in
// markdown fences or otherwise not a bare JSON array is a parse error, not
// something this function tries to repair. This function itself never
// touches the network: it depends only on the injected StoryboardGenerator
// interface, same separation as generate-slide-content-usecase.
func GenerateStoryboard(input GenerateStoryboardInput) (GenerateStoryboardResult, Report) {
	report := Report{}

	if input.Objective == "" {
		report.Errors = append(report.Errors, "objective is required")
	}
	if input.Count <= 0 {
		report.Errors = append(report.Errors, "count must be positive")
	}
	if len(report.Errors) != 0 {
		return GenerateStoryboardResult{}, report
	}

	raw, err := input.Generator.GenerateStoryboard(GenerateStoryboardRequest{
		Objective: input.Objective,
		Audience:  input.Audience,
		Context:   input.Context,
		Count:     input.Count,
	})
	if err != nil {
		report.Errors = append(report.Errors, err.Error())
		return GenerateStoryboardResult{}, report
	}

	var dtos []storyboardSlideDTO
	if err := json.Unmarshal([]byte(raw), &dtos); err != nil {
		report.Errors = append(report.Errors, fmt.Sprintf("invalid storyboard JSON: %s", err))
		return GenerateStoryboardResult{}, report
	}
	if len(dtos) == 0 {
		report.Errors = append(report.Errors, "generator returned no slides")
		return GenerateStoryboardResult{}, report
	}

	slides := make([]StoryboardSlide, len(dtos))
	for i, dto := range dtos {
		if dto.Title == "" {
			report.Errors = append(report.Errors, fmt.Sprintf("slide[%d]: title is required", i))
		}
		if dto.Intent == "" {
			report.Errors = append(report.Errors, fmt.Sprintf("slide[%d]: intent is required", i))
		}
		slides[i] = StoryboardSlide{Title: dto.Title, Intent: dto.Intent}
	}
	if len(report.Errors) != 0 {
		return GenerateStoryboardResult{}, report
	}

	return GenerateStoryboardResult{Slides: slides}, report
}
