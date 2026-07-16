---
type: 'Task Contract'
title: 'Caso de uso: guardar un Project'
description: 'Construye un Project via el modelo de dominio y lo persiste como JSON en el filesystem, sin escribir nada si es invalido.'
tags: ['showme', 'go', 'usecase', 'storage', 'project']

task: save-project-usecase
intent: "Persistir un Project valido como JSON en un directorio del filesystem."
target: internal/storage/project_store.go
signature: "func SaveProject(request SaveProjectRequest) (string, domain.Report, error)"
test_command: "go test ./internal/storage"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 8
  max_nesting_depth: 3
  cyclomatic_max: 8
  nesting_max: 3
  params_max: 1
  lines_max: 140
tests: internal/storage/project_store_test.go
tests_sha256: "4a45e3731df96465f0b348a0538302af0ad15ad10994af117a6f4927c4de1a5e"
touch_only: ['internal/storage/project_store.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: save-project-usecase

## Intent

Extiende el storage de `showme` de `Deck` a `Project`: toma un
`domain.ProjectInput`, lo construye con
[`domain.NewProject`](./project-model.md) y lo persiste como JSON, siguiendo
exactamente el mismo patron que [save-deck-usecase](./save-deck-usecase.md).
Un `Project` valido incluye su `Deck`, sus rutas de identidad/conocimiento y
su version, todo en un unico archivo.

## Interface

```go
type SaveProjectRequest struct {
    Dir   string
    Input domain.ProjectInput
}

func SaveProject(request SaveProjectRequest) (path string, report domain.Report, err error)
```

## Invariants

- Si `domain.NewProject` produce errores, `SaveProject` no escribe ningun
  archivo: devuelve `path` vacio y el mismo `Report` con los errores de
  validacion.
- El nombre de archivo es un slug deterministico del `Name` del proyecto
  (mismo algoritmo de slug que `save-deck-usecase`) mas la extension
  `.json`.
- Si el slug resultante queda vacio, se agrega el error
  `name produces an empty slug` a `Report` y no se escribe archivo.
- El archivo escrito es JSON valido que deserializa a `domain.Project`
  preservando `Name`, `DesignPath`, `KnowledgePath`, `Version` y el `Deck`
  completo (incluidas sus slides en orden).
- Un error de I/O (ej. el directorio `Dir` no existe) se devuelve por `err`,
  nunca por `Report`.
- La funcion no crea directorios: `Dir` debe existir de antemano.
- No hace red, subprocess ni llamadas a un proveedor de IA.

## Examples

- `Name: "Presentacion Q3"` con un deck valido y rutas no vacias -> escribe
  `<Dir>/presentacion-q3.json`, `Report.Errors` vacio, `err` nil.
- `Name: ""` -> `Report.Errors` incluye `name is required` (via
  `domain.NewProject`), `path` vacio, no se escribe archivo.
- `Name: "!!!"` con el resto de los campos validos -> `Report.Errors`
  incluye `name produces an empty slug`, no se escribe archivo.
- `Dir` inexistente con un proyecto valido -> `err` no nil,
  `Report.Errors` vacio.

## Do / Don't

- DO: delegar toda la validacion estructural a `domain.NewProject`; este
  contrato no reimplementa esas reglas.
- DO: mantener la misma convencion de slug y de serializacion JSON legible
  que `save-deck-usecase`, para que ambos formatos de archivo sean
  consistentes.
- DON'T: crear el directorio destino ni resolver rutas relativas a un
  proyecto — eso es responsabilidad del llamador.
- DON'T: implementar lectura (`LoadProject`) en este archivo; es un
  contrato separado.

## Tests

Los tests estan en `internal/storage/project_store_test.go` y cubren:
proyecto valido con round-trip JSON y ruta de archivo esperada, proyecto
invalido que no escribe nada, nombre que produce un slug vacio, y
directorio inexistente que devuelve error de I/O.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta crear directorios, leer un Project
  existente o tocar `internal/domain/project.go` para cumplir el intent —
  eso excede el alcance de este contrato.
