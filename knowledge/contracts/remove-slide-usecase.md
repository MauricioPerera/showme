---
type: 'Task Contract'
title: 'Caso de uso: quitar una slide de un Deck'
description: 'Valida que una slide exista y que quitarla no deje el Deck vacio, y devuelve una copia del Deck sin esa slide.'
tags: ['showme', 'go', 'usecase', 'domain', 'deck', 'slide']

task: remove-slide-usecase
intent: "Quitar una slide existente de un Deck, sin dejarlo sin slides."
target: internal/domain/remove_slide.go
signature: "func RemoveSlide(input RemoveSlideInput) (Deck, Report)"
test_command: "go test ./internal/domain"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 6
  max_nesting_depth: 3
  cyclomatic_max: 6
  nesting_max: 3
  params_max: 1
  lines_max: 70
tests: internal/domain/remove_slide_test.go
tests_sha256: "d58b6987f3f420ebbf1d14e0efe7614405e10f0bc52cd21d74ccf3c55d217a50"
touch_only: ['internal/domain/remove_slide.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: remove-slide-usecase

## Intent

Complemento simetrico de [add-slide-usecase](./add-slide-usecase.md): dado
un `Deck` existente y el `ID` de una de sus slides, la quita. Reafirma la
invariante de [deck-slide-model](./deck-slide-model.md) de que un `Deck`
siempre tiene al menos una slide, rechazando la remocion que dejaria el
deck vacio en vez de permitir un estado que despues fallaria silenciosamente
al intentar guardarlo.

## Interface

```go
type RemoveSlideInput struct {
    Deck    Deck
    SlideID string
}

func RemoveSlide(input RemoveSlideInput) (Deck, Report)
```

## Invariants

- Si `SlideID` no corresponde a ninguna slide de `Deck.Slides`, es un error
  `slide not found: <id>` y el `Deck` devuelto queda sin cambios.
- Si `Deck.Slides` tiene exactamente una slide y es la que se pide quitar,
  es un error `deck must have at least one slide` (mismo mensaje que
  `deck-slide-model`) y el `Deck` devuelto queda sin cambios.
- En cualquier otro caso valido, se devuelve una copia de `Deck` sin esa
  slide, preservando el orden y contenido del resto.
- El `Deck` de entrada nunca se muta.
- La funcion es determinista, no hace I/O, red, subprocess ni llamadas a un
  proveedor de IA.

## Examples

- Deck con slides `intro` y `plan`, `SlideID: "plan"` -> deck resultante
  con solo `intro`, `Report.Errors` vacio.
- `SlideID: "missing"` -> error `slide not found: missing`, deck sin
  cambios.
- Deck con una unica slide `intro`, `SlideID: "intro"` -> error
  `deck must have at least one slide`, deck sin cambios (sigue con su
  unica slide).

## Do / Don't

- DO: preservar el orden de las slides restantes tal cual estaban.
- DO: copiar `Slides` antes de quitar cualquier elemento, para no mutar el
  `Deck` recibido.
- DON'T: permitir vaciar el deck — esa invariante se reafirma aqui, no se
  delega a un chequeo posterior en `NewProject`/`SaveProject`.
- DON'T: persistir el deck actualizado aqui; eso es responsabilidad de
  `save-deck-usecase`/`save-project-usecase`.

## Tests

Los tests estan en `internal/domain/remove_slide_test.go` y cubren:
remocion valida, slide no encontrada, remocion de la unica slide rechazada,
y no-mutacion del deck original.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta reordenar slides para cumplir el intent —
  eso excede el alcance de este contrato.
