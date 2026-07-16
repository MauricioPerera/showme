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
<form method="post" action="/projects/view/{{.Slug}}/archive" style="display:inline">
  {{if .Project.Archived}}
    <input type="hidden" name="archived" value="false">
    <button type="submit">Desarchivar</button>
  {{else}}
    <input type="hidden" name="archived" value="true">
    <button type="submit">Archivar</button>
  {{end}}
</form>
<form method="post" action="/projects/view/{{.Slug}}/duplicate" style="display:inline">
  <input type="text" name="new_name" placeholder="Nombre de la copia" required>
  <button type="submit">Duplicar</button>
</form>
<form method="post" action="/projects/view/{{.Slug}}/rename" style="display:inline">
  <input type="text" name="new_name" placeholder="Nuevo nombre" required>
  <button type="submit">Renombrar</button>
</form>
{{if .Errors}}
<ul>{{range .Errors}}<li>ERROR: {{.}}</li>{{end}}</ul>
{{end}}
<p>Objetivo/audiencia: {{.Project.Deck.Audience}}</p>
{{$slug := .Slug}}
<form method="post" action="/projects/view/{{.Slug}}/generate-all" style="border:1px solid #ccc;padding:8px;display:inline-block">
  <label>Base URL IA <input type="text" name="base_url" placeholder="http://127.0.0.1:8080/v1" required></label>
  <label>Modelo <input type="text" name="model" required></label>
  <button type="submit">Generar todas las slides pendientes</button>
</form>
<ul>
{{range .Project.Deck.Slides}}
  <li>
    [{{.Status}}] <strong>{{.Title}}</strong>{{if .Content}}<br>{{.Content}}{{end}}
    <form method="post" action="/projects/view/{{$slug}}/review" style="display:inline">
      <input type="hidden" name="slide_id" value="{{.ID}}">
      <button type="submit" name="decision" value="accepted">Aceptar</button>
      <button type="submit" name="decision" value="rejected">Rechazar</button>
    </form>
    <form method="post" action="/projects/view/{{$slug}}/generate" style="display:inline">
      <input type="hidden" name="slide_id" value="{{.ID}}">
      <input type="text" name="base_url" placeholder="Base URL IA" required>
      <input type="text" name="model" placeholder="Modelo" required>
      <button type="submit">Generar</button>
    </form>
    <form method="post" action="/projects/view/{{$slug}}/remove-slide" style="display:inline">
      <input type="hidden" name="slide_id" value="{{.ID}}">
      <button type="submit">Quitar</button>
    </form>
  </li>
{{end}}
</ul>
<form method="post" action="/projects/view/{{.Slug}}/add-slide" style="border:1px solid #ccc;padding:8px;display:inline-block">
  <label>ID <input type="text" name="slide_id" required></label>
  <label>Título <input type="text" name="title" required></label>
  <label>Intent <input type="text" name="intent"></label>
  <button type="submit">Agregar slide</button>
