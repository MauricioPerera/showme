---
type: 'Task Contract'
title: 'CLI: comando project update-slide'
description: 'Logica pura del comando project update-slide: reemplaza los campos de una slide existente en un Project guardado y lo re-persiste.'
tags: ['showme', 'go', 'cli', 'project', 'slide']

task: cli-update-slide-command
intent: "Ejecutar el comando 'project update-slide': reemplazar los campos de una slide existente y guardar el proyecto actualizado."
target: internal/cli/update_slide_command.go
signature: "func RunUpdateSlideCommand(input UpdateSlideCommandInput) (UpdateSlideCommandResult, error)"
test_command: "go test ./internal/cli"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 8
  max_nesting_depth: 3
  cyclomatic_max: 8
  nesting_max: 3
  params_max: 1
  lines_max: 100
tests: internal/cli/update_slide_command_test.go
tests_sha256: "f29a0d47ff31875657bda42d9a9f9ea27b8c5502f202e03230bbbfda8132dffd"
touch_only: ['internal/cli/update_slide_command.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: cli-update-slide-command

## Intent

Expone [update-slide-usecase](./update-slide-usecase.md) por CLI, cerrando
el trio `add-slide`/`remove-slide`/`update-slide` que ya tienen
[cli-add-slide-command](./cli-add-slide-command.md) y
[cli-remove-slide-command](./cli-remove-slide-command.md). Carga un
`Project` guardado, reemplaza una de sus slides con `domain.UpdateSlide` y,
si es valido, re-guarda el proyecto con
[save-project-usecase](./save-project-usecase.md).

## Interface

```go
type UpdateSlideCommandInput struct {
    Path, SlideID, Title, Intent, Content, Status, OutDir string
}

type UpdateSlideCommandResult struct {
    OK       bool
    Path     string
    Errors, Warnings []string
}

func RunUpdateSlideCommand(input UpdateSlideCommandInput) (UpdateSlideCommandResult, error)
```

## Invariants

- Un error cargando `Path` (archivo inexistente o JSON invalido) se
  propaga tal cual via `err`; no se intenta actualizar ninguna slide.
- La actualizacion se valida con `domain.UpdateSlide` sobre el `Deck`
  cargado; si reporta errores (id/titulo vacio, slide no encontrada,
  status invalido), el resultado tiene `OK: false`, `Path: ""` y esos
  errores en `Errors` — el proyecto en disco NO se toca.
- Si `Status` (flag) queda vacio, `domain.UpdateSlide` preserva el status
  previo de la slide en vez de resetearlo — igual comportamiento que el
  contrato de dominio.
- Si la actualizacion es valida, el `Project` actualizado (mismo `Name`,
  `DesignPath`, `KnowledgePath`, `Version`; `Deck` con la slide
  reemplazada) se guarda con `storage.SaveProject` bajo `OutDir`. Si
  `OutDir` y `Name` coinciden con el archivo original, esto sobreescribe
  el mismo archivo.
- Un error de I/O al guardar se propaga via `err`.
- No hace red, subprocess ni llamadas a un proveedor de IA.

## Examples

- Proyecto guardado con slide `intro` ya `accepted` (via `project
  review`), `SlideID: "intro"`, `Title: "Introduccion revisada"` sin
  `Status` -> `OK: true`; al recargar, `intro` tiene el titulo nuevo pero
  sigue `accepted`.
- `SlideID: "missing"` -> `OK: false`, `Errors` incluye
  `slide not found: missing`.
- `Path` inexistente -> `err` no nil.

## Do / Don't

- DO: reusar `storage.LoadProject`, `domain.UpdateSlide` y
  `storage.SaveProject` tal cual; este comando es orquestacion pura.
- DO: preservar `Name`, `DesignPath`, `KnowledgePath` y `Version` del
  proyecto original al re-guardar; solo el `Deck` cambia.
- DON'T: permitir cambiar el `ID` de la slide via este comando.
- DON'T: imprimir a stdout ni parsear flags aqui — eso vive en
  `cmd/showme/main.go`.

## Tests

Los tests estan en `internal/cli/update_slide_command_test.go` y cubren:
actualizacion que preserva el status previo, slide no encontrada, y
archivo origen inexistente.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta permitir cambiar el `ID` de una slide
  para cumplir el intent — eso excede el alcance de este contrato.
