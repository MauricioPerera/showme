package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/MauricioPerera/showme/internal/ai"
	"github.com/MauricioPerera/showme/internal/domain"
	"github.com/MauricioPerera/showme/internal/knowledge"
	"github.com/MauricioPerera/showme/internal/storage"
)

const defaultGenerationContextLimit = 3

// GenerateSlideContentCommandInput is the raw data needed to run the
// "project generate-slide" CLI command.
type GenerateSlideContentCommandInput struct {
	Path    string
	SlideID string
	BaseURL string
	Model   string
	OutDir  string
}

// GenerateSlideContentCommandResult is the JSON-stable result of running
// the "project generate-slide" CLI command.
type GenerateSlideContentCommandResult struct {
	OK       bool
	Path     string
	Content  string
	Errors   []string
	Warnings []string
}

func findSlide(deck domain.Deck, slideID string) (domain.Slide, bool) {
	for _, slide := range deck.Slides {
		if slide.ID == slideID {
			return slide, true
		}
	}
	return domain.Slide{}, false
}

func selectionQuery(slide domain.Slide) string {
	if slide.Intent != "" {
		return slide.Intent
	}
	return slide.Title
}

func joinConceptBodies(concepts []knowledge.Concept) string {
	bodies := make([]string, len(concepts))
	for i, concept := range concepts {
		bodies[i] = concept.Body
	}
	return strings.Join(bodies, "\n\n---\n\n")
}

// RunGenerateSlideContentCommand loads the Project at Path, selects OKF
// context for the slide identified by SlideID (via knowledge.Select over
// the project's KnowledgePath bundle, queried by the slide's Intent or
// Title as a fallback), generates its Content with an OpenAIClient
// (BaseURL/Model), applies it via domain.UpdateSlide (preserving the
// slide's current Status), records a domain.GenerationRun (with a real
// timestamp -- this is the one place in the CLI layer allowed to call
// time.Now(), since domain.NewGenerationRun itself stays pure) in the
// project's Runs history, and, if valid, saves the updated Project under
// OutDir.
//
// A file-system error loading Path, loading the knowledge bundle, or
// saving under OutDir is returned via err. Validation/generation problems
// (slide not found, empty context query, an AI provider error) are
// returned in the result's Errors, with OK false and the project left
// untouched.
func RunGenerateSlideContentCommand(input GenerateSlideContentCommandInput) (GenerateSlideContentCommandResult, error) {
	proj, err := storage.LoadProject(input.Path)
	if err != nil {
		return GenerateSlideContentCommandResult{}, err
	}

	slide, found := findSlide(proj.Deck, input.SlideID)
	if !found {
		return GenerateSlideContentCommandResult{
			Errors: []string{fmt.Sprintf("slide not found: %s", input.SlideID)},
		}, nil
	}

	bundle, knowledgeReport := knowledge.Load(proj.KnowledgePath)
	result := GenerateSlideContentCommandResult{
		Errors:   append([]string{}, knowledgeReport.Errors...),
		Warnings: append([]string{}, knowledgeReport.Warnings...),
	}

	concepts, selectReport := knowledge.Select(bundle, selectionQuery(slide), defaultGenerationContextLimit)
	result.Errors = append(result.Errors, selectReport.Errors...)
	result.Warnings = append(result.Warnings, selectReport.Warnings...)
	if len(result.Errors) != 0 {
		return result, nil
	}

	context := joinConceptBodies(concepts)
	client := ai.NewOpenAIClient(input.BaseURL, input.Model)
	genResult, genReport := ai.GenerateSlideContent(ai.GenerateSlideContentInput{
		Generator: client,
		Intent:    slide.Intent,
		Context:   context,
	})
	result.Errors = append(result.Errors, genReport.Errors...)
	result.Warnings = append(result.Warnings, genReport.Warnings...)
	if len(result.Errors) != 0 {
		return result, nil
	}

	updatedDeck, updateReport := domain.UpdateSlide(domain.UpdateSlideInput{
		Deck: proj.Deck,
		Slide: domain.Slide{
			ID:      slide.ID,
			Title:   slide.Title,
			Intent:  slide.Intent,
			Content: genResult.Content,
			Status:  slide.Status,
		},
	})
	result.Errors = append(result.Errors, updateReport.Errors...)
	if len(result.Errors) != 0 {
		return result, nil
	}

	run, runReport := domain.NewGenerationRun(domain.GenerationRunInput{
		SlideID:   slide.ID,
		Model:     input.Model,
		Provider:  input.BaseURL,
		Intent:    slide.Intent,
		Context:   context,
		Output:    genResult.Content,
		Warnings:  result.Warnings,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})
	result.Errors = append(result.Errors, runReport.Errors...)
	if len(result.Errors) != 0 {
		return result, nil
	}
	projWithRun := domain.AppendGenerationRun(domain.AppendGenerationRunInput{Project: proj, Run: run})

	path, saveReport, err := storage.SaveProject(storage.SaveProjectRequest{
		Dir: input.OutDir,
		Input: domain.ProjectInput{
			Name:          proj.Name,
			Deck:          updatedDeck,
			DesignPath:    proj.DesignPath,
			KnowledgePath: proj.KnowledgePath,
			Version:       proj.Version,
			Archived:      proj.Archived,
			Runs:          projWithRun.Runs,
		},
	})
	if err != nil {
		return GenerateSlideContentCommandResult{}, err
	}

	result.Errors = append(result.Errors, saveReport.Errors...)
	result.Warnings = append(result.Warnings, saveReport.Warnings...)
	result.Path = path
	result.Content = genResult.Content
	result.OK = len(result.Errors) == 0
	return result, nil
}
