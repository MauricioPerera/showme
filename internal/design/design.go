package design

import "strings"

// Report contains deterministic findings for a DESIGN.md document.
type Report struct {
	Errors   []string
	Warnings []string
}

// Validate checks the minimum structure required by showme before using a
// design system for slide rendering.
func Validate(content string) Report {
	var report Report
	frontmatter, body, ok := splitFrontmatter(content)
	if !ok {
		report.Errors = append(report.Errors, "frontmatter is required")
		return report
	}
	if !hasField(frontmatter, "name:") {
		report.Errors = append(report.Errors, "name is required")
	}
	if !hasNestedField(frontmatter, "colors:", "primary:") {
		report.Errors = append(report.Errors, "colors.primary is required")
	}
	seen := map[string]bool{}
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "## ") {
			continue
		}
		section := strings.TrimSpace(strings.TrimPrefix(line, "## "))
		if seen[section] {
			report.Errors = append(report.Errors, "duplicate section: "+section)
			continue
		}
		seen[section] = true
	}
	return report
}

func splitFrontmatter(content string) (string, string, bool) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "", content, false
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return strings.Join(lines[1:i], "\n"), strings.Join(lines[i+1:], "\n"), true
		}
	}
	return "", content, false
}

func hasField(frontmatter, field string) bool {
	for _, line := range strings.Split(frontmatter, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), field) {
			return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), field)) != ""
		}
	}
	return false
}

func hasNestedField(frontmatter, parent, child string) bool {
	lines := strings.Split(frontmatter, "\n")
	inside := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == parent {
			inside = true
			continue
		}
		if inside && !strings.HasPrefix(line, " ") && trimmed != "" {
			inside = false
		}
		if inside && strings.HasPrefix(trimmed, child) {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, child)) != ""
		}
	}
	return false
}
