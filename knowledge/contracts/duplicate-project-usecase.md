---
type: 'Task Contract'
title: 'Caso de uso: duplicar un Project'
description: 'Carga un Project existente y guarda una copia con un nuevo nombre y version reiniciada, sin modificar el original.'
tags: ['showme', 'go', 'usecase', 'project', 'storage']

task: duplicate-project-usecase
intent: "Duplicar un Project existente con un nuevo nombre y version reiniciada."
target: internal/project/duplicate.go
signature: "func DuplicateProject(input DuplicateProjectInput) (string, domain.Report, error)"
test_command: "go test ./internal/project"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 5
  max_nesting_depth: 2
  cyclomatic_max: 5
  nesting_max: 2
  params_max: 1
  lines_max: 60
tests: internal/project/duplicate_test.go
tests_sha256: "09aa26f3c72c18101267dcaf4c73574e8cad426d632c69df24cec26cb0a837f9"
touch_only: ['internal/project/duplicate.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: duplicate-project-usecase

## Intent

Cubre la capacidad "duplicar" listada en `DEFINITION.md` ("Producto y
capacidades" > Webapp: "crear, listar, abrir, duplicar y archivar
presentaciones"). Reusa [load-project-usecase](./load-project-usecase.md) y
[save-project-usecase](./save-project-usecase.md) tal cual: carga el
`Project` original, arma un `domain.ProjectInput` nuevo con el `Deck` y las
rutas de identidad/conocimiento preservadas pero un nombre distinto y sin
`Version` explicita (se reinicia a 1 vía `domain.NewProject`), y lo guarda
como un archivo separado.

## Interface

```go
type DuplicateProjectInput struct {
    SourcePath, NewName, Dir string
}

func DuplicateProject(input DuplicateProjectInput) (path string, report domain.Report, err error)
```

## Invariants

- El archivo en `SourcePath` nunca se modifica ni se borra.
- El `Deck`, `DesignPath` y `KnowledgePath` del proyecto original se
  preservan tal cual en la copia.
- La copia siempre arranca en `Version == 1`, sin importar la version del
  original (es un proyecto nuevo, no una revision del mismo).
- Un error al leer `SourcePath` (archivo inexistente o JSON invalido) se
  propaga tal cual vía `err`; no se intenta guardar nada.
- Un `NewName` invalido (ej. vacio) no escribe ningun archivo: el error de
  `domain.NewProject` (ej. `name is required`) llega en `Report`, igual que
  en `save-project-usecase`.
- Un `Dir` destino inexistente devuelve un error de I/O vía `err`.
- No hace red, subprocess ni llamadas a un proveedor de IA.

## Examples

- Proyecto original `{Name: "Original", Version: 3}` duplicado con
  `NewName: "Copia"` -> nuevo archivo `<Dir>/copia.json` con
  `Name: "Copia"`, `Version: 1`, mismo `Deck`/rutas; el archivo original
  sigue con `Name: "Original"` y `Version: 3`.
- `NewName: ""` -> `path` vacio, `Report.Errors` incluye `name is required`,
  no se escribe archivo.
- `SourcePath` inexistente -> `err` no nil.
- `Dir` destino inexistente -> `err` no nil.

## Do / Don't

- DO: reusar `storage.LoadProject` y `storage.SaveProject` sin
  reimplementar su logica de lectura/escritura o validacion.
- DO: dejar que `domain.NewProject` decida el `Version` por defecto (1);
  este contrato no fija el numero explicitamente.
- DON'T: copiar el `Version` del original — una copia es un proyecto
  nuevo, no una revision.
- DON'T: implementar "archivar" en este archivo; es una capacidad separada
  con su propio contrato.

## Tests

Los tests estan en `internal/project/duplicate_test.go` y cubren:
duplicado valido (nombre nuevo, version reiniciada, original intacto),
nombre nuevo vacio, archivo origen inexistente y directorio destino
inexistente.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta implementar "archivar" o "listar
  proyectos" para cumplir el intent — eso excede el alcance de este
  contrato.
