# presets

Niche business presets and seed configurations for SiteCraft SMB verticals.

## Purpose

Owns the collection of pre-configured business templates (legal, medical, auto, beauty, delivery, notary, repair, cyber, etc.) used to bootstrap industry-specific sites and populate default database schemas.

## Ownership

- `presets/*.json` — Industry-specific configuration files defining site metadata, themes, navigation, content blocks, and initial records.

## Local Contracts

- JSON Schema: each preset must strictly conform to the structure expected by `internal/data/preset.go` (`PresetData`, `SectionConfig`, etc.).
- Theme tokens: must supply valid hexadecimal or CSS color values for `primary_color`, `accent_color`, `background_color`, and a recognized `font_family`.
- Section Registry mapping: all entries in the `sections` array must specify a `type` that is registered in `ui/registry/registry.go`.
- Self-contained content: all content sections must provide sensible defaults for required copy, titles, and labels to prevent nil/empty render panics.

## Work Guidance

- When introducing a new preset or section key, ensure corresponding Go models in `internal/models/` and UI blocks in `ui/sections/` support the schema.
- Validate JSON formatting before saving.
- Test presets by running the preset test suite or seeding into a test database.

## Verification

- Run preset validation tests: `go test -v ./internal/data`

## Child DOX Index

None.
