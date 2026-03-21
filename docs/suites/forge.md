# Forge Suite

## Purpose

Forge is the user-composed suite mode. It does not introduce unique container images.
Instead, it lets the user select components from Boosting, Neural, and Flow, then
generates a tailored Compose file from the shared Forge template.

## Component Sources

Forge can enable components drawn from:

- Boosting
- Neural
- Flow

Each candidate service is present in `templates/forge.compose.yml` with
`profiles: ["disabled"]`. Selection logic in `internal/suite/forge.go` removes the
profile gate from chosen services before writing the final Compose output.

## Selection Flow

1. `concave` presents the available components through `ui.Checklist`.
2. The chosen set is mapped to concrete container definitions.
3. Shared port logic checks conflicts before a file is written.
4. The generated Compose file is validated with `docker compose config --quiet`.
5. The final file is stored in `~/gradient/compose/forge.compose.yml`.

## Port and Service Rules

- port assignments come from the source suite definitions
- MLflow on `5000` is deduplicated through shared port management
- conflicting selections must fail before Compose output is installed
- Forge never invents new host paths or custom ports outside the registered suite data

## Documentation Rule

Because Forge reuses services from the other suites, detailed container internals remain
documented in:

- [Boosting](boosting.md)
- [Neural](neural.md)
- [Flow](flow.md)

This document focuses on selection, generation, and conflict behavior rather than
duplicating the service reference from those suite docs.
