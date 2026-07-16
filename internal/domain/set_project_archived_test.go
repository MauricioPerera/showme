package domain

import "testing"

func validProjectForArchiving(t *testing.T) Project {
	t.Helper()
	proj, report := NewProject(ProjectInput{
		Name:          "Roadmap Q3",
		Deck:          validDeck(t),
		DesignPath:    "DESIGN.md",
		KnowledgePath: "knowledge/showme",
	})
	if len(report.Errors) != 0 {
		t.Fatalf("setup: unexpected project errors: %v", report.Errors)
	}
	return proj
}

func TestSetProjectArchived_ToTrue(t *testing.T) {
	proj := validProjectForArchiving(t)

	updated := SetProjectArchived(SetProjectArchivedInput{Project: proj, Archived: true})

	if !updated.Archived {
		t.Fatalf("expected Archived true, got %+v", updated)
	}
	if updated.Name != proj.Name || updated.Version != proj.Version {
		t.Fatalf("expected other fields preserved, got %+v", updated)
	}
}

func TestSetProjectArchived_ToFalse(t *testing.T) {
	proj := validProjectForArchiving(t)
	proj.Archived = true

	updated := SetProjectArchived(SetProjectArchivedInput{Project: proj, Archived: false})

	if updated.Archived {
		t.Fatalf("expected Archived false, got %+v", updated)
	}
}

func TestSetProjectArchived_DoesNotMutateOriginal(t *testing.T) {
	proj := validProjectForArchiving(t)

	_ = SetProjectArchived(SetProjectArchivedInput{Project: proj, Archived: true})

	if proj.Archived {
		t.Fatalf("expected original project untouched, got %+v", proj)
	}
}
