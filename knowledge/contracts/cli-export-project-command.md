---
type: 'Task Contract'
title: 'CLI: comando project export'
description: 'Logica pura del comando project export: renderiza un Project guardado como HTML y lo escribe en un archivo.'
tags: ['showme', 'go', 'cli', 'project', 'export', 'html']

task: cli-export-project-command
intent: "Ejecutar el comando 'project export': renderizar un proyecto guardado como HTML y escribirlo en un archivo."
target: internal/cli/export_project_command.go
signature: "func RunExportProjectCommand(input ExportProjectCommandInput) (ExportProjectCommandResult, error)"
test_command: "go test ./internal/cli"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 4
  max_nesting_depth: 2
  cyclomatic_max: 4
  nesting_max: 2
  params_max: 1
  lines_max: 50
tests: internal/cli/export_project_command_test.go
tests_sha256: "eda56885540d9d7b103b4808eb56e0fd965c2384320ebf4a82ed2cab42d151e0"
touch_only: ['internal/cli/export_project_command.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: cli-export-project-command

## Intent

Primer comando de exportacion de la CLI: carga un `Project` guardado con
[load-project-usecase](./load-project-usecase.md), lo renderiza con
[export-project-html](./export-project-html.md) y escribe el resultado en
`OutPath`. Es el punto de entrada por el que un agente obtiene un artefacto
portable de una presentacion.

## Interface

```go
type ExportProjectCommandInput struct {
    Path, OutPath string
}

type ExportProjectCommandResult struct {
    OK   bool
    Path string
}

func RunExportProjectCommand(input ExportProjectCommandInput) (ExportProjectCommandResult, error)
```

## Invariants

- Un error cargando `Path` (archivo inexistente o JSON invalido) se
  propaga tal cual via `err`.
- `export.ExportProjectHTML` no puede fallar (es una funcion pura sobre un
  `Project` ya cargado), asi que un `err` no nil siempre es un problema de
  I/O de `storage.LoadProject` o de `os.WriteFile`, nunca de renderizado.
- El archivo escrito en `OutPath` es exactamente el string devuelto por
  `export.ExportProjectHTML(proj)`.
- La funcion no crea directorios: el directorio de `OutPath` debe existir
  de antemano.
- No hace red, subprocess ni llamadas a un proveedor de IA.

## Examples

- `Path` de un proyecto valido con 2 slides, `OutPath` en un directorio
  existente -> `OK: true`, `Path == OutPath`, y el archivo escrito
  contiene `<!doctype html>` y el contenido de las slides.
- `Path` inexistente -> `err` no nil.
- Directorio de `OutPath` inexistente -> `err` no nil.

## Do / Don't

- DO: reusar `storage.LoadProject` y `export.ExportProjectHTML` tal cual;
  este comando es orquestacion pura (cargar, renderizar, escribir).
- DO: escribir el archivo exacto que devuelve el renderer, sin
  post-procesarlo.
- DON'T: soportar otros formatos (PDF/PPTX) en este comando — son
  contratos de exportador separados que este comando podria invocar mas
  adelante segun la extension de `OutPath`.
- DON'T: imprimir a stdout ni parsear flags aqui — eso vive en
  `cmd/showme/main.go`.

## Tests

Los tests estan en `internal/cli/export_project_command_test.go` y cubren:
export valido con archivo HTML escrito y contenido de slides presente,
archivo origen inexistente, y directorio de salida inexistente.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta soportar otro formato de exportacion para
  cumplir el intent — eso excede el alcance de este contrato.
