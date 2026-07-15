package design

import "testing"

func TestValidateAcceptsShowmeDesign(t *testing.T) {
	content := "---\nname: Showme\ncolors:\n  primary: \"#111827\"\ntypography:\n  body: {}\n---\n\n## Overview\n\n## Colors\n\n## Typography\n"
	report := Validate(content)
	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
}

func TestValidateRequiresPrimaryColor(t *testing.T) {
	content := "---\nname: Showme\ncolors:\n  secondary: \"#475569\"\n---\n\n## Overview\n"
	report := Validate(content)
	if !contains(report.Errors, "colors.primary is required") {
		t.Fatalf("expected missing primary error, got %v", report.Errors)
	}
}

func TestValidateRejectsDuplicateSections(t *testing.T) {
	content := "---\nname: Showme\ncolors:\n  primary: \"#111827\"\n---\n\n## Colors\n\n## Colors\n"
	report := Validate(content)
	if !contains(report.Errors, "duplicate section: Colors") {
		t.Fatalf("expected duplicate section error, got %v", report.Errors)
	}
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
