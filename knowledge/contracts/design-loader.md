---
type: 'Task Contract'
title: 'Validador Go de DESIGN.md'
description: 'Valida la estructura mínima de una identidad visual antes de usarla para renderizar una presentación.'
tags: ['showme', 'go', 'design', 'validation']

task: design-loader
intent: "Validar la estructura mínima de un documento DESIGN.md."
target: internal/design/design.go
signature: "func Validate(content string) Report"
test_command: "go test ./internal/design"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 8
  max_nesting_depth: 3
  cyclomatic_max: 8
  nesting_max: 3
  params_max: 1
  lines_max: 100
tests: internal/design/design_test.go
tests_sha256: "6f337ddcc1cc1e8ba4e63291b8b54a458d9e642878fb4c66771601128b7e9662"
touch_only: ['internal/design/design.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: design-loader

## Intent

Proporciona una validación determinista y sin red para que la webapp, la CLI y
el servidor MCP puedan rechazar una identidad visual incompleta antes de
usarla. El validador implementa el mínimo necesario para showme y mantiene el
documento completo fuera del dominio de la UI.

## Interface

```go
type Report struct {
    Errors   []string
    Warnings []string
}

func Validate(content string) Report
```

## Invariants

- Un documento válido contiene frontmatter delimitado por `---`.
- El frontmatter válido contiene `name` y `colors.primary`.
- Una sección Markdown `##` no puede aparecer dos veces.
- La función es determinista y no realiza I/O, red, subprocess ni llamadas a
  modelos.
- Los errores se devuelven en el reporte y no se lanzan excepciones.

## Examples

- Frontmatter con `name`, `colors.primary` y secciones únicas -> `Errors` vacío.
- Sin `colors.primary` -> error `colors.primary is required`.
- Dos secciones `## Colors` -> error `duplicate section: Colors`.
- Entrada vacía -> al menos un error de frontmatter.

## Do / Don't

- DO: conservar el orden de aparición de errores.
- DO: tratar advertencias como no bloqueantes.
- DON'T: convertir el validador en un renderer o parser de layout.
- DON'T: depender de npm, red o un proveedor de IA.

## Tests

Los tests están en `internal/design/design_test.go` y cubren documento válido,
campo primario ausente y secciones duplicadas.

## Constraints

- PARAR y reportar si la implementación requiere modificar el contrato,
  tests, `DESIGN.md` o archivos fuera de `touch_only`.
