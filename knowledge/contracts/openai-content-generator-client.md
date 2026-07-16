---
type: 'Task Contract'
title: 'Cliente OpenAI-compatible para generar contenido de slides y storyboards'
description: 'Implementa ContentGenerator y StoryboardGenerator llamando a un endpoint HTTP OpenAI-compatible (chat completions) compartido.'
tags: ['showme', 'go', 'ai', 'http', 'openai']

task: openai-content-generator-client
intent: "Generar contenido de una slide o un storyboard llamando al endpoint de chat completions de un servidor OpenAI-compatible."
target: internal/ai/openai_client.go
signature: "func (c *OpenAIClient) GenerateContent(request GenerateContentRequest) (string, error)"
test_command: "go test ./internal/ai"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 8
  max_nesting_depth: 3
  cyclomatic_max: 8
  nesting_max: 3
  params_max: 1
  lines_max: 140
tests: internal/ai/openai_client_test.go
tests_sha256: "111ccc3b9ef1baeff6976930a5085eeee5aa1b9be50f42b67d7ea133333c423c"
touch_only: ['internal/ai/openai_client.go']
deps_allowed: []
forbids: ['subprocess']
---

# Contract: openai-content-generator-client

## Intent

Implementacion concreta de DOS puertos: `ContentGenerator`
([generate-slide-content-usecase](./generate-slide-content-usecase.md)) y
`StoryboardGenerator`
([generate-storyboard-usecase](./generate-storyboard-usecase.md)). Ambos
metodos comparten el mismo endpoint `/chat/completions` de un servidor
OpenAI-compatible (ej. llama.cpp, LM Studio, vLLM corriendo en localhost) y
la misma logica de request/response (`chatCompletion`), solo cambia el
prompt system y el mensaje user. Es, junto con
[mcp-gate-dispatch](./mcp-gate-dispatch.md) y `test-command-gate.md`, uno de
los pocos contratos del repo que declara explicitamente una excepcion a la
convencion `forbids` — aca se dropea `network` y `llm` porque ES,
literalmente, el adaptador que habla con un proveedor de IA por HTTP; ambos
puertos permanecen libres de red porque solo dependen de sus interfaces
(`ContentGenerator`/`StoryboardGenerator`), no de esta implementacion.

## Interface

```go
type OpenAIClient struct { /* no exportado: baseURL, model, httpClient */ }

func NewOpenAIClient(baseURL, model string) *OpenAIClient

func (c *OpenAIClient) GenerateContent(request GenerateContentRequest) (string, error)
func (c *OpenAIClient) GenerateStoryboard(request GenerateStoryboardRequest) (string, error)
```

## Invariants

- Ambos metodos envian un POST a `<baseURL>/chat/completions` con
  `{"model": <model>, "messages": [...], "max_tokens": 512,
  "chat_template_kwargs": {"enable_thinking": false}}` via el helper
  compartido `chatCompletion`. `chat_template_kwargs.enable_thinking` va
  en `false` porque el modelo de referencia expone razonamiento en un
  canal separado que no queremos en la salida final.
- `GenerateContent` arma `messages` con un system prompt sobre escribir
  contenido de slide y un user prompt con `Intent`/`Context`.
  `GenerateStoryboard` arma `messages` con un system prompt que exige
  responder SOLO un array JSON `[{"title", "intent"}, ...]` (sin fences ni
  texto adicional) y un user prompt con `Objective`/`Audience`/`Context`/
  `Count`.
- Un status HTTP distinto de 200 es un error
  (`ai provider returned status <codigo>`), igual en ambos metodos.
- Una respuesta con `choices` vacio es un error
  (`ai provider returned no choices`), igual en ambos metodos.
- Una respuesta que no es JSON valido (a nivel de la respuesta HTTP, no del
  contenido del mensaje) devuelve el error de decodificacion tal cual.
- Con una respuesta valida, ambos devuelven `choices[0].message.content`
  tal cual, sin post-procesarlo — es responsabilidad de
  `generate-storyboard-usecase` parsear ese contenido como JSON de
  slides.
- Usa un `http.Client` con timeout de 120s (los modelos locales pueden
  tardar); no reintenta automaticamente.
- No lee variables de entorno para autenticacion: el servidor de
  referencia (OpenAI-compatible local) no exige API key. Si un proveedor
  futuro la exige, es un cambio a este mismo contrato o uno nuevo, no una
  extension implicita.

## Examples

- Servidor que devuelve
  `{"choices":[{"message":{"content":"Contenido generado."}}]}` -> el
  cliente devuelve `"Contenido generado."`, `err` nil.
- Servidor que responde con status 500 -> `err` no nil
  (`ai provider returned status 500`), en ambos metodos.
- Servidor que responde `{"choices":[]}` -> `err` no nil
  (`ai provider returned no choices`).
- Servidor que responde un body no-JSON -> `err` no nil (error de
  decodificacion).
- `GenerateStoryboard` con `Objective: "Presentar el roadmap"` -> el
  mensaje `user` enviado al servidor contiene ese texto (verificable
  inspeccionando el request en el servidor fake).

## Do / Don't

- DO: testear este contrato con `net/http/httptest.NewServer` (stdlib,
  sin dependencias nuevas) que simula el shape real confirmado contra el
  servidor de referencia — nunca contra un servidor real en la suite
  automatizada, para que CI siga siendo hermetico.
- DO: mantener el prompt system fijo y explicito sobre no inventar mas
  alla del contexto dado (principio de `DEFINITION.md`: "toda afirmacion
  factual debe tener una cita o quedar marcada para revision").
- DON'T: hardcodear el nombre del modelo o el `baseURL` — ambos son
  parametros de `NewOpenAIClient`, decididos por quien lo invoca (CLI/MCP),
  nunca un default silencioso en este archivo.
- DON'T: loguear ni exponer el body completo de la request/response en
  errores — solo el status code o el mensaje de decodificacion.

## Tests

Los tests estan en `internal/ai/openai_client_test.go` y cubren, para
`GenerateContent`: request bien formada (path, model, messages,
chat_template_kwargs) contra un servidor fake, status no-200, respuesta
sin choices, y JSON invalido en la respuesta; para `GenerateStoryboard`:
request bien formada (objetivo presente en el mensaje user, contenido
crudo devuelto tal cual) y status no-200.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta soportar autenticacion, streaming o
  reintentos para cumplir el intent — eso excede el alcance de este
  contrato.
