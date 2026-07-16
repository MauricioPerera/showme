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
	"strconv"
	"strings"

	"github.com/MauricioPerera/showme/internal/ai"
	"github.com/MauricioPerera/showme/internal/cli"
	"github.com/MauricioPerera/showme/internal/domain"
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
<p><a href="/projects">Ver proyectos</a> · <a href="/projects/new/storyboard">Proponer storyboard con IA</a></p>
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

const storyboardFormPage = `<!doctype html>
<html lang="es">
<head>
<meta charset="utf-8">
<title>showme — proponer storyboard</title>
</head>
<body>
<h1>Proponer storyboard con IA</h1>
<p><a href="/projects/new">&larr; Volver a crear proyecto</a></p>
{{if .Errors}}
<ul>{{range .Errors}}<li>ERROR: {{.}}</li>{{end}}</ul>
{{end}}
<form method="post" action="/projects/new/storyboard/propose">
  <p><label>Nombre<br><input type="text" name="name" required></label></p>
  <p><label>Ruta a DESIGN.md<br><input type="text" name="design_path" required></label></p>
  <p><label>Directorio del bundle OKF<br><input type="text" name="knowledge_root" required></label></p>
  <p><label>Título del deck<br><input type="text" name="deck_title" required></label></p>
  <p><label>Audiencia<br><input type="text" name="deck_audience"></label></p>
  <p><label>Objetivo de la presentación<br><input type="text" name="objective" required></label></p>
  <p><label>Cantidad de slides<br><input type="number" name="count" value="5" required></label></p>
  <p><label>Base URL IA<br><input type="text" name="base_url" placeholder="http://127.0.0.1:8080/v1" required></label></p>
  <p><label>Modelo<br><input type="text" name="model" required></label></p>
  <p><button type="submit">Proponer storyboard</button></p>
</form>
</body>
</html>
`

