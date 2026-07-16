---
type: 'Task Contract'
title: 'CLI: comandos project archive / project unarchive'
description: 'Logica pura de los comandos project archive/unarchive: cambia el campo Archived de un Project guardado y lo re-persiste.'
tags: ['showme', 'go', 'cli', 'project', 'archive']

task: cli-archive-project-command
intent: "Ejecutar los comandos 'project archive'/'project unarchive': cambiar el estado archivado de un proyecto guardado."
target: internal/cli/archive_project_command.go
signature: "func RunArchiveProjectCommand(input ArchiveProjectCommandInput) (ArchiveProjectCommandResult, error)"
test_command: "go test ./internal/cli"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 4
  max_nesting_depth: 2
  cyclomatic_max: 4
  nesting_max: 2
  params_max: 1
  lines_max: 70
tests: internal/cli/archive_project_command_test.go
tests_sha256: "1f8a4bb93842a8399830c0da6caf99b0d3c7400a8f3ac469a0c321b6a9e8c319"
touch_only: ['internal/cli/archive_project_command.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: cli-archive-project-command

## Intent

Expone [set-project-archived-usecase](./set-project-archived-usecase.md)
por CLI, con un unico comando parametrizado por un booleano (`Archived`)
que cubre tanto `project archive` como `project unarchive`: carga un
`Project` guardado, cambia su `Archived` con `domain.SetProjectArchived` y
re-guarda con [save-project-usecase](./save-project-usecase.md).

## Interface

```go
type ArchiveProjectCommandInput struct {
    Path     string
    Archived bool
    OutDir   string
}

type ArchiveProjectCommandResult struct {
    OK       bool
    Path     string
    Errors, Warnings []string
}

func RunArchiveProjectCommand(input ArchiveProjectCommandInput) (ArchiveProjectCommandResult, error)
```

## Invariants

- Un error cargando `Path` (archivo inexistente o JSON invalido) se
  propaga tal cual via `err`.
- `domain.SetProjectArchived` no puede fallar (ver su contrato), asi que
  un `err` no nil siempre es un problema de I/O de `storage.LoadProject` o
  `storage.SaveProject`, nunca una validacion.
- El `Project` actualizado (mismo `Name`, `Deck`, `DesignPath`,
  `KnowledgePath`, `Version`; `Archived` igual al valor pedido) se guarda
  con `storage.SaveProject` bajo `OutDir`. Si `OutDir` y `Name` coinciden
  con el archivo original, esto sobreescribe el mismo archivo.
- `OK` es `true` salvo que `storage.SaveProject` reporte un error de
  validacion (en la practica, solo si `Name` produce un slug vacio —
  situacion ya imposible si el proyecto se guardo antes con
  `save-project-usecase`).
- No hace red, subprocess ni llamadas a un proveedor de IA.

## Examples

- Proyecto guardado con `Archived: false`, `Archived: true` en el input,
  `OutDir` igual al directorio original -> `OK: true`; al recargarlo
  `Archived == true`.
- El mismo proyecto, `Archived: false` en el input -> `Archived == false`
  (desarchivar).
- `Path` inexistente -> `err` no nil.

## Do / Don't

- DO: reusar `storage.LoadProject`, `domain.SetProjectArchived` y
  `storage.SaveProject` tal cual; este comando es orquestacion pura.
- DO: mantener un solo comando parametrizado por el booleano `Archived` en
  vez de dos contratos casi identicos para "archivar" y "desarchivar".
- DON'T: filtrar `project list` por `Archived` desde este archivo — si se
  decide, es responsabilidad de `cli-list-projects-command`.
- DON'T: imprimir a stdout ni parsear flags aqui — eso vive en
  `cmd/showme/main.go` (`project archive` mapea a `Archived: true`,
  `project unarchive` a `Archived: false`).

## Tests

Los tests estan en `internal/cli/archive_project_command_test.go` y
cubren: archivar, desarchivar, y archivo origen inexistente.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta filtrar listados por `Archived` para
  cumplir el intent — eso excede el alcance de este contrato.
