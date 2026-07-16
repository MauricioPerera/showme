---
type: 'Task Contract'
title: 'Cliente OpenAI-compatible para generar contenido de slides'
description: 'Implementa ContentGenerator llamando a un endpoint HTTP OpenAI-compatible (chat completions) para generar el contenido de una slide.'
tags: ['showme', 'go', 'ai', 'http', 'openai']

task: openai-content-generator-client
intent: "Generar contenido de una slide llamando al endpoint de chat completions de un servidor OpenAI-compatible."
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
  lines_max: 110
tests: internal/ai/openai_client_test.go
tests_sha256: "ad46b10063a60b96fbf26c3e1e9824957b54d245b18eaa745c418a38c5f9efc3"
touch_only: ['internal/ai/openai_client.go']
deps_allowed: []
forbids: ['subprocess']
---

# Contract: openai-content-generator-client

## Intent

Implementacion concreta del puerto `ContentGenerator` definido en
[generate-slide-content-usecase](./generate-slide-content-usecase.md):
llama al endpoint `/chat/completions` de un servidor OpenAI-compatible
(ej. llama.cpp, LM Studio, vLLM corriendo en localhost) para generar el
contenido de una slide a partir de su intent y contexto. Es, junto con
[mcp-gate-dispatch](./mcp-gate-dispatch.md) y `test-command-gate.md`, uno de
los pocos contratos del repo que declara explicitamente una excepcion a la
convencion `forbids` — aca se dropea `network` y `llm` porque ES,
literalmente, el adaptador que habla con un proveedor de IA por HTTP; el
puerto (`generate-slide-content-usecase`) permanece libre de red porque solo
depende de la interfaz `ContentGenerator`, no de esta implementacion.

## Interface

```go
type OpenAIClient struct { /* no exportado: baseURL, model, httpClient */ }

func NewOpenAIClient(baseURL, model string) *OpenAIClient

func (c *OpenAIClient) GenerateContent(request GenerateContentRequest) (string, error)
```

## Invariants

- Envia un POST a `<baseURL>/chat/completions` con
  `{"model": <model>, "messages": [...], "max_tokens": 512,
  "chat_template_kwargs": {"enable_thinking": false}}`, donde `messages`
  incluye un mensaje `system` fijo y un mensaje `user` con el prompt
  construido a partir de `Intent`/`Context`.
  `chat_template_kwargs.enable_thinking` va en `false` porque el modelo de
  referencia expone razonamiento en un canal separado que no queremos en
  el contenido final de la slide.
- Un status HTTP distinto de 200 es un error
  (`ai provider returned status <codigo>`).
- Una respuesta con `choices` vacio es un error
  (`ai provider returned no choices`).
- Una respuesta que no es JSON valido devuelve el error de decodificacion
  tal cual.
- Con una respuesta valida, devuelve `choices[0].message.content` tal
  cual, sin post-procesarlo.
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
  (`ai provider returned status 500`).
- Servidor que responde `{"choices":[]}` -> `err` no nil
  (`ai provider returned no choices`).
- Servidor que responde un body no-JSON -> `err` no nil (error de
  decodificacion).

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

Los tests estan en `internal/ai/openai_client_test.go` y cubren: request
bien formada (path, model, messages, chat_template_kwargs) contra un
servidor fake, status no-200, respuesta sin choices, y JSON invalido en la
respuesta.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta soportar autenticacion, streaming o
  reintentos para cumplir el intent — eso excede el alcance de este
  contrato.
