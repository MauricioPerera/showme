package project

import (
	"github.com/MauricioPerera/showme/internal/design"
	"github.com/MauricioPerera/showme/internal/domain"
	"github.com/MauricioPerera/showme/internal/knowledge"
)

// CreateProjectInput is the raw data needed to assemble a Project.
type CreateProjectInput struct {
	Name          string
	DesignContent string
	DesignPath    string
	KnowledgeRoot string
	DeckInput     domain.DeckInput
	Version       int
}

// Report aggregates the findings from every step of assembling a Project.
type Report struct {
	Errors   []string
	Warnings []string
}

func appendReport(target *Report, errors, warnings []string) {
	target.Errors = append(target.Errors, errors...)
	target.Warnings = append(target.Warnings, warnings...)
}

// CreateProject validates a DESIGN.md, loads an OKF bundle, builds a Deck and
// assembles them into a Project, aggregating every step's findings.
func CreateProject(input CreateProjectInput) (domain.Project, Report) {
	var report Report

	designReport := design.Validate(input.DesignContent)
	appendReport(&report, designReport.Errors, designReport.Warnings)

	_, knowledgeReport := knowledge.Load(input.KnowledgeRoot)
	appendReport(&report, knowledgeReport.Errors, knowledgeReport.Warnings)

	deck, deckReport := domain.NewDeck(input.DeckInput)
	appendReport(&report, deckReport.Errors, deckReport.Warnings)

	proj, projectReport := domain.NewProject(domain.ProjectInput{
		Name:          input.Name,
		Deck:          deck,
		DesignPath:    input.DesignPath,
		KnowledgePath: input.KnowledgeRoot,
		Version:       input.Version,
	})
	appendReport(&report, projectReport.Errors, projectReport.Warnings)

	return proj, report
}
