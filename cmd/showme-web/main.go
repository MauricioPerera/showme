package main

// main is the thin wiring for the showme webapp: parses flags, renders
// HTML templates and serves HTTP. It is glue, not covered by a frozen
// oracle -- same criterion as cmd/showme/main.go. The actual form-handling
// logic lives in internal/web (see web-create-project-handler.md).
//
// Usage:
//
//	showme-web --dir <data-dir> [--addr :8080]
//
// --dir is required: it is where created projects are saved and has no
// hardcoded default, by design (see web-create-project-handler.md).

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/MauricioPerera/showme/internal/cli"
	"github.com/MauricioPerera/showme/internal/export"
	"github.com/MauricioPerera/showme/internal/storage"
	"github.com/MauricioPerera/showme/internal/web"
)

const formPage = `<!doctype html>
<html lang="es">
<head>
<meta charset="utf-8">
<title>showme — crear proyecto</title>
</head>
<body>
<h1>Crear presentación</h1>
<p><a href="/projects">Ver proyectos</a></p>
{{if .Result}}
  {{if .Result.OK}}
    <p>Proyecto creado: <code>{{.Result.Path}}</code></p>
  {{else}}
    <ul>
    {{range .Result.Errors}}<li>ERROR: {{.}}</li>{{end}}
    </ul>
  {{end}}
{{end}}
<form method="post" action="/projects/new">
  <p><label>Nombre<br><input type="text" name="name" required></label></p>
  <p><label>Ruta a DESIGN.md<br><input type="text" name="design_path" required></label></p>
  <p><label>Directorio del bundle OKF<br><input type="text" name="knowledge_root" required></label></p>
  <p><label>Título del deck<br><input type="text" name="deck_title" required></label></p>
  <p><label>Audiencia<br><input type="text" name="deck_audience"></label></p>
  <p><label>Título de la primera slide<br><input type="text" name="slide_title" required></label></p>
  <p><label>Intent de la primera slide<br><input type="text" name="slide_intent" required></label></p>
  <p><button type="submit">Crear</button></p>
</form>
</body>
</html>
`

const listPage = `<!doctype html>
<html lang="es">
<head>
<meta charset="utf-8">
<title>showme — proyectos</title>
</head>
<body>
<h1>Proyectos</h1>
<p><a href="/projects/new">Crear proyecto</a></p>
{{if .Errors}}
<ul>{{range .Errors}}<li>ERROR: {{.}}</li>{{end}}</ul>
{{end}}
<ul>
{{range .Projects}}
  <li><a href="/projects/view/{{.Slug}}">{{.Name}}</a>{{if .Archived}} [archived]{{end}}</li>
{{end}}
</ul>
</body>
</html>
`

const showPage = `<!doctype html>
<html lang="es">
<head>
<meta charset="utf-8">
<title>showme — {{.Project.Name}}</title>
</head>
<body>
<h1>{{.Project.Name}} (v{{.Project.Version}}){{if .Project.Archived}} [archived]{{end}}</h1>
<p><a href="/projects">&larr; Volver</a></p>
<p><a href="/projects/view/{{.Slug}}/export">Exportar a HTML</a></p>
<p>Objetivo/audiencia: {{.Project.Deck.Audience}}</p>
{{$slug := .Slug}}
<ul>
{{range .Project.Deck.Slides}}
  <li>
    [{{.Status}}] <strong>{{.Title}}</strong>{{if .Content}}<br>{{.Content}}{{end}}
    <form method="post" action="/projects/view/{{$slug}}/review" style="display:inline">
      <input type="hidden" name="slide_id" value="{{.ID}}">
      <button type="submit" name="decision" value="accepted">Aceptar</button>
      <button type="submit" name="decision" value="rejected">Rechazar</button>
    </form>
  </li>
{{end}}
</ul>
</body>
</html>
`

var formTemplate = template.Must(template.New("form").Parse(formPage))
var listTemplate = template.Must(template.New("list").Parse(listPage))
var showTemplate = template.Must(template.New("show").Parse(showPage))

