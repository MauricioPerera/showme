---
type: 'Task Contract'
title: 'Modelo de dominio GenerationRun'
description: 'Construye un GenerationRun que registra los inputs, configuracion y salida de una ejecucion de IA, para trazabilidad.'
tags: ['showme', 'go', 'domain', 'ai', 'generation-run']

task: generation-run-model
intent: "Construir un GenerationRun valido que registre una ejecucion de IA para trazabilidad."
target: internal/domain/generation_run.go
signature: "func NewGenerationRun(input GenerationRunInput) (GenerationRun, Report)"
test_command: "go test ./internal/domain"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 6
  max_nesting_depth: 2
  cyclomatic_max: 6
  nesting_max: 2
  params_max: 1
  lines_max: 60
tests: internal/domain/generation_run_test.go
tests_sha256: "96c5308a0b05d5d266ff4909a45a02a6e3bdd9acf9792470968bbf565d956303"
touch_only: ['internal/domain/generation_run.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: generation-run-model

## Intent

Define `GenerationRun` del "Modelo conceptual minimo" de `DEFINITION.md`:
"inputs, modelo, parametros, salida, advertencias y referencias de una
ejecucion de IA... debe conservar suficiente informacion para explicar que
contexto, tokens y configuracion produjo la salida." Hasta ahora
[cli-generate-slide-content-command](./cli-generate-slide-content-command.md)
generaba contenido sin dejar ningun rastro auditable de como se produjo;
este contrato es el registro que lo cierra.

## Interface

```go
type GenerationRunInput struct {
    SlideID, Model, Provider, Intent, Context, Output, CreatedAt string
    Warnings []string
}

type GenerationRun struct {
    SlideID, Model, Provider, Intent, Context, Output, CreatedAt string
    Warnings []string
}

func NewGenerationRun(input GenerationRunInput) (GenerationRun, Report)
```

## Invariants

- `SlideID` no puede estar vacio (`slide id is required`).
- `Model` no puede estar vacio (`model is required`) — sin saber que
  modelo produjo la salida, el registro no es auditable.
- `Output` no puede estar vacio (`output is required`) — un run sin salida
  no aporta nada al historial.
- `CreatedAt` no puede estar vacio (`created at is required`); si esta
  presente debe parsear como RFC3339 (`time.Parse(time.RFC3339, ...)`), o
  es un error `created at must be a valid RFC3339 timestamp`.
- `CreatedAt` es siempre provisto por el llamador, NUNCA generado dentro
  de esta funcion (no hay `time.Now()` aca): esto mantiene el constructor
  puro y deterministico, testeable con timestamps fijos. Quien lo invoca
  en produccion (la CLI) es responsable de pasar la hora real.
- `Provider`, `Intent`, `Context` y `Warnings` son informativos, sin
  invariante propia (pueden quedar vacios).
- La funcion es determinista, no hace I/O, red, subprocess ni llamadas a
  un proveedor de IA por si misma.

## Examples

- `SlideID: "intro"`, `Model: "..."`, `Output: "..."`,
  `CreatedAt: "2026-07-16T12:00:00Z"` -> `Report.Errors` vacio.
- `SlideID: ""` -> error `slide id is required`.
- `Model: ""` -> error `model is required`.
- `Output: ""` -> error `output is required`.
- `CreatedAt: ""` -> error `created at is required`.
- `CreatedAt: "not-a-timestamp"` -> error
  `created at must be a valid RFC3339 timestamp`.

## Do / Don't

- DO: mantener `CreatedAt` como parametro de entrada, nunca generado con
  `time.Now()` dentro de esta funcion.
- DO: acumular todos los errores encontrados en `Report`, no cortar en el
  primero.
- DON'T: persistir el `GenerationRun` aqui — eso es responsabilidad de un
  caso de uso de storage separado.
- DON'T: validar el contenido de `Context`/`Intent` — son texto libre
  informativo, no datos estructurales del dominio.

## Tests

Los tests estan en `internal/domain/generation_run_test.go` y cubren: run
valido, slide id vacio, model vacio, output vacio, created at invalido y
created at vacio.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta persistir el `GenerationRun` o adjuntarlo
  a un `Project` para cumplir el intent — eso excede el alcance de este
  contrato.
