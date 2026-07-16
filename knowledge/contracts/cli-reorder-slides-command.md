---
type: 'Task Contract'
title: 'CLI: comando project reorder-slides'
description: 'Logica pura del comando project reorder-slides: reordena las slides de un Project guardado y lo re-persiste.'
tags: ['showme', 'go', 'cli', 'project', 'slide']

task: cli-reorder-slides-command
intent: "Ejecutar el comando 'project reorder-slides': reordenar las slides de un proyecto guardado y guardarlo actualizado."
target: internal/cli/reorder_slides_command.go
signature: "func RunReorderSlidesCommand(input ReorderSlidesCommandInput) (ReorderSlidesCommandResult, error)"
test_command: "go test ./internal/cli"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 8
  max_nesting_depth: 3
  cyclomatic_max: 8
  nesting_max: 3
  params_max: 1
  lines_max: 100
tests: internal/cli/reorder_slides_command_test.go
tests_sha256: "2b9c19decd8cbb0f6684d36e93871d6a5dbfac221d3bd5f42673de9ac4f188d6"
touch_only: ['internal/cli/reorder_slides_command.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: cli-reorder-slides-command

## Intent

Expone [reorder-slides-usecase](./reorder-slides-usecase.md) por CLI,
cerrando el conjunto de comandos de edicion de slides junto a `add-slide`,
`remove-slide` y `update-slide`. Carga un `Project` guardado, reordena las
slides de su `Deck` con `domain.ReorderSlides` y, si el orden dado es una
permutacion valida, re-guarda el proyecto con
[save-project-usecase](./save-project-usecase.md).

## Interface

```go
type ReorderSlidesCommandInput struct {
    Path   string
    Order  []string
    OutDir string
}

type ReorderSlidesCommandResult struct {
    OK       bool
    Path     string
    Errors, Warnings []string
}

func RunReorderSlidesCommand(input ReorderSlidesCommandInput) (ReorderSlidesCommandResult, error)
```

## Invariants

- Un error cargando `Path` (archivo inexistente o JSON invalido) se
  propaga tal cual via `err`; no se intenta reordenar nada.
- El reordenamiento se valida con `domain.ReorderSlides` sobre el `Deck`
  cargado; si reporta errores (`Order` incompleto, con ids desconocidos o
  duplicados), el resultado tiene `OK: false`, `Path: ""` y esos errores
  en `Errors` — el proyecto en disco NO se toca.
- Si `Order` es una permutacion valida, el `Project` actualizado (mismo
  `Name`, `DesignPath`, `KnowledgePath`, `Version`; `Deck` con las slides
  en el nuevo orden) se guarda con `storage.SaveProject` bajo `OutDir`. Si
  `OutDir` y `Name` coinciden con el archivo original, esto sobreescribe
  el mismo archivo.
- Un error de I/O al guardar se propaga via `err`.
- No hace red, subprocess ni llamadas a un proveedor de IA.

## Examples

- Proyecto guardado con slides `intro, plan` (en ese orden),
  `Order: ["plan", "intro"]`, `OutDir` igual al directorio original ->
  `OK: true`; al recargarlo el deck queda `plan, intro`.
- `Order: ["intro"]` (falta `plan`) -> `OK: false`, `Errors` incluye
  `missing slide id in order: plan`.
- `Path` inexistente -> `err` no nil.

## Do / Don't

- DO: reusar `storage.LoadProject`, `domain.ReorderSlides` y
  `storage.SaveProject` tal cual; este comando es orquestacion pura.
- DO: preservar `Name`, `DesignPath`, `KnowledgePath` y `Version` del
  proyecto original al re-guardar; solo el orden de `Deck.Slides` cambia.
- DON'T: aceptar un `Order` parcial — debe cubrir todas las slides, igual
  que exige `domain.ReorderSlides`.
- DON'T: imprimir a stdout ni parsear flags aqui — eso vive en
  `cmd/showme/main.go` (el flag de linea de comando, una lista separada
  por comas, se parsea ahi antes de llamar a esta funcion).

## Tests

Los tests estan en `internal/cli/reorder_slides_command_test.go` y cubren:
reordenamiento valido que persiste el cambio en el mismo archivo, orden
incompleto, y archivo origen inexistente.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta soportar un `Order` parcial para cumplir
  el intent — eso excede el alcance de este contrato.
