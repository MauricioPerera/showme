package ai

// GenerateContentRequest is the input to a ContentGenerator.
type GenerateContentRequest struct {
	Intent  string
	Context string
}

// ContentGenerator produces slide content from an intent and its supporting
// context. Implementations may call an external AI provider; this
// interface itself makes no assumption about that.
type ContentGenerator interface {
	GenerateContent(request GenerateContentRequest) (string, error)
}

// GenerateSlideContentInput is the raw data used by GenerateSlideContent.
type GenerateSlideContentInput struct {
	Generator ContentGenerator
	Intent    string
	Context   string
}

// GenerateSlideContentResult is the outcome of GenerateSlideContent.
type GenerateSlideContentResult struct {
	Content string
}

// Report collects the problems found while generating slide content.
type Report struct {
	Errors   []string
	Warnings []string
}

// GenerateSlideContent validates Intent and delegates to Generator to
// produce a slide's content. This function itself never touches the
// network: it depends only on the injected ContentGenerator interface, so
// the core can be tested with a deterministic fake (see
// generate-slide-content-usecase's Do/Don't) instead of a real provider.
func GenerateSlideContent(input GenerateSlideContentInput) (GenerateSlideContentResult, Report) {
	report := Report{}

	if input.Intent == "" {
		report.Errors = append(report.Errors, "intent is required")
		return GenerateSlideContentResult{}, report
	}

	content, err := input.Generator.GenerateContent(GenerateContentRequest{
		Intent:  input.Intent,
		Context: input.Context,
	})
	if err != nil {
		report.Errors = append(report.Errors, err.Error())
		return GenerateSlideContentResult{}, report
	}
	if content == "" {
		report.Errors = append(report.Errors, "generator returned empty content")
		return GenerateSlideContentResult{}, report
	}

	return GenerateSlideContentResult{Content: content}, report
}