</form>
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
	mux.HandleFunc("POST /projects/view/{slug}/archive", func(w http.ResponseWriter, r *http.Request) {
		handleArchiveProject(w, r, r.PathValue("slug"), *dir)
	})
	mux.HandleFunc("POST /projects/view/{slug}/generate", func(w http.ResponseWriter, r *http.Request) {
		handleGenerateSlide(w, r, r.PathValue("slug"), *dir)
	})
	mux.HandleFunc("POST /projects/view/{slug}/generate-all", func(w http.ResponseWriter, r *http.Request) {
		handleGenerateAllSlides(w, r, r.PathValue("slug"), *dir)
	})
	mux.HandleFunc("POST /projects/view/{slug}/add-slide", func(w http.ResponseWriter, r *http.Request) {
		handleAddSlide(w, r, r.PathValue("slug"), *dir)
	})
	mux.HandleFunc("POST /projects/view/{slug}/remove-slide", func(w http.ResponseWriter, r *http.Request) {
		handleRemoveSlide(w, r, r.PathValue("slug"), *dir)
	})
	mux.HandleFunc("POST /projects/view/{slug}/duplicate", func(w http.ResponseWriter, r *http.Request) {
		handleDuplicateProject(w, r, r.PathValue("slug"), *dir)
	})
	mux.HandleFunc("POST /projects/view/{slug}/rename", func(w http.ResponseWriter, r *http.Request) {
		handleRenameProject(w, r, r.PathValue("slug"), *dir)
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
		entries[i] = projectListEntry{Name: p.Name, Slug: slugFromPath(p.Path), Archived: p.Archived}
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

	renderShowPage(w, result, slug, nil)
}

func renderShowPage(w http.ResponseWriter, result cli.ShowProjectCommandResult, slug string, errs []string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := struct {
		cli.ShowProjectCommandResult
		Slug   string
		Errors []string
	}{ShowProjectCommandResult: result, Slug: slug, Errors: errs}
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

func handleArchiveProject(w http.ResponseWriter, r *http.Request, slug, dir string) {
	path, err := web.ProjectFilePath(dir, slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = cli.RunArchiveProjectCommand(cli.ArchiveProjectCommandInput{
		Path:     path,
		Archived: r.FormValue("archived") == "true",
		OutDir:   dir,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/projects/view/"+slug, http.StatusFound)
}

func handleGenerateSlide(w http.ResponseWriter, r *http.Request, slug, dir string) {
	path, err := web.ProjectFilePath(dir, slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := cli.RunGenerateSlideContentCommand(cli.GenerateSlideContentCommandInput{
		Path:    path,
		SlideID: r.FormValue("slide_id"),
		BaseURL: r.FormValue("base_url"),
		Model:   r.FormValue("model"),
		OutDir:  dir,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !result.OK {
		showResult, showErr := cli.RunShowProjectCommand(cli.ShowProjectCommandInput{Path: path})
		if showErr != nil {
			http.Error(w, showErr.Error(), http.StatusInternalServerError)
			return
		}
		renderShowPage(w, showResult, slug, result.Errors)
		return
	}

	http.Redirect(w, r, "/projects/view/"+slug, http.StatusFound)
}

func handleGenerateAllSlides(w http.ResponseWriter, r *http.Request, slug, dir string) {
	path, err := web.ProjectFilePath(dir, slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := cli.RunGenerateAllSlidesCommand(cli.GenerateAllSlidesCommandInput{
		Path:    path,
		BaseURL: r.FormValue("base_url"),
		Model:   r.FormValue("model"),
		OutDir:  dir,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !result.OK {
		showResult, showErr := cli.RunShowProjectCommand(cli.ShowProjectCommandInput{Path: path})
		if showErr != nil {
			http.Error(w, showErr.Error(), http.StatusInternalServerError)
			return
		}
		renderShowPage(w, showResult, slug, result.Errors)
		return
	}

	http.Redirect(w, r, "/projects/view/"+slug, http.StatusFound)
}

func handleAddSlide(w http.ResponseWriter, r *http.Request, slug, dir string) {
	path, err := web.ProjectFilePath(dir, slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := cli.RunAddSlideCommand(cli.AddSlideCommandInput{
		Path:    path,
		SlideID: r.FormValue("slide_id"),
		Title:   r.FormValue("title"),
		Intent:  r.FormValue("intent"),
		OutDir:  dir,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !result.OK {
		showResult, showErr := cli.RunShowProjectCommand(cli.ShowProjectCommandInput{Path: path})
		if showErr != nil {
			http.Error(w, showErr.Error(), http.StatusInternalServerError)
			return
		}
		renderShowPage(w, showResult, slug, result.Errors)
		return
	}

	http.Redirect(w, r, "/projects/view/"+slug, http.StatusFound)
}

func handleRemoveSlide(w http.ResponseWriter, r *http.Request, slug, dir string) {
	path, err := web.ProjectFilePath(dir, slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := cli.RunRemoveSlideCommand(cli.RemoveSlideCommandInput{
		Path:    path,
		SlideID: r.FormValue("slide_id"),
		OutDir:  dir,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !result.OK {
		showResult, showErr := cli.RunShowProjectCommand(cli.ShowProjectCommandInput{Path: path})
		if showErr != nil {
			http.Error(w, showErr.Error(), http.StatusInternalServerError)
			return
		}
		renderShowPage(w, showResult, slug, result.Errors)
		return
	}

	http.Redirect(w, r, "/projects/view/"+slug, http.StatusFound)
}

func handleDuplicateProject(w http.ResponseWriter, r *http.Request, slug, dir string) {
	path, err := web.ProjectFilePath(dir, slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := cli.RunDuplicateProjectCommand(cli.DuplicateProjectCommandInput{
		SourcePath: path,
		NewName:    r.FormValue("new_name"),
		OutDir:     dir,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !result.OK {
		showResult, showErr := cli.RunShowProjectCommand(cli.ShowProjectCommandInput{Path: path})
		if showErr != nil {
			http.Error(w, showErr.Error(), http.StatusInternalServerError)
			return
		}
		renderShowPage(w, showResult, slug, result.Errors)
		return
	}

	newSlug := slugFromPath(result.Path)
	http.Redirect(w, r, "/projects/view/"+newSlug, http.StatusFound)
}

func handleRenameProject(w http.ResponseWriter, r *http.Request, slug, dir string) {
	path, err := web.ProjectFilePath(dir, slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := cli.RunRenameProjectCommand(cli.RenameProjectCommandInput{
		SourcePath: path,
		NewName:    r.FormValue("new_name"),
		OutDir:     dir,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !result.OK {
		showResult, showErr := cli.RunShowProjectCommand(cli.ShowProjectCommandInput{Path: path})
		if showErr != nil {
			http.Error(w, showErr.Error(), http.StatusInternalServerError)
			return
		}
		renderShowPage(w, showResult, slug, result.Errors)
		return
	}

	newSlug := slugFromPath(result.Path)
	http.Redirect(w, r, "/projects/view/"+newSlug, http.StatusFound)
}

func slugFromPath(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
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
