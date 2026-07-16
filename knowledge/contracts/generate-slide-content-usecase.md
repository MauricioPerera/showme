---
type: 'Task Contract'
title: 'Caso de uso: generar el contenido de una slide'
description: 'Valida un intent y delega en un ContentGenerator inyectado la generacion del contenido de una slide, sin tocar la red directamente.'
tags: ['showme', 'go', 'ai', 'usecase', 'slide']

task: generate-slide-content-usecase
intent: "Generar el contenido de una slide a partir de su intent y contexto, delegando en un ContentGenerator inyectado."
target: internal/ai/generate_slide_content.go
signature: "func GenerateSlideContent(input GenerateSlideContentInput) (GenerateSlideContentResult, Report)"
test_command: "go test ./internal/ai"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 6
  max_nesting_depth: 2
  cyclomatic_max: 6
  nesting_max: 2
  params_max: 1
  lines_max: 70
tests: internal/ai/generate_slide_content_test.go
tests_sha256: "d29018aee480a0c8385af10af4c03ecd97fb9e1f4d56c9df4b42f0305e458a89"
touch_only: ['internal/ai/generate_slide_content.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: generate-slide-content-usecase

## Intent

Primer caso de uso de IA de `showme` (`DEFINITION.md`: "El proveedor de IA
se integra detras de un puerto para poder probar el core con respuestas
deterministas y cambiar de proveedor sin cambiar los casos de uso"). Define
el puerto `ContentGenerator` y la logica de validacion/orquestacion sobre
el; la implementacion real que llama a un proveedor de IA por HTTP vive en
un contrato separado ([openai-content-generator-client](./openai-content-generator-client.md))
que si declara `network`/`llm`.

## Interface

```go
type GenerateContentRequest struct {
    Intent, Context string
}

type ContentGenerator interface {
    GenerateContent(request GenerateContentRequest) (string, error)
}

type GenerateSlideContentInput struct {
    Generator ContentGenerator
    Intent, Context string
}

type GenerateSlideContentResult struct {
    Content string
}

type Report struct { Errors, Warnings []string }

func GenerateSlideContent(input GenerateSlideContentInput) (GenerateSlideContentResult, Report)
```

## Invariants

- `Intent` no puede estar vacio (`intent is required`); si lo esta, el
  `Generator` NUNCA se invoca.
- Si `Generator.GenerateContent` devuelve un error, su mensaje se agrega a
  `Report.Errors` tal cual (sin envolverlo).
- Si `Generator.GenerateContent` devuelve contenido vacio sin error, es un
  error `generator returned empty content` (una slide nunca queda con
  contenido generado vacio silenciosamente).
- Esta funcion en si misma no hace red, subprocess ni llama a un proveedor
  de IA directamente: solo invoca la interfaz `ContentGenerator` inyectada,
  por eso conserva `forbids: network` a pesar de ser el punto de entrada a
  IA del sistema — quien realmente toca la red es la implementacion
  concreta que el llamador elija inyectar.

## Examples

- `Generator` que devuelve `"Contenido generado."` con `Intent` no vacio
  -> `Report.Errors` vacio, `Result.Content == "Contenido generado."`.
- `Intent: ""` -> error `intent is required`; el generador nunca se llama
  (verificable contando invocaciones en un fake).
- `Generator` que devuelve un error -> ese mismo mensaje aparece en
  `Report.Errors`.
- `Generator` que devuelve `""` sin error -> error
  `generator returned empty content`.

## Do / Don't

- DO: probar este contrato exclusivamente con un `ContentGenerator` fake
  deterministico (ver `fakeGenerator` en los tests) — nunca con un
  proveedor real, siguiendo la separacion puerto/adaptador de
  `DEFINITION.md`.
- DO: mantener `Intent`/`Context` como texto plano, explicito y acotado
  (principio de `DEFINITION.md`: "el contexto enviado al modelo debe ser
  explicito, acotado y auditable por diapositiva").
- DON'T: importar `net/http` ni ningun cliente HTTP en este archivo — eso
  rompe la separacion puerto/adaptador y contradice `forbids: network`.
- DON'T: aplicar la respuesta generada a un `Deck`/`Slide` aqui — esta
  funcion solo genera texto; aplicarlo es responsabilidad de un caso de
  uso separado (ej. combinarlo con `update-slide-usecase`).

## Tests

Los tests estan en `internal/ai/generate_slide_content_test.go` y cubren:
generacion valida con intent/context reenviados al generador, intent
vacio que nunca invoca al generador, error del generador propagado, y
contenido generado vacio tratado como error.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta importar un cliente HTTP o tocar la red
  para cumplir el intent — eso pertenece a un contrato de adaptador
  separado, no a este.
