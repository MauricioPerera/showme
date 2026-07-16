---
type: 'Task Contract'
title: 'CLI: comando project list'
description: 'Logica pura del comando project list: lista los proyectos guardados en un directorio con su nombre y ruta.'
tags: ['showme', 'go', 'cli', 'project']

task: cli-list-projects-command
intent: "Ejecutar el comando 'project list': listar los proyectos guardados en un directorio con su nombre y ruta."
target: internal/cli/list_projects_command.go
signature: "func RunListProjectsCommand(input ListProjectsCommandInput) (ListProjectsCommandResult, error)"
test_command: "go test ./internal/cli"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 6
  max_nesting_depth: 2
  cyclomatic_max: 6
  nesting_max: 2
  params_max: 1
  lines_max: 80
tests: internal/cli/list_projects_command_test.go
tests_sha256: "82c497ef42a6e1d0259ae165c9e29d48ae454f84ab6d927b04a5ea0dae23e083"
touch_only: ['internal/cli/list_projects_command.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: cli-list-projects-command

## Intent

Segundo comando de la CLI Go de `showme` (`DEFINITION.md`, "CLI": "crear y
listar proyectos"), companero de lectura de
[cli-create-project-command](./cli-create-project-command.md). Reusa
[list-decks-usecase](./list-decks-usecase.md) para encontrar los archivos y
[load-project-usecase](./load-project-usecase.md) para leer el `Name` de
cada uno, sin reimplementar ninguna de las dos.

## Interface

```go
type ListProjectsCommandInput struct {
    Dir string
}

type ProjectSummary struct {
    Name, Path string
    Archived   bool
}

type ListProjectsCommandResult struct {
    Projects []ProjectSummary
    Errors   []string
}

func RunListProjectsCommand(input ListProjectsCommandInput) (ListProjectsCommandResult, error)
```

## Invariants

- `Dir` inexistente se propaga como `err` (via `storage.ListDecks`); el
  resultado queda en su valor cero.
- Cada archivo `.json` de `Dir` que `storage.LoadProject` puede leer
  aparece en `Projects` con su `Name`, su `Path` completo y su `Archived`
  (ver [set-project-archived-usecase](./set-project-archived-usecase.md)).
- Un archivo que falla al cargar (JSON invalido, por ejemplo) NO aborta el
  comando: se omite de `Projects` y se agrega a `Errors` como
  `"<path>: <error subyacente>"`, para que un archivo roto no oculte el
  resto de los proyectos validos.
- `Dir` vacio devuelve `Projects` vacio, `Errors` vacio, `err` nil.
- No hace red, subprocess ni llamadas a un proveedor de IA.

## Examples

- `Dir` con `roadmap-q3.json` (`Name: "Roadmap Q3"`) y
  `onboarding.json` (`Name: "Onboarding"`) -> `Projects` con ambos,
  cada uno con su `Path` correspondiente.
- `Dir` vacio -> `Projects` vacio, `err` nil.
- `Dir` inexistente -> `err` no nil.
- `Dir` con un proyecto valido mas un `broken.json` con contenido invalido
  -> `Projects` solo trae el valido; `Errors` incluye una entrada que
  empieza con `"<ruta de broken.json>: "`.

## Do / Don't

- DO: reusar `storage.ListDecks` y `storage.LoadProject` tal cual; este
  comando es orquestacion de lectura, no reimplementa listado ni parseo.
- DO: seguir cargando el resto de los archivos aunque uno falle — no
  cortar en el primer error.
- DON'T: imprimir a stdout ni parsear flags aqui — eso vive en
  `cmd/showme/main.go`.
- DON'T: ordenar `Projects` por `Name`; el orden que importa es el de
  `storage.ListDecks` (alfabetico por archivo), documentado ahi.

## Tests

Los tests estan en `internal/cli/list_projects_command_test.go` y cubren:
listado con nombres y rutas correctos, reflejo del estado `Archived`,
directorio vacio, directorio inexistente, y un archivo illegible que se
omite sin abortar el listado.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta filtrar, ordenar o paginar la lista para
  cumplir el intent — eso excede el alcance de este contrato.
