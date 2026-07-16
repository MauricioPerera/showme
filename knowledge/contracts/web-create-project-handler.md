---
type: 'Task Contract'
title: 'Webapp: handler del formulario crear proyecto'
description: 'Convierte los campos de un formulario web en un deck.json temporal y delega en cli.RunCreateProjectCommand.'
tags: ['showme', 'go', 'web', 'project']

task: web-create-project-handler
intent: "Convertir los campos del formulario web de creacion de proyecto en un deck.json temporal y crear el proyecto."
target: internal/web/create_project_handler.go
signature: "func HandleCreateProjectForm(input CreateProjectFormInput) (CreateProjectFormResult, error)"
test_command: "go test ./internal/web"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 8
  max_nesting_depth: 3
  cyclomatic_max: 8
  nesting_max: 3
  params_max: 1
  lines_max: 110
tests: internal/web/create_project_handler_test.go
tests_sha256: "42a98c776e49890bdf5d1e370b976abe13899e63875093600307aed04003edbf"
touch_only: ['internal/web/create_project_handler.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: web-create-project-handler

## Intent

Primera pagina de la Piel 1 (webapp) de `showme` (`DEFINITION.md`, "Producto
y capacidades" > Webapp: "crear... presentaciones"). Cubre la logica pura
(sin HTTP) de convertir los campos de un formulario en el mismo insumo que
ya consume [cli-create-project-command](./cli-create-project-command.md) —
NO reimplementa esa validacion ni la creacion del proyecto: escribe un
`deck.json` temporal y delega enteramente en `cli.RunCreateProjectCommand`.
El wiring HTTP real (parseo de `multipart/form`, renderizado de la
respuesta) vive en `cmd/showme-web/main.go`, glue no cubierta por oraculo,
mismo criterio que `cmd/showme/main.go`.

## Interface

```go
type CreateProjectFormInput struct {
    Name, DesignPath, KnowledgeRoot string
    DeckTitle, DeckAudience string
    SlideTitle, SlideIntent string
    Dir string
}

type CreateProjectFormResult struct {
    OK       bool
    Path     string
    Errors, Warnings []string
}

func HandleCreateProjectForm(input CreateProjectFormInput) (CreateProjectFormResult, error)
```

## Invariants

- Arma un `deck.json` con exactamente una slide (`ID: "slide-1"`,
  `Title: SlideTitle`, `Intent: SlideIntent`) — la primera version del
  formulario solo soporta una slide inicial; agregar mas es
  responsabilidad de `project add-slide` (CLI) despues de crear el
  proyecto, o de un formulario dinamico futuro.
- El `deck.json` se escribe en un archivo temporal (`os.CreateTemp`) que
  SIEMPRE se borra antes de retornar (`defer os.Remove`), exista o no un
  error.
- Delega enteramente en `cli.RunCreateProjectCommand` con ese archivo
  temporal como `DeckPath`; no reimplementa ninguna validacion de
  `DESIGN.md`, del bundle OKF ni del deck.
- Un error de I/O (leer `DesignPath`, escribir el archivo temporal, o
  guardar bajo `Dir`) se propaga tal cual via `err`.
- Problemas de validacion (ej. `DESIGN.md` sin frontmatter) se devuelven
  en `Errors`, con `OK: false` y `Path` vacio; no se persiste nada.
- No hace red, subprocess ni llamadas a un proveedor de IA.

## Examples

- Formulario con `Name`, `DesignPath`, `KnowledgeRoot`, `DeckTitle` y una
  slide validos -> `OK: true`, `Path` apuntando al proyecto guardado bajo
  `Dir`.
- `DesignPath` inexistente -> `err` no nil.
- `DESIGN.md` sin frontmatter -> `OK: false`, `Errors` incluye
  `frontmatter is required`, `Dir` queda sin archivos nuevos.
- `Dir` inexistente -> `err` no nil.

## Do / Don't

- DO: reusar `cli.RunCreateProjectCommand` tal cual; este handler es
  puramente adaptacion de formulario a deck JSON.
- DO: borrar siempre el archivo temporal, incluso en el camino de error.
- DON'T: parsear el `http.Request` ni renderizar HTML aqui — eso vive en
  `cmd/showme-web/main.go`.
- DON'T: soportar mas de una slide inicial en este formulario todavia —
  es una limitacion explicita de esta primera version, no un descuido.

## Tests

Los tests estan en `internal/web/create_project_handler_test.go` y
cubren: formulario valido con proyecto persistido y campos correctos,
archivo de diseno inexistente, errores de validacion que no escriben
archivo, y directorio de datos inexistente.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta soportar multiples slides o carga de
  archivos (en vez de rutas de texto) para cumplir el intent — eso excede
  el alcance de este contrato.