type projectListEntry struct {
	Name     string
	Slug     string
	Archived bool
}

func main() {
	dir := flag.String("dir", "", "directory where created projects are saved (required)")
	addr := flag.String("addr", ":8080", "address to listen on")
	flag.Parse()

	if *dir == "" {
		fmt.Fprintln(os.Stderr, "usage: showme-web --dir <data-dir> [--addr :8080]")
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /projects/new", func(w http.ResponseWriter, r *http.Request) {
		renderForm(w, nil)
	})
	mux.HandleFunc("POST /projects/new", func(w http.ResponseWriter, r *http.Request) {
		handleCreateProject(w, r, *dir)
	})
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/projects/new", http.StatusFound)
	})
	mux.HandleFunc("GET /projects", func(w http.ResponseWriter, r *http.Request) {
		handleListProjects(w, *dir)
	})
	mux.HandleFunc("GET /projects/view/{slug}", func(w http.ResponseWriter, r *http.Request) {
		handleShowProject(w, r.PathValue("slug"), *dir)
	})
	mux.HandleFunc("POST /projects/view/{slug}/review", func(w http.ResponseWriter, r *http.Request) {
		handleReviewSlide(w, r, r.PathValue("slug"), *dir)
	})
	mux.HandleFunc("GET /projects/view/{slug}/export", func(w http.ResponseWriter, r *http.Request) {
		handleExportProject(w, r.PathValue("slug"), *dir)
	})

	log.Printf("showme-web listening on %s (data dir: %s)", *addr, *dir)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func handleCreateProject(w http.ResponseWriter, r *http.Request, dir string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := web.HandleCreateProjectForm(web.CreateProjectFormInput{
		Name:          r.FormValue("name"),
		DesignPath:    r.FormValue("design_path"),
		KnowledgeRoot: r.FormValue("knowledge_root"),
		DeckTitle:     r.FormValue("deck_title"),
		DeckAudience:  r.FormValue("deck_audience"),
		SlideTitle:    r.FormValue("slide_title"),
		SlideIntent:   r.FormValue("slide_intent"),
		Dir:           dir,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderForm(w, &result)
}

func handleListProjects(w http.ResponseWriter, dir string) {
	result, err := cli.RunListProjectsCommand(cli.ListProjectsCommandInput{Dir: dir})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	entries := make([]projectListEntry, len(result.Projects))
	for i, p := range result.Projects {
		base := filepath.Base(p.Path)
		slug := strings.TrimSuffix(base, filepath.Ext(base))
		entries[i] = projectListEntry{Name: p.Name, Slug: slug, Archived: p.Archived}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := struct {
		Projects []projectListEntry
		Errors   []string
	}{Projects: entries, Errors: result.Errors}
	if err := listTemplate.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleShowProject(w http.ResponseWriter, slug, dir string) {
	path, err := web.ProjectFilePath(dir, slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := cli.RunShowProjectCommand(cli.ShowProjectCommandInput{Path: path})
	if err != nil {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := struct {
		cli.ShowProjectCommandResult
		Slug string
	}{ShowProjectCommandResult: result, Slug: slug}
	if err := showTemplate.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleReviewSlide(w http.ResponseWriter, r *http.Request, slug, dir string) {
	path, err := web.ProjectFilePath(dir, slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = cli.RunReviewProjectCommand(cli.ReviewProjectCommandInput{
		Path:     path,
		SlideID:  r.FormValue("slide_id"),
		Decision: r.FormValue("decision"),
		OutDir:   dir,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/projects/view/"+slug, http.StatusFound)
}

func handleExportProject(w http.ResponseWriter, slug, dir string) {
	path, err := web.ProjectFilePath(dir, slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	proj, err := storage.LoadProject(path)
	if err != nil {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}

	rendered := export.ExportProjectHTML(proj)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+slug+`.html"`)
	_, _ = w.Write([]byte(rendered))
}

func renderForm(w http.ResponseWriter, result *web.CreateProjectFormResult) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := struct {
		Result *web.CreateProjectFormResult
	}{Result: result}
	if err := formTemplate.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
