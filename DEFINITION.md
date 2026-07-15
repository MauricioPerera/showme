# showme — Definición del proyecto

Este documento define el producto que se construirá con KDD. Es una
definición de producto y arquitectura, no una implementación: cada capacidad
se convertirá después en un contrato KDD independiente, con tests congelados,
implementación y verificación.

## Qué es

`showme` es una webapp para crear presentaciones asistidas por IA. La persona
define el objetivo de la presentación, aporta contexto verificable y elige una
identidad visual; showme propone una estructura de diapositivas, genera el
contenido y permite revisar, editar, regenerar y exportar el resultado.

La IA no decide por sí sola la identidad ni el contexto. La identidad visual
vive en un archivo `DESIGN.md` y el conocimiento que puede utilizarse para
redactar cada diapositiva vive en un bundle OKF. Ambos son artefactos legibles,
versionables y reutilizables fuera de showme.

## Objetivos

- Reducir el tiempo entre una idea y un primer borrador coherente de una
  presentación.
- Mantener una separación visible entre contexto aportado, contenido generado
  y decisiones aceptadas por la persona.
- Aplicar una identidad visual consistente a toda la presentación y a cada
  diapositiva.
- Permitir el mismo flujo desde navegador, línea de comandos y agentes de IA.
- Generar resultados reproducibles a partir de los mismos inputs, modelo y
  configuración registrada.

## Usuarios y flujo principal

El usuario crea un proyecto de presentación y proporciona:

- propósito, audiencia, idioma, duración y número objetivo de diapositivas;
- un bundle de conocimiento OKF o una selección de conceptos OKF;
- un `DESIGN.md` con la identidad visual de la marca;
- restricciones opcionales, como tono, claims prohibidos o fuentes
  obligatorias.

showme genera un storyboard inicial. Cada diapositiva se puede inspeccionar
como una unidad con título, objetivo, contenido, layout, elementos visuales,
fuentes y estado de revisión. El usuario puede aceptar, editar o regenerar una
diapositiva sin perder el contexto ni la versión anterior.

## Fuentes normativas

### Identidad visual: `DESIGN.md`

