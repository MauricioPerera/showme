---
type: 'Task Contract'
title: 'Caso de uso: cargar un Deck'
description: 'Lee un archivo JSON de deck escrito por save-deck-usecase y lo deserializa de vuelta a domain.Deck.'
tags: ['showme', 'go', 'usecase', 'storage', 'deck']

task: load-deck-usecase
intent: "Leer un Deck persistido como JSON desde el filesystem."
target: internal/storage/deck_load.go
signature: "func LoadDeck(path string) (domain.Deck, error)"
test_command: "go test ./internal/storage"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 5
  max_nesting_depth: 2
  cyclomatic_max: 5
  nesting_max: 2
  params_max: 1
  lines_max: 60
tests: internal/storage/deck_load_test.go
tests_sha256: "273b19ed0b069fba68cc0e7a137d20314aef0603fdca75a8d2ab968ff74badfe"
touch_only: ['internal/storage/deck_load.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: load-deck-usecase

## Intent

Cierra el round-trip de persistencia iniciado en
[save-deck-usecase](./save-deck-usecase.md): dado el path de un archivo
escrito por `SaveDeck`, lo lee y lo deserializa de vuelta a `domain.Deck`
para que webapp, CLI y MCP puedan reabrir un proyecto ya guardado.

## Interface

```go
func LoadDeck(path string) (domain.Deck, error)
```

## Invariants

- Lee el archivo en `path` y deserializa su contenido JSON a `domain.Deck`.
- No escribe ni modifica el archivo leido.
- Preserva `Title`, `Audience` y el orden y contenido de `Slides` tal como
  fueron escritos por `SaveDeck`.
- Un archivo inexistente devuelve un error (el del sistema de archivos), no
  un `domain.Deck` vacio silencioso.
- Un contenido que no es JSON valido devuelve un error de parseo.
- No hace red, subprocess ni llamadas a un proveedor de IA.
- No revalida las invariantes de `domain.NewDeck`: asume que el archivo fue
  escrito por `SaveDeck` (que ya solo persiste decks validos).

## Examples

- `SaveDeck` seguido de `LoadDeck(path)` -> mismo `Title`, `Audience` y
  `Slides` (mismo orden, mismo `Status`) que el `DeckInput` original.
- Path a un archivo que no existe -> `err` no nil.
- Archivo con contenido `"{not json"` -> `err` no nil (fallo de parseo).

## Do / Don't

- DO: mantener la funcion simetrica con `SaveDeck` (mismo formato JSON,
  mismo tipo `domain.Deck`).
- DO: devolver el error del sistema de archivos o de parseo tal cual, sin
  envolverlo en una etiqueta propia todavia (no hay convencion de error
  wrapping definida aun para el proyecto).
- DON'T: revalidar ni "reparar" el contenido leido — si el archivo esta
  corrupto, el error debe propagarse.
- DON'T: buscar el archivo por titulo o slug; el llamador ya tiene el path
  (devuelto por `SaveDeck` o listado por un contrato futuro de listado).

## Tests

Los tests estan en `internal/storage/deck_load_test.go` y cubren:
round-trip completo `SaveDeck` -> `LoadDeck` con dos slides y estados
distintos, archivo inexistente y contenido JSON invalido.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta listar decks existentes o buscar por
  titulo/slug para cumplir el intent — eso excede el alcance de este
  contrato.
