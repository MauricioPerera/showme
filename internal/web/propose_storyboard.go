package web

import (
	"strings"

	"github.com/MauricioPerera/showme/internal/ai"
	"github.com/MauricioPerera/showme/internal/knowledge"
)

const proposeStoryboardContextLimit = 3

// ProposeStoryboardInput is the raw data submitted by the "propose
// storyboard" web form.
type ProposeStoryboardInput struct {
	Objective     string
	Audience      string
	KnowledgeRoot string
	BaseURL       string
	Model         string
	Count         int
}

// ProposeStoryboardResult is the outcome of proposing a storyboard for
// review before creating a project.
type ProposeStoryboardResult struct {
	Slides   []ai.StoryboardSlide
	Errors   []string
	Warnings []string
}

func joinConceptBodies(concepts []knowledge.Concept) string {
	bodies := make([]string, len(concepts))
	for i, concept := range concepts {
		bodies[i] = concept.Body
	}
	return strings.Join(bodies, "\n\n---\n\n")
}

// HandleProposeStoryboard selects OKF context relevant to Objective (if
// KnowledgeRoot is non-empty) and generates a storyboard proposal via
// ai.GenerateStoryboard (backed by an OpenAIClient), for the caller to
// render as an editable review form before actually creating a project
// (see web-create-project-with-slides-handler). It never writes anything
// to disk.
//
// A file-system error loading the knowledge bundle is returned via err.
// Generation problems (empty objective, non-positive Count, an AI provider
// error, invalid JSON from the provider) are returned in the result's
// Errors, with Slides empty.
func HandleProposeStoryboard(input ProposeStoryboardInput) (ProposeStoryboardResult, error) {
	result := ProposeStoryboardResult{}

	var context string
	if input.KnowledgeRoot != "" {
		bundle, knowledgeReport := knowledge.Load(input.KnowledgeRoot)
		result.Errors = append(result.Errors, knowledgeReport.Errors...)
		result.Warnings = append(result.Warnings, knowledgeReport.Warnings...)

		concepts, selectReport := knowledge.Select(bundle, input.Objective, proposeStoryboardContextLimit)
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

	result.Slides = genResult.Slides
	return result, nil
}
