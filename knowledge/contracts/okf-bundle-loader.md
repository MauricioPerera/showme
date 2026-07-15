---
type: 'Task Contract'
title: 'Lector Go de bundles OKF'
description: 'Carga conceptos Markdown con frontmatter desde un bundle OKF para usarlos como contexto de una diapositiva.'
tags: ['showme', 'go', 'okf', 'knowledge']

task: okf-bundle-loader
intent: "Cargar conceptos OKF desde un directorio de conocimiento."
target: internal/knowledge/bundle.go
signature: "func Load(root string) (Bundle, Report)"
test_command: "go test ./internal/knowledge"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 8
  max_nesting_depth: 3
  cyclomatic_max: 8
  nesting_max: 3
  params_max: 1
  lines_max: 120
tests: internal/knowledge/bundle_test.go
tests_sha256: "7149ec35fc3b6991b8546bb653ae8edf8b5b294d41df3fe94e9ed85da7973e20"
touch_only: ['internal/knowledge/bundle.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: okf-bundle-loader

## Intent

Carga el contexto textual que showme podrá seleccionar para generar una
diapositiva. El lector trata el bundle como un directorio portable de Markdown
con frontmatter, omite `index.md` y conserva tipos OKF desconocidos para no
acoplar el producto a una taxonomía cerrada.

## Interface

```go
type Concept struct {
    ID, Type, Title, Description, Path string
    Tags []string
    Body string
}

type Bundle struct { Concepts []Concept }
type Report struct { Errors, Warnings []string }

func Load(root string) (Bundle, Report)
```

## Invariants

- Solo se consideran archivos `.md` dentro de `root` y sus subdirectorios.
- `index.md` no se carga como concepto.
- Cada concepto debe tener frontmatter y un `type` no vacío.
- El ID es la ruta relativa al bundle sin la extensión `.md`, con separadores
  `/`.
- Los tipos desconocidos se conservan; no son un error.
- La función no realiza red, subprocess ni llamadas a modelos.
- Los conceptos se devuelven ordenados por ID para obtener resultados
  deterministas.

## Examples

- `audience.md` con `type: Audience` -> concepto `audience`.
- `slides/slide-01.md` -> concepto `slides/slide-01`.
- `index.md` -> no aparece en `Bundle.Concepts`.
- Documento sin `type` -> error `<path>: type is required`.

## Do / Don't

- DO: conservar el body Markdown para el ensamblado de contexto posterior.
- DO: aceptar campos de frontmatter adicionales sin perder el concepto.
- DON'T: imponer la lista local de tipos del gate OKF al lector de producto.
- DON'T: resolver enlaces ni consultar recursos remotos en este contrato.

## Tests

Los tests están en `internal/knowledge/bundle_test.go` y cubren carga,
subdirectorios, omisión de índices, tipos desconocidos y errores de `type`.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o archivos
  fuera de `touch_only`.
