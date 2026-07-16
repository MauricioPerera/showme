---
type: 'Task Contract'
title: 'Exportador HTML de un Project'
description: 'Renderiza un Project como un documento HTML autocontenido, una seccion por slide, con todo el texto escapado.'
tags: ['showme', 'go', 'export', 'html']

task: export-project-html
intent: "Renderizar un Project como un documento HTML autocontenido."
target: internal/export/html.go
signature: "func ExportProjectHTML(proj domain.Project) string"
test_command: "go test ./internal/export"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 6
  max_nesting_depth: 2
  cyclomatic_max: 6
  nesting_max: 2
  params_max: 1
  lines_max: 60
tests: internal/export/html_test.go
tests_sha256: "93bf70f8641de032810e5d87ef6dbd40aab9cfb077b16591c5ab99ae84a94d98"
touch_only: ['internal/export/html.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: export-project-html

## Intent

Primer formato de exportacion de `showme` (`DEFINITION.md`, "Entregables
iniciales": "Render de slide y exportador inicial definidos por contrato";
"Decisiones que quedan abiertas": "estrategia de renderizado
PDF/PPTX/HTML"). Se elige HTML como primer formato por ser portable y no
requerir dependencias externas (a diferencia de PDF/PPTX, que si las
pedirian). Este contrato cubre solo el render puro a `string`; escribirlo a
disco es responsabilidad del llamador (CLI), no de esta funcion.

## Interface

```go
func ExportProjectHTML(proj domain.Project) string
```

## Invariants

- El documento empieza con `<!doctype html>`.
- Contiene un `<title>` con `proj.Deck.Title`.
- Hay exactamente un `<section>` por cada elemento de `proj.Deck.Slides`,
  en el mismo orden.
- Todo texto proveniente del `Project` (titulo del deck, audiencia,
  titulo/contenido de cada slide) se escapa con `html.EscapeString` antes
  de insertarse — ningun campo puede inyectar markup ni salirse de su
  elemento.
- El `id` de cada `<section>` es el `ID` de la slide (tambien escapado); su
  atributo `data-status` refleja `Slide.Status`.
- La funcion es pura y deterministica: la misma entrada produce siempre el
  mismo string. No hace I/O, red, subprocess ni llamadas a un proveedor de
  IA.

## Examples

- Deck con `Title: "Roadmap Q3"` y slides `intro`/`plan` -> el HTML
  resultante contiene `<title>Roadmap Q3</title>` y dos `<section>`, en
  orden `intro` antes que `plan`.
- Slide con `Title: "<script>alert(1)</script>"` -> el HTML resultante
  contiene `&lt;script&gt;alert(1)&lt;/script&gt;`, nunca la etiqueta sin
  escapar.
- Slide con `Content: "Tom & Jerry"` -> aparece como `Tom &amp; Jerry`.
- Misma llamada dos veces con el mismo `Project` -> mismo string byte a
  byte.

## Do / Don't

- DO: escapar TODO texto proveniente del usuario/proyecto con
  `html.EscapeString` antes de concatenarlo al documento.
- DO: mantener el HTML minimo y semantico (un `<section>` por slide, sin
  CSS embebido todavia).
- DON'T: aplicar los tokens de `DESIGN.md` (colores, tipografia) en esta
  primera version — `internal/design` hoy solo valida estructura, no
  expone los valores parseados; aplicar la identidad visual real es un
  contrato futuro que depende de eso.
- DON'T: escribir a disco ni aceptar una ruta de archivo — eso es
  responsabilidad de quien llama a esta funcion.

## Tests

Los tests estan en `internal/export/html_test.go` y cubren: doctype,
titulo y secciones presentes, orden de slides preservado, escape de
contenido con caracteres HTML especiales, y salida deterministica.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta aplicar tokens de `DESIGN.md` o generar
  PDF/PPTX para cumplir el intent — eso excede el alcance de este
  contrato.
