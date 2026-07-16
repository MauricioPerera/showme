---
type: 'Task Contract'
title: 'CLI: comando project rename'
description: 'Logica pura del comando project rename: envuelve project.RenameProject para la CLI.'
tags: ['showme', 'go', 'cli', 'project']

task: cli-rename-project-command
intent: "Ejecutar el comando 'project rename': cambiar el nombre de un proyecto existente in-place."
target: internal/cli/rename_project_command.go
signature: "func RunRenameProjectCommand(input RenameProjectCommandInput) (RenameProjectCommandResult, error)"
test_command: "go test ./internal/cli"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 4
  max_nesting_depth: 2
  cyclomatic_max: 4
  nesting_max: 2
  params_max: 1
  lines_max: 50
tests: internal/cli/rename_project_command_test.go
tests_sha256: "d4e334119d793c7739dd2b7b1f01be42d4a9fb6f6d83d0690258d125e7983366"
touch_only: ['internal/cli/rename_project_command.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: cli-rename-project-command

## Intent

Expone [rename-project-usecase](./rename-project-usecase.md) por CLI, mismo
patron que [cli-duplicate-project-command](./cli-duplicate-project-command.md):
wrapper delgado que no reimplementa `project.RenameProject`, solo adapta su
firma a un input/output JSON-estable.

## Interface

```go
type RenameProjectCommandInput struct {
    SourcePath, NewName, OutDir string
}

type RenameProjectCommandResult struct {
    OK       bool
    Path     string
    Errors, Warnings []string
}

func RunRenameProjectCommand(input RenameProjectCommandInput) (RenameProjectCommandResult, error)
```

## Invariants

- Delega enteramente en `project.RenameProject`; un error de I/O (fuente
  inexistente, fallo al guardar o al borrar el archivo viejo) se propaga
  tal cual via `err`.
- Si `project.RenameProject` reporta errores de validacion (`NewName`
  vacio, o colision con otro proyecto ya existente), el resultado tiene
  `OK: false`, `Path: ""` y esos errores en `Errors`.
- `OK` es `true` unicamente cuando `Errors` queda vacio.
- No hace red, subprocess ni llamadas a un proveedor de IA.

## Examples

- `SourcePath` de un proyecto valido, `NewName: "Renombrado"`, `OutDir`
  existente -> `OK: true`, `Path` apuntando al archivo renombrado; el
  archivo original ya no existe.
- `NewName` que colisiona con otro proyecto ya guardado -> `OK: false`,
  `Errors` incluye `a project already exists at that name`.
- `SourcePath` inexistente -> `err` no nil.

## Do / Don't

- DO: mantener la funcion como un wrapper directo de
  `project.RenameProject`, sin logica adicional.
- DON'T: imprimir a stdout ni parsear flags aqui — eso vive en
  `cmd/showme/main.go`.
- DON'T: resolver colisiones automaticamente (ej. agregando un sufijo) —
  eso queda explicitamente fuera de `rename-project-usecase`.

## Tests

Los tests estan en `internal/cli/rename_project_command_test.go` y cubren:
renombrado valido, colision rechazada, y archivo origen inexistente.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta resolver colisiones automaticamente para
  cumplir el intent — eso excede el alcance de este contrato.
