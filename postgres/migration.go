package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	spicemigration "github.com/spice-framework/spice/migration"
	spicepostgres "github.com/spice-framework/starter-postgres"
)

const (
	petclinicModuleID = "github.com/spice-framework/petclinic"
	schemaVersion     = 202607280101
	seedVersion       = 202607280102
)

// Migration owns the PostgreSQL schema transition plan.
type Migration struct {
	native *sql.DB
}

// NewMigration constructs a PostgreSQL migration runner.
func NewMigration(native *sql.DB) *Migration {
	return &Migration{native: native}
}

func (migration *Migration) plan() (*spicemigration.Plan, error) {
	return spicemigration.NewPlan([]spicemigration.Spec{
		{
			Version: schemaVersion,
			Module:  petclinicModuleID,
			Name:    "create Petclinic schema",
			SQL:     schemaSQL,
		},
		{
			Version: seedVersion,
			Module:  petclinicModuleID,
			Name:    "load canonical Petclinic data",
			SQL:     seedSQL,
		},
	})
}

// Run applies the module-owned schema and seed plan.
func (migration *Migration) Run(ctx context.Context) error {
	if ctx == nil {
		return errors.New("migrate Petclinic PostgreSQL database: context is nil")
	}
	if migration == nil || migration.native == nil {
		return errors.New("migrate Petclinic PostgreSQL database: database is nil")
	}
	plan, err := migration.plan()
	if err != nil {
		return fmt.Errorf(
			"construct Petclinic PostgreSQL migration plan: %w",
			err,
		)
	}
	backend, err := spicepostgres.NewMigrationBackend(
		migration.native,
		spicepostgres.MigrationOptions{},
	)
	if err != nil {
		return fmt.Errorf(
			"construct Petclinic PostgreSQL migration backend: %w",
			err,
		)
	}
	runner, err := spicemigration.NewRunner(backend)
	if err != nil {
		return fmt.Errorf(
			"construct Petclinic PostgreSQL migration runner: %w",
			err,
		)
	}
	if _, err := runner.Run(ctx, plan); err != nil {
		return fmt.Errorf(
			"migrate Petclinic PostgreSQL database: %w",
			err,
		)
	}
	return nil
}
