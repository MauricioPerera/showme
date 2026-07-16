---
type: 'Task Contract'
title: 'MCP: tools de gestion de proyecto completo'
description: 'Expone archive_project, duplicate_project, rename_project, review_project y export_project como tools MCP sobre los comandos ya existentes de internal/cli.'
tags: ['showme', 'go', 'mcp', 'agent']

task: mcp-showme-project-management-tools
intent: "Exponer archive_project, duplicate_project, rename_project, review_project y export_project como tools MCP."
target: internal/mcpserver/project_tools.go
signature: "func ProjectManagementTools() []server.ServerTool"
test_command: "go test ./internal/mcpserver"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 10
  max_nesting_depth: 3
  cyclomatic_max: 10
  nesting_max: 3
  params_max: 0
  lines_max: 200
tests: internal/mcpserver/extra_tools_test.go
tests_sha256: "429350138c250ef640a8021c14cd5b20ffcd99cfd26d04700b9a5c0ac805fe37"
touch_only: ['internal/mcpserver/project_tools.go']
deps_allowed: ['github.com/mark3labs/mcp-go']
forbids: ['subprocess', 'llm']
---

# Contract: mcp-showme-project-management-tools

## Intent

Tercer grupo de tools del servidor MCP de showme: dejar que un agente
gestione un proyecto ya guardado como unidad completa (archivar,
duplicar, renombrar, aplicar una decision de review sobre una slide, y
exportar a HTML), delegando exactamente en los mismos comandos que ya
usan la CLI y la webapp (`cli.RunArchiveProjectCommand`,
`cli.RunDuplicateProjectCommand`, `cli.RunRenameProjectCommand`,
`cli.RunReviewProjectCommand`, `cli.RunExportProjectCommand`).

## Interface

```go
func ProjectManagementTools() []server.ServerTool
```

Se agrega al resultado de [Tools()](./mcp-showme-tools.md), no lo
reemplaza.

## Invariants

- `archive_project` requiere `path`, `archived` (boolean, via
  `mcp.WithBoolean`/`request.RequireBool`), `out_dir`. Delega en
  `cli.RunArchiveProjectCommand`.
- `duplicate_project` requiere `source_path`, `new_name`, `out_dir`.
  Delega en `cli.RunDuplicateProjectCommand`.
- `rename_project` requiere `source_path`, `new_name`, `out_dir`. Delega
  en `cli.RunRenameProjectCommand`.
- `review_project` requiere `path`, `slide_id`, `decision`, `out_dir`;
  `notes` opcional. Delega en `cli.RunReviewProjectCommand`.
- `export_project` requiere `path`, `out_path`. Delega en
  `cli.RunExportProjectCommand`.
- Un argumento requerido ausente produce `mcp.NewToolResultError` sin
  llamar al comando subyacente.
- Un error de archivo/E-S del comando subyacente produce un tool error
  via `mcp.NewToolResultErrorFromErr`. Errores de validacion del dominio
  (decision de review invalida, nombre duplicado, etc.) se devuelven
  como resultado exitoso con `OK: false` y `Errors` poblado en el JSON --
  mismo criterio que [mcp-showme-tools](./mcp-showme-tools.md).

## Examples

- `archive_project` con `archived: true` sobre un proyecto valido ->
  `OK: true`.
- `review_project` con `decision: "not-a-real-decision"` -> `OK: false`,
  `Errors` no vacio, `IsError: false`.
- `export_project` sobre un proyecto valido -> `OK: true` y el archivo en
  `out_path` existe en disco.
- `duplicate_project` sin `new_name` -> tool error (`IsError: true`), el
  comando no se ejecuta.

## Do / Don't

- DO: testear via `mcptest.NewServer(t, Tools()...)` — protocolo MCP
  real, no llamar los handlers directamente.
- DON'T: reimplementar logica de archivado/duplicado/renombrado/review/
  exportado aqui — siempre delegar a `internal/cli`.

## Tests

Los tests estan en `internal/mcpserver/extra_tools_test.go` (compartido
con [mcp-showme-slide-tools](./mcp-showme-slide-tools.md) y
[mcp-showme-ai-tools](./mcp-showme-ai-tools.md)) y cubren: flujo
review/archive/export/duplicate/rename sobre un proyecto creado en el
propio test, y un `review_project` con decision invalida.

## Constraints

- PARAR y reportar si se necesita agregar tools que no correspondan 1:1
  a un `cli.RunXCommand` ya existente.
- PARAR y reportar si hace falta modificar `internal/cli` para dar
  soporte a estas tools; este contrato solo agrega el adaptador.
