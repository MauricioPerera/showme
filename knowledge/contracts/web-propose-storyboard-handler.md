---
type: 'Task Contract'
title: 'Webapp: handler de proponer un storyboard con IA'
description: 'Selecciona contexto OKF opcional y genera una propuesta de storyboard con IA, sin escribir nada a disco, para revision previa a crear el proyecto.'
tags: ['showme', 'go', 'web', 'ai', 'storyboard']

task: web-propose-storyboard-handler
intent: "Generar una propuesta de storyboard con IA para revisar antes de crear el proyecto."
target: internal/web/propose_storyboard.go
signature: "func HandleProposeStoryboard(input ProposeStoryboardInput) (ProposeStoryboardResult, error)"
test_command: "go test ./internal/web"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 8
  max_nesting_depth: 3
  cyclomatic_max: 8
  nesting_max: 3
  params_max: 1
  lines_max: 90
tests: internal/web/propose_storyboard_test.go
tests_sha256: "5a6f4ecc0dbb9c3a0792ee6f5f57998158bbafe2e18ac6b4e953b8daf6f12c80"
touch_only: ['internal/web/propose_storyboard.go']
deps_allowed: []
forbids: ['subprocess']
---

# Contract: web-propose-storyboard-handler

## Intent

Primer paso del flujo "crear proyecto con storyboard generado" de la
webapp: dado un objetivo, audiencia y contexto OKF opcional, propone una
lista de slides con IA — igual logica de seleccion de contexto y
generacion que
[cli-generate-storyboard-command](./cli-generate-storyboard-command.md),
pero SIN escribir ningun archivo. El resultado se renderiza como un
formulario editable; solo cuando la persona confirma se llama a
[web-create-project-with-slides-handler](./web-create-project-with-slides-handler.md)
para persistir.

## Interface

```go
type ProposeStoryboardInput struct {
    Objective, Audience, KnowledgeRoot, BaseURL, Model string
    Count int
}

type ProposeStoryboardResult struct {
    Slides   []ai.StoryboardSlide
    Errors, Warnings []string
}

func HandleProposeStoryboard(input ProposeStoryboardInput) (ProposeStoryboardResult, error)
```

## Invariants

- Si `KnowledgeRoot` es no vacio, se carga con `knowledge.Load` y se
  seleccionan hasta 3 conceptos relevantes al `Objective` con
  `knowledge.Select`; sus `Body` se concatenan como contexto. Un
  `KnowledgeRoot` inexistente no es un error (`knowledge.Load` tolera un
  directorio ausente devolviendo un bundle vacio).
- Si `KnowledgeRoot` es vacio, no se selecciona contexto (`Context: ""`).
- Cualquier error de `knowledge.Load`, `knowledge.Select` o
  `ai.GenerateStoryboard` (objetivo vacio, count no positivo, error del
  proveedor, JSON invalido) se agrega a `Errors`; `Slides` queda vacio en
  ese caso.
- No escribe ningun archivo ni crea ningun `Project` — es exclusivamente
  el paso de generacion/propuesta.
- No hace red directamente mas alla de invocar `ai.GenerateStoryboard`
  (que si la hace via el `ContentGenerator`/`StoryboardGenerator`
  inyectado); no hace subprocess.

## Examples

- `Objective: "Presentar el roadmap"`, `Count: 2`, proveedor de IA que
  responde `[{"title":"Introduccion","intent":"..."},{"title":"Plan","intent":"..."}]`
  -> `Errors` vacio, `Slides` con las 2 propuestas en orden.
- `KnowledgeRoot` con un concepto relacionado al objetivo -> el mensaje
  `user` enviado al proveedor de IA incluye ese contexto (verificable
  contra un servidor fake).
- El proveedor de IA responde status 500 -> `Errors` no vacio, `Slides`
  vacio.
- `KnowledgeRoot` apuntando a un directorio inexistente -> `err` nil (se
  tolera), `Slides` generado igual sin contexto adicional.

## Do / Don't

- DO: testear con `net/http/httptest.NewServer` simulando el proveedor de
  IA — nunca un servidor real en la suite automatizada.
- DO: mantener esta funcion de solo-generacion; la persistencia es
  responsabilidad exclusiva de
  `web-create-project-with-slides-handler`.
- DON'T: asignar `ID` a las slides propuestas aqui — eso ocurre recien en
  `web-create-project-with-slides-handler`, cuando la persona confirma.
- DON'T: reintentar automaticamente si el proveedor de IA falla.

## Tests

Los tests estan en `internal/web/propose_storyboard_test.go` y cubren:
propuesta valida sin contexto OKF, propuesta con contexto OKF
(verificando que el mensaje enviado lo incluye), error del proveedor de
IA, y un `KnowledgeRoot` inexistente tolerado sin error.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta persistir el proyecto o asignar ids a
  las slides para cumplir el intent — eso excede el alcance de este
  contrato.
