---
type: 'Task Contract'
title: 'CLI: comando project generate-slide'
description: 'Selecciona contexto OKF para una slide, genera su contenido con un proveedor de IA y lo aplica al Project guardado.'
tags: ['showme', 'go', 'cli', 'ai', 'project', 'slide']

task: cli-generate-slide-content-command
intent: "Ejecutar el comando 'project generate-slide': generar el contenido de una slide con IA y guardarlo en el proyecto."
target: internal/cli/generate_slide_content_command.go
signature: "func RunGenerateSlideContentCommand(input GenerateSlideContentCommandInput) (GenerateSlideContentCommandResult, error)"
test_command: "go test ./internal/cli"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 12
  max_nesting_depth: 3
  cyclomatic_max: 12
  nesting_max: 3
  params_max: 1
  lines_max: 180
tests: internal/cli/generate_slide_content_command_test.go
tests_sha256: "b955221173c25b208eb19e0ba47fdd6b14c71cea1d838c57850294ee7451c37c"
touch_only: ['internal/cli/generate_slide_content_command.go']
deps_allowed: []
forbids: ['subprocess']
---

# Contract: cli-generate-slide-content-command

## Intent

Primer comando de IA de la CLI: ata
[context-selector](./context-selector.md) (`knowledge.Select`),
[generate-slide-content-usecase](./generate-slide-content-usecase.md) (con
[openai-content-generator-client](./openai-content-generator-client.md)
como `ContentGenerator` real),
[update-slide-usecase](./update-slide-usecase.md) y
[append-generation-run-usecase](./append-generation-run-usecase.md) en un
solo comando: dado un proyecto guardado y el `ID` de una de sus slides,
selecciona contexto OKF relevante, genera el contenido con un proveedor de
IA, lo aplica a esa slide preservando su `Status` actual, y deja un
`GenerationRun` auditable en el historial del proyecto. Hereda la excepcion
a `forbids: network/llm` de sus dependencias de IA (declara solo `forbids:
subprocess`, igual que ellas).

## Interface

```go
type GenerateSlideContentCommandInput struct {
    Path, SlideID, BaseURL, Model, OutDir string
}

type GenerateSlideContentCommandResult struct {
    OK      bool
    Path    string
    Content string
    Errors, Warnings []string
}

func RunGenerateSlideContentCommand(input GenerateSlideContentCommandInput) (GenerateSlideContentCommandResult, error)
```

## Invariants

- Un error cargando `Path` se propaga tal cual via `err`; no se genera ni
  se selecciona contexto.
- Si `SlideID` no corresponde a ninguna slide del `Deck`, el resultado
  tiene `OK: false` y `Errors` incluye `slide not found: <id>` — no se
  carga el bundle OKF ni se llama al proveedor de IA.
- El contexto se selecciona con `knowledge.Select` sobre el bundle en
  `proj.KnowledgePath`, usando como query el `Intent` de la slide o, si
  esta vacio, su `Title`. Los `Body` de los conceptos seleccionados
  (hasta 3) se concatenan separados por `\n\n---\n\n`.
- Cualquier error de `knowledge.Load`, `knowledge.Select` o
  `ai.GenerateSlideContent` se agrega a `Errors`; si hay al menos un error
  en cualquiera de esos tres pasos, el proceso se corta ahi: no se llama
  al paso siguiente y el proyecto en disco NO se toca.
- Si la generacion es exitosa, la slide se actualiza con
  `domain.UpdateSlide` preservando su `Status` actual (se pasa
  explicitamente, no vacio) — generar contenido no cambia el estado de
  revision de la slide.
- Se construye un `domain.GenerationRun` (con `SlideID`, `Model`,
  `Provider: BaseURL`, `Intent`, `Context` seleccionado, `Output`,
  `Warnings` acumulados y `CreatedAt` con la hora real —
  `time.Now().UTC().Format(time.RFC3339)`, el UNICO uso de tiempo real en
  todo el comando; el resto de la logica sigue siendo determinista) y se
  agrega al historial del proyecto con `domain.AppendGenerationRun` antes
  de guardar.
- Si todo es valido, el `Project` actualizado (mismo `Name`, `DesignPath`,
  `KnowledgePath`, `Version`, `Archived`; `Deck` con el `Content` nuevo;
  `Runs` con el `GenerationRun` nuevo agregado) se guarda con
  `storage.SaveProject` bajo `OutDir`.
- Un error de I/O al guardar se propaga via `err`.
- No hace subprocess.

## Examples

- Proyecto con slide `intro` (`Intent: "Dar la bienvenida..."`,
  `Status: accepted`) y un bundle OKF con un concepto relacionado; server
  de IA responde `"Contenido generado."` -> `OK: true`,
  `Content == "Contenido generado."`; al recargar el proyecto, `intro`
  tiene ese `Content` y sigue `Status: accepted`.
- `SlideID: "missing"` -> `OK: false`, `Errors` incluye
  `slide not found: missing`.
- El servidor de IA responde status 500 -> `OK: false`, `Errors` incluye
  el error de `openai-content-generator-client`; la slide y el archivo
  quedan sin cambios.
- `Path` inexistente -> `err` no nil.

## Do / Don't

- DO: testear con `net/http/httptest.NewServer` simulando el servidor de
  IA — nunca contra un proveedor real en la suite automatizada.
- DO: preservar el `Status` de la slide; generar contenido no es lo mismo
  que revisarlo (`apply-review-usecase` sigue siendo el unico camino para
  cambiar `Status`).
- DON'T: reintentar automaticamente si el proveedor de IA falla — un
  fallo se reporta y el llamador decide si reintentar.
- DON'T: imprimir a stdout ni parsear flags aqui — eso vive en
  `cmd/showme/main.go`.

## Tests

Los tests estan en `internal/cli/generate_slide_content_command_test.go` y
cubren: generacion valida con contenido aplicado y status preservado,
registro del `GenerationRun` con sus datos correctos, slide no encontrada,
error del proveedor de IA que deja el proyecto sin cambios, y archivo
origen inexistente.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta generar un storyboard completo (multiples
  slides) para cumplir el intent — eso excede el alcance de este
  contrato, que opera sobre una slide a la vez.
