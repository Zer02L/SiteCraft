# presets

Niche business presets and seed configurations for SiteCraft SMB verticals.

## Purpose

Owns the collection of pre-configured business templates (legal, medical, auto, beauty, delivery, notary, repair, cyber, etc.) used to bootstrap industry-specific sites and populate default database schemas.

## Ownership

- `presets/*.json` — Industry-specific configuration files defining site metadata, themes, navigation, content blocks, and initial records.

## Local Contracts

- JSON Schema: each preset must strictly conform to the structure expected by `internal/data/preset.go` (`PresetData`, `SiteSettings`, `SiteContent`, `ServiceItem`, `PageBlock`).
- Theme tokens: must supply valid hexadecimal or CSS color values for `primary_color`, `accent_color`, `background_color`, and a recognized `font_family`.
- Section Registry mapping: all entries in the `blocks` array must specify a `type` registered in `ui/registry/registry.go`.
- Self-contained content: all content sections must provide sensible defaults for required copy, titles, labels, and structured payloads (for `features`, `faq`, `quiz`, `before_after`, etc.) to prevent nil/empty render panics.

## Preset Classification

- **B2B Presets** (`b2b_legal.json`, `b2b_legal_corporate.json`, `b2b_legal_swiss.json`, `b2b_legal_british.json`, `b2b_legal_milano.json`, `b2b_notary.json`, `b2b_repair.json`, `b2b_cyber.json`, `b2b_academy.json`):
  - Prioritize informational clarity, compliance, and risk mitigation.
  - Include legal entity details (`legal_entity`, `inn`, `kpp`, `ogrn`, `bank_*`, `contact_email`).
  - Feature comparison tables (`comparison`), metric odometers (`cinematic_odometer`), spotlight practices (`cinematic_spotlight`), corporate retainers (`cinematic_sticky_cards`), careers (`careers`), and SLA roadmaps (`cinematic_sticky_stack`).
- **B2C Presets** (`b2c_auto.json`, `b2c_beauty.json`, `b2c_delivery.json`, `b2c_medical.json`, `b2c_craft_coffee.json`):
  - Prioritize visual rhythm, emotional engagement, and tactile interaction.
  - Feature interactive cost/preference calculators (`quiz`), before/after transformations (`before_after`), galleries, client reviews (`reviews`), cart drawers (`enable_cart`), instant messenger buttons (`whatsapp_phone`, `telegram_contact`), and slot booking (`booking_mode: slots`).
- **Comprehensive Test Suites** (`test_all_modules.json`, `test.json`):
  - Validates full integration across all 20 commercial modules.
- **Prefix Resolver Fallback**: `internal/data/preset.go` automatically resolves legacy preset names (e.g. `legal`, `auto`, `notary`) to their `b2b_` or `b2c_` prefixed files.

## Work Guidance

- Differentiate presets by layout structure and density, not merely `primary_color`: B2B presets must prioritize informational clarity, comparison tables, and quick conversion funnels; B2C presets must prioritize visual rhythm, showcase sections, and tactile interaction.
- When introducing a new preset or section key, ensure corresponding Go models in `internal/models/` and UI blocks in `ui/sections/` support the schema.
- Validate JSON formatting before saving.
- Test presets by running the preset test suite or seeding into a test database.

## Verification

- Run preset validation tests: `go test -v ./internal/data`

## Child DOX Index

None.
