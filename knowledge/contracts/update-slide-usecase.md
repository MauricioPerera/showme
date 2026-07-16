---
type: 'Task Contract'
title: 'Caso de uso: actualizar una slide de un Deck'
description: 'Reemplaza los campos de una slide existente en un Deck, preservando su status si no se indica uno nuevo.'
tags: ['showme', 'go', 'usecase', 'domain', 'deck', 'slide']

task: update-slide-usecase
intent: "Reemplazar los campos de una slide existente de un Deck por sus nuevos valores."
target: internal/domain/update_slide.go
signature: "func UpdateSlide(input UpdateSlideInput) (Deck, Report)"
test_command: "go test ./internal/domain"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 8
  max_nesting_depth: 3
  cyclomatic_max: 8
  nesting_max: 3
  params_max: 1
  lines_max: 90
tests: internal/domain/update_slide_test.go
tests_sha256: "425595176c808c51c2de141ccc0b04b468a15de8edc8d1787982cda5517f0a0a"
touch_only: ['internal/domain/update_slide.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: update-slide-usecase

## Intent

Cierra el CRUD de slides junto a
[add-slide-usecase](./add-slide-usecase.md) y
[remove-slide-usecase](./remove-slide-usecase.md): dado un `Deck` existente
y los nuevos valores de una de sus slides (identificada por `ID`), reemplaza
esa slide en el mismo lugar de `Slides`. A diferencia de `AddSlide` (una
slide nueva sin historia, `Status` vacio se normaliza a `draft`), aca
`Status` vacio preserva el valor previo de la slide — actualizar el texto de
una slide ya `accepted` no deberia resetear su revision silenciosamente.

## Interface

```go
func UpdateSlide(input UpdateSlideInput) (Deck, Report)
```

Donde `UpdateSlideInput{Deck Deck; Slide Slide}` y `Slide.ID` identifica cual
slide reemplazar; el resto de los campos de `Slide` son sus nuevos valores.

## Invariants

- `Slide.ID` no puede estar vacio; si no corresponde a ninguna slide
  existente de `Deck.Slides`, es un error `slide not found: <id>`.
- `Slide.Title` no puede estar vacio (`slide title is required`).
- `Slide.Status` vacio preserva el `Status` que tenia la slide antes de la
  actualizacion (NO se normaliza a `draft`, a diferencia de `AddSlide`); un
  valor no vacio fuera de `{draft, accepted, rejected}` es un error
  `invalid status: <value>`.
- `Intent` y `Content` se reemplazan tal cual por los valores dados
  (pueden quedar vacios; no tienen invariante propia).
- Si hay algun error, se devuelve una copia de `Deck` sin cambios.
- Si es valida, la slide en la posicion encontrada se reemplaza entera por
  los nuevos valores (incluido el `Status` resuelto); el resto de las
  slides y su orden se preservan.
- El `Deck` de entrada nunca se muta.
- La funcion es determinista, no hace I/O, red, subprocess ni llamadas a un
  proveedor de IA.

## Examples

- Deck con `intro` (`accepted`) y `plan` (`draft`); actualizar `intro` con
  nuevo `Title`/`Content` y `Status` vacio -> `intro` queda con el texto
  nuevo pero `Status: accepted` (preservado); `plan` sin cambios.
- Actualizar `intro` con `Status: rejected` explicito -> `intro` queda
  `rejected`.
- `Slide.ID: "missing"` -> error `slide not found: missing`, deck sin
  cambios.
- `Slide.ID: ""` -> error `slide id is required`.
- `Slide.Title: ""` -> error `slide title is required`.
- `Slide.Status: "archived"` -> error `invalid status: archived`.

## Do / Don't

- DO: preservar el `Status` previo cuando no se especifica uno nuevo — es
  la diferencia deliberada de comportamiento respecto a `AddSlide`.
- DO: copiar `Slides` antes de reemplazar cualquier elemento, para no mutar
  el `Deck` recibido.
- DON'T: permitir cambiar el `ID` de una slide via este contrato — `ID` es
  el criterio de busqueda, no un campo editable aqui.
- DON'T: persistir el deck actualizado aqui; eso es responsabilidad de
  `save-deck-usecase`/`save-project-usecase`.

## Tests

Los tests estan en `internal/domain/update_slide_test.go` y cubren: status
preservado cuando no se indica, status explicito que sobreescribe, slide no
encontrada, id vacio, titulo vacio, status invalido y no-mutacion del deck
original.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta permitir cambiar el `ID` de una slide para
  cumplir el intent — eso excede el alcance de este contrato.
