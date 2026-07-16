---
type: 'Task Contract'
title: 'Caso de uso: actualizar titulo/audiencia de un Deck'
description: 'Reemplaza el Title y Audience de un Deck existente, preservando sus slides tal cual.'
tags: ['showme', 'go', 'usecase', 'domain', 'deck']

task: update-deck-info-usecase
intent: "Reemplazar el titulo y la audiencia de un Deck existente sin tocar sus slides."
target: internal/domain/update_deck_info.go
signature: "func UpdateDeckInfo(input UpdateDeckInfoInput) (Deck, Report)"
test_command: "go test ./internal/domain"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 4
  max_nesting_depth: 2
  cyclomatic_max: 4
  nesting_max: 2
  params_max: 1
  lines_max: 50
tests: internal/domain/update_deck_info_test.go
tests_sha256: "3ae3784c8adaa5ad67acbcddaeaa6049f94b38e031ffb03e2cb140269f483a8d"
touch_only: ['internal/domain/update_deck_info.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: update-deck-info-usecase

## Intent

Hasta ahora `Title` y `Audience` de un `Deck` solo se fijaban al crear el
proyecto ([create-project-usecase](./create-project-usecase.md)); no habia
forma de cambiarlos despues sin recrear todo el deck. Este contrato cubre
"configurar objetivo, audiencia" fuera de la creacion inicial (`DEFINITION.md`,
"Producto y capacidades" > Webapp), reemplazando solo esos dos campos y
preservando las slides intactas.

## Interface

```go
type UpdateDeckInfoInput struct {
    Deck     Deck
    Title    string
    Audience string
}

func UpdateDeckInfo(input UpdateDeckInfoInput) (Deck, Report)
```

## Invariants

- `Title` no puede estar vacio (`title is required`), misma regla que
  [deck-slide-model](./deck-slide-model.md).
- `Audience` es libre y opcional: puede reemplazarse por una cadena vacia
  sin que sea un error.
- `Slides` se preserva exactamente igual (mismo orden y contenido); esta
  funcion nunca las toca.
- Si hay un error, se devuelve una copia de `Deck` sin cambios.
- El `Deck` de entrada nunca se muta.
- La funcion es determinista, no hace I/O, red, subprocess ni llamadas a un
  proveedor de IA.

## Examples

- Deck con `Title: "Roadmap Q3"` y 2 slides; `Title: "Roadmap Q4"`,
  `Audience: "Equipo ejecutivo"` -> deck resultante con el titulo y
  audiencia nuevos, mismas 2 slides en el mismo orden.
- `Title: ""` -> error `title is required`, deck sin cambios.

## Do / Don't

- DO: preservar `Slides` sin ninguna transformacion.
- DO: reusar el mismo mensaje de error (`title is required`) que
  `deck-slide-model` para mantener consistencia entre validaciones de
  `Deck`.
- DON'T: tocar `Slides` de ninguna forma — agregar, quitar, reordenar o
  editar una slide son casos de uso separados
  (`add-slide-usecase`/`remove-slide-usecase`/`update-slide-usecase`/`reorder-slides-usecase`).
- DON'T: persistir el deck actualizado aqui; eso es responsabilidad de
  `save-deck-usecase`/`save-project-usecase`.

## Tests

Los tests estan en `internal/domain/update_deck_info_test.go` y cubren:
actualizacion valida con slides preservadas, titulo vacio, y no-mutacion
del deck original.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta tocar `Slides` para cumplir el intent —
  eso excede el alcance de este contrato.
