package knowledge

import "testing"

func TestSelectRanksRelevantConceptsDeterministically(t *testing.T) {
	bundle := Bundle{Concepts: []Concept{
		{ID: "zeta", Title: "Audience", Body: "A technical audience."},
		{ID: "alpha", Title: "Technical audience", Body: "Audience guidance."},
		{ID: "other", Title: "Brand", Body: "Visual identity."},
	}}

	got, report := Select(bundle, "technical audience", 2)
	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if len(got) != 2 || got[0].ID != "alpha" || got[1].ID != "zeta" {
		t.Fatalf("unexpected selection: %#v", got)
	}
}

func TestSelectUsesIDTagsAndBodyAndHonorsLimit(t *testing.T) {
	bundle := Bundle{Concepts: []Concept{
		{ID: "slides/opening", Tags: []string{"review"}, Body: "Citations support review."},
		{ID: "brand", Tags: []string{"design"}, Body: "Visual tokens."},
	}}

	got, report := Select(bundle, "opening review citations", 1)
	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if len(got) != 1 || got[0].ID != "slides/opening" {
		t.Fatalf("expected one matching concept, got %#v", got)
	}
}

func TestSelectRejectsEmptyQueryAndNonPositiveLimit(t *testing.T) {
	bundle := Bundle{Concepts: []Concept{{ID: "one", Title: "One"}}}

	if got, report := Select(bundle, "", 1); len(got) != 0 || len(report.Errors) != 1 {
		t.Fatalf("expected empty-query error, got %#v, %v", got, report.Errors)
	}
	if got, report := Select(bundle, "one", 0); len(got) != 0 || len(report.Errors) != 1 {
		t.Fatalf("expected limit error, got %#v, %v", got, report.Errors)
	}
}

