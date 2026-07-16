---
type: 'Task Contract'
title: 'Caso de uso: renombrar un Project in-place'
description: 'Cambia el Name de un Project guardado, moviendo/renombrando su archivo, sin resetear Version ni Archived.'
tags: ['showme', 'go', 'usecase', 'project', 'storage']

task: rename-project-usecase
intent: "Renombrar un Project existente in-place, sin crear una copia ni resetear su version."
target: internal/project/rename.go
signature: "func RenameProject(input RenameProjectInput) (string, domain.Report, error)"
test_command: "go test ./internal/project"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 8
  max_nesting_depth: 3
  cyclomatic_max: 8
  nesting_max: 3
  params_max: 1
  lines_max: 90
tests: internal/project/rename_test.go
tests_sha256: "3bf283c7421981afc0bd15365b61f840ea72a36cedfe4ce2bd97416f99d4c2f8"
touch_only: ['internal/project/rename.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: rename-project-usecase

## Intent

Cierra el hueco que dejaron explicitamente diferido
[duplicate-project-usecase](./duplicate-project-usecase.md) y
[cli-update-deck-info-command](./cli-update-deck-info-command.md): cambiar
el `Name` de un `Project` sin crear una copia y sin resetear `Version` ni
`Archived` (a diferencia de `DuplicateProject`, que si los resetea porque
una copia es un proyecto nuevo). Como `Name` determina el slug del archivo
(via `storage.Slugify`), renombrar puede implicar mover el contenido a un
archivo distinto y borrar el original.

## Interface

```go
type RenameProjectInput struct {
    SourcePath, NewName, Dir string
}

func RenameProject(input RenameProjectInput) (path string, report domain.Report, err error)
```

## Invariants

- Carga el `Project` en `SourcePath`; un error de lectura o parseo se
  propaga tal cual via `err`.
- El path resultante se calcula igual que `storage.SaveProject`:
  `filepath.Join(Dir, storage.Slugify(NewName)+".json")`.
- Si el path resultante coincide con `SourcePath` (el nombre no cambio de
  slug), la operacion es un no-op exitoso: se re-guarda el mismo archivo,
  no se borra nada.
- Si el path resultante es distinto de `SourcePath` y YA EXISTE un archivo
  ahi, la operacion se rechaza con el error
  `a project already exists at that name` — ni el archivo origen ni el que
  ya existia se tocan. Esto evita perder datos de otro proyecto por una
  colision de slug silenciosa.
- Si `NewName` es invalido (ej. vacio), `storage.SaveProject` devuelve el
  error de `domain.NewProject` correspondiente (ej. `name is required`) en
  `Report`; el archivo origen no se borra.
- Si el renombrado es valido y no hay colision, se guarda el `Project` con
  el nuevo `Name` (preservando `Deck`, `DesignPath`, `KnowledgePath`,
  `Version` y `Archived` tal cual) y, si el path cambio, se borra
  `SourcePath`.
- Un error de I/O al guardar o al borrar el archivo original se propaga
  via `err`.
- No hace red, subprocess ni llamadas a un proveedor de IA.

## Examples

- `SourcePath` de `{Name: "Original", Version: 3}`, `NewName:
  "Renombrado"` -> nuevo archivo `<Dir>/renombrado.json` con
  `Name: "Renombrado"`, `Version: 3` (preservado); `SourcePath` ya no
  existe.
- `NewName` igual al nombre actual -> mismo path devuelto, el archivo
  sigue existiendo, sin error.
- `NewName` que produce el mismo slug que un proyecto YA existente y
  distinto -> `path` vacio, error `a project already exists at that name`,
  ningun archivo se modifica.
- `NewName: ""` -> `path` vacio, `Report.Errors` incluye
  `name is required`, `SourcePath` intacto.
- `SourcePath` inexistente -> `err` no nil.

## Do / Don't

- DO: reusar `storage.LoadProject`, `storage.Slugify` y
  `storage.SaveProject` tal cual; este contrato es orquestacion pura mas
  la logica de colision y borrado.
- DO: preservar `Version` y `Archived` — a diferencia de
  `duplicate-project-usecase`, esto es el MISMO proyecto, no uno nuevo.
- DON'T: sobreescribir un archivo ya existente distinto del origen — la
  deteccion de colision es obligatoria antes de guardar.
- DON'T: borrar `SourcePath` si `storage.SaveProject` reporto errores o si
  el path resultante es el mismo que el origen.

## Tests

Los tests estan en `internal/project/rename_test.go` y cubren:
renombrado valido con version/borrado correctos, mismo nombre como no-op,
colision rechazada sin tocar ningun archivo, nombre nuevo vacio, y archivo
origen inexistente.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta resolver colisiones automaticamente (ej.
  sufijo numerico) para cumplir el intent — eso excede el alcance de este
  contrato; hoy se rechaza explicitamente.
