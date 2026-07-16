---
type: 'Task Contract'
title: 'MCP: tools create_project, list_projects, show_project'
description: 'Expone los casos de uso principales de showme como tools MCP, reusando los mismos comandos que ya usan la CLI y la webapp, para que un agente pueda conectarse por MCP.'
tags: ['showme', 'go', 'mcp', 'agent']

task: mcp-showme-tools
intent: "Exponer create_project, list_projects y show_project como tools MCP sobre los comandos ya existentes de internal/cli."
target: internal/mcpserver/tools.go
signature: "func Tools() []server.ServerTool"
test_command: "go test ./internal/mcpserver"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 8
  max_nesting_depth: 3
  cyclomatic_max: 8
  nesting_max: 3
  params_max: 0
  lines_max: 120
tests: internal/mcpserver/tools_test.go
tests_sha256: "0ff4ee54746d60a285137ea23a6d23585e03f69ec1954413c2fb1411b3d24d3c"
touch_only: ['internal/mcpserver/tools.go']
deps_allowed: ['github.com/mark3labs/mcp-go']
forbids: ['subprocess', 'llm']
---

# Contract: mcp-showme-tools

## Intent

Primer paso del soporte MCP de showme: dejar que un agente conectado por
MCP cree, liste y muestre proyectos, delegando en los mismos
`cli.RunCreateProjectCommand`, `cli.RunListProjectsCommand` y
`cli.RunShowProjectCommand` que ya usan la CLI y la webapp — ningun caso
de uso nuevo, solo un adaptador de protocolo mas, igual que
`internal/web` lo es para HTTP.

Este es el primer contrato del repo con `deps_allowed` no vacio: importa
`github.com/mark3labs/mcp-go`, la primera dependencia externa del
proyecto, necesaria para hablar el protocolo MCP (transporte,
serializacion JSON-RPC, definicion de tools). Ningun otro paquete debe
depender de esta libreria.

## Interface

```go
func Tools() []server.ServerTool
```

Cada `server.ServerTool` define un `mcp.Tool` (nombre, descripcion,
parametros string requeridos) y un `Handler` que:
1. Extrae los argumentos requeridos con `request.RequireString`.
2. Llama al `cli.RunXCommand` correspondiente.
3. Devuelve el resultado JSON-encoded via `mcp.NewToolResultText`, o un
   `mcp.NewToolResultErrorFromErr` si el comando falla.

## Invariants

- `create_project` requiere `name`, `design_path`, `knowledge_root`,
  `deck_path`, `out_dir`; delega en
  `cli.RunCreateProjectCommand(cli.CreateProjectCommandInput{...})`.
- `list_projects` requiere `dir`; delega en
  `cli.RunListProjectsCommand(cli.ListProjectsCommandInput{Dir: dir})`.
- `show_project` requiere `path`; delega en
  `cli.RunShowProjectCommand(cli.ShowProjectCommandInput{Path: path})`.
- Un argumento requerido ausente produce `mcp.NewToolResultError` (tool
  error), sin llamar al comando subyacente.
- Un error del comando subyacente (archivo faltante, JSON invalido, etc.)
  produce un tool error via `mcp.NewToolResultErrorFromErr`, nunca un
  error Go a nivel de protocolo (el segundo valor devuelto por el
  handler es siempre `nil` en estos casos).
- El resultado exitoso se serializa a JSON exactamente como la estructura
  que devuelve el `cli.RunXCommand` correspondiente (mismo shape que
  `--json` en la CLI).
- No hace subprocess ni llama directamente a un proveedor de IA; toda
  logica de negocio vive en `internal/cli`, este archivo solo traduce.

## Examples

- `create_project` con los 5 argumentos validos apuntando a un
  `DESIGN.md`, bundle OKF y deck JSON existentes -> resultado sin error,
  JSON con `OK: true` y un `Path` no vacio.
- `create_project` con `design_path` a un archivo inexistente -> tool
  error (`IsError: true`).
- `list_projects` con `dir` inexistente -> tool error.
- `show_project` con `path` a un archivo inexistente -> tool error.
- `list_projects` sobre un directorio con un proyecto guardado -> JSON
  con un elemento, incluyendo nombre y path del proyecto.

## Do / Don't

- DO: testear via `mcptest.NewServer(t, Tools()...)` y el cliente MCP
  real (`srv.Client().CallTool`) — protocolo real, no llamar los
  handlers directamente.
- DO: mantener el `jsonResult` helper como unico punto de serializacion.
- DON'T: importar `github.com/mark3labs/mcp-go` desde ningun otro
  paquete fuera de `internal/mcpserver` y `cmd/showme-mcp`.
- DON'T: reimplementar logica de creacion/listado/lectura de proyectos
  aqui — siempre delegar a `internal/cli`.

## Tests

Los tests estan en `internal/mcpserver/tools_test.go` y cubren: flujo
completo create->list->show con datos validos, error de validacion en
`create_project` (DESIGN.md faltante), `list_projects` con directorio
inexistente, y `show_project` con archivo inexistente. Usan
`mcptest.NewServer` (cliente y servidor MCP reales en proceso).

## Constraints

- PARAR y reportar si se necesita agregar tools que no correspondan 1:1
  a un `cli.RunXCommand` ya existente — eso excede el alcance de este
  contrato.
- PARAR y reportar si hace falta modificar `internal/cli` para dar
  soporte a MCP; este contrato solo agrega el adaptador.
