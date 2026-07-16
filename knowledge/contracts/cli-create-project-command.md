---
type: 'Task Contract'
title: 'CLI: comando project create'
description: 'Logica pura del comando project create: lee DESIGN.md y un deck JSON, ensambla y persiste el Project.'
tags: ['showme', 'go', 'cli', 'project']

task: cli-create-project-command
intent: "Ejecutar el comando 'project create': leer DESIGN.md y un deck JSON, ensamblar el Project y persistirlo si es valido."
target: internal/cli/create_project_command.go
signature: "func RunCreateProjectCommand(input CreateProjectCommandInput) (CreateProjectCommandResult, error)"
test_command: "go test ./internal/cli"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 10
  max_nesting_depth: 3
  cyclomatic_max: 10
  nesting_max: 3
  params_max: 1
  lines_max: 160
tests: internal/cli/create_project_command_test.go
tests_sha256: "3432a31db7d94b297d26c618c8841ee571da94f5004377456631f638f0f90a7b"
touch_only: ['internal/cli/create_project_command.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: cli-create-project-command

## Intent

Primer comando de la CLI Go de `showme` (`DEFINITION.md`, "CLI": "crear y
listar proyectos... emitir salida legible para humanos y JSON estable para
automatizacion"), pensado para que un agente pueda invocarlo directamente:
recibe rutas a `DESIGN.md` y a un archivo JSON de deck, y usa
[create-project-usecase](./create-project-usecase.md) +
[save-project-usecase](./save-project-usecase.md) para ensamblar y persistir
el `Project`. Este contrato cubre solo la logica pura (parseo + orquestacion);
el wiring de flags/stdout (`cmd/showme/main.go`) es glue no cubierta por
oraculo, mismo criterio que `tui/main.go`.

## Interface

```go
type CreateProjectCommandInput struct {
    Name, DesignPath, KnowledgeRoot, DeckPath, OutDir string
}

type CreateProjectCommandResult struct {
    OK       bool
    Path     string
    Errors, Warnings []string
}

func RunCreateProjectCommand(input CreateProjectCommandInput) (CreateProjectCommandResult, error)
```

El archivo en `DeckPath` es JSON con el shape:
```json
{"title": "...", "audience": "...", "slides": [{"id": "...", "title": "...", "intent": "...", "content": "...", "status": "..."}]}
```

## Invariants

- Un error leyendo `DesignPath` o `DeckPath`, o un JSON de deck invalido, se
  devuelve por `err`; `CreateProjectCommandResult` queda en su valor cero.
- Si `project.CreateProject` reporta errores, el resultado tiene `OK: false`,
  `Path: ""` y esos errores en `Errors`; no se llama a `storage.SaveProject`
  (no se escribe ningun archivo).
- Si el proyecto es valido pero `storage.SaveProject` reporta errores (ej.
  `name produces an empty slug`), el resultado agrega esos errores a los ya
  acumulados, `OK` queda en `false` y `Path` vacio.
- Un error de I/O de `storage.SaveProject` (ej. `OutDir` inexistente) se
  propaga por `err`.
- `OK` es `true` unicamente cuando `Errors` queda vacio tras agregar los
  hallazgos de ensamblado y de guardado.
- No hace red, subprocess ni llamadas a un proveedor de IA.

## Examples

- `DESIGN.md` valido, bundle OKF valido, deck JSON valido y `Name` no vacio
  -> `OK: true`, `Path` apuntando al archivo guardado bajo `OutDir`.
- `DesignPath` inexistente -> `err` no nil.
- Contenido de `DeckPath` no es JSON valido -> `err` no nil.
- `DESIGN.md` sin frontmatter -> `OK: false`, `Path: ""`, `Errors` incluye
  `frontmatter is required`, `OutDir` queda sin archivos nuevos.
- `Name: "!!!"` (produce un slug vacio) -> `OK: false`, `Errors` incluye
  `name produces an empty slug`.

## Do / Don't

- DO: reusar `project.CreateProject` y `storage.SaveProject` tal cual; este
  comando es orquestacion + parseo del deck JSON, no reimplementa
  validaciones.
- DO: usar DTOs con tags JSON en minuscula (`slideDTO`/`deckInputDTO`) para
  el archivo de deck, en vez de exponer los nombres de campo Go de
  `domain.Slide`/`domain.DeckInput`.
- DON'T: imprimir a stdout ni parsear flags aqui — eso vive en
  `cmd/showme/main.go`, fuera de este contrato.
- DON'T: agregar mas comandos a este archivo; cada comando CLI nuevo es un
  contrato separado.

## Tests

Los tests estan en `internal/cli/create_project_command_test.go` y cubren:
comando valido con proyecto persistido, archivo de diseno inexistente, JSON
de deck invalido, errores de validacion que no escriben archivo, y nombre
que produce un slug vacio.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta implementar el wiring de `main.go` o mas
  comandos para cumplir el intent — eso excede el alcance de este
  contrato.
