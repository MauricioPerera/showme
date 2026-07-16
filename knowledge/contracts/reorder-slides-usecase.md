---
type: 'Task Contract'
title: 'Caso de uso: reordenar las slides de un Deck'
description: 'Valida que un nuevo orden de IDs sea una permutacion exacta de las slides de un Deck y devuelve una copia reordenada.'
tags: ['showme', 'go', 'usecase', 'domain', 'deck', 'slide']

task: reorder-slides-usecase
intent: "Reordenar las slides de un Deck segun una lista completa de sus IDs en el orden deseado."
target: internal/domain/reorder_slides.go
signature: "func ReorderSlides(input ReorderSlidesInput) (Deck, Report)"
test_command: "go test ./internal/domain"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 8
  max_nesting_depth: 3
  cyclomatic_max: 8
  nesting_max: 3
  params_max: 1
  lines_max: 90
tests: internal/domain/reorder_slides_test.go
tests_sha256: "5802992959187f849966b858c660eb6ae92b0ee2ae0ee986a799dd58121ad637"
touch_only: ['internal/domain/reorder_slides.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: reorder-slides-usecase

## Intent

Cubre reordenar, la operacion de edicion de slides diferida explicitamente
por [add-slide-usecase](./add-slide-usecase.md) y
[remove-slide-usecase](./remove-slide-usecase.md) ("DON'T: permitir insertar
en una posicion especifica" / "reordenar es un caso de uso separado"). El
llamador provee el orden final completo como una lista de `ID`; este
contrato valida que sea exactamente una permutacion de las slides existentes
antes de aplicarlo.

## Interface

```go
type ReorderSlidesInput struct {
    Deck  Deck
    Order []string
}

func ReorderSlides(input ReorderSlidesInput) (Deck, Report)
```

## Invariants

- `Order` debe contener cada `ID` de `Deck.Slides` exactamente una vez: ni
  de mas ni de menos.
- Un `ID` en `Order` que no exista en `Deck.Slides` es un error
  `unknown slide id: <id>`.
- Un `ID` repetido dentro de `Order` es un error
  `duplicate slide id in order: <id>`.
- Un `ID` de `Deck.Slides` ausente en `Order` es un error
  `missing slide id in order: <id>`.
- Si hay algun error, se devuelve una copia de `Deck` sin cambios (mismo
  orden que tenia).
- Si `Order` es una permutacion valida, se devuelve una copia de `Deck` con
  `Slides` en ese orden exacto; el contenido de cada slide no cambia, solo
  su posicion.
- El `Deck` de entrada nunca se muta.
- La funcion es determinista, no hace I/O, red, subprocess ni llamadas a un
  proveedor de IA.

## Examples

- Deck con `intro, plan` (en ese orden), `Order: ["plan", "intro"]` ->
  deck resultante con `plan, intro`; el contenido de cada slide (titulo,
  status) no cambia.
- `Order: ["intro"]` (falta `plan`) -> error
  `missing slide id in order: plan`, deck sin cambios.
- `Order: ["intro", "plan", "closing"]` (`closing` no existe) -> error
  `unknown slide id: closing`.
- `Order: ["intro", "intro"]` -> error
  `duplicate slide id in order: intro`.

## Do / Don't

- DO: exigir el orden completo (todas las slides), no una operacion de
  "mover una posicion" parcial — mas simple de validar y mas explicito
  para quien llama.
- DO: copiar `Slides` (o construir la lista reordenada) sin mutar el
  `Deck` recibido.
- DON'T: permitir un `Order` parcial que solo mueva algunas slides dejando
  las demas en un orden implicito — ambiguo y dificil de razonar.
- DON'T: persistir el deck reordenado aqui; eso es responsabilidad de
  `save-deck-usecase`/`save-project-usecase`.

## Tests

Los tests estan en `internal/domain/reorder_slides_test.go` y cubren:
reordenamiento valido con contenido preservado, orden incompleto, id
desconocido, id duplicado en el orden, y no-mutacion del deck original.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta soportar un `Order` parcial para cumplir
  el intent — eso excede el alcance de este contrato.
