# Spice Petclinic Implementation Contract

This repository is the standalone reference application for the Spice
Framework. It proves that a real application can consume a published Spice
module without a local replacement or monorepo-only behavior.

## Delivery model

- Work directly on local `main` in single-writer mode.
- Fetch and inspect `origin/main` before work and immediately before pushing.
- Use bounded, reviewable commits and never overwrite unexpected remote work.
- Preserve generated source and `.spice/*.manifest.json`; never hand-edit them.
- Pin Spice core and toolchain independently to immutable released or
  pseudo-versions.
- Do not add a local `replace` to the committed module graph.

## Required verification

Go 1.26.5 is exact and mandatory. Every commit must pass `make verify`, which
is implemented by the cross-platform Go program in `internal/qualitygate` and
also runs directly on Windows as `go run ./internal/qualitygate`.

The gate checks formatting, module and vendor reproducibility, vet, the
canonical Spice application style, the allowlisted linter policy, NilAway,
gosec, govulncheck, shuffled and race tests,
an 85% business-source coverage floor, vendor-offline tests, all three Spice
generation targets, normal Go builds, and executable `-check` smoke paths.

Database integration tests additionally run against PostgreSQL 18 and MySQL
8.4 in hosted Linux jobs. Local verification may use those tests when their
explicit test URLs are present; it never starts or downloads services.

## Application invariants

- Handwritten application source conforms to `.spice/style.json` and the
  canonical Spice `CODE_STYLE.md`; `spicestyle` runs offline in every check.
- Handwritten source remains ordinary valid Go.
- Annotations remain declaration comments resolved through explicit imports.
- Generated code mirrors source ownership under `internal/spicegen/<target>`
  and remains inspectable, committed, deterministic Go.
- Interfaces and cleanup behavior remain explicit in generated direct calls;
  no runtime reflection, package scanning, or global service locator is added.
- The in-memory target remains the zero-network default.
- PostgreSQL and MySQL retain secure defaults and require explicit local
  opt-in for disabled TLS.
- Tests cover success, validation, missing data, deterministic ordering,
  cancellation, and integration behavior appropriate to each change.
