---
type: 'Task Contract'
title: 'Webapp: handler de crear proyecto con multiples slides'
description: 'Convierte un nombre, deck y una lista arbitraria de slides en un deck.json temporal con ids unicos, y crea el proyecto via cli.RunCreateProjectCommand.'
tags: ['showme', 'go', 'web', 'project', 'storyboard']

task: web-create-project-with-slides-handler
intent: "Crear un proyecto a partir de un numero arbitrario de slides (ej. un storyboard revisado), generando ids unicos por titulo."
target: internal/web/create_project_with_slides.go
signature: "func HandleCreateProjectWithSlides(input CreateProjectWithSlidesInput) (CreateProjectFormResult, error)"
test_command: "go test ./internal/web"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 6
  max_nesting_depth: 2
  cyclomatic_max: 6
  nesting_max: 2
  params_max: 1
  lines_max: 100
tests: internal/web/create_project_with_slides_test.go
tests_sha256: "ec18e61d41e71c9e9a590432fe700edb5f0df8f80346475c24ca4170c1476194"
touch_only: ['internal/web/create_project_with_slides.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: web-create-project-with-slides-handler

## Intent

Complemento de [web-create-project-handler](./web-create-project-handler.md)
(que solo soporta exactamente una slide inicial): permite crear un proyecto
con un numero arbitrario de slides, tipicamente las propuestas por
[web-propose-storyboard-handler](./web-propose-storyboard-handler.md) y
revisadas por la persona antes de confirmar la creacion. Mismo patron de
delegacion: arma un `deck.json` temporal (siempre borrado) y delega
enteramente en `cli.RunCreateProjectCommand`.

## Interface

```go
type SlideInput struct {
    Title, Intent string
}

type CreateProjectWithSlidesInput struct {
    Name, DesignPath, KnowledgeRoot string
    DeckTitle, DeckAudience string
    Slides []SlideInput
    Dir string
}

func HandleCreateProjectWithSlides(input CreateProjectWithSlidesInput) (CreateProjectFormResult, error)
```

## Invariants

- Cada `SlideInput` recibe un `ID` deterministico via `storage.Slugify`
  sobre su `Title`; colisiones entre titulos que producen el mismo slug se
  resuelven agregando un sufijo numerico (`-2`, `-3`, ...), mismo criterio
  que [cli-generate-storyboard-command](./cli-generate-storyboard-command.md).
- `Slides` vacio no es un error de esta funcion: se delega a
  `cli.RunCreateProjectCommand`, que via `domain.NewDeck` devuelve
  `at least one slide is required` en `Errors`.
- El `deck.json` temporal SIEMPRE se borra antes de retornar
  (`defer os.Remove`), exista o no un error.
- Un error de I/O (leer `DesignPath`, escribir el archivo temporal, o
  guardar bajo `Dir`) se propaga tal cual via `err`.
- No hace red, subprocess ni llamadas a un proveedor de IA — la
  generacion con IA ya ocurrio en un paso previo
  (`web-propose-storyboard-handler`); esta funcion solo persiste.

## Examples

- `Slides: [{Title: "Introduccion", ...}, {Title: "Plan", ...}]` -> `OK:
  true`, proyecto guardado con 2 slides con ids unicos (`introduccion`,
  `plan`).
- Dos slides con el mismo `Title` (`"Introduccion"`) -> ids
  `introduccion` y `introduccion-2`.
- `Slides: nil` -> `OK: false`, `Errors` incluye
  `at least one slide is required`.
- `Dir` inexistente -> `err` no nil.

## Do / Don't

- DO: reusar `cli.RunCreateProjectCommand` y los tipos
  `formDeckSlide`/`formDeckInput` ya definidos en
  `web-create-project-handler` — no duplicar el shape del deck JSON.
- DO: garantizar ids unicos entre las slides antes de escribir el
  archivo temporal.
- DON'T: llamar a un proveedor de IA desde aqui — eso ya paso en
  `web-propose-storyboard-handler`; esta funcion solo recibe texto ya
  generado/revisado.
- DON'T: reemplazar `web-create-project-handler` (una sola slide) — este
  es un flujo adicional, no un reemplazo.

## Tests

Los tests estan en `internal/web/create_project_with_slides_test.go` y
cubren: creacion valida con ids unicos, colision de titulos deduplicada,
lista de slides vacia, y directorio de datos inexistente.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta invocar un proveedor de IA desde este
  archivo para cumplir el intent — eso excede el alcance de este
  contrato.
