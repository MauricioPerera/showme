---
type: 'Task Contract'
title: 'Caso de uso: crear un Project'
description: 'Valida un DESIGN.md, carga un bundle OKF, construye un Deck y los ensambla en un Project, agregando los hallazgos de cada paso.'
tags: ['showme', 'go', 'usecase', 'project', 'design', 'okf']

task: create-project-usecase
intent: "Ensamblar un Project validando en un solo paso su identidad visual, su bundle de conocimiento y su Deck."
target: internal/project/create.go
signature: "func CreateProject(input CreateProjectInput) (domain.Project, Report)"
test_command: "go test ./internal/project"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 8
  max_nesting_depth: 3
  cyclomatic_max: 8
  nesting_max: 3
  params_max: 1
  lines_max: 120
tests: internal/project/create_test.go
tests_sha256: "eccf0ab883bfcef64d9b2a6b4564558786534606612213f3421e10cf7d739ce5"
touch_only: ['internal/project/create.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: create-project-usecase

## Intent

Primer caso de uso que ata los cuatro contratos previos en un solo punto de
entrada: valida `DESIGN.md` con [design-loader](./design-loader.md), carga el
bundle OKF con [okf-bundle-loader](./okf-bundle-loader.md), construye el
`Deck` con [deck-slide-model](./deck-slide-model.md) y ensambla todo en un
`Project` con [project-model](./project-model.md). Es el paso que webapp,
CLI y MCP podran invocar directamente para "crear un proyecto" (ver
`DEFINITION.md`, "Usuarios y flujo principal").

## Interface

```go
type CreateProjectInput struct {
    Name, DesignContent, DesignPath, KnowledgeRoot string
    DeckInput domain.DeckInput
    Version   int
}

type Report struct { Errors, Warnings []string }

func CreateProject(input CreateProjectInput) (domain.Project, Report)
```

## Invariants

- Cada paso corre siempre, incluso si un paso anterior fallo: el `Report`
  final agrega los errores/warnings de `design.Validate`, `knowledge.Load`,
  `domain.NewDeck` y `domain.NewProject`, en ese orden.
- Un `Deck` invalido (ej. sin slides) produce tanto el error de
  `domain.NewDeck` (`at least one slide is required`) como el error
  cascadeado de `domain.NewProject` (`deck must have at least one slide`)
  sobre el `Deck` ya invalido — no se deduplican.
- La funcion no hace red, subprocess ni llamadas a un proveedor de IA
  directamente: delega toda I/O de lectura de archivos a `knowledge.Load`
  (unico paso que toca el filesystem).
- El `Project` devuelto siempre refleja el mejor esfuerzo de construccion
  (nunca es un valor cero arbitrario), igual que el resto de los
  constructores del dominio.

## Examples

- `DesignContent` valido, `KnowledgeRoot` con un concepto con `type`,
  `DeckInput` valido y `Name` no vacio -> `Report.Errors` vacio,
  `Project.Version == 1`.
- `DesignContent` sin frontmatter -> `Report.Errors` incluye
  `frontmatter is required`.
- Un archivo en `KnowledgeRoot` sin `type` -> `Report.Errors` incluye
  `<archivo>: type is required`.
- `DeckInput` sin slides -> `Report.Errors` incluye tanto
  `at least one slide is required` como `deck must have at least one slide`.
- `Name` vacio -> `Report.Errors` incluye `name is required`.

## Do / Don't

- DO: reusar `design.Validate`, `knowledge.Load`, `domain.NewDeck` y
  `domain.NewProject` tal cual; este contrato es pura orquestacion, no
  reimplementa ninguna de esas validaciones.
- DO: preservar el orden de agregacion (design, knowledge, deck, project)
  para que los mensajes de error sean predecibles.
- DON'T: cortar la ejecucion en el primer error — todos los pasos corren
  siempre para dar el diagnostico completo en una sola llamada.
- DON'T: agregar generacion de contenido, IA o persistencia aqui; eso son
  casos de uso separados (ver `save-deck-usecase` para el lado de storage).

## Tests

Los tests estan en `internal/project/create_test.go` y cubren: proyecto
valido con todos los pasos en verde, `DESIGN.md` invalido, concepto OKF sin
`type`, deck invalido cascadeando a un error de proyecto, y nombre vacio.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta persistir el `Project` resultante o generar
  contenido con IA para cumplir el intent — eso excede el alcance de este
  contrato.
