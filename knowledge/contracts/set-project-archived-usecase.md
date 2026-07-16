---
type: 'Task Contract'
title: 'Caso de uso: marcar un Project como archivado o activo'
description: 'Cambia el campo Archived de un Project, sin tocar el resto de sus campos.'
tags: ['showme', 'go', 'usecase', 'domain', 'project', 'archive']

task: set-project-archived-usecase
intent: "Marcar un Project como archivado o activo."
target: internal/domain/set_project_archived.go
signature: "func SetProjectArchived(input SetProjectArchivedInput) Project"
test_command: "go test ./internal/domain"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 2
  max_nesting_depth: 1
  cyclomatic_max: 2
  nesting_max: 1
  params_max: 1
  lines_max: 25
tests: internal/domain/set_project_archived_test.go
tests_sha256: "8cf8503296a809dc65e771ece4928e9e667be5af4288dc32c34903e6595fa756"
touch_only: ['internal/domain/set_project_archived.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: set-project-archived-usecase

## Intent

Cubre "archivar" del `DEFINITION.md` ("Producto y capacidades" > Webapp:
"crear, listar, abrir, duplicar y archivar presentaciones"), representacion
elegida: un campo `Archived bool` en [`Project`](./project-model.md)
(alternativa a mover el archivo a un subdirectorio). `project-model.md` se
extiende con este campo tanto en `Project` como en `ProjectInput` (para que
los casos de uso de edicion existentes puedan preservarlo al re-guardar);
por defecto nace en `false` y el cambio explicito de estado vive en este
contrato, no en `NewProject`.

## Interface

```go
type SetProjectArchivedInput struct {
    Project  Project
    Archived bool
}

func SetProjectArchived(input SetProjectArchivedInput) Project
```

## Invariants

- Devuelve una copia de `Project` con `Archived` igual al valor dado; todos
  los demas campos (`Name`, `Deck`, `DesignPath`, `KnowledgePath`,
  `Version`) se preservan tal cual.
- No devuelve `Report`: a diferencia del resto de los constructores del
  dominio, un booleano no tiene valores invalidos, asi que esta operacion
  no puede fallar.
- El `Project` de entrada nunca se muta.
- La funcion es determinista, no hace I/O, red, subprocess ni llamadas a un
  proveedor de IA.

## Examples

- `Project` con `Archived: false`, `Archived: true` en el input -> copia
  con `Archived: true`, resto de los campos igual.
- `Project` con `Archived: true`, `Archived: false` en el input -> copia
  con `Archived: false` (desarchivar).

## Do / Don't

- DO: mantener la funcion sin `Report` — agregar uno solo para seguir un
  patron uniforme seria ruido, ya que nunca tendria contenido.
- DO: preservar `Deck` (incluidas sus slides) sin ninguna transformacion.
- DON'T: mover ni renombrar el archivo del proyecto en disco desde aqui —
  eso es responsabilidad de la capa de storage/CLI que lo invoque.
- DON'T: filtrar proyectos archivados de ningun listado desde este
  contrato — eso, si se decide, es responsabilidad de quien consuma
  `Project.Archived` (ej. `list-decks-usecase` o su CLI).

## Tests

Los tests estan en `internal/domain/set_project_archived_test.go` y
cubren: marcar como archivado, desarchivar, y no-mutacion del proyecto
original.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta persistir el proyecto actualizado o
  filtrar listados por `Archived` para cumplir el intent — eso excede el
  alcance de este contrato.