La identidad de la marca se almacenará en `DESIGN.md`, siguiendo el formato de
[`google-labs-code/design.md`](https://github.com/google-labs-code/design.md):

- frontmatter YAML para tokens de colores, tipografía, espaciado, formas y
  componentes;
- cuerpo Markdown para la intención de marca, jerarquía visual y reglas de
  uso;
- referencias entre tokens con la forma `{path.to.token}`;
- secciones ordenadas como Overview, Colors, Typography, Layout, Elevation &
  Depth, Shapes, Components y Do's and Don'ts cuando apliquen.

Los tokens son la fuente normativa para renderizar slides. El texto narrativo
explica la intención y sirve de contexto para los agentes. El sistema debe
validar referencias rotas y contraste antes de usar una identidad en una
presentación; el formato está en versión alpha, por lo que el parser se
mantendrá aislado y tolerante a extensiones compatibles.

### Contexto de conocimiento: OKF

El contexto se almacenará como un knowledge bundle conforme a
[`OKF/SPEC.md`](https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md):

- documentos Markdown UTF-8 con frontmatter YAML;
- `type` no vacío como campo obligatorio;
- `title`, `description`, `tags`, `timestamp` y `resource` cuando sean
  relevantes;
- `index.md` para descubrimiento progresivo y `log.md` para historial cuando
  se necesiten;
- enlaces Markdown entre conceptos y una sección `# Citations` para respaldar
  afirmaciones.

Para el perfil de showme, cada fuente usada por una diapositiva debe quedar
registrada en sus metadatos o en su sección de citas. El perfil puede definir
tipos descriptivos como `Deck`, `Slide`, `Claim`, `Metric`, `Reference` o
`Audience`, pero los consumidores deben tolerar tipos OKF desconocidos.

Una diapositiva no debe inventar hechos que no estén en el contexto autorizado
o marcarlos explícitamente como propuesta, hipótesis o contenido pendiente de
verificación.

## Producto y capacidades

### Webapp

- crear, listar, abrir, duplicar y archivar presentaciones;
- configurar objetivo, audiencia, idioma, longitud, tono y fuentes;
- seleccionar o cargar `DESIGN.md` y un bundle OKF;
- generar storyboard y contenido por diapositiva;
- editar texto y estructura, aceptar/rechazar cambios y regenerar una slide;
- visualizar la presentación en modo lienzo y modo presentación;
- mostrar fuentes y contexto usado para cada afirmación generada;
- guardar versiones y exportar, como mínimo, un formato portable definido por
  contrato.

La interfaz será server-rendered con Go y HTML semántico. `htmx` se usará para
actualizaciones parciales, formularios y acciones de generación sin convertir
la aplicación en una SPA. JavaScript adicional será progresivo y limitado a
interacciones que no puedan resolverse con HTML/htmx.

### CLI

La CLI será un binario Go y compartirá el mismo core de dominio que la
webapp. Deberá permitir, como mínimo:

- validar `DESIGN.md` y bundles OKF;
- crear y listar proyectos;
- generar un storyboard o una diapositiva;
- inspeccionar el contexto y las fuentes usados;
- exportar una presentación;
- emitir salida legible para humanos y JSON estable para automatización.

Los nombres exactos de comandos, flags y esquema JSON se definirán en
contratos posteriores. La CLI no debe duplicar reglas de negocio del servidor.

### Servidor MCP

El servidor MCP será también Go y expondrá operaciones para agentes, con
entradas y salidas estructuradas. La primera superficie prevista incluye:

- descubrir proyectos, identidades y bundles de conocimiento;
- validar y resumir `DESIGN.md` y OKF;
- crear un proyecto o storyboard;
- generar, revisar y regenerar una diapositiva;
- consultar citas y trazabilidad de una slide;
- exportar o recuperar el artefacto final.

Las herramientas MCP deben reutilizar los casos de uso del core. No habrá una
segunda implementación de generación ni una ruta que permita saltarse las
validaciones de contexto, permisos o trazabilidad.

## Arquitectura

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

El core de Go contendrá el dominio de decks y slides, casos de uso,
validadores, trazabilidad y puertos para IA, almacenamiento y exportación. La
webapp, CLI y MCP serán adaptadores delgados. Los adaptadores no deben
reimplementar reglas de generación, validación o permisos.

La infraestructura KDD heredada en `scripts/`, `knowledge/contracts/` y
`.agents/` sigue gobernando el desarrollo. El producto nuevo se implementa en
Go; Python se conserva únicamente para los gates y herramientas KDD existentes
hasta que un contrato autorice una migración.

## Modelo conceptual mínimo

- **Project**: contenedor de una presentación, configuración y versiones.
- **Deck**: título, objetivo, audiencia, configuración de generación y orden de
  slides.
- **Slide**: unidad editable con intención, contenido, layout, elementos,
  citas, estado y versión.
- **DesignSystem**: `DESIGN.md` parseado, tokens y diagnóstico de validación.
- **KnowledgeBundle**: colección OKF y sus índices, conceptos y enlaces.
- **GenerationRun**: inputs, modelo, parámetros, salida, advertencias y
  referencias de una ejecución de IA.
- **Review**: decisión humana sobre una propuesta o una slide.

Cada `GenerationRun` debe conservar suficiente información para explicar qué
contexto, tokens y configuración produjo la salida. Los secretos y
credenciales nunca forman parte de los artefactos exportados ni de los logs.

## Principios de IA y calidad

- La generación es una propuesta; la aprobación final es humana.
- Toda afirmación factual debe tener una cita o quedar marcada para revisión.
- El contexto enviado al modelo debe ser explícito, acotado y auditable por
  diapositiva.
- La identidad visual se aplica mediante tokens, no mediante valores sueltos
  inventados por cada prompt.
- Los errores de validación se muestran como findings accionables, no se
  silencian para producir una salida aparentemente correcta.
- El proveedor de IA se integra detrás de un puerto para poder probar el core
  con respuestas deterministas y cambiar de proveedor sin cambiar los casos de
  uso.

## Entregables iniciales

1. Core Go con modelo de dominio, almacenamiento local y puertos de IA.
2. Parser/validador de `DESIGN.md` y perfil OKF de showme.
3. Webapp mínima con creación de proyecto, carga de fuentes y storyboard.
4. CLI Go con comandos equivalentes y JSON estable.
5. Servidor MCP Go con herramientas equivalentes a los casos de uso.
6. Render de slide y exportador inicial definidos por contrato.
7. Pruebas de trazabilidad, seguridad, validación y flujos end-to-end.

El orden exacto se decide mediante contratos KDD. Antes de delegar cada pieza
se escribirá su test-oráculo, se sellará su hash y se comprobarán los gates de
Nivel 1.

## Fuera de alcance inicial

- Entrenamiento o fine-tuning de modelos propios.
- Editor visual libre tipo PowerPoint con cientos de controles manuales.
- Colaboración multiusuario en tiempo real.
- Marketplace de plantillas o identidades.
- Publicación automática en servicios externos.
- Soporte de formatos de exportación no definidos por contrato.
- Permitir que un agente modifique el bundle OKF o `DESIGN.md` sin revisión y
  sin historial de cambios.

## Decisiones que quedan abiertas para contratos

- proveedor o proveedores de IA y política de selección de modelo;
- almacenamiento inicial y formato persistido del deck;
- formato de exportación prioritario;
- autenticación, autorización y aislamiento entre proyectos;
- protocolo de transporte MCP y estrategia de versionado de herramientas;
- esquema exacto de slide, layout y elementos visuales;
- límites de tamaño, coste, tiempo de espera y reintentos de generación;
- estrategia de renderizado PDF/PPTX/HTML.

Estas decisiones no deben resolverse implícitamente en una primera
implementación: cada una que afecte la interfaz o el comportamiento observable
debe tener un contrato KDD o una decisión documentada enlazada desde él.
