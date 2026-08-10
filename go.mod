module github.com/spice-framework/petclinic

go 1.26.0

toolchain go1.26.5

tool (
	github.com/spice-framework/toolchain/cmd/spice
	github.com/spice-framework/toolchain/cmd/spice-annotation-core
	github.com/spice-framework/toolchain/cmd/spicestyle
)

require (
	github.com/spice-framework/spice v0.1.0-preview.1.0.20260807050649-46ba4660cfb0
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
	github.com/spice-framework/toolchain v0.1.0-preview.2.0.20260810170203-57cf59f2d895 // indirect
	golang.org/x/mod v0.38.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	golang.org/x/tools v0.48.0 // indirect
)
