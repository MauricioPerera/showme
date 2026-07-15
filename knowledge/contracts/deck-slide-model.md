---
type: 'Task Contract'
title: 'Modelo de dominio Deck/Slide'
description: 'Construye un Deck a partir de su titulo, audiencia y slides, validando las invariantes estructurales minimas.'
tags: ['showme', 'go', 'domain', 'deck', 'slide']

task: deck-slide-model
intent: "Construir un Deck valido a partir de un titulo, audiencia y sus slides."
target: internal/domain/deck.go
signature: "func NewDeck(input DeckInput) (Deck, Report)"
test_command: "go test ./internal/domain"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 8
  max_nesting_depth: 3
  cyclomatic_max: 8
  nesting_max: 3
  params_max: 1
  lines_max: 140
tests: internal/domain/deck_test.go
tests_sha256: "a3c4126d0dd574c8ae988278eac70ecb24f7e7d96e4c3e2f76082fdf8fa3f99e"
touch_only: ['internal/domain/deck.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: deck-slide-model

## Intent

Define el modelo de dominio minimo de `showme`: `Deck` como coleccion ordenada
de `Slide` con un titulo y una audiencia. Es la base de la que dependen
webapp, CLI y servidor MCP (ver `DEFINITION.md`, "Modelo conceptual minimo").
Este contrato cubre solo `Deck`/`Slide`; `Project`, `GenerationRun` y `Review`
quedan fuera de alcance y requieren contratos propios.

## Interface

```go
type SlideStatus string

const (
    SlideStatusDraft    SlideStatus = "draft"
    SlideStatusAccepted SlideStatus = "accepted"
    SlideStatusRejected SlideStatus = "rejected"
)

type Slide struct {
    ID, Title, Intent, Content string
    Status SlideStatus
}

type DeckInput struct {
    Title, Audience string
    Slides []Slide
}

type Deck struct {
    Title, Audience string
    Slides []Slide
}

type Report struct { Errors, Warnings []string }

func NewDeck(input DeckInput) (Deck, Report)
```

## Invariants

- `Title` no puede estar vacio.
- El deck debe tener al menos una slide.
- Cada slide debe tener `ID` y `Title` no vacios.
- Los `ID` de slide son unicos dentro del deck.
- El orden de `Slides` se preserva tal cual el input; la funcion no reordena.
- `Status` vacio se normaliza a `SlideStatusDraft`; cualquier otro valor fuera
  de `{draft, accepted, rejected}` es un error.
- La funcion es determinista, no hace I/O, red, subprocess ni llamadas a un
  proveedor de IA, y nunca lanza panic: los problemas se acumulan en `Report`.

## Examples

- Titulo, audiencia y 2 slides con IDs unicos -> `Report.Errors` vacio y
  `Deck.Slides` en el mismo orden del input.
- `Title` vacio -> error `title is required`.
- `Slides` vacio -> error `at least one slide is required`.
- Dos slides con el mismo `ID` -> error `duplicate slide id: <id>`.
- Slide sin `Status` -> el `Deck` devuelto trae `Status: SlideStatusDraft`.

## Do / Don't

- DO: preservar el orden de `Slides` tal cual el input.
- DO: acumular todos los errores encontrados en `Report`, no cortar en el
  primero.
- DON'T: incluir generacion de contenido, seleccion de contexto OKF, render o
  storage aqui — eso son casos de uso y contratos separados.
- DON'T: importar `internal/design` o `internal/knowledge` desde este
  archivo; el modelo de dominio no depende de los validadores de identidad ni
  de conocimiento.

## Tests

Los tests estan en `internal/domain/deck_test.go` y cubren: deck valido con
orden preservado, titulo vacio, deck sin slides, slide sin id, slide sin
titulo, id de slide duplicado, status invalido y normalizacion de status
vacio a `draft`.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si el modelo requiere referenciar `Project` o
  `GenerationRun` para cumplir el intent — eso implica que el alcance de este
  contrato quedo corto y hace falta uno nuevo.
