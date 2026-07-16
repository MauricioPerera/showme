package domain

// ProjectInput is the raw data used to build a Project.
type ProjectInput struct {
	Name          string
	Deck          Deck
	DesignPath    string
	KnowledgePath string
	Version       int
	Archived      bool
	Runs          []GenerationRun
}

// Project is the container for a presentation, its identity/knowledge
// references, its version, its archived state and its AI generation
// history.
type Project struct {
	Name          string
	Deck          Deck
	DesignPath    string
	KnowledgePath string
	Version       int
	Archived      bool
	Runs          []GenerationRun
}

// NewProject builds a Project from input, enforcing its structural invariants.
func NewProject(input ProjectInput) (Project, Report) {
	report := Report{}

	if input.Name == "" {
		report.Errors = append(report.Errors, "name is required")
	}
	if len(input.Deck.Slides) == 0 {
		report.Errors = append(report.Errors, "deck must have at least one slide")
	}
	if input.DesignPath == "" {
		report.Errors = append(report.Errors, "design path is required")
	}
	if input.KnowledgePath == "" {
		report.Errors = append(report.Errors, "knowledge path is required")
	}

	version := input.Version
	if version < 0 {
		report.Errors = append(report.Errors, "version must be positive")
	} else if version == 0 {
		version = 1
	}

	project := Project{
		Name:          input.Name,
		Deck:          input.Deck,
		DesignPath:    input.DesignPath,
		KnowledgePath: input.KnowledgePath,
		Version:       version,
		Archived:      input.Archived,
		Runs:          input.Runs,
	}
	return project, report
}
