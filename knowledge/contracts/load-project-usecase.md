---
type: 'Task Contract'
title: 'Caso de uso: cargar un Project'
description: 'Lee un archivo JSON de project escrito por save-project-usecase y lo deserializa de vuelta a domain.Project.'
tags: ['showme', 'go', 'usecase', 'storage', 'project']

task: load-project-usecase
intent: "Leer un Project persistido como JSON desde el filesystem."
target: internal/storage/project_load.go
signature: "func LoadProject(path string) (domain.Project, error)"
test_command: "go test ./internal/storage"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 5
  max_nesting_depth: 2
  cyclomatic_max: 5
  nesting_max: 2
  params_max: 1
  lines_max: 60
tests: internal/storage/project_load_test.go
tests_sha256: "0201b5bcbe82eb17f199505c72a92f955839441ce9e3066b72abce0e547a432c"
touch_only: ['internal/storage/project_load.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: load-project-usecase

## Intent

Cierra el round-trip de persistencia de `Project` iniciado en
[save-project-usecase](./save-project-usecase.md), igual que
[load-deck-usecase](./load-deck-usecase.md) lo hace para `Deck`: dado el path
de un archivo escrito por `SaveProject`, lo lee y lo deserializa de vuelta a
`domain.Project` para que webapp, CLI y MCP puedan reabrir un proyecto ya
guardado, incluido su `Deck` completo.

## Interface

```go
func LoadProject(path string) (domain.Project, error)
```

## Invariants

- Lee el archivo en `path` y deserializa su contenido JSON a
  `domain.Project`.
- No escribe ni modifica el archivo leido.
- Preserva `Name`, `DesignPath`, `KnowledgePath`, `Version` y el `Deck`
  completo (orden y contenido de `Slides` incluido) tal como fueron
  escritos por `SaveProject`.
- Un archivo inexistente devuelve un error del sistema de archivos, no un
  `domain.Project` vacio silencioso.
- Un contenido que no es JSON valido devuelve un error de parseo.
- No hace red, subprocess ni llamadas a un proveedor de IA.
- No revalida las invariantes de `domain.NewProject`: asume que el archivo
  fue escrito por `SaveProject` (que ya solo persiste proyectos validos).

## Examples

- `SaveProject` seguido de `LoadProject(path)` -> mismo `Name`,
  `DesignPath`, `KnowledgePath`, `Version` y `Deck` que el
  `domain.ProjectInput` original.
- Path a un archivo que no existe -> `err` no nil.
- Archivo con contenido `"{not json"` -> `err` no nil (fallo de parseo).

## Do / Don't

- DO: mantener la funcion simetrica con `SaveProject` (mismo formato JSON,
  mismo tipo `domain.Project`).
- DO: devolver el error del sistema de archivos o de parseo tal cual, igual
  que `LoadDeck`.
- DON'T: revalidar ni "reparar" el contenido leido — si el archivo esta
  corrupto, el error debe propagarse.
- DON'T: buscar el archivo por nombre o slug; el llamador ya tiene el path
  (devuelto por `SaveProject` o listado por un futuro contrato de listado
  de proyectos).

## Tests

Los tests estan en `internal/storage/project_load_test.go` y cubren:
round-trip completo `SaveProject` -> `LoadProject` con version explicita,
archivo inexistente y contenido JSON invalido.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta listar proyectos existentes para cumplir
  el intent — eso excede el alcance de este contrato.
