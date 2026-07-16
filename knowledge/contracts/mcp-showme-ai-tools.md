---
type: 'Task Contract'
title: 'MCP: tools de generacion con IA'
description: 'Expone generate_slide_content, generate_storyboard y generate_all_slides como tools MCP sobre los comandos ya existentes de internal/cli, con base_url/model como parametros por llamada.'
tags: ['showme', 'go', 'mcp', 'agent', 'ai']

task: mcp-showme-ai-tools
intent: "Exponer generate_slide_content, generate_storyboard y generate_all_slides como tools MCP."
target: internal/mcpserver/ai_tools.go
signature: "func AITools() []server.ServerTool"
test_command: "go test ./internal/mcpserver"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 10
  max_nesting_depth: 3
  cyclomatic_max: 10
  nesting_max: 3
  params_max: 0
  lines_max: 170
tests: internal/mcpserver/extra_tools_test.go
tests_sha256: "429350138c250ef640a8021c14cd5b20ffcd99cfd26d04700b9a5c0ac805fe37"
touch_only: ['internal/mcpserver/ai_tools.go']
deps_allowed: ['github.com/mark3labs/mcp-go']
forbids: ['subprocess']
---

# Contract: mcp-showme-ai-tools

## Intent

Cuarto grupo de tools del servidor MCP de showme: dejar que un agente
dispare generacion con IA sobre un proyecto ya guardado (contenido de
una slide, un storyboard completo, o todas las slides pendientes),
delegando exactamente en los mismos comandos que ya usan la CLI y la
webapp (`cli.RunGenerateSlideContentCommand`,
`cli.RunGenerateStoryboardCommand`, `cli.RunGenerateAllSlidesCommand`).
Igual que en la webapp, `base_url`/`model` son siempre parametros de la
llamada, nunca configuracion del servidor MCP.

## Interface

```go
func AITools() []server.ServerTool
```

Se agrega al resultado de [Tools()](./mcp-showme-tools.md), no lo
reemplaza. Este es el unico de los cuatro grupos que hace red (a traves
del `OpenAIClient` que construye cada comando delegado); por eso
`forbids` no incluye `llm` a diferencia de los otros tres contratos de
`internal/mcpserver`.

## Invariants

- `generate_slide_content` requiere `path`, `slide_id`, `base_url`,
  `model`, `out_dir`. Delega en `cli.RunGenerateSlideContentCommand`.
- `generate_storyboard` requiere `objective`, `base_url`, `model`,
  `deck_title`, `count` (number, via `mcp.WithNumber`/`RequireInt`),
  `out_path`; `audience`, `knowledge_root` opcionales. Delega en
  `cli.RunGenerateStoryboardCommand`.
- `generate_all_slides` requiere `path`, `base_url`, `model`, `out_dir`.
  Delega en `cli.RunGenerateAllSlidesCommand`.
- Un argumento requerido ausente produce `mcp.NewToolResultError` sin
  llamar al comando subyacente.
- Un error de archivo/E-S del comando subyacente produce un tool error
  via `mcp.NewToolResultErrorFromErr`. Errores del proveedor de IA o de
  validacion (objetivo vacio, count no positivo, JSON invalido del
  proveedor, slide no encontrada) se devuelven como resultado exitoso
  con `OK: false` y `Errors` poblado en el JSON -- mismo criterio que
  [mcp-showme-tools](./mcp-showme-tools.md).

## Examples

- `generate_slide_content` sobre una slide con `Intent` no vacio y un
  proveedor de IA que responde 200 -> `OK: true`, `Content` no vacio.
- `generate_slide_content` con un proveedor que responde 500 -> `OK:
  false`, `Errors` no vacio, `IsError: false`.
- `generate_storyboard` con `count: 2` y un proveedor que responde un
  array JSON de 2 slides -> `OK: true`, el archivo en `out_path` existe.
- `generate_all_slides` sobre un proyecto con una slide ya generada y
  otra sin contenido -> `Skipped` incluye la primera, `Generated` la
  segunda.

## Do / Don't

- DO: testear con `net/http/httptest.NewServer` simulando el proveedor
  de IA — nunca un servidor real en la suite automatizada.
- DO: testear via `mcptest.NewServer(t, Tools()...)` — protocolo MCP
  real, no llamar los handlers directamente.
- DON'T: aceptar `base_url`/`model` como configuracion del servidor MCP
  (flags de arranque, variables de entorno) — siempre parametros de la
  llamada, igual que en la webapp.
- DON'T: reimplementar logica de generacion aqui — siempre delegar a
  `internal/cli`/`internal/ai`.

## Tests

Los tests estan en `internal/mcpserver/extra_tools_test.go` (compartido
con [mcp-showme-slide-tools](./mcp-showme-slide-tools.md) y
[mcp-showme-project-management-tools](./mcp-showme-project-management-tools.md))
y cubren: `generate_slide_content` exitoso y con error del proveedor,
`generate_all_slides` distinguiendo generado/saltado, y
`generate_storyboard` exitoso escribiendo el deck JSON.

## Constraints

- PARAR y reportar si se necesita agregar tools que no correspondan 1:1
  a un `cli.RunXCommand` ya existente.
- PARAR y reportar si hace falta modificar `internal/cli`/`internal/ai`
  para dar soporte a estas tools; este contrato solo agrega el adaptador.
