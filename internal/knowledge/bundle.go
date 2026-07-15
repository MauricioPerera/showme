package knowledge

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Concept is one OKF document loaded from a bundle.
type Concept struct {
	ID, Type, Title, Description, Path string
	Tags                               []string
	Body                               string
}

// Bundle contains the concepts available to showme generation.
type Bundle struct {
	Concepts []Concept
}

// Report contains deterministic, non-fatal diagnostics from loading a bundle.
type Report struct {
	Errors   []string
	Warnings []string
}

// Load reads Markdown concepts from root, excluding reserved index.md files.
func Load(root string) (Bundle, Report) {
	paths := markdownPaths(root)
	sort.Strings(paths)
	var bundle Bundle
	var report Report
	for _, path := range paths {
		if filepath.Base(path) == "index.md" {
			continue
		}
		concept, err := loadConcept(root, path)
		if err != "" {
			report.Errors = append(report.Errors, err)
			continue
		}
		bundle.Concepts = append(bundle.Concepts, concept)
	}
	return bundle, report
}

func markdownPaths(root string) []string {
	var paths []string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err == nil && info != nil && !info.IsDir() && strings.EqualFold(filepath.Ext(path), ".md") {
			paths = append(paths, path)
		}
		return nil
	})
	return paths
}

func loadConcept(root, path string) (Concept, string) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Concept{}, filepath.ToSlash(relative(root, path)) + ": " + err.Error()
	}
	fm, body, ok := frontmatter(string(content))
	rel := filepath.ToSlash(relative(root, path))
	if !ok {
		return Concept{}, rel + ": frontmatter is required"
	}
	typ := fm["type"]
	if typ == "" {
		return Concept{}, rel + ": type is required"
	}
	return Concept{
		ID:          strings.TrimSuffix(rel, filepath.Ext(rel)),
		Type:        typ,
		Title:       fm["title"],
		Description: fm["description"],
		Tags:        parseTags(fm["tags"]),
		Path:        rel,
		Body:        body,
	}, ""
}

func relative(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.Base(path)
	}
	return rel
}

func frontmatter(content string) (map[string]string, string, bool) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil, content, false
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "---" {
			continue
		}
		fields := map[string]string{}
		for _, line := range lines[1:i] {
			key, value, found := strings.Cut(line, ":")
			if found {
				fields[strings.TrimSpace(key)] = unquote(strings.TrimSpace(value))
			}
		}
		return fields, strings.Join(lines[i+1:], "\n"), true
	}
	return nil, content, false
}

func parseTags(value string) []string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := unquote(strings.TrimSpace(part))
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

func unquote(value string) string {
	if len(value) >= 2 && ((value[0] == '\'' && value[len(value)-1] == '\'') || (value[0] == '"' && value[len(value)-1] == '"')) {
		return value[1 : len(value)-1]
	}
	return value
}
