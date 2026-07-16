---
type: 'Task Contract'
title: 'CLI: comando project duplicate'
description: 'Logica pura del comando project duplicate: envuelve project.DuplicateProject para la CLI.'
tags: ['showme', 'go', 'cli', 'project']

task: cli-duplicate-project-command
intent: "Ejecutar el comando 'project duplicate': duplicar un proyecto existente con un nuevo nombre."
target: internal/cli/duplicate_project_command.go
signature: "func RunDuplicateProjectCommand(input DuplicateProjectCommandInput) (DuplicateProjectCommandResult, error)"
test_command: "go test ./internal/cli"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 4
  max_nesting_depth: 2
  cyclomatic_max: 4
  nesting_max: 2
  params_max: 1
  lines_max: 50
tests: internal/cli/duplicate_project_command_test.go
tests_sha256: "17af859268b803e39bd86f14115e971b0e828f3edff51ef0f93015e36bb7d9a3"
touch_only: ['internal/cli/duplicate_project_command.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: cli-duplicate-project-command

## Intent

Expone [duplicate-project-usecase](./duplicate-project-usecase.md) como
comando de la CLI (`DEFINITION.md`, capacidad "duplicar" presentaciones),
para que un agente pueda duplicar un proyecto sin invocar Go directamente.
Es un wrapper delgado: no reimplementa `project.DuplicateProject`, solo
adapta su firma a un input/output JSON-estable.

## Interface

```go
type DuplicateProjectCommandInput struct {
    SourcePath, NewName, OutDir string
}

type DuplicateProjectCommandResult struct {
    OK       bool
    Path     string
    Errors, Warnings []string
}

func RunDuplicateProjectCommand(input DuplicateProjectCommandInput) (DuplicateProjectCommandResult, error)
```

## Invariants

- Delega enteramente en `project.DuplicateProject`; un error de I/O (fuente
  inexistente, `OutDir` inexistente) se propaga tal cual via `err`.
- Si `project.DuplicateProject` reporta errores de validacion (ej.
  `NewName` vacio), el resultado tiene `OK: false`, `Path: ""` y esos
  errores en `Errors`.
- `OK` es `true` unicamente cuando `Errors` queda vacio.
- No hace red, subprocess ni llamadas a un proveedor de IA.

## Examples

- `SourcePath` de un proyecto valido, `NewName: "Copia"`, `OutDir`
  existente -> `OK: true`, `Path` apuntando al nuevo archivo.
- `NewName: ""` -> `OK: false`, `Errors` incluye `name is required`.
- `SourcePath` inexistente -> `err` no nil.

## Do / Don't

- DO: mantener la funcion como un wrapper directo de
  `project.DuplicateProject`, sin logica adicional.
- DON'T: imprimir a stdout ni parsear flags aqui — eso vive en
  `cmd/showme/main.go`.
- DON'T: reimplementar el reset de `Version` ni la preservacion de
  `Deck`/rutas — eso ya lo hace `project.DuplicateProject`.

## Tests

Los tests estan en `internal/cli/duplicate_project_command_test.go` y
cubren: duplicado valido, nombre nuevo vacio y archivo origen inexistente.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta implementar "archivar" para cumplir el
  intent — eso excede el alcance de este contrato.
