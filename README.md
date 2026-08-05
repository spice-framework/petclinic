# Spice Petclinic

Spice Petclinic is the standalone reference application for the
[Spice Framework](https://github.com/spice-framework/spice). It is a
behavior-first Go port of
[Spring Petclinic](https://github.com/spring-projects/spring-petclinic) at
commit `f182358d02e4a68e52bdbabf55ca7800288511e7`.

This repository is intentionally a real consuming module. Its `go.mod` pins an
immutable Spice version, authorizes the Spice CLI and annotation tool through
standard Go `tool` directives, and contains no local `replace`. A clone builds,
tests, generates, and runs without a Spice monorepo checkout.

## What it proves

Petclinic exercises a complete, inspectable Spice application rather than a
toy dependency-injection graph:

- owners, pets, visits, pet types, veterinarians, and specialties;
- deterministic validation and stable query ordering;
- explicit generated interface bindings and direct constructor calls;
- dependency-first startup, reverse cleanup, and graceful HTTP shutdown;
- complete owner, pet, visit, veterinarian, and welcome HTTP workflows;
- responsive embedded views, localization, security headers, and RFC 9457
  problem responses;
- loopback-only generated management endpoints;
- a zero-network, concurrency-safe in-memory application target;
- separate PostgreSQL 18 and MySQL 8.4 application graphs with module-owned
  migrations, secure defaults, transactions, and real integration tests.

The application source uses one named domain type per file where practical.
Generated code is committed beneath `internal/spicegen/<target>` and mirrors
the handwritten package/file ownership in `sources/`. Target-wide contracts,
providers, lifecycle, features, and HTTP wiring are split into named files.
`.spice/*.manifest.json` records exact source relationships and SHA-256
ownership, so generation rejects manual edits and stale foreign output.

Use the standard Go debugger on the complete package. The generated calls are
ordinary Go: there is no runtime reflection container, package scan, or hidden
service locator. `spice generated --source <file>` maps handwritten source to
its generated unit when deeper graph debugging is useful.

## Requirements

- Go 1.26.5 exactly.
- No database for the default in-memory target.
- PostgreSQL 18 or MySQL 8.4 only for the corresponding integration workflow.

Fetch explicitly selected dependencies once, then verify offline:

```text
make bootstrap
make verify
```

On Windows without `make`, use the equivalent commands:

```text
go mod tidy -diff
go -C tools mod tidy -diff
go mod download
go -C tools mod download
go mod vendor
go run ./internal/qualitygate
```

The two online tidy checks prime the complete selected closure without changing
the module files, including transitive test-only modules. The quality gate then
forces `GOPROXY=off` so verification cannot hide a missing bootstrap dependency
behind a network download.

The cross-platform quality gate enforces formatting, tidy/vendor
reproducibility, vet, allowlisted linting, nil safety, security and vulnerability
analysis, shuffled/race tests, at least 85% business-source coverage,
vendor-offline builds, generated freshness, and executable checks for all
three targets.

For an equivalent warm body-edit comparison against Spring Petclinic, first
prime the reference checkout's Maven dependencies and then run:

```text
make benchmark-spring SPRING_PETCLINIC=<checkout>
```

The repository-owned manifest fixes the Spring commit, Spring Boot and Java
versions, warmups, samples, p90 budgets, and exact source files. The comparison
runs Maven offline and restores both edited sources even when a build fails.

## Generate and run

The root target uses only memory and has no external I/O requirement:

```text
go tool github.com/spice-framework/spice/cmd/spice generate --check --target Petclinic .
go tool github.com/spice-framework/spice/cmd/spice run --target Petclinic . -- -check
```

Start the web application in PowerShell:

```powershell
$env:SPICE_PETCLINIC_ADDRESS = "127.0.0.1:8080"
go tool github.com/spice-framework/spice/cmd/spice run --target Petclinic .
```

Or in a POSIX shell:

```sh
export SPICE_PETCLINIC_ADDRESS=127.0.0.1:8080
go tool github.com/spice-framework/spice/cmd/spice run --target Petclinic .
```

Open `http://127.0.0.1:8080/`. Generated `/actuator/*` routes accept only a
direct loopback peer and do not trust forwarding headers.

For the watched development loop:

```text
go tool github.com/spice-framework/spice/cmd/spice dev --target Petclinic .
```

An invalid annotation produces the same source-positioned compiler/LSP
diagnostic and preserves the last-known-good process. Fixing and saving causes
guarded regeneration and a graceful complete-package restart; source always
retains valid-Go `// @...` comments.

## Persistence targets

PostgreSQL generation and execution:

```powershell
go tool github.com/spice-framework/spice/cmd/spice generate --check --target Postgres ./cmd/postgres
$env:SPICE_PETCLINIC_POSTGRES_URL = "postgres://petclinic:petclinic@127.0.0.1:5432/petclinic?sslmode=disable"
$env:SPICE_PETCLINIC_POSTGRES_ALLOW_INSECURE = "true"
go tool github.com/spice-framework/spice/cmd/spice run --target Postgres ./cmd/postgres
```

Real PostgreSQL repository test:

```powershell
$env:SPICE_POSTGRES_TEST_URL = "postgres://petclinic:petclinic@127.0.0.1:5432/petclinic?sslmode=disable"
go test -tags=integration -count=1 ./postgres
```

MySQL generation and execution:

```powershell
go tool github.com/spice-framework/spice/cmd/spice generate --check --target Mysql ./cmd/mysql
$env:SPICE_PETCLINIC_MYSQL_URL = "mysql://petclinic:petclinic@127.0.0.1:3306/petclinic?tls=disable"
$env:SPICE_PETCLINIC_MYSQL_ALLOW_INSECURE = "true"
go tool github.com/spice-framework/spice/cmd/spice run --target Mysql ./cmd/mysql
```

Real MySQL repository test:

```powershell
$env:SPICE_MYSQL_TEST_URL = "mysql://petclinic:petclinic@127.0.0.1:3306/petclinic?tls=disable"
go test -tags=integration -count=1 ./mysql
```

Verified TLS is the production default for both database targets. Disabled TLS
requires the explicit local-development opt-in shown above. MySQL migrations
are advisory-locked, checksum-verified, idempotent, and resumable because MySQL
DDL cannot truthfully provide cross-statement transactional behavior.

## Public routes

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/` | Welcome page |
| `GET` | `/owners/find` | Owner search form |
| `GET` | `/owners` | Prefix search and paginated results |
| `GET`, `POST` | `/owners/new` | Create an owner |
| `GET` | `/owners/{ownerId}` | Owner, pets, and visits |
| `GET`, `POST` | `/owners/{ownerId}/edit` | Edit an owner |
| `GET`, `POST` | `/owners/{ownerId}/pets/new` | Add a pet |
| `GET`, `POST` | `/owners/{ownerId}/pets/{petId}/edit` | Edit a pet |
| `GET`, `POST` | `/owners/{ownerId}/pets/{petId}/visits/new` | Add a visit |
| `GET` | `/vets.html` | Paginated veterinarian view |
| `GET` | `/vets` | Stable veterinarian JSON collection |
| `GET` | `/actuator/*` | Generated loopback-only management |

Framework documentation: [getting started](https://github.com/spice-framework/spice/blob/main/docs/getting-started.md),
[generated-code architecture](https://github.com/spice-framework/spice/blob/main/docs/compiler.md),
and [developer loop](https://github.com/spice-framework/spice/blob/main/docs/development-loop.md).
