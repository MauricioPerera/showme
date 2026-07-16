---
type: 'Task Contract'
title: 'CLI: comando project generate-all'
description: 'Genera el contenido de todas las slides sin contenido de un Project guardado, reusando RunGenerateSlideContentCommand.'
tags: ['showme', 'go', 'cli', 'ai', 'project', 'slide']

task: cli-generate-all-slides-command
intent: "Ejecutar el comando 'project generate-all': generar el contenido de cada slide sin contenido de un proyecto guardado."
target: internal/cli/generate_all_slides_command.go
signature: "func RunGenerateAllSlidesCommand(input GenerateAllSlidesCommandInput) (GenerateAllSlidesCommandResult, error)"
test_command: "go test ./internal/cli"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 8
  max_nesting_depth: 3
  cyclomatic_max: 8
  nesting_max: 3
  params_max: 1
  lines_max: 90
tests: internal/cli/generate_all_slides_command_test.go
tests_sha256: "5729dc9d8fedda3b24ef73e639d6763931dfa2c0468e2f7f2c15e1af082b996f"
touch_only: ['internal/cli/generate_all_slides_command.go']
deps_allowed: []
forbids: ['subprocess']
---

# Contract: cli-generate-all-slides-command

## Intent

Reduce la friccion de tener que llamar
[cli-generate-slide-content-command](./cli-generate-slide-content-command.md)
una vez por slide despues de `project generate-storyboard` +
`project create`: dado un proyecto guardado, genera el contenido de cada
slide cuyo `Content` este vacio, dejando intactas las que ya tienen
contenido (ya generado antes o editado a mano). Es pura composicion: no
reimplementa ninguna logica de generacion, solo itera y reusa
`RunGenerateSlideContentCommand`.

## Interface

```go
type GenerateAllSlidesCommandInput struct {
    Path, BaseURL, Model, OutDir string
}

type GenerateAllSlidesCommandResult struct {
    OK        bool
    Generated []string
    Skipped   []string
    Errors    []string
}

func RunGenerateAllSlidesCommand(input GenerateAllSlidesCommandInput) (GenerateAllSlidesCommandResult, error)
```

## Invariants

- Un error cargando `Path` se propaga tal cual via `err`.
- Cada slide con `Content` no vacio se agrega a `Skipped` y NO se toca —
  generar contenido nunca sobreescribe una slide ya generada o editada a
  mano.
- Cada slide con `Content` vacio se procesa llamando a
  `RunGenerateSlideContentCommand` con el mismo `Path`/`BaseURL`/`Model`/
  `OutDir`; como ese comando recarga el proyecto en cada llamada, cada
  slide se genera sobre el estado mas reciente (incluidas las slides ya
  generadas en llamadas previas de esta misma corrida).
- Si una slide falla al generarse (error de `err` o `Result.Errors` no
  vacio), se agrega a `Errors` como `"<slideID>: <mensaje>"` y el proceso
  CONTINUA con la siguiente slide — una falla no aborta el resto.
- `OK` es `true` unicamente cuando `Errors` queda vacio; tener entradas en
  `Skipped` no afecta `OK`.
- No hace subprocess.

## Examples

- Deck con `intro` (sin contenido) y `plan` (ya con contenido); proveedor
  de IA responde bien -> `Generated: ["intro"]`, `Skipped: ["plan"]`,
  `Errors` vacio, `OK: true`; al recargar, `intro` tiene contenido nuevo y
  `plan` sigue con el suyo.
- Ambas slides ya tienen contenido -> `Generated` vacio, `Skipped` con
  ambas, `OK: true` (nada que hacer no es un error).
- Ambas slides sin contenido, proveedor de IA caido -> `Errors` con una
  entrada por cada slide (`"intro: ..."`, `"plan: ..."`), `OK: false`; se
  intentaron ambas, ninguna fallo aborto a la otra.
- `Path` inexistente -> `err` no nil.

## Do / Don't

- DO: reusar `RunGenerateSlideContentCommand` tal cual para cada slide;
  este comando es orquestacion pura, no reimplementa seleccion de
  contexto ni llamadas al proveedor de IA.
- DO: seguir procesando el resto de las slides aunque una falle.
- DON'T: sobreescribir una slide que ya tiene `Content` — el unico camino
  para regenerar una slide ya generada es
  `cli-generate-slide-content-command` directamente sobre esa slide.
- DON'T: imprimir a stdout ni parsear flags aqui — eso vive en
  `cmd/showme/main.go`.

## Tests

Los tests estan en `internal/cli/generate_all_slides_command_test.go` y
cubren: genera solo las slides vacias y preserva las que ya tienen
contenido, todas las slides ya tienen contenido (nada que hacer), un
fallo del proveedor de IA no aborta el resto de las slides, y archivo
origen inexistente.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta forzar la regeneracion de slides ya
  generadas para cumplir el intent — eso ya lo cubre
  `cli-generate-slide-content-command` sobre una slide especifica.
