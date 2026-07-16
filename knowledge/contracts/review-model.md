---
type: 'Task Contract'
title: 'Modelo de dominio Review'
description: 'Construye una Review con la decision humana sobre una slide, validando que la decision sea una de las conocidas.'
tags: ['showme', 'go', 'domain', 'review']

task: review-model
intent: "Construir una Review valida a partir de una slide y la decision humana sobre ella."
target: internal/domain/review.go
signature: "func NewReview(input ReviewInput) (Review, Report)"
test_command: "go test ./internal/domain"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 6
  max_nesting_depth: 2
  cyclomatic_max: 6
  nesting_max: 2
  params_max: 1
  lines_max: 70
tests: internal/domain/review_test.go
tests_sha256: "6ccf8c70e2863ee820f1c7781871bc9915e35f8558c5cbf3eec06df2b0145a4c"
touch_only: ['internal/domain/review.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: review-model

## Intent

Define `Review` del "Modelo conceptual minimo" de `DEFINITION.md`: la
decision humana sobre una slide o una propuesta generada. Es deliberadamente
independiente de `GenerationRun` (que depende de un puerto de IA todavia sin
decidir, ver `DEFINITION.md` "Decisiones que quedan abiertas") — una
`Review` puede existir y validarse sin que exista todavia ese puerto.

## Interface

```go
type ReviewDecision string

const (
    ReviewDecisionAccepted ReviewDecision = "accepted"
    ReviewDecisionEdited   ReviewDecision = "edited"
    ReviewDecisionRejected ReviewDecision = "rejected"
)

type ReviewInput struct {
    SlideID  string
    Decision ReviewDecision
    Notes    string
}

type Review struct {
    SlideID  string
    Decision ReviewDecision
    Notes    string
}

func NewReview(input ReviewInput) (Review, Report)
```

## Invariants

- `SlideID` no puede estar vacio.
- `Decision` no puede estar vacia; si esta presente debe ser una de
  `{accepted, edited, rejected}`.
- `Notes` es libre y opcional: ninguna decision la exige (no se infiere una
  regla de negocio no documentada en `DEFINITION.md`).
- La funcion es determinista, no hace I/O, red, subprocess ni llamadas a un
  proveedor de IA, y nunca lanza panic: los problemas se acumulan en
  `Report`, igual que `deck-slide-model` y `project-model`.

## Examples

- `SlideID: "intro"`, `Decision: accepted` -> `Report.Errors` vacio.
- `SlideID` vacio -> error `slide id is required`.
- `Decision` vacia -> error `decision is required`.
- `Decision: "archived"` (fuera del enum) -> error
  `invalid decision: archived`.
- `Decision: rejected` con `Notes` -> `Report.Errors` vacio, `Notes`
  preservado.

## Do / Don't

- DO: mantener el mismo patron de `Report` acumulativo que el resto de los
  constructores del dominio.
- DO: tratar `Notes` como texto libre sin validacion adicional.
- DON'T: inventar una regla que exija `Notes` para `rejected` u otra
  decision — no esta especificada en `DEFINITION.md`; si se necesita,
  amerita su propio contrato.
- DON'T: referenciar `GenerationRun` ni ningun puerto de IA desde este
  archivo.

## Tests

Los tests estan en `internal/domain/review_test.go` y cubren: review
valida, slide id vacio, decision vacia, decision invalida, decision
`rejected` con notas y decision `edited`.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si el modelo requiere referenciar `GenerationRun` o un
  puerto de IA para cumplir el intent — esa decision todavia esta abierta
  en `DEFINITION.md`.
