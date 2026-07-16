---
type: 'Task Contract'
title: 'Caso de uso: agregar un GenerationRun al historial de un Project'
description: 'Agrega un GenerationRun ya construido al historial Runs de un Project, preservando el orden y sin mutar el original.'
tags: ['showme', 'go', 'usecase', 'domain', 'project', 'ai', 'generation-run']

task: append-generation-run-usecase
intent: "Agregar un GenerationRun al historial de generaciones de IA de un Project."
target: internal/domain/append_generation_run.go
signature: "func AppendGenerationRun(input AppendGenerationRunInput) Project"
test_command: "go test ./internal/domain"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 2
  max_nesting_depth: 1
  cyclomatic_max: 2
  nesting_max: 1
  params_max: 1
  lines_max: 30
tests: internal/domain/append_generation_run_test.go
tests_sha256: "5f64bf83679f46f67270d8f2c662fc84b8f635e7f6cda7aa04a78636262399e9"
touch_only: ['internal/domain/append_generation_run.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: append-generation-run-usecase

## Intent

Conecta [generation-run-model](./generation-run-model.md) con
[project-model](./project-model.md): dado un `Project` y un
`GenerationRun` ya construido y validado, lo agrega al final de
`Project.Runs`. Es el paso que
[cli-generate-slide-content-command](./cli-generate-slide-content-command.md)
usara para dejar un rastro auditable de cada generacion antes de guardar el
proyecto.

## Interface

```go
type AppendGenerationRunInput struct {
    Project Project
    Run     GenerationRun
}

func AppendGenerationRun(input AppendGenerationRunInput) Project
```

## Invariants

- Devuelve una copia de `Project` con `Run` agregado al final de `Runs`,
  preservando el orden y contenido de los runs previos.
- No valida `Run`: asume que ya paso por `domain.NewGenerationRun` (que ya
  garantiza sus invariantes). No devuelve `Report` por la misma razon que
  [set-project-archived-usecase](./set-project-archived-usecase.md): un
  append sobre un valor ya valido no tiene forma de fallar.
- El resto de los campos de `Project` (`Name`, `Deck`, `DesignPath`,
  `KnowledgePath`, `Version`, `Archived`) se preservan tal cual.
- El `Project` de entrada nunca se muta.
- La funcion es determinista, no hace I/O, red, subprocess ni llamadas a
  un proveedor de IA.

## Examples

- `Project` con `Runs` vacio, un `Run` valido -> `Project.Runs` con ese
  unico run.
- `Project` con un run ya presente, un segundo `Run` -> `Project.Runs` con
  ambos, en el orden en que se agregaron.

## Do / Don't

- DO: copiar `Runs` antes de agregar cualquier elemento, para no mutar el
  `Project` recibido.
- DO: confiar en que `Run` ya fue validado por `domain.NewGenerationRun`;
  no reimplementar esas reglas aqui.
- DON'T: persistir el proyecto actualizado aqui; eso es responsabilidad de
  `save-project-usecase`.
- DON'T: limitar o rotar el historial (ej. quedarse solo con los ultimos
  N runs) — eso, si se decide, es un contrato separado.

## Tests

Los tests estan en `internal/domain/append_generation_run_test.go` y
cubren: agregar el primer run, agregar un segundo run preservando el
orden, y no-mutacion del proyecto original.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta validar o limitar el historial de runs
  para cumplir el intent — eso excede el alcance de este contrato.
