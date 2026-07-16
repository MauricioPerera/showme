package cli

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/MauricioPerera/showme/internal/ai"
	"github.com/MauricioPerera/showme/internal/knowledge"
	"github.com/MauricioPerera/showme/internal/storage"
)

// GenerateStoryboardCommandInput is the raw data needed to run the
// "project generate-storyboard" CLI command.
type GenerateStoryboardCommandInput struct {
	Objective     string
	Audience      string
	KnowledgeRoot string
	BaseURL       string
	Model         string
	DeckTitle     string
	Count         int
	OutPath       string
}

// GenerateStoryboardCommandResult is the JSON-stable result of running the
// "project generate-storyboard" CLI command.
type GenerateStoryboardCommandResult struct {
	OK       bool
	Path     string
	Errors   []string
	Warnings []string
}

func uniqueSlideID(seen map[string]bool, title string) string {
	base := storage.Slugify(title)
	if base == "" {
		base = "slide"
	}
	id := base
	for suffix := 2; seen[id]; suffix++ {
		id = base + "-" + strconv.Itoa(suffix)
	}
	seen[id] = true
	return id
}

// RunGenerateStoryboardCommand proposes a storyboard for a presentation
// (via ai.GenerateStoryboard, backed by an OpenAIClient) and writes it as
// a deck JSON file at OutPath, in the same shape RunCreateProjectCommand
// expects for its --deck flag. If KnowledgeRoot is non-empty, OKF context
// relevant to Objective is selected and included in the generation prompt.
//
// A file-system error loading the knowledge bundle or writing OutPath is
// returned via err. Validation/generation problems (empty objective,
// non-positive Count, an AI provider error, invalid JSON from the
// provider) are returned in the result's Errors, with OK false and no
// file written.
func RunGenerateStoryboardCommand(input GenerateStoryboardCommandInput) (GenerateStoryboardCommandResult, error) {
	result := GenerateStoryboardCommandResult{}

	var context string
	if input.KnowledgeRoot != "" {
		bundle, knowledgeReport := knowledge.Load(input.KnowledgeRoot)
		result.Errors = append(result.Errors, knowledgeReport.Errors...)
		result.Warnings = append(result.Warnings, knowledgeReport.Warnings...)

		concepts, selectReport := knowledge.Select(bundle, input.Objective, defaultGenerationContextLimit)
		result.Errors = append(result.Errors, selectReport.Errors...)
		result.Warnings = append(result.Warnings, selectReport.Warnings...)
		if len(result.Errors) != 0 {
			return result, nil
		}
		context = joinConceptBodies(concepts)
	}

	client := ai.NewOpenAIClient(input.BaseURL, input.Model)
	genResult, genReport := ai.GenerateStoryboard(ai.GenerateStoryboardInput{
		Generator: client,
		Objective: input.Objective,
		Audience:  input.Audience,
		Context:   context,
		Count:     input.Count,
	})
	result.Errors = append(result.Errors, genReport.Errors...)
	result.Warnings = append(result.Warnings, genReport.Warnings...)
	if len(result.Errors) != 0 {
		return result, nil
	}

	seen := map[string]bool{}
	slides := make([]slideDTO, len(genResult.Slides))
	for i, s := range genResult.Slides {
		slides[i] = slideDTO{ID: uniqueSlideID(seen, s.Title), Title: s.Title, Intent: s.Intent}
	}

	encoded, err := json.MarshalIndent(deckInputDTO{
		Title:    input.DeckTitle,
		Audience: input.Audience,
		Slides:   slides,
	}, "", "  ")
	if err != nil {
		return GenerateStoryboardCommandResult{}, err
	}

	if err := os.WriteFile(input.OutPath, encoded, 0o644); err != nil {
		return GenerateStoryboardCommandResult{}, err
	}

	result.OK = true
	result.Path = input.OutPath
	return result, nil
}
