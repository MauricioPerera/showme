---
type: 'Task Contract'
title: 'Caso de uso: agregar una slide a un Deck'
description: 'Valida una slide nueva contra un Deck existente y devuelve una copia del Deck con la slide agregada al final, sin mutar el original.'
tags: ['showme', 'go', 'usecase', 'domain', 'deck', 'slide']

task: add-slide-usecase
intent: "Agregar una slide nueva y valida al final de un Deck existente."
target: internal/domain/add_slide.go
signature: "func AddSlide(input AddSlideInput) (Deck, Report)"
test_command: "go test ./internal/domain"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 8
  max_nesting_depth: 3
  cyclomatic_max: 8
  nesting_max: 3
  params_max: 1
  lines_max: 90
tests: internal/domain/add_slide_test.go
tests_sha256: "33d25b271cbe5bf9ed1f22821dc02ba6c7f93d36d5e1915320ea1aad2fcd79fb"
touch_only: ['internal/domain/add_slide.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: add-slide-usecase

## Intent

Hasta ahora la unica forma de poblar las slides de un `Deck` era pasarlas
todas juntas en `DeckInput` al crear el proyecto
([create-project-usecase](./create-project-usecase.md)). Este contrato
cubre agregar una slide nueva a un `Deck` ya existente, reusando las mismas
invariantes de [deck-slide-model](./deck-slide-model.md) (id/titulo
requeridos, id unico, status valido) pero validando contra las slides ya
presentes en vez de reconstruir el deck entero.

## Interface

```go
type AddSlideInput struct {
    Deck  Deck
    Slide Slide
}

func AddSlide(input AddSlideInput) (Deck, Report)
```

## Invariants

- La slide nueva debe tener `ID` y `Title` no vacios.
- El `ID` de la slide nueva no puede coincidir con el de ninguna slide ya
  presente en `Deck.Slides`.
- `Status` vacio se normaliza a `SlideStatusDraft`; cualquier otro valor
  fuera de `{draft, accepted, rejected}` es un error.
- Si hay algun error, se devuelve una copia de `Deck` sin cambios (la slide
  nueva NO se agrega) junto con `Report`.
- Si es valida, se devuelve una copia de `Deck` con la slide nueva agregada
  al final de `Slides`, preservando el orden y contenido de las slides
  existentes.
- El `Deck` de entrada nunca se muta.
- La funcion es determinista, no hace I/O, red, subprocess ni llamadas a un
  proveedor de IA.

## Examples

- Deck con slides `intro` y `plan`, slide nueva `{ID: "closing", Title:
  "Cierre"}` -> deck resultante con 3 slides en orden `intro, plan,
  closing`, la nueva con `Status: draft`.
- Slide nueva con `ID: ""` -> error `slide id is required`, deck sin
  cambios.
- Slide nueva con `Title: ""` -> error `slide title is required`, deck sin
  cambios.
- Slide nueva con `ID: "intro"` (ya existente) -> error
  `duplicate slide id: intro`, deck sin cambios.
- Slide nueva con `Status: "archived"` -> error `invalid status: archived`.

## Do / Don't

- DO: reusar el mismo enum `SlideStatus` y la misma logica de validacion de
  estado que `deck-slide-model` (via la funcion compartida
  `isValidSlideStatus`).
- DO: copiar `Slides` antes de agregar cualquier elemento, para no mutar el
  `Deck` recibido.
- DON'T: permitir insertar en una posicion especifica — la slide nueva
  siempre va al final; reordenar es un caso de uso separado.
- DON'T: persistir el deck actualizado aqui; eso es responsabilidad de
  `save-deck-usecase`/`save-project-usecase`.

## Tests

Los tests estan en `internal/domain/add_slide_test.go` y cubren: slide
valida agregada al final, id vacio, titulo vacio, id duplicado, status
invalido y no-mutacion del deck original.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta eliminar o reordenar slides para cumplir
  el intent — eso excede el alcance de este contrato.
