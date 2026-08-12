# Spice MySQL starter

`github.com/spice-framework/starter-mysql` is the independently versioned,
opt-in MySQL integration for Spice. It creates secure, instance-owned
`database/sql` pools with the maintained go-sql-driver/mysql driver. Importing
Spice core alone never selects this driver.

```go
database, err := mysql.Open(mysql.Options{
    URL:             configuration.DatabaseURL,
    ApplicationName: "orders-service",
})
if err != nil {
    return nil, err
}
```

`Open` validates configuration and creates a caller-owned pool without network
I/O or global driver/TLS registration. `Ping` performs the explicit readiness
check with the caller's context. TLS 1.2 hostname verification is the default;
disabled TLS requires both `tls=disable` in the URL and explicit
`AllowInsecure`, which is intended only for isolated local/container tests.
Unknown driver options, incomplete identities, unbounded pool settings, and
unsafe application metadata fail before construction without exposing URLs or
passwords.

## Install

```text
go get github.com/spice-framework/starter-mysql@latest
```

During preview development, applications should pin the exact compatible
commit recorded in [support metadata](docs/support.md).

## Verify

Go 1.26.5 is mandatory:

```text
make check
make compatibility
make release-parity
make verify
make verify-release
```

The normal verifier checks formatting, module/vendor reproducibility, vet,
allowlisted lint and nil safety, gosec, govulncheck, shuffled race tests, at
least 85% product coverage, and offline vendor builds.

Core compatibility exercises the distinct minimum and current exact versions
in `spice-compatibility.json`. The minimum must equal the direct Spice
requirement in `go.mod`; temporary alternate module files keep both checks from
rewriting `go.mod`, `go.sum`, or `vendor`. `make verify` runs both compatibility
lines before the repository quality gate, and hosted CI runs them independently.

Release parity runs the exact `spice-dev` tool authorized by `go.mod` and the
retained repository builder twice each, entirely from `vendor` with network and
workspace resolution disabled. It compares archive entries after normalizing
only the builders' documented root-directory spelling, requires equivalent
SBOM package and dependency facts, verifies both checksum files, and forbids
rehearsal signatures on Windows and Linux.

Real-system acceptance uses the immutable official MySQL 8.4.11 image index.
It proves two isolated pools, live queries, bounded query cancellation,
connection recovery, independent cleanup, and credential-safe authentication
failures:

```text
docker run --detach --name spice-mysql --publish 53306:3306 \
  --env MYSQL_DATABASE=spice --env MYSQL_USER=spice \
  --env MYSQL_PASSWORD=spice-test --env MYSQL_ROOT_PASSWORD=spice-root \
  mysql@sha256:b3b90af2a6552ae30c266fdb7d5dd55f3afb72404bb78d37fe8a23eb857fd3fb
SPICE_MYSQL_TEST_URL='mysql://spice:spice-test@127.0.0.1:53306/spice?tls=disable' \
  make integration
docker rm --force spice-mysql
```

MySQL DDL implicitly commits, so this starter does not claim a transactional
migration backend. Applications own explicit resumable migration plans.

See [the dependency review](docs/dependency-review.md) and
[support contract](docs/support.md) before production adoption.

## Releases

The repository builds deterministic source-only releases with an SPDX 2.3
SBOM, SHA-256 checksums, and Ed25519 signatures. See the exact artifact and
clean-tag ceremony in [the release guide](docs/releasing.md).
The reviewed public trust anchor is committed with fingerprint
`1c3d374cfe4465f6c04ad7afe22acad2bb05d51bc6d249745c92a817265d08a9`;
no signed release is implied until the protected release ceremony completes.
The retained repository builder and signed production workflow remain the
release authority while the centrally rendered unsigned candidate is held to
the dual-builder parity contract.
