# showme

Proyecto basado en [KDD](https://github.com/MauricioPerera/KDD), usando [`MauricioPerera/lazykdd`](https://github.com/MauricioPerera/lazykdd) como infraestructura operativa.

TUI + CLI para operar un repo [KDD](https://github.com/MauricioPerera/KDD) (Knowledge-Driven Development) sin memorizar los `scripts/*.py` ni leer texto plano de gates en la consola — la alternativa liviana a una GUI de escritorio, para quien vive en la terminal. Mismo espíritu que `lazygit`/`lazydocker`.

Este proyecto conserva el flujo KDD completo: contratos ejecutables, gates de validación, CLI JSON y TUI. Las decisiones del proyecto están en [`DEFINITION.md`](DEFINITION.md).

## Qué hace

- **Correr los gates de Nivel 1** de un repo KDD y ver el resultado (11 gates: `validate_contracts`, `validate_specs`, `validate_okf`, `lint_ascii`, `validate_rules`, `validate_skills`, `validate_changelog`, `validate_ux_page`, `validate_diagrams`, `validate_test_commands`, `scan_secrets`).
- **Listar y leer** los contratos de `knowledge/contracts/`.
- **Crear un contrato nuevo** desde la plantilla (`TEMPLATE-task-contract.md`), validando kebab-case y sin sobreescribir nunca uno existente.
- **Ver el estado de ciclo de vida** de cada contrato: `draft` → `validated` → `implemented` → `verified`.

## Arquitectura: un core, tres pieles

```
        scripts/mcp_gate_dispatch.py   (Python)
        lógica pura de despacho de gates
                    |
      +-------------+-------------+
      |             |             |
   MCP server     CLI --json    (el CLI es lo que consume el TUI)
   (14 tools)     (kdd_cli.py)
                      |
                   TUI (Go + Bubble Tea)
                   shellea al CLI, parsea JSON,
                   CERO lógica de KDD propia
```

El TUI nunca reimplementa un check: siempre shellea al CLI Python y parsea su JSON (misma tradición que `lazygit` shelleando a `git`). Detalle completo en [`DEFINITION.md`](DEFINITION.md).

## Uso

### CLI (Python, stdlib puro)

```bash
python scripts/kdd_cli.py gates run-all --json
python scripts/kdd_cli.py contracts list --json
python scripts/kdd_cli.py contracts scaffold mi-tarea-nueva --json
python scripts/kdd_cli.py contracts status --json
```

### TUI (Go)

```bash
cd tui && go build -o showme .
cd ..  # el binario asume cwd = raíz del repo KDD
./tui/showme
```

| Tecla | Acción |
|---|---|
| `g` | panel de gates |
| `c` | panel de contratos (navegable con ↑/↓) |
| `Enter` | ver el contenido completo del contrato seleccionado |
| `Esc` | volver de la vista de detalle |
| `r` | refrescar ambos paneles |
| `n` | crear un contrato nuevo (tipear el nombre, `Enter` confirma, `Esc` cancela) |
| `q` / `Ctrl+C` | salir |

## Estructura del repo

Este repo es una instancia de la [plantilla KDD](https://github.com/MauricioPerera/KDD) — hereda su metodología y su tooling de validación (`scripts/validate_*.py`, `knowledge/contracts/`, `.agents/`). Antes de tocar código: `python scripts/validate_contracts.py knowledge/contracts`.

- `scripts/kdd_cli.py` + `tests/test_kdd_cli.py`: la Piel 2 (CLI).
- `tui/`: la Piel 3 (TUI), módulo Go independiente (`tui/go.mod`). `tui/internal/kdd/`: parseo puro del JSON del CLI. `tui/internal/ui/`: arquitectura Elm (Bubble Tea) — lógica pura (`UpdateModel`/`View`) separada del wiring de I/O.
- `knowledge/contracts/`: los task contracts CCDD que gobiernan cada función, con oráculo congelado y sellado por hash.
- `DEFINITION.md` / `CASE-STUDY-LOG.md`: la definición y el diario del caso de estudio (no son parte de la metodología KDD en sí, son específicos de este proyecto).

## Changelog

<a id="español">

Este README documenta la instancia `showme`; la infraestructura heredada se mantiene para que los gates de KDD sigan siendo reproducibles.
