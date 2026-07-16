---
type: 'Task Contract'
title: 'Caso de uso: listar Decks guardados'
description: 'Lista las rutas de los archivos JSON de deck presentes directamente en un directorio, ordenadas.'
tags: ['showme', 'go', 'usecase', 'storage', 'deck']

task: list-decks-usecase
intent: "Listar las rutas de los decks guardados como JSON en un directorio."
target: internal/storage/deck_list.go
signature: "func ListDecks(dir string) ([]string, error)"
test_command: "go test ./internal/storage"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 5
  max_nesting_depth: 3
  cyclomatic_max: 5
  nesting_max: 3
  params_max: 1
  lines_max: 60
tests: internal/storage/deck_list_test.go
tests_sha256: "ba31d0da207969683517005359bccdeea51ae63834bc5986367185d8c46b691d"
touch_only: ['internal/storage/deck_list.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: list-decks-usecase

## Intent

Cierra el trio minimo de storage junto a
[save-deck-usecase](./save-deck-usecase.md) y
[load-deck-usecase](./load-deck-usecase.md): dado un directorio, devuelve las
rutas de los decks guardados para que webapp, CLI y MCP puedan listar
proyectos existentes antes de abrir uno con `LoadDeck`.

## Interface

```go
func ListDecks(dir string) (paths []string, err error)
```

## Invariants

- Devuelve solo archivos con extension `.json` ubicados directamente en
  `dir` (no recorre subdirectorios).
- Un directorio con nombre terminado en `.json` no se incluye en el
  resultado.
- Las rutas devueltas estan ordenadas (mismo orden que `os.ReadDir`, que ya
  ordena por nombre de archivo).
- Un directorio vacio devuelve una lista vacia, no un error.
- Un directorio inexistente devuelve un error.
- No hace red, subprocess ni llamadas a un proveedor de IA; no modifica el
  directorio ni los archivos listados.

## Examples

- Directorio con `kickoff.json`, `onboarding.json`, `roadmap-q3.json` y un
  `notes.txt` -> devuelve los tres `.json`, ordenados alfabeticamente, sin
  `notes.txt`.
- Directorio vacio -> `[]string{}` (o `nil`), `err` nil.
- Directorio inexistente -> `err` no nil.

## Do / Don't

- DO: mantener la funcion de solo lectura, sin efectos secundarios sobre el
  filesystem.
- DO: devolver rutas completas (`filepath.Join(dir, nombre)`), listas para
  pasarle directamente a `LoadDeck`.
- DON'T: leer o parsear el contenido de cada archivo — eso es
  responsabilidad de `LoadDeck`.
- DON'T: recorrer subdirectorios; `SaveDeck` siempre escribe en un directorio
  plano.

## Tests

Los tests estan en `internal/storage/deck_list_test.go` y cubren: listado
ordenado ignorando archivos no-JSON y un directorio con nombre `.json`,
directorio vacio, y directorio inexistente.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta leer metadata de cada deck (ej. titulo o
  fecha) para cumplir el intent — eso implica combinar este contrato con
  `LoadDeck` en un caso de uso nuevo, no extender este.
