---
type: 'Task Contract'
title: 'Modelo de dominio Project'
description: 'Construye un Project que contiene un Deck valido junto a las referencias a su DESIGN.md y bundle OKF, validando sus invariantes minimas.'
tags: ['showme', 'go', 'domain', 'project']

task: project-model
intent: "Construir un Project valido a partir de un nombre, un Deck y sus referencias de identidad/conocimiento."
target: internal/domain/project.go
signature: "func NewProject(input ProjectInput) (Project, Report)"
test_command: "go test ./internal/domain"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 8
  max_nesting_depth: 3
  cyclomatic_max: 8
  nesting_max: 3
  params_max: 1
  lines_max: 100
tests: internal/domain/project_test.go
tests_sha256: "e7f335a220e30d892c3cb5b5741eaa0984d2bc4b96cfeb4566b865b0ca73979c"
touch_only: ['internal/domain/project.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: project-model

## Intent

Define `Project`, el contenedor de una presentacion segun `DEFINITION.md`
("Modelo conceptual minimo"): agrupa un [`Deck`](./deck-slide-model.md) ya
construido con las rutas a su identidad visual (`DESIGN.md`) y su bundle de
conocimiento (OKF), mas un numero de version. Este contrato NO resuelve ni
valida el contenido de esas rutas — eso ya lo hacen
[design-loader](./design-loader.md) y [okf-bundle-loader](./okf-bundle-loader.md)
por separado; `Project` solo exige que las referencias existan como texto no
vacio.

## Interface

```go
type ProjectInput struct {
    Name, DesignPath, KnowledgePath string
    Deck     Deck
    Version  int
    Archived bool
    Runs     []GenerationRun
}

type Project struct {
    Name, DesignPath, KnowledgePath string
    Deck     Deck
    Version  int
    Archived bool
    Runs     []GenerationRun
}

func NewProject(input ProjectInput) (Project, Report)
```

`Archived` fue agregado por
[set-project-archived-usecase](./set-project-archived-usecase.md) y `Runs`
por [append-generation-run-usecase](./append-generation-run-usecase.md):
ninguno tiene invariante propia (un bool no puede ser invalido, un slice ya
construido por `AppendGenerationRun` tampoco), se copian tal cual de
`ProjectInput` a `Project`, y por defecto son `false`/`nil` cuando el
llamador no los especifica. Ambos se incluyen en `ProjectInput` (y no solo
en `Project`) por la misma razon: los casos de uso que releen y re-guardan
un proyecto ya existente (`review`, `add-slide`, `remove-slide`,
`update-slide`, `reorder-slides`, `update-info`, `generate-slide`) deben
preservarlos pasando `Archived: proj.Archived` y `Runs: proj.Runs` — de lo
contrario, cualquier edicion posterior resetearia el archivado o borraria
el historial de generaciones silenciosamente.

## Invariants

- `Name` no puede estar vacio.
- `Deck.Slides` debe tener al menos un elemento (un `Project` nunca envuelve
  un deck vacio).
- `DesignPath` y `KnowledgePath` no pueden estar vacios.
- `Version` menor a 0 es un error; `Version == 0` se normaliza a `1`; un
  valor positivo explicito se preserva tal cual.
- La funcion es determinista, no hace I/O, red, subprocess ni llamadas a un
  proveedor de IA, y nunca lanza panic: los problemas se acumulan en
  `Report`, igual que en `deck-slide-model`.

## Examples

- Nombre, deck valido (via `NewDeck`), `DesignPath` y `KnowledgePath` no
  vacios, sin `Version` -> `Report.Errors` vacio y `Project.Version == 1`.
- `Name` vacio -> error `name is required`.
- `Deck` sin slides -> error `deck must have at least one slide`.
- `DesignPath` vacio -> error `design path is required`.
- `KnowledgePath` vacio -> error `knowledge path is required`.
- `Version: -1` -> error `version must be positive`.
- `Version: 3` -> `Project.Version == 3` (no se normaliza).

## Do / Don't

- DO: recibir un `Deck` ya construido por `NewDeck`; no reimplementar sus
  invariantes aqui, solo exigir que tenga al menos una slide.
- DO: acumular todos los errores encontrados en `Report`, no cortar en el
  primero.
- DON'T: leer `DESIGN.md` ni el bundle OKF desde este archivo — `DesignPath`
  y `KnowledgePath` son referencias de texto, no contenido validado.
- DON'T: incluir `GenerationRun` ni `Review` en este modelo; son conceptos
  con su propio contrato futuro.

## Tests

Los tests estan en `internal/domain/project_test.go` y cubren: proyecto
valido con version por defecto, nombre vacio, deck sin slides, ruta de
diseno vacia, ruta de conocimiento vacia, version negativa y version
explicita preservada.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si el modelo requiere resolver o leer `DesignPath` /
  `KnowledgePath`, o modelar `GenerationRun`/`Review`, para cumplir el
  intent — eso excede el alcance de este contrato.
