package export

import (
	"html"
	"strings"

	"github.com/MauricioPerera/showme/internal/domain"
)

// ExportProjectHTML renders a Project's Deck as a single self-contained
// HTML document: one <section> per slide, in order. All slide and deck
// text is HTML-escaped, so arbitrary user content can never break out of
// its element or inject markup. The output is deterministic for the same
// input.
func ExportProjectHTML(proj domain.Project) string {
	var b strings.Builder

	b.WriteString("<!doctype html>\n<html lang=\"es\">\n<head>\n")
	b.WriteString("<meta charset=\"utf-8\">\n")
	b.WriteString("<title>" + html.EscapeString(proj.Deck.Title) + "</title>\n")
	b.WriteString("</head>\n<body>\n")
	b.WriteString("<header><h1>" + html.EscapeString(proj.Deck.Title) + "</h1>")
	if proj.Deck.Audience != "" {
		b.WriteString("<p>" + html.EscapeString(proj.Deck.Audience) + "</p>")
	}
	b.WriteString("</header>\n")

	for _, slide := range proj.Deck.Slides {
		b.WriteString("<section id=\"" + html.EscapeString(slide.ID) + "\" data-status=\"" + html.EscapeString(string(slide.Status)) + "\">\n")
		b.WriteString("<h2>" + html.EscapeString(slide.Title) + "</h2>\n")
		if slide.Content != "" {
			b.WriteString("<p>" + html.EscapeString(slide.Content) + "</p>\n")
		}
		b.WriteString("</section>\n")
	}

	b.WriteString("</body>\n</html>\n")
	return b.String()
}