const storyboardReviewPage = `<!doctype html>
<html lang="es">
<head>
<meta charset="utf-8">
<title>showme — revisar storyboard</title>
</head>
<body>
<h1>Revisar storyboard propuesto</h1>
{{if .Errors}}
<ul>{{range .Errors}}<li>ERROR: {{.}}</li>{{end}}</ul>
<p><a href="/projects/new/storyboard">&larr; Volver a intentar</a></p>
{{else}}
<form method="post" action="/projects/new/storyboard/create">
  <input type="hidden" name="name" value="{{.Name}}">
  <input type="hidden" name="design_path" value="{{.DesignPath}}">
  <input type="hidden" name="knowledge_root" value="{{.KnowledgeRoot}}">
  <input type="hidden" name="deck_title" value="{{.DeckTitle}}">
  <input type="hidden" name="deck_audience" value="{{.DeckAudience}}">
  {{range .Slides}}
    <fieldset>
      <p><label>Título<br><input type="text" name="slide_title" value="{{.Title}}" required></label></p>
      <p><label>Intent<br><input type="text" name="slide_intent" value="{{.Intent}}"></label></p>
    </fieldset>
  {{end}}
  <p><button type="submit">Crear proyecto con estas slides</button></p>
</form>
{{end}}
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
<form method="post" action="/projects/view/{{.Slug}}/update-info" style="border:1px solid #ccc;padding:8px;display:inline-block">
  <label>Título del deck<br><input type="text" name="title" value="{{.Project.Deck.Title}}" required></label>
  <label>Audiencia<br><input type="text" name="audience" value="{{.Project.Deck.Audience}}"></label>
  <button type="submit">Actualizar info</button>
</form>
<form method="post" action="/projects/view/{{.Slug}}/reorder-slides" style="border:1px solid #ccc;padding:8px;display:inline-block">
  <label>Orden de slides (IDs separados por coma)<br>
    <input type="text" name="order" value="{{slideIDs .Project.Deck.Slides}}" size="60" required>
  </label>
  <button type="submit">Reordenar</button>
</form>
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
    <details style="display:inline-block;vertical-align:top">
      <summary>Editar</summary>
      <form method="post" action="/projects/view/{{$slug}}/update-slide">
        <input type="hidden" name="slide_id" value="{{.ID}}">
        <p><label>Título<br><input type="text" name="title" value="{{.Title}}" required></label></p>
        <p><label>Intent<br><input type="text" name="intent" value="{{.Intent}}"></label></p>
        <p><label>Contenido<br><textarea name="content">{{.Content}}</textarea></label></p>
        <p><label>Status<br>
          <select name="status">
            <option value="">(mantener actual)</option>
            <option value="draft">draft</option>
            <option value="accepted">accepted</option>
            <option value="rejected">rejected</option>
          </select>
        </label></p>
        <button type="submit">Guardar</button>
      </form>
    </details>
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
var storyboardFormTemplate = template.Must(template.New("storyboardForm").Parse(storyboardFormPage))
var storyboardReviewTemplate = template.Must(template.New("storyboardReview").Parse(storyboardReviewPage))
var listTemplate = template.Must(template.New("list").Parse(listPage))
var showTemplate = template.Must(template.New("show").Funcs(template.FuncMap{
	"slideIDs": func(slides []domain.Slide) string {
		ids := make([]string, len(slides))
		for i, s := range slides {
			ids[i] = s.ID
		}
		return strings.Join(ids, ",")
	},
}).Parse(showPage))

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
	mux.HandleFunc("GET /projects/new/storyboard", func(w http.ResponseWriter, r *http.Request) {
		renderStoryboardForm(w, nil)
	})
	mux.HandleFunc("POST /projects/new/storyboard/propose", func(w http.ResponseWriter, r *http.Request) {
		handleProposeStoryboard(w, r)
	})
	mux.HandleFunc("POST /projects/new/storyboard/create", func(w http.ResponseWriter, r *http.Request) {
		handleCreateProjectFromStoryboard(w, r, *dir)
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
	mux.HandleFunc("POST /projects/view/{slug}/update-slide", func(w http.ResponseWriter, r *http.Request) {
		handleUpdateSlide(w, r, r.PathValue("slug"), *dir)
	})
	mux.HandleFunc("POST /projects/view/{slug}/update-info", func(w http.ResponseWriter, r *http.Request) {
		handleUpdateDeckInfo(w, r, r.PathValue("slug"), *dir)
	})
	mux.HandleFunc("POST /projects/view/{slug}/reorder-slides", func(w http.ResponseWriter, r *http.Request) {
		handleReorderSlides(w, r, r.PathValue("slug"), *dir)
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

func renderStoryboardForm(w http.ResponseWriter, errs []string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := struct{ Errors []string }{Errors: errs}
	if err := storyboardFormTemplate.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleProposeStoryboard(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	count, _ := strconv.Atoi(r.FormValue("count"))
	result, err := web.HandleProposeStoryboard(web.ProposeStoryboardInput{
		Objective:     r.FormValue("objective"),
		Audience:      r.FormValue("deck_audience"),
		KnowledgeRoot: r.FormValue("knowledge_root"),
		BaseURL:       r.FormValue("base_url"),
		Model:         r.FormValue("model"),
		Count:         count,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(result.Errors) != 0 {
		renderStoryboardForm(w, result.Errors)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := struct {
		Name, DesignPath, KnowledgeRoot, DeckTitle, DeckAudience string
		Slides                                                   []ai.StoryboardSlide
		Errors                                                   []string
	}{
		Name:          r.FormValue("name"),
		DesignPath:    r.FormValue("design_path"),
		KnowledgeRoot: r.FormValue("knowledge_root"),
		DeckTitle:     r.FormValue("deck_title"),
		DeckAudience:  r.FormValue("deck_audience"),
		Slides:        result.Slides,
	}
	if err := storyboardReviewTemplate.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleCreateProjectFromStoryboard(w http.ResponseWriter, r *http.Request, dir string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	titles := r.PostForm["slide_title"]
	intents := r.PostForm["slide_intent"]
	slides := make([]web.SlideInput, len(titles))
	for i, title := range titles {
		intent := ""
		if i < len(intents) {
			intent = intents[i]
		}
		slides[i] = web.SlideInput{Title: title, Intent: intent}
	}

	result, err := web.HandleCreateProjectWithSlides(web.CreateProjectWithSlidesInput{
		Name:          r.FormValue("name"),
		DesignPath:    r.FormValue("design_path"),
		KnowledgeRoot: r.FormValue("knowledge_root"),
		DeckTitle:     r.FormValue("deck_title"),
		DeckAudience:  r.FormValue("deck_audience"),
		Slides:        slides,
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

func handleUpdateSlide(w http.ResponseWriter, r *http.Request, slug, dir string) {
	path, err := web.ProjectFilePath(dir, slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := cli.RunUpdateSlideCommand(cli.UpdateSlideCommandInput{
		Path:    path,
		SlideID: r.FormValue("slide_id"),
		Title:   r.FormValue("title"),
		Intent:  r.FormValue("intent"),
		Content: r.FormValue("content"),
		Status:  r.FormValue("status"),
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

func handleUpdateDeckInfo(w http.ResponseWriter, r *http.Request, slug, dir string) {
	path, err := web.ProjectFilePath(dir, slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := cli.RunUpdateDeckInfoCommand(cli.UpdateDeckInfoCommandInput{
		Path:     path,
		Title:    r.FormValue("title"),
		Audience: r.FormValue("audience"),
		OutDir:   dir,
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

func handleReorderSlides(w http.ResponseWriter, r *http.Request, slug, dir string) {
	path, err := web.ProjectFilePath(dir, slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var order []string
	if raw := r.FormValue("order"); raw != "" {
		order = strings.Split(raw, ",")
	}

	result, err := cli.RunReorderSlidesCommand(cli.ReorderSlidesCommandInput{
		Path:   path,
		Order:  order,
		OutDir: dir,
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
