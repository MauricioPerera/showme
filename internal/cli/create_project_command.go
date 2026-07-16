package cli

import (
	"encoding/json"
	"os"

	"github.com/MauricioPerera/showme/internal/domain"
	"github.com/MauricioPerera/showme/internal/project"
	"github.com/MauricioPerera/showme/internal/storage"
)

// CreateProjectCommandInput is the raw data needed to run the "project
// create" CLI command.
type CreateProjectCommandInput struct {
	Name          string
	DesignPath    string
	KnowledgeRoot string
	DeckPath      string
	OutDir        string
}

// CreateProjectCommandResult is the JSON-stable result of running the
// "project create" CLI command.
type CreateProjectCommandResult struct {
	OK       bool
	Path     string
	Errors   []string
	Warnings []string
}

type slideDTO struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Intent  string `json:"intent"`
	Content string `json:"content"`
	Status  string `json:"status"`
}

type deckInputDTO struct {
	Title    string     `json:"title"`
	Audience string     `json:"audience"`
	Slides   []slideDTO `json:"slides"`
}

func parseDeckInput(data []byte) (domain.DeckInput, error) {
	var dto deckInputDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return domain.DeckInput{}, err
	}

	slides := make([]domain.Slide, len(dto.Slides))
	for i, s := range dto.Slides {
		slides[i] = domain.Slide{
			ID:      s.ID,
			Title:   s.Title,
			Intent:  s.Intent,
			Content: s.Content,
			Status:  domain.SlideStatus(s.Status),
		}
	}

	return domain.DeckInput{Title: dto.Title, Audience: dto.Audience, Slides: slides}, nil
}

// RunCreateProjectCommand reads a DESIGN.md file and a deck JSON file,
// assembles a Project via project.CreateProject and, if valid, persists it
// under OutDir via storage.SaveProject.
//
// A file-system or parse error (missing/unreadable design or deck file,
// invalid deck JSON) is returned via err. Validation problems from
// assembling or saving the project are returned in the result's
// Errors/Warnings, with OK false and Path empty.
func RunCreateProjectCommand(input CreateProjectCommandInput) (CreateProjectCommandResult, error) {
	designContent, err := os.ReadFile(input.DesignPath)
	if err != nil {
		return CreateProjectCommandResult{}, err
	}

	deckContent, err := os.ReadFile(input.DeckPath)
	if err != nil {
		return CreateProjectCommandResult{}, err
	}

	deckInput, err := parseDeckInput(deckContent)
	if err != nil {
		return CreateProjectCommandResult{}, err
	}

	proj, report := project.CreateProject(project.CreateProjectInput{
		Name:          input.Name,
		DesignContent: string(designContent),
		DesignPath:    input.DesignPath,
		KnowledgeRoot: input.KnowledgeRoot,
		DeckInput:     deckInput,
	})

	result := CreateProjectCommandResult{
		Errors:   append([]string{}, report.Errors...),
		Warnings: append([]string{}, report.Warnings...),
	}
	if len(report.Errors) != 0 {
		return result, nil
	}

	path, saveReport, err := storage.SaveProject(storage.SaveProjectRequest{
		Dir: input.OutDir,
		Input: domain.ProjectInput{
			Name:          proj.Name,
			Deck:          proj.Deck,
			DesignPath:    proj.DesignPath,
			KnowledgePath: proj.KnowledgePath,
			Version:       proj.Version,
		},
	})
	if err != nil {
		return CreateProjectCommandResult{}, err
	}

	result.Errors = append(result.Errors, saveReport.Errors...)
	result.Warnings = append(result.Warnings, saveReport.Warnings...)
	result.Path = path
	result.OK = len(result.Errors) == 0
	return result, nil
}
