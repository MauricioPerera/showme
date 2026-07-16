---
type: 'Task Contract'
title: 'CLI: comando project add-slide'
description: 'Logica pura del comando project add-slide: agrega una slide a un Project guardado y lo re-persiste.'
tags: ['showme', 'go', 'cli', 'project', 'slide']

task: cli-add-slide-command
intent: "Ejecutar el comando 'project add-slide': agregar una slide a un proyecto guardado y guardarlo actualizado."
target: internal/cli/add_slide_command.go
signature: "func RunAddSlideCommand(input AddSlideCommandInput) (AddSlideCommandResult, error)"
test_command: "go test ./internal/cli"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 8
  max_nesting_depth: 3
  cyclomatic_max: 8
  nesting_max: 3
  params_max: 1
  lines_max: 100
tests: internal/cli/add_slide_command_test.go
tests_sha256: "b8464fd974d87237fbbcc735fb0f429530d5e8ba8f08244554a87cedb07d5827"
touch_only: ['internal/cli/add_slide_command.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: cli-add-slide-command

## Intent

Expone [add-slide-usecase](./add-slide-usecase.md) por CLI, mismo patron
que [cli-review-project-command](./cli-review-project-command.md): carga un
`Project` guardado, agrega una slide nueva a su `Deck` con
`domain.AddSlide` y, si es valida, re-guarda el proyecto con
[save-project-usecase](./save-project-usecase.md). Permite a un agente
poblar un deck incrementalmente en vez de tener que declarar todas las
slides en `project create`.

## Interface

```go
type AddSlideCommandInput struct {
    Path, SlideID, Title, Intent, Content, Status, OutDir string
}

type AddSlideCommandResult struct {
    OK       bool
    Path     string
    Errors, Warnings []string
}

func RunAddSlideCommand(input AddSlideCommandInput) (AddSlideCommandResult, error)
```

## Invariants

- Un error cargando `Path` (archivo inexistente o JSON invalido) se
  propaga tal cual via `err`; no se intenta agregar ninguna slide.
- La slide nueva se valida y agrega con `domain.AddSlide` sobre el `Deck`
  cargado; si `AddSlide` reporta errores (id/titulo vacio, id duplicado,
  status invalido), el resultado tiene `OK: false`, `Path: ""` y esos
  errores en `Errors` — el proyecto en disco NO se toca.
- Si la slide es valida, el `Project` actualizado (mismo `Name`,
  `DesignPath`, `KnowledgePath`, `Version`, `Archived`; `Deck` con la slide
  nueva al final) se guarda con `storage.SaveProject` bajo `OutDir`. Si
  `OutDir` y `Name` coinciden con el archivo original, esto sobreescribe el
  mismo archivo.
- `Archived` se preserva tal cual estaba (no se resetea a `false`): este
  comando pasa `Archived: proj.Archived` al reconstruir el `ProjectInput`
  de guardado, siguiendo la convencion fijada por
  [project-model](./project-model.md).
- Un error de I/O al guardar se propaga via `err`.
- No hace red, subprocess ni llamadas a un proveedor de IA.

## Examples

- Proyecto guardado con una slide `intro`, `SlideID: "closing"`,
  `Title: "Cierre"`, `OutDir` igual al directorio original -> `OK: true`,
  `Path` igual al archivo original; al recargarlo el deck tiene 2 slides,
  la segunda `closing`.
- `SlideID: "intro"` (ya existente) -> `OK: false`, `Errors` incluye
  `duplicate slide id: intro`, el archivo original queda sin cambios.
- `Path` inexistente -> `err` no nil.

## Do / Don't

- DO: reusar `storage.LoadProject`, `domain.AddSlide` y
  `storage.SaveProject` tal cual; este comando es orquestacion pura.
- DO: preservar `Name`, `DesignPath`, `KnowledgePath` y `Version` del
  proyecto original al re-guardar; solo el `Deck` cambia.
- DON'T: permitir insertar en una posicion especifica ni reordenar slides
  — la slide nueva siempre va al final.
- DON'T: imprimir a stdout ni parsear flags aqui — eso vive en
  `cmd/showme/main.go`.

## Tests

Los tests estan en `internal/cli/add_slide_command_test.go` y cubren:
slide valida que persiste el cambio en el mismo archivo, id duplicado,
preservacion de `Archived` a traves de la edicion, y archivo origen
inexistente.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta eliminar o reordenar slides para cumplir
  el intent — eso excede el alcance de este contrato.
