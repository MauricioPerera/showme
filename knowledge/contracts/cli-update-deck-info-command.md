---
type: 'Task Contract'
title: 'CLI: comando project update-info'
description: 'Logica pura del comando project update-info: reemplaza titulo/audiencia de un Project guardado y lo re-persiste.'
tags: ['showme', 'go', 'cli', 'project', 'deck']

task: cli-update-deck-info-command
intent: "Ejecutar el comando 'project update-info': reemplazar titulo y audiencia de un proyecto guardado."
target: internal/cli/update_deck_info_command.go
signature: "func RunUpdateDeckInfoCommand(input UpdateDeckInfoCommandInput) (UpdateDeckInfoCommandResult, error)"
test_command: "go test ./internal/cli"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 6
  max_nesting_depth: 2
  cyclomatic_max: 6
  nesting_max: 2
  params_max: 1
  lines_max: 90
tests: internal/cli/update_deck_info_command_test.go
tests_sha256: "9711d56366a6d9a2908b02ab99374ad1bb8abc6e974df4925b6eb4eeb3409aae"
touch_only: ['internal/cli/update_deck_info_command.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: cli-update-deck-info-command

## Intent

Expone [update-deck-info-usecase](./update-deck-info-usecase.md) por CLI:
carga un `Project` guardado, reemplaza el `Title`/`Audience` de su `Deck` con
`domain.UpdateDeckInfo` y, si es valido, re-guarda el proyecto con
[save-project-usecase](./save-project-usecase.md). El `Name` del `Project`
(usado para el slug del archivo) no se toca — cambiarlo es
[duplicate-project-usecase](./duplicate-project-usecase.md) o un futuro
contrato de renombrado.

## Interface

```go
type UpdateDeckInfoCommandInput struct {
    Path, Title, Audience, OutDir string
}

type UpdateDeckInfoCommandResult struct {
    OK       bool
    Path     string
    Errors, Warnings []string
}

func RunUpdateDeckInfoCommand(input UpdateDeckInfoCommandInput) (UpdateDeckInfoCommandResult, error)
```

## Invariants

- Un error cargando `Path` (archivo inexistente o JSON invalido) se
  propaga tal cual via `err`; no se intenta actualizar nada.
- La actualizacion se valida con `domain.UpdateDeckInfo`; si `Title` esta
  vacio, el resultado tiene `OK: false`, `Path: ""` y `Errors` incluye
  `title is required` — el proyecto en disco NO se toca.
- Si es valida, el `Project` actualizado (mismo `Name`, `DesignPath`,
  `KnowledgePath`, `Version`, `Archived`, `Deck.Slides`;
  `Deck.Title`/`Deck.Audience` reemplazados) se guarda con
  `storage.SaveProject` bajo `OutDir`. Si `OutDir` y `Name` coinciden con
  el archivo original, esto sobreescribe el mismo archivo.
- `Archived` se preserva tal cual estaba (no se resetea a `false`), misma
  convencion que [cli-add-slide-command](./cli-add-slide-command.md).
- `Runs` (el historial de generaciones de IA) tambien se preserva tal
  cual, por la misma razon (ver
  [append-generation-run-usecase](./append-generation-run-usecase.md)).
- Un error de I/O al guardar se propaga via `err`.
- No hace red, subprocess ni llamadas a un proveedor de IA.

## Examples

- Proyecto guardado con `Name: "Roadmap Q3"` y `Deck.Title: "Roadmap Q3"`;
  `Title: "Roadmap Q4"`, `Audience: "Equipo ejecutivo"`, `OutDir` igual al
  directorio original -> `OK: true`; al recargarlo `Deck.Title`/`Audience`
  cambiaron pero `Project.Name` sigue `"Roadmap Q3"` y las slides no se
  tocaron.
- `Title: ""` -> `OK: false`, `Errors` incluye `title is required`.
- `Path` inexistente -> `err` no nil.

## Do / Don't

- DO: reusar `storage.LoadProject`, `domain.UpdateDeckInfo` y
  `storage.SaveProject` tal cual; este comando es orquestacion pura.
- DO: preservar `Name`, `DesignPath`, `KnowledgePath`, `Version` y
  `Deck.Slides` del proyecto original al re-guardar; solo
  `Title`/`Audience` cambian.
- DON'T: cambiar `Project.Name` ni el slug del archivo — eso queda para
  `duplicate-project-usecase` o un contrato de renombrado futuro.
- DON'T: imprimir a stdout ni parsear flags aqui — eso vive en
  `cmd/showme/main.go`.

## Tests

Los tests estan en `internal/cli/update_deck_info_command_test.go` y
cubren: actualizacion valida con slides y `Name` preservados, titulo vacio,
y archivo origen inexistente.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta cambiar `Project.Name` o el archivo
  destino para cumplir el intent — eso excede el alcance de este
  contrato.
