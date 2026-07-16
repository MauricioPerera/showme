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

var formTemplate = template.Must(template.New("form").Parse(formPage))

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

func renderForm(w http.ResponseWriter, result *web.CreateProjectFormResult) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := struct {
		Result *web.CreateProjectFormResult
	}{Result: result}
	if err := formTemplate.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
