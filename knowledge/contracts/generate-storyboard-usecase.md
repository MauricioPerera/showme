---
type: 'Task Contract'
title: 'Caso de uso: generar el storyboard de un Deck'
description: 'Valida objetivo/cantidad y delega en un StoryboardGenerator inyectado la propuesta de slides, parseando su JSON sin intentar repararlo.'
tags: ['showme', 'go', 'ai', 'usecase', 'storyboard']

task: generate-storyboard-usecase
intent: "Generar una lista de slides propuestas a partir del objetivo y audiencia de una presentacion, delegando en un StoryboardGenerator inyectado."
target: internal/ai/generate_storyboard.go
signature: "func GenerateStoryboard(input GenerateStoryboardInput) (GenerateStoryboardResult, Report)"
test_command: "go test ./internal/ai"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 10
  max_nesting_depth: 3
  cyclomatic_max: 10
  nesting_max: 3
  params_max: 1
  lines_max: 100
tests: internal/ai/generate_storyboard_test.go
tests_sha256: "0955644635880c0bce0b1f7c54e4300cc4d01dd6734213821ed65b54b665551a"
touch_only: ['internal/ai/generate_storyboard.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: generate-storyboard-usecase

## Intent

Cubre "showme genera un storyboard inicial" de `DEFINITION.md` ("Usuarios y
flujo principal"), la pieza que
[cli-generate-slide-content-command](./cli-generate-slide-content-command.md)
dejo explicitamente diferida (opera sobre una slide a la vez, no un deck
entero). Define el puerto `StoryboardGenerator` (JSON crudo: un array de
`{"title", "intent"}`) y la logica de validacion/parseo sobre el; la
implementacion real que llama a un proveedor de IA vive en
[openai-content-generator-client](./openai-content-generator-client.md)
(mismo cliente, un metodo nuevo). Decision de formato: JSON estricto,
elegida sobre una lista linea-por-linea porque separa titulo de intent; si
el modelo no devuelve JSON valido (ej. lo envuelve en fences de markdown),
es un error de parseo — esta funcion NO intenta reparar el JSON.

## Interface

```go
type GenerateStoryboardRequest struct {
    Objective, Audience, Context string
    Count int
}

type StoryboardGenerator interface {
    GenerateStoryboard(request GenerateStoryboardRequest) (string, error)
}

type StoryboardSlide struct {
    Title, Intent string
}

type GenerateStoryboardInput struct {
    Generator StoryboardGenerator
    Objective, Audience, Context string
    Count int
}

type GenerateStoryboardResult struct {
    Slides []StoryboardSlide
}

func GenerateStoryboard(input GenerateStoryboardInput) (GenerateStoryboardResult, Report)
```

## Invariants

- `Objective` no puede estar vacio (`objective is required`); `Count` debe
  ser positivo (`count must be positive`). Si cualquiera falla, el
  `Generator` NUNCA se invoca.
- Si `Generator.GenerateStoryboard` devuelve un error, su mensaje se agrega
  a `Report.Errors` tal cual.
- La respuesta cruda se parsea con `encoding/json` como un array de
  `{"title", "intent"}`. Un parseo fallido (incluido texto envuelto en
  markdown, prosa antes/despues del array, etc.) es el error
  `invalid storyboard JSON: <detalle>` — no se intenta extraer ni reparar
  el JSON.
- Un array vacio (`[]`) es el error `generator returned no slides`.
- Cada slide con `title` o `intent` vacio agrega
  `slide[i]: title is required` / `slide[i]: intent is required`
  (indexado, igual que `deck-slide-model`); si hay algun error de este
  tipo, `Result.Slides` queda vacio.
- Esta funcion en si misma no hace red, subprocess ni llama a un proveedor
  de IA directamente: solo invoca la interfaz `StoryboardGenerator`
  inyectada, por eso conserva `forbids: network`.

## Examples

- `Generator` que devuelve
  `[{"title":"Introduccion","intent":"Dar la bienvenida"},{"title":"Plan","intent":"..."}]`
  con `Objective`/`Count` validos -> `Report.Errors` vacio,
  `Result.Slides` con las 2 slides en orden.
- `Objective: ""` -> error `objective is required`; el generador nunca se
  llama.
- `Count: 0` -> error `count must be positive`.
- `Generator` que devuelve un error -> ese mismo mensaje en
  `Report.Errors`.
- `Generator` que devuelve texto envuelto en fences de markdown
  (` ```json\n[...]\n``` `) -> error `invalid storyboard JSON: ...`
  (fallo de parseo, no reparado).
- `Generator` que devuelve `[]` -> error `generator returned no slides`.
- Una slide con `title`/`intent` vacios -> errores
  `slide[0]: title is required` / `slide[0]: intent is required`.

## Do / Don't

- DO: probar este contrato exclusivamente con un `StoryboardGenerator` fake
  deterministico — nunca con un proveedor real, misma separacion
  puerto/adaptador que `generate-slide-content-usecase`.
- DO: exigir JSON estricto y fallar ruidosamente si no lo es — es una
  decision de producto explicita, no un descuido.
- DON'T: intentar extraer un array JSON de un string mas largo (regex,
  strip de fences, etc.) — eso es "reparar" el JSON, prohibido por esta
  decision.
- DON'T: asignar `ID` a las slides ni construir un `domain.Deck` aqui —
  eso es responsabilidad del llamador (CLI), que decide como generar IDs
  unicos a partir de los titulos.

## Tests

Los tests estan en `internal/ai/generate_storyboard_test.go` y cubren:
generacion valida con request reenviado, objetivo vacio que nunca invoca
al generador, count no positivo, error del generador propagado, JSON
invalido (incluido el caso de fences de markdown), lista vacia, y slides
con campos faltantes.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta reparar o tolerar JSON malformado para
  cumplir el intent — esa decision ya se tomo explicitamente en contra.
