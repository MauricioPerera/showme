---
type: 'Task Contract'
title: 'Selector Go de contexto de diapositiva'
description: 'Selecciona conceptos relevantes de un bundle OKF mediante puntuacion lexica determinista.'
tags: ['showme', 'go', 'okf', 'context']

task: context-selector
intent: "Seleccionar el contexto relevante de una diapositiva desde un bundle OKF cargado."
target: internal/knowledge/context.go
signature: "func Select(bundle Bundle, query string, limit int) ([]Concept, Report)"
test_command: "go test ./internal/knowledge"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 8
  max_nesting_depth: 3
  cyclomatic_max: 8
  nesting_max: 3
  params_max: 3
  lines_max: 120
tests: internal/knowledge/context_test.go
tests_sha256: "b6b60b1402e7d4b247409ee1b5021ea4c5bbe9d1518bf627b964282c54a62f1a"
touch_only: ['internal/knowledge/context.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: context-selector

## Intent

Selecciona un subconjunto pequeño y auditable de conceptos ya cargados por
[el lector de bundles OKF](./okf-bundle-loader.md) para alimentar una
diapositiva. La selección no modifica el bundle ni resuelve enlaces remotos.

## Interface

```go
func Select(bundle Bundle, query string, limit int) ([]Concept, Report)
```

## Invariants

- La selección es determinista para el mismo bundle, query y límite.
- Solo se devuelven conceptos del bundle original, sin mutarlos.
- Las coincidencias se puntúan en ID, título, descripción, tags y cuerpo; los
  empates se ordenan por ID ascendente.
- Nunca se devuelven más de `limit` conceptos ni se realizan operaciones de
  red, subprocess o llamadas a modelos.
- Query vacío o límite no positivo devuelve cero conceptos y un error.

## Examples

- Query `technical audience` prioriza el concepto cuyo título contiene ambos términos.
- Query `opening review citations` encuentra un concepto por ID, tags y cuerpo.
- Query vacío o límite `0` devuelve un reporte con un error.

## Do / Don't

- DO: normalizar mayúsculas y puntuación para comparar tokens.
- DO: preservar cada `Concept` seleccionado tal como fue cargado.
- DON'T: leer archivos, resolver enlaces o consultar servicios externos.
- DON'T: introducir ranking probabilístico o dependencia de IA.

## Tests

Los tests en `internal/knowledge/context_test.go` congelan ranking, fuentes de
coincidencia, límite y validación de entradas.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, sus tests o cualquier
  archivo fuera de `touch_only` para implementar la función.
- Mantener el selector dentro de `internal/knowledge` y reutilizar `Concept`
  y `Report` existentes.
