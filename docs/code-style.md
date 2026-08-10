# Petclinic code-style contract

Petclinic is a strict consumer of the canonical Spice application standard in
`CODE_STYLE.md` from the independently versioned Spice Toolchain. The exact
machine-readable policy is committed as `.spice/style.json`, and the module
pins `github.com/spice-framework/toolchain/cmd/spicestyle` through a Go `tool`
directive. Normal verification is offline and vendor-backed:

```text
go tool github.com/spice-framework/toolchain/cmd/spicestyle \
  --config=.spice/style.json ./...
```

`make check` and `make verify` run that command. A handwritten application
file cannot silently drift to a second primary type, a loose package function,
implicit construction, mutable package state, or a mismatched filename.
Generated Spice output and the narrow language/framework boundaries below are
classified explicitly instead of being misrepresented as ordinary application
objects.

## Exact interoperability boundaries

- `presentation/assets.go`, `presentation/catalog.go`, and
  `presentation/renderer.go` retain one typed `embed.FS` package variable each;
  Go's `//go:embed` directive requires package scope.
- `owner/errors.go` retains `ErrOwnerNotFound` and `ErrPetNotFound` as typed
  `error` package variables so external repositories and controllers preserve
  standard `errors.Is` interoperability.
- `mysql/owner_sql.go`, `postgres/owner_sql.go`, and `postgres/schema.go` are
  immutable constant-only SQL carrier boundaries.
- Go process entrypoints, tests, Spice provider/topic declarations, and
  generated ownership roots use the standard exact boundary classifications.

Every variable exception records an exact path, symbol, resolved Go type,
reason, and review issue. Globs never authorize a family of mutable variables,
and declaration comments cannot suppress analyzer findings.

## Refactoring rule

New application behavior belongs to the owning named type. Helpers become
private methods of that type; deterministic seed, migration, schema, and loader
behavior is constructed explicitly. A new exception is appropriate only when
Go or an external interoperability contract requires package identity. It must
be narrower than the code it authorizes and must include a durable review
identifier.
