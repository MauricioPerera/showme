package knowledge

import (
	"sort"
	"strings"
	"unicode"
)

type scoredConcept struct {
	concept Concept
	score   int
}

// Select returns the highest-scoring concepts for a slide query.
func Select(bundle Bundle, query string, limit int) ([]Concept, Report) {
	queryTokens := tokenSet(query)
	report := Report{}
	if len(queryTokens) == 0 {
		report.Errors = append(report.Errors, "query is required")
		return nil, report
	}
	if limit <= 0 {
		report.Errors = append(report.Errors, "limit must be positive")
		return nil, report
	}

	candidates := make([]scoredConcept, 0, len(bundle.Concepts))
	for _, concept := range bundle.Concepts {
		score := conceptScore(concept, queryTokens)
		if score > 0 {
			candidates = append(candidates, scoredConcept{concept: concept, score: score})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		return candidates[i].concept.ID < candidates[j].concept.ID
	})
	if limit > len(candidates) {
		limit = len(candidates)
	}
	selected := make([]Concept, limit)
	for i := range selected {
		selected[i] = candidates[i].concept
	}
	return selected, report
}

func conceptScore(concept Concept, queryTokens map[string]struct{}) int {
	score := 0
	for token := range queryTokens {
		score += fieldScore(token, concept.ID, 2)
		score += fieldScore(token, concept.Title, 5)
		score += fieldScore(token, concept.Description, 4)
		for _, tag := range concept.Tags {
			score += fieldScore(token, tag, 3)
		}
		score += fieldScore(token, concept.Body, 1)
	}
	return score
}

func fieldScore(token, field string, weight int) int {
	if _, ok := tokenSet(field)[token]; ok {
		return weight
	}
	return 0
}

func tokenSet(value string) map[string]struct{} {
	set := map[string]struct{}{}
	for _, raw := range strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		set[raw] = struct{}{}
	}
	return set
}
