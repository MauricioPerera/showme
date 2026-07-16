---
type: 'Task Contract'
title: 'CLI: comando project generate-storyboard'
description: 'Propone la estructura de un deck (lista de slides) con IA y la escribe como un archivo deck.json compatible con project create.'
tags: ['showme', 'go', 'cli', 'ai', 'storyboard', 'deck']

task: cli-generate-storyboard-command
intent: "Ejecutar el comando 'project generate-storyboard': proponer la estructura de un deck con IA y escribirla como deck.json."
target: internal/cli/generate_storyboard_command.go
signature: "func RunGenerateStoryboardCommand(input GenerateStoryboardCommandInput) (GenerateStoryboardCommandResult, error)"
test_command: "go test ./internal/cli"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 10
  max_nesting_depth: 3
  cyclomatic_max: 10
  nesting_max: 3
  params_max: 1
  lines_max: 140
tests: internal/cli/generate_storyboard_command_test.go
tests_sha256: "acb621674116e4ab924cb7798cf7ef2ce1680f0d9f398aed27ed60a6ec3ff29d"
touch_only: ['internal/cli/generate_storyboard_command.go']
deps_allowed: []
forbids: ['subprocess']
---

# Contract: cli-generate-storyboard-command

## Intent

Cubre la pieza que
[cli-generate-slide-content-command](./cli-generate-slide-content-command.md)
dejo explicitamente diferida: proponer la estructura completa de un deck en
vez de una slide a la vez. Ata
[context-selector](./context-selector.md) (opcional, si se da
`KnowledgeRoot`),
[generate-storyboard-usecase](./generate-storyboard-usecase.md) (con
[openai-content-generator-client](./openai-content-generator-client.md)
como `StoryboardGenerator` real) y escribe el resultado como un archivo
deck JSON en el mismo shape que espera `--deck` en
[cli-create-project-command](./cli-create-project-command.md) — este
comando NO crea un proyecto, produce el insumo para `project create`.

## Interface

```go
type GenerateStoryboardCommandInput struct {
    Objective, Audience, KnowledgeRoot, BaseURL, Model, DeckTitle, OutPath string
    Count int
}

type GenerateStoryboardCommandResult struct {
    OK       bool
    Path     string
    Errors, Warnings []string
}

func RunGenerateStoryboardCommand(input GenerateStoryboardCommandInput) (GenerateStoryboardCommandResult, error)
```

## Invariants

- Si `KnowledgeRoot` es no vacio, se carga con `knowledge.Load` y se
  seleccionan hasta 3 conceptos relevantes al `Objective` con
  `knowledge.Select`; sus `Body` se concatenan como contexto. Errores de
  carga/seleccion se agregan a `Errors` y cortan el proceso antes de
  llamar al proveedor de IA.
- Si `KnowledgeRoot` es vacio, no se selecciona contexto (`Context: ""`).
- Cualquier error de `ai.GenerateStoryboard` (objetivo vacio, count no
  positivo, error del proveedor, JSON invalido) se agrega a `Errors`; si
  hay al menos un error, NO se escribe ningun archivo.
- Cada slide propuesta recibe un `ID` deterministico via `storage.Slugify`
  sobre su `Title`; colisiones entre titulos que producen el mismo slug se
  resuelven agregando un sufijo numerico (`-2`, `-3`, ...) para garantizar
  ids unicos, requisito de `deck-slide-model`.
- El archivo escrito en `OutPath` es JSON con el shape
  `{"title", "audience", "slides": [{"id", "title", "intent"}]}`, el mismo
  que `cli-create-project-command` parsea desde su flag `--deck`.
- Un error de I/O escribiendo `OutPath` se propaga via `err`.
- No hace subprocess.

## Examples

- `Objective: "Presentar el roadmap"`, `Count: 2`, proveedor de IA que
  responde `[{"title":"Introduccion","intent":"..."},{"title":"Plan","intent":"..."}]`
  -> `OK: true`, archivo en `OutPath` con 2 slides, ids `introduccion` y
  `plan`.
- Dos slides propuestas con el mismo `Title` (`"Introduccion"`) -> ids
  `introduccion` y `introduccion-2`.
- El proveedor de IA responde status 500 -> `OK: false`, `Errors` incluye
  el error propagado; no se escribe archivo.
- Directorio de `OutPath` inexistente -> `err` no nil.

## Do / Don't

- DO: reusar `knowledge.Load`/`knowledge.Select`, `ai.GenerateStoryboard`,
  `storage.Slugify` y los DTOs (`slideDTO`/`deckInputDTO`) ya definidos en
  `cli-create-project-command` — no reimplementar el shape del deck JSON.
- DO: garantizar ids unicos entre las slides propuestas antes de escribir
  el archivo.
- DON'T: crear ni guardar un `Project` aqui — este comando solo produce el
  archivo `--deck` que consume `project create`.
- DON'T: imprimir a stdout ni parsear flags aqui — eso vive en
  `cmd/showme/main.go`.

## Tests

Los tests estan en `internal/cli/generate_storyboard_command_test.go` y
cubren: generacion valida con ids unicos y deck JSON parseable, colision
de titulos deduplicada, error del proveedor de IA que no escribe archivo,
y directorio de salida inexistente.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta crear el `Project` directamente (en vez
  de solo el deck JSON) para cumplir el intent — eso excede el alcance de
  este contrato.
