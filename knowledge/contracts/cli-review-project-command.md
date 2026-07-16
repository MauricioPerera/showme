---
type: 'Task Contract'
title: 'CLI: comando project review'
description: 'Logica pura del comando project review: aplica una decision a una slide de un Project guardado y lo re-persiste.'
tags: ['showme', 'go', 'cli', 'project', 'review']

task: cli-review-project-command
intent: "Ejecutar el comando 'project review': aplicar una decision a una slide de un proyecto guardado y guardarlo actualizado."
target: internal/cli/review_project_command.go
signature: "func RunReviewProjectCommand(input ReviewProjectCommandInput) (ReviewProjectCommandResult, error)"
test_command: "go test ./internal/cli"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 8
  max_nesting_depth: 3
  cyclomatic_max: 8
  nesting_max: 3
  params_max: 1
  lines_max: 100
tests: internal/cli/review_project_command_test.go
tests_sha256: "694040d940dcc901f4bd76c7d7c37d3be40eb5801bc180db08cb76ef39ee77c8"
touch_only: ['internal/cli/review_project_command.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: cli-review-project-command

## Intent

Primer comando de la CLI que cierra el ciclo revision-persistencia: carga un
`Project` guardado, aplica una decision a una de sus slides con
[apply-review-usecase](./apply-review-usecase.md) y, si es valida, guarda el
`Project` actualizado con [save-project-usecase](./save-project-usecase.md).
Cubre "aceptar, editar o regenerar una diapositiva" de `DEFINITION.md`
("Usuarios y flujo principal") para el canal CLI/agente.

## Interface

```go
type ReviewProjectCommandInput struct {
    Path, SlideID, Decision, Notes, OutDir string
}

type ReviewProjectCommandResult struct {
    OK       bool
    Path     string
    Errors, Warnings []string
}

func RunReviewProjectCommand(input ReviewProjectCommandInput) (ReviewProjectCommandResult, error)
```

## Invariants

- Un error cargando `Path` (archivo inexistente o JSON invalido) se
  propaga tal cual via `err`; no se intenta aplicar ninguna review.
- La review se valida y aplica con `domain.ApplyReview` sobre el `Deck`
  cargado; si `ApplyReview` reporta errores (decision invalida, slide no
  encontrada, `SlideID` vacio), el resultado tiene `OK: false`,
  `Path: ""` y esos errores en `Errors` — el proyecto en disco NO se
  toca.
- Si la review es valida, el `Project` actualizado (mismo `Name`,
  `DesignPath`, `KnowledgePath`, `Version`, `Archived`; `Deck` con la
  slide actualizada) se guarda con `storage.SaveProject` bajo `OutDir`. Si
  `OutDir` y `Name` coinciden con el archivo original, esto sobreescribe
  el mismo archivo (mismo slug determinista).
- `Archived` se preserva tal cual estaba (no se resetea a `false`), misma
  convencion que [cli-add-slide-command](./cli-add-slide-command.md).
- Un error de I/O al guardar (ej. `OutDir` inexistente) se propaga via
  `err`.
- No hace red, subprocess ni llamadas a un proveedor de IA.

## Examples

- Proyecto guardado con slide `intro` en `draft`, `SlideID: "intro"`,
  `Decision: "accepted"`, `OutDir` igual al directorio original -> `OK:
  true`, `Path` igual al archivo original, y al recargarlo la slide
  `intro` queda `accepted`.
- `Decision: "archived"` (fuera del enum) -> `OK: false`, `Errors` incluye
  `invalid decision: archived`, el archivo original queda sin cambios.
- `SlideID: "missing"` -> `OK: false`, `Errors` incluye
  `slide not found: missing`.
- `Path` inexistente -> `err` no nil.

## Do / Don't

- DO: reusar `storage.LoadProject`, `domain.ApplyReview` y
  `storage.SaveProject` tal cual; este comando es orquestacion pura.
- DO: preservar `Name`, `DesignPath`, `KnowledgePath` y `Version` del
  proyecto original al re-guardar; solo el `Deck` cambia.
- DON'T: incrementar `Version` automaticamente al revisar — no hay una
  decision de producto tomada sobre versionado de reviews todavia.
- DON'T: imprimir a stdout ni parsear flags aqui — eso vive en
  `cmd/showme/main.go`.

## Tests

Los tests estan en `internal/cli/review_project_command_test.go` y cubren:
review valida que persiste el cambio en el mismo archivo, decision
invalida, slide no encontrada, y archivo origen inexistente.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta decidir una politica de versionado para
  las reviews (ej. incrementar `Version`) para cumplir el intent — eso
  excede el alcance de este contrato y amerita una decision de producto
  aparte.
