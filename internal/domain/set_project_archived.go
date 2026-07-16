package domain

// SetProjectArchivedInput is the raw data used by SetProjectArchived.
type SetProjectArchivedInput struct {
	Project  Project
	Archived bool
}

// SetProjectArchived returns a copy of Project with its Archived field set
// to Archived. A boolean flag has no invalid values, so this operation
// cannot fail and returns no Report, unlike the rest of the domain's
// constructors. The original Project is never mutated.
func SetProjectArchived(input SetProjectArchivedInput) Project {
	return Project{
		Name:          input.Project.Name,
		Deck:          input.Project.Deck,
		DesignPath:    input.Project.DesignPath,
		KnowledgePath: input.Project.KnowledgePath,
		Version:       input.Project.Version,
		Archived:      input.Archived,
		Runs:          input.Project.Runs,
	}
}
