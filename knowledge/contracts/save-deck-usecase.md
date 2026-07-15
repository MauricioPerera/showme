---
type: 'Task Contract'
title: 'Caso de uso: guardar un Deck'
description: 'Construye un Deck via el modelo de dominio y lo persiste como JSON en el filesystem, sin escribir nada si es invalido.'
tags: ['showme', 'go', 'usecase', 'storage', 'deck']

task: save-deck-usecase
intent: "Persistir un Deck valido como JSON en un directorio del filesystem."
target: internal/storage/deck_store.go
signature: "func SaveDeck(request SaveDeckRequest) (string, domain.Report, error)"
test_command: "go test ./internal/storage"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 8
  max_nesting_depth: 3
  cyclomatic_max: 8
  nesting_max: 3
  params_max: 1
  lines_max: 140
tests: internal/storage/deck_store_test.go
tests_sha256: "8132f1256131abff7568511da4be7ce48c6f7d3aedb8e3a6aca848f704e9f9ab"
touch_only: ['internal/storage/deck_store.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: save-deck-usecase

## Intent

Primer caso de uso real de `showme`: tomar un `DeckInput`, construirlo con
[`domain.NewDeck`](./deck-slide-model.md) y persistirlo como archivo JSON en
un directorio local. Es el puerto de storage minimo que webapp, CLI y MCP
podran compartir para crear y guardar una presentacion (ver `DEFINITION.md`,
seccion "Arquitectura"). Cargar un Deck de vuelta desde disco queda fuera de
alcance y requiere su propio contrato.

## Interface

```go
type SaveDeckRequest struct {
    Dir   string
    Input domain.DeckInput
}

func SaveDeck(request SaveDeckRequest) (path string, report domain.Report, err error)
```

## Invariants

- Si `domain.NewDeck` produce errores, `SaveDeck` no escribe ningun archivo:
  devuelve `path` vacio y el mismo `Report` con los errores de validacion.
- El nombre de archivo es un slug deterministico del `Title` del deck
  (minusculas, caracteres fuera de `a-z0-9` colapsados a `-`, sin `-` al
  inicio o al final) mas la extension `.json`.
- Si el slug resultante queda vacio (ej. un titulo compuesto solo por
  puntuacion), se agrega el error `title produces an empty slug` a `Report`
  y no se escribe archivo.
- El archivo escrito es JSON valido que deserializa a `domain.Deck`
  preservando `Title`, `Audience` y el orden de `Slides`.
- Un error de I/O (ej. el directorio `Dir` no existe) se devuelve por `err`,
  nunca por `Report` — `Report` es exclusivamente para errores de validacion
  del dominio.
- La funcion no crea directorios: `Dir` debe existir de antemano.
- No hace red, subprocess ni llamadas a un proveedor de IA.

## Examples

- `Title: "Roadmap Q3"` con una slide valida -> escribe `<Dir>/roadmap-q3.json`,
  `Report.Errors` vacio, `err` nil.
- `Title: ""` -> `Report.Errors` incluye `title is required` (via
  `domain.NewDeck`), `path` vacio, no se escribe archivo.
- `Title: "!!!"` con slides validas -> `Report.Errors` incluye
  `title produces an empty slug`, no se escribe archivo.
- `Dir` inexistente con un deck valido -> `err` no nil, `Report.Errors` vacio.

## Do / Don't

- DO: delegar toda la validacion estructural a `domain.NewDeck`; este
  contrato no reimplementa esas reglas.
- DO: mantener la serializacion JSON legible (`MarshalIndent`) para que el
  archivo sea diffable y auditable.
- DON'T: crear el directorio destino ni resolver rutas relativas a un
  proyecto — eso es responsabilidad del llamador.
- DON'T: implementar lectura (`LoadDeck`) en este archivo; es un contrato
  separado.

## Tests

Los tests estan en `internal/storage/deck_store_test.go` y cubren: deck
valido con round-trip JSON y ruta de archivo esperada, deck invalido que no
escribe nada, titulo que produce un slug vacio, y directorio inexistente que
devuelve error de I/O.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta crear directorios, leer un Deck existente o
  tocar `internal/domain/deck.go` para cumplir el intent — eso excede el
  alcance de este contrato.
