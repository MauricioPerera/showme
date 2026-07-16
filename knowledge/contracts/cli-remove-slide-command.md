---
type: 'Task Contract'
title: 'CLI: comando project remove-slide'
description: 'Logica pura del comando project remove-slide: quita una slide de un Project guardado y lo re-persiste.'
tags: ['showme', 'go', 'cli', 'project', 'slide']

task: cli-remove-slide-command
intent: "Ejecutar el comando 'project remove-slide': quitar una slide de un proyecto guardado y guardarlo actualizado."
target: internal/cli/remove_slide_command.go
signature: "func RunRemoveSlideCommand(input RemoveSlideCommandInput) (RemoveSlideCommandResult, error)"
test_command: "go test ./internal/cli"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 8
  max_nesting_depth: 3
  cyclomatic_max: 8
  nesting_max: 3
  params_max: 1
  lines_max: 100
tests: internal/cli/remove_slide_command_test.go
tests_sha256: "72af62febc76ac782e36fe616be0e6e675dd3c15120da0b855d563920ef2e4d3"
touch_only: ['internal/cli/remove_slide_command.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: cli-remove-slide-command

## Intent

Expone [remove-slide-usecase](./remove-slide-usecase.md) por CLI, mismo
patron que [cli-add-slide-command](./cli-add-slide-command.md): carga un
`Project` guardado, quita una de sus slides con `domain.RemoveSlide` y, si
es valido, re-guarda el proyecto con
[save-project-usecase](./save-project-usecase.md).

## Interface

```go
type RemoveSlideCommandInput struct {
    Path, SlideID, OutDir string
}

type RemoveSlideCommandResult struct {
    OK       bool
    Path     string
    Errors, Warnings []string
}

func RunRemoveSlideCommand(input RemoveSlideCommandInput) (RemoveSlideCommandResult, error)
```

## Invariants

- Un error cargando `Path` (archivo inexistente o JSON invalido) se
  propaga tal cual via `err`; no se intenta quitar ninguna slide.
- La remocion se valida con `domain.RemoveSlide` sobre el `Deck` cargado;
  si reporta errores (`slide not found: <id>` o
  `deck must have at least one slide`), el resultado tiene `OK: false`,
  `Path: ""` y esos errores en `Errors` — el proyecto en disco NO se toca.
- Si la remocion es valida, el `Project` actualizado (mismo `Name`,
  `DesignPath`, `KnowledgePath`, `Version`, `Archived`; `Deck` sin la slide
  quitada) se guarda con `storage.SaveProject` bajo `OutDir`. Si `OutDir` y
  `Name` coinciden con el archivo original, esto sobreescribe el mismo
  archivo.
- `Archived` se preserva tal cual estaba (no se resetea a `false`), misma
  convencion que [cli-add-slide-command](./cli-add-slide-command.md).
- `Runs` (el historial de generaciones de IA) tambien se preserva tal
  cual, por la misma razon (ver
  [append-generation-run-usecase](./append-generation-run-usecase.md)).
- Un error de I/O al guardar se propaga via `err`.
- No hace red, subprocess ni llamadas a un proveedor de IA.

## Examples

- Proyecto guardado con slides `intro` y `closing`, `SlideID: "closing"`,
  `OutDir` igual al directorio original -> `OK: true`, `Path` igual al
  archivo original; al recargarlo el deck tiene solo `intro`.
- Proyecto con una unica slide, `SlideID` de esa slide -> `OK: false`,
  `Errors` incluye `deck must have at least one slide`.
- `Path` inexistente -> `err` no nil.

## Do / Don't

- DO: reusar `storage.LoadProject`, `domain.RemoveSlide` y
  `storage.SaveProject` tal cual; este comando es orquestacion pura.
- DO: preservar `Name`, `DesignPath`, `KnowledgePath` y `Version` del
  proyecto original al re-guardar; solo el `Deck` cambia.
- DON'T: permitir quitar todas las slides de un proyecto — eso lo rechaza
  `domain.RemoveSlide`, no se relaja aqui.
- DON'T: imprimir a stdout ni parsear flags aqui — eso vive en
  `cmd/showme/main.go`.

## Tests

Los tests estan en `internal/cli/remove_slide_command_test.go` y cubren:
remocion valida que persiste el cambio en el mismo archivo, remocion de la
unica slide rechazada, y archivo origen inexistente.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta reordenar slides para cumplir el intent —
  eso excede el alcance de este contrato.
