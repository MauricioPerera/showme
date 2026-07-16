---
type: 'Task Contract'
title: 'Caso de uso: aplicar una Review a un Deck'
description: 'Valida una Review y actualiza el Status de la slide correspondiente en una copia del Deck, sin mutar el original.'
tags: ['showme', 'go', 'usecase', 'domain', 'review', 'deck']

task: apply-review-usecase
intent: "Actualizar el estado de una slide de un Deck segun la decision de una Review valida."
target: internal/domain/apply_review.go
signature: "func ApplyReview(input ApplyReviewInput) (Deck, Report)"
test_command: "go test ./internal/domain"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 8
  max_nesting_depth: 3
  cyclomatic_max: 8
  nesting_max: 3
  params_max: 1
  lines_max: 90
tests: internal/domain/apply_review_test.go
tests_sha256: "a7c53eb95cde4a375ef990e9fc2ac51db560405a76f6c83e25046178b5cbb9e6"
touch_only: ['internal/domain/apply_review.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: apply-review-usecase

## Intent

Conecta [review-model](./review-model.md) con
[deck-slide-model](./deck-slide-model.md): dado un `Deck` y los datos de una
`Review`, valida la review y, si es valida, devuelve una copia del `Deck` con
el `Status` de la slide referenciada actualizado segun la decision. Es el
caso de uso que la webapp/CLI/MCP invocaran cuando una persona revisa una
slide generada (ver `DEFINITION.md`, "Usuarios y flujo principal": "aceptar,
editar o regenerar una diapositiva sin perder el contexto ni la version
anterior").

## Interface

```go
type ApplyReviewInput struct {
    Deck   Deck
    Review ReviewInput
}

func ApplyReview(input ApplyReviewInput) (Deck, Report)
```

## Invariants

- La review se valida primero con `domain.NewReview`; si es invalida, se
  devuelve una copia del `Deck` sin cambios y `Report` con los errores de
  la review.
- Si la review es valida pero `SlideID` no corresponde a ninguna slide del
  deck, se devuelve una copia del `Deck` sin cambios y el error
  `slide not found: <id>`.
- El mapeo de decision a estado es fijo: `accepted` -> `SlideStatusAccepted`,
  `rejected` -> `SlideStatusRejected`, `edited` -> `SlideStatusDraft` (una
  slide editada vuelve a `draft` porque necesita una revision nueva sobre el
  contenido cambiado).
- El `Deck` de entrada nunca se muta: la funcion siempre trabaja sobre una
  copia de `Slides`; el llamador puede seguir usando su `Deck` original sin
  que este cambie.
- Solo la slide referenciada cambia de `Status`; el resto de las slides y
  los campos `Title`/`Audience` del deck se preservan tal cual.
- La funcion es determinista, no hace I/O, red, subprocess ni llamadas a un
  proveedor de IA.

## Examples

- Deck con slides `intro` (accepted) y `plan` (draft), review
  `{SlideID: "plan", Decision: accepted}` -> `plan` pasa a `accepted`,
  `intro` sigue `accepted`, `Report.Errors` vacio.
- Review `{SlideID: "intro", Decision: rejected}` -> `intro` pasa a
  `rejected`.
- Review `{SlideID: "intro", Decision: edited}` -> `intro` pasa a `draft`
  (sin importar su estado previo).
- Review con `SlideID: "missing"` -> error `slide not found: missing`, deck
  devuelto sin cambios.
- Review con `SlideID: ""` -> error `slide id is required` (via
  `domain.NewReview`), deck devuelto sin cambios.
- Deck original pasado por valor a `ApplyReview` -> tras la llamada, sus
  slides conservan su `Status` previo (no se mutan).

## Do / Don't

- DO: delegar toda la validacion de la review a `domain.NewReview`; este
  contrato no reimplementa esas reglas.
- DO: copiar `Slides` antes de modificar cualquier elemento, para no mutar
  el `Deck` recibido.
- DON'T: reordenar ni eliminar slides — solo se actualiza el `Status` de la
  slide encontrada.
- DON'T: persistir el `Deck` actualizado aqui; eso es responsabilidad de
  `save-deck-usecase`/`save-project-usecase`.

## Tests

Los tests estan en `internal/domain/apply_review_test.go` y cubren: decision
`accepted`, decision `rejected`, decision `edited` (reset a `draft`), slide
no encontrada, review invalida (deck sin cambios) y no-mutacion del deck
original.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta persistir el deck actualizado o registrar
  la review en algun lado para cumplir el intent — eso excede el alcance de
  este contrato.
