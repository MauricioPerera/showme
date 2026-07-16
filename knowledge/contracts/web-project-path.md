---
type: 'Task Contract'
title: 'Webapp: resolver un slug de URL a un path de proyecto seguro'
description: 'Valida un slug recibido por URL y lo resuelve al path de archivo del proyecto bajo un directorio, rechazando cualquier intento de path traversal.'
tags: ['showme', 'go', 'web', 'security']

task: web-project-path
intent: "Resolver un slug de proyecto recibido por URL a su path de archivo, rechazando separadores de path y secuencias .."
target: internal/web/project_path.go
signature: "func ProjectFilePath(dir, slug string) (string, error)"
test_command: "go test ./internal/web"
test_cwd: ../..
budget:
  max_cyclomatic_complexity: 4
  max_nesting_depth: 2
  cyclomatic_max: 4
  nesting_max: 2
  params_max: 1
  lines_max: 30
tests: internal/web/project_path_test.go
tests_sha256: "f7fbb27ce9fa035c5d45e043fa4f3aa7d62083019e0056e325e9f6961b8958d2"
touch_only: ['internal/web/project_path.go']
deps_allowed: []
forbids: ['network', 'subprocess', 'llm']
---

# Contract: web-project-path

## Intent

Es el unico punto donde una URL de la webapp (`GET /projects/view/{slug}`)
se convierte en una ruta de filesystem. Sin esta validacion, un slug
adversario (`../../etc/passwd`, `..\\..\\config`) permitiria leer archivos
arbitrarios fuera del directorio de datos configurado — este contrato es la
barrera de path traversal, no una conveniencia.

## Interface

```go
func ProjectFilePath(dir, slug string) (string, error)
```

## Invariants

- `slug` vacio es un error (`slug is required`).
- `slug` que contiene `/` o `\` es un error (`invalid slug: <slug>`).
- `slug` que contiene la secuencia `..` (en cualquier posicion, no solo
  como componente completo) es un error — rechaza tanto `..` como
  `a/../b` como `..secret`.
- Un `slug` valido devuelve `filepath.Join(dir, slug+".json")`, el mismo
  convenio que usa `storage.SaveProject`/`storage.Slugify` al guardar.
- La funcion es determinista, no hace I/O (no verifica que el archivo
  exista), red, subprocess ni llamadas a un proveedor de IA.

## Examples

- `ProjectFilePath("/data", "roadmap-q3")` ->
  `"/data/roadmap-q3.json"`, `err` nil.
- `ProjectFilePath("/data", "")` -> error `slug is required`.
- `ProjectFilePath("/data", "a/b")`,
  `ProjectFilePath("/data", "a\\b")`,
  `ProjectFilePath("/data", "../secret")`,
  `ProjectFilePath("/data", "..")`,
  `ProjectFilePath("/data", "a/../b")` -> todos son error.

## Do / Don't

- DO: rechazar `..` como substring, no solo como componente de path
  completo — `a/../b` ya fue descartado por el chequeo de separadores,
  pero `..secret` (sin separador) tambien debe rechazarse.
- DO: mantener esta funcion como el UNICO lugar que traduce un slug de URL
  a un path de filesystem en toda la webapp.
- DON'T: verificar si el archivo resultante existe — eso es
  responsabilidad de `storage.LoadProject`, llamado despues con el path ya
  validado.
- DON'T: aceptar un path completo del cliente en ningun handler HTTP —
  siempre un slug que pasa por esta funcion primero.

## Tests

Los tests estan en `internal/web/project_path_test.go` y cubren: slug
valido, slug vacio, y slugs con separadores de path o secuencias `..` en
distintas formas.

## Constraints

- PARAR y reportar si se necesita modificar el contrato, los tests o
  archivos fuera de `touch_only`.
- PARAR y reportar si hace falta verificar la existencia del archivo o
  normalizar el path (`filepath.Clean` sobre el resultado) para cumplir
  el intent — eso excede el alcance de este contrato; el chequeo de
  substring ya cubre los casos de ataque conocidos.
