# showme

A webapp for building AI-assisted presentations. Full product definition,
architecture and Spanish docs below (see [`CHANGELOG.md`](CHANGELOG.md) for
release history) — jump to [la versión en español](#español).

`showme` keeps visual identity and factual context out of the model's hands:
identity lives in a versionable [`DESIGN.md`](DESIGN.md), and knowledge lives
in a citable [OKF](https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md)
bundle. The AI proposes a slide deck; a human reviews, edits and approves it.

<a id="español">

## showme (español)

Webapp para crear presentaciones asistidas por IA. La persona define el
objetivo, aporta contexto verificable y elige una identidad visual; `showme`
propone una estructura de diapositivas, genera el contenido y permite
revisar, editar, regenerar y exportar el resultado.

La IA no decide por sí sola la identidad ni el contexto:

- la **identidad visual** vive en un [`DESIGN.md`](DESIGN.md) versionable
  (tokens de color, tipografía, espaciado y componentes);
- el **conocimiento** disponible para redactar cada diapositiva vive en un
  bundle [OKF](https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md),
  con citas obligatorias por afirmación.

La definición completa de producto y arquitectura está en
[`DEFINITION.md`](DEFINITION.md). El historial de cambios está en
[`CHANGELOG.md`](CHANGELOG.md).

### Principios

- La generación es una propuesta; la aprobación final es humana.
- Toda afirmación factual debe tener una cita o quedar marcada para revisión.
- La identidad visual se aplica por tokens, nunca por valores sueltos
  inventados en el prompt.
- El mismo flujo funciona desde navegador, línea de comandos y agentes de IA.

### Arquitectura

Un core de dominio en Go (decks, slides, validadores, trazabilidad) con tres
adaptadores delgados que comparten los mismos casos de uso:

```text
                         +------------------+
                         |  AI provider(s)  |
                         +--------+---------+
                                  |
 +-------------+          +------v-------+          +----------------+
 | Webapp Go   |---------->|              |<----------| CLI Go         |
 | HTML + htmx |           |  showme core |           | human/JSON     |
 +-------------+          |              |           +----------------+
                          |  use cases   |
 +-------------+          |  domain      |           +----------------+
 | MCP server  |---------->|  ports       |<----------| Storage        |
 | Go          |           +------+-------+           +----------------+
 +-------------+                  |
                         +--------v---------+
                         | DESIGN.md + OKF |
                         | validation      |
                         +------------------+
```

- **Webapp**: server-rendered con Go + HTML semántico, `htmx` para
  actualizaciones parciales.
- **CLI**: mismo core, salida humana y JSON estable para automatización.
- **Servidor MCP**: expone los mismos casos de uso a agentes de IA.

Ningún adaptador reimplementa reglas de generación, validación o permisos —
todo vive en `internal/`.

### Estado

En construcción. El core Go (`internal/design`, `internal/knowledge`) ya
implementa el parser/validador de `DESIGN.md` y el loader/selector de
contexto sobre bundles OKF. Webapp, CLI y servidor MCP están definidos en
`DEFINITION.md` pero pendientes de implementación.

### Metodología

Este repo se desarrolla con [KDD](https://github.com/MauricioPerera/KDD)
(Knowledge-Driven Development): cada capacidad nace como un contrato en
[`knowledge/contracts/`](knowledge/contracts/) con test-oráculo congelado,
implementación y verificación antes de mezclarse. Antes de tocar código:

```bash
python scripts/validate_contracts.py knowledge/contracts
```

La infraestructura de gates y contratos (`scripts/`, `.agents/`) se heredó de
[`lazykdd`](https://github.com/MauricioPerera/lazykdd) y sigue gobernando el
desarrollo; el producto en sí se construye en Go bajo `internal/`.
