package domain

// AppendGenerationRunInput is the raw data used by AppendGenerationRun.
type AppendGenerationRunInput struct {
	Project Project
	Run     GenerationRun
}

// AppendGenerationRun returns a copy of Project with Run appended to its
// Runs history, preserving the order and content of any existing runs. It
// takes an already-built GenerationRun (see generation-run-model) and does
// not itself validate it, since a boolean-free append has no invalid
// values to reject. The original Project is never mutated.
func AppendGenerationRun(input AppendGenerationRunInput) Project {
	runs := make([]GenerationRun, len(input.Project.Runs), len(input.Project.Runs)+1)
	copy(runs, input.Project.Runs)
	runs = append(runs, input.Run)

	return Project{
		Name:          input.Project.Name,
		Deck:          input.Project.Deck,
		DesignPath:    input.Project.DesignPath,
		KnowledgePath: input.Project.KnowledgePath,
		Version:       input.Project.Version,
		Archived:      input.Project.Archived,
		Runs:          runs,
	}
}
