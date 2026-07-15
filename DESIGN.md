---
version: alpha
name: Showme Editorial Signal
description: Identidad visual para crear presentaciones claras, expresivas y verificables.
colors:
  primary: "#111827"
  secondary: "#475569"
  tertiary: "#7C3AED"
  neutral: "#F8FAFC"
  surface: "#FFFFFF"
  on-surface: "#111827"
  success: "#047857"
  warning: "#B45309"
  error: "#B91C1C"
typography:
  headline-display:
    fontFamily: "Space Grotesk"
    fontSize: 48px
    fontWeight: 600
    lineHeight: 1.05
    letterSpacing: -0.02em
  headline-lg:
    fontFamily: "Space Grotesk"
    fontSize: 32px
    fontWeight: 600
    lineHeight: 1.1
  body-md:
    fontFamily: "Inter"
    fontSize: 16px
    fontWeight: 400
    lineHeight: 1.5
  body-sm:
    fontFamily: "Inter"
    fontSize: 14px
    fontWeight: 400
    lineHeight: 1.45
  label-md:
    fontFamily: "Inter"
    fontSize: 12px
    fontWeight: 600
    lineHeight: 1.1
    letterSpacing: 0.06em
rounded:
  sm: 6px
  md: 12px
  lg: 20px
  full: 9999px
spacing:
  xs: 4px
  sm: 8px
  md: 16px
  lg: 24px
  xl: 48px
  xxl: 96px
components:
  button-primary:
    backgroundColor: "{colors.tertiary}"
    textColor: "{colors.surface}"
    rounded: "{rounded.md}"
    padding: 12px
  button-primary-hover:
    backgroundColor: "#6D28D9"
    textColor: "{colors.surface}"
    rounded: "{rounded.md}"
    padding: 12px
  card:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.on-surface}"
    rounded: "{rounded.md}"
    padding: 24px
  citation:
    backgroundColor: "{colors.neutral}"
    textColor: "{colors.secondary}"
    rounded: "{rounded.sm}"
    padding: 8px
  validation-success:
    backgroundColor: "{colors.success}"
    textColor: "{colors.surface}"
    rounded: "{rounded.sm}"
    padding: 8px
  validation-warning:
    backgroundColor: "{colors.warning}"
    textColor: "{colors.surface}"
    rounded: "{rounded.sm}"
    padding: 8px
---

## Overview

Showme feels like an editorial studio for ideas: calm enough for careful
thinking, expressive enough to make a point memorable. The visual language is
structured, high-contrast and generous with space. AI output is presented as a
draft with visible provenance, never as an opaque final answer.

## Colors

Deep ink is used for primary text and slide headlines. Slate carries secondary
information and metadata. Violet is the single action accent: it marks the
next meaningful action, generation state and selected focus. The neutral
background keeps dense source and review information readable. Green, amber
and red are reserved for validation status and must not replace explanatory
labels.

## Typography

Space Grotesk gives presentation headlines a distinct voice without becoming
decorative. Inter carries body copy, controls, citations and system messages.
Headlines should be short and confident; body text should remain scannable.
Uppercase labels are reserved for metadata and status, not paragraphs.

## Layout

Use a presentation canvas with a stable 16:9 stage and a responsive review
workspace around it. Favor a clear two-column review layout: the slide preview
is primary and the context, citations and generation controls are secondary.
Use the spacing scale consistently and preserve generous safe areas around
slide content.

## Elevation & Depth

Depth comes from tonal layers, borders and restrained shadows. A slide preview
may sit on a neutral workspace, while editable panels use white surfaces. Do
not use gradients or heavy shadows to communicate hierarchy.

## Shapes

Showme uses soft editorial geometry: modest radii for controls and cards,
larger radii only for prominent containers. A slide itself remains a clean
rectangle so its output transfers naturally to export formats.

## Components

Primary actions include generating a deck, generating a slide and accepting a
proposal. Citation blocks are always visually distinct from generated prose.
Validation findings use the semantic colors but also include text and icons so
meaning never depends on color alone.

## Do's and Don'ts

- Do show the source context and citations next to generated claims.
- Do use violet for one primary action per view.
- Do preserve readable contrast and visible focus states.
- Do keep slide layouts sparse enough to support presentation at a distance.
- Don't present unreviewed AI output as approved content.
- Don't invent brand colors, fonts or component styles outside these tokens.
- Don't use more than two headline treatments on one slide.
- Don't use color alone to communicate validation, approval or error.
