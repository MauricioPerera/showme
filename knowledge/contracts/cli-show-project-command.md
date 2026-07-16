---
type: 'Task Contract'
title: 'CLI: comando project show'
description: 'Logica pura del comando project show: carga y devuelve un Project completo dado su path.'
tags: ['showme', 'go', 'cli', 'project']

task: cli-show-project-command
intent: "Ejecutar el comando 'project show': cargar y devolver un Project completo dado su path."
target: internal/cli/show_project_command.go
signature: "func RunShowProjectCommand(input ShowProjectCommandInput) (ShowProjectCommandResult, error)"
test_command: "go test ./internal/cli"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 3
  max_nesting_depth: 2
  cyclomatic_max: 3
  nesting_max: 2
  params_max: 1
  lines_max: 40
tests: internal/cli/show_project_command_test.go
tests_sha256: "69300defbc2afb6a8e1f48eaf6e127f23a46754a2667202ddde6b0ed2a209918"
touch_only: ['internal/cli/show_project_command.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: cli-show-project-command

## Intent

Cierra el trio de lectura de la CLI junto a
[cli-create-project-command](./cli-create-project-command.md) y
[cli-list-projects-command](./cli-list-projects-command.md): dado el `Path`
de un proyecto (tipicamente uno devuelto por `project list`), lo carga
completo con [load-project-usecase](./load-project-usecase.md) para que un
agente pueda inspeccionarlo (deck, slides, rutas, version) sin reimplementar
la lectura.

## Interface

```go
type ShowProjectCommandInput struct {
    Path string
}

type ShowProjectCommandResult struct {
    Project domain.Project
}

func RunShowProjectCommand(input ShowProjectCommandInput) (ShowProjectCommandResult, error)
```

## Invariants

- Delega enteramente en `storage.LoadProject`: un archivo inexistente o con
  JSON invalido se propaga tal cual via `err`.
- El `Project` devuelto preserva `Name`, `Deck` (slides en orden), `DesignPath`,
  `KnowledgePath` y `Version` tal como estan en el archivo.
- No hace red, subprocess ni llamadas a un proveedor de IA; no modifica el
  archivo leido.

## Examples

- `Path` de un proyecto guardado con `Name: "Roadmap Q3"` y una slide ->
  `Project.Name == "Roadmap Q3"`, `Project.Deck.Slides` con esa slide.
- `Path` inexistente -> `err` no nil.
- Archivo con contenido `"{not json"` -> `err` no nil (fallo de parseo).

## Do / Don't

- DO: mantener la funcion como un wrapper directo de `storage.LoadProject`,
  sin logica adicional.
- DON'T: buscar el proyecto por nombre o slug; el llamador ya tiene el
  `Path` (de `project create` o `project list`).
- DON'T: imprimir a stdout ni parsear flags aqui — eso vive en
  `cmd/showme/main.go`.

## Tests

Los tests estan en `internal/cli/show_project_command_test.go` y cubren:
proyecto valido con deck preservado, archivo inexistente y contenido JSON
invalido.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta buscar el proyecto por nombre en vez de
  por path para cumplir el intent — eso excede el alcance de este
  contrato.
