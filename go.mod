module github.com/spice-framework/petclinic

go 1.26.0

toolchain go1.26.5

tool (
	github.com/spice-framework/spice/cmd/spice
	github.com/spice-framework/spice/cmd/spice-annotation-core
)

require (
	github.com/spice-framework/spice v0.0.0-20260805185924-ee45e0aa386e
	github.com/spice-framework/starter-mysql v0.0.0-20260805191832-cc9a2c5034d0
	github.com/spice-framework/starter-postgres v0.0.0-20260805182646-6310c4b4b95e
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/go-sql-driver/mysql v1.10.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.10.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/mod v0.38.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	golang.org/x/tools v0.48.0 // indirect
)
