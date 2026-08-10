package mysql

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	schemaVersion            uint64 = 202607280101
	seedVersion              uint64 = 202607280102
	migrationLockName               = "spice.petclinic.schema"
	migrationLockWaitSeconds        = 30
	migrationCleanupTimeout         = 5 * time.Second
)

// Migration owns MySQL's resumable schema transition lifecycle.
type Migration struct {
	native *sql.DB
}

// NewMigration constructs a resumable MySQL migration runner.
func NewMigration(native *sql.DB) *Migration {
	return &Migration{native: native}
}

// Run applies every locked schema transition exactly once.
func (migration *Migration) Run(ctx context.Context) (err error) {
	if ctx == nil {
		return errors.New("migrate Petclinic MySQL database: context is nil")
	}
	if migration == nil || migration.native == nil {
		return errors.New("migrate Petclinic MySQL database: database is nil")
	}
	connection, err := migration.native.Conn(ctx)
	if err != nil {
		return fmt.Errorf("migrate Petclinic MySQL database: acquire connection: %w", err)
	}
	defer func() {
		err = errors.Join(err, connection.Close())
	}()
	if lockErr := migration.acquireLock(ctx, connection); lockErr != nil {
		return lockErr
	}
	defer func() {
		err = errors.Join(err, migration.releaseLock(ctx, connection))
	}()
	if _, createErr := connection.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS spice_schema_history (
		version BIGINT UNSIGNED NOT NULL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		checksum CHAR(64) NOT NULL,
		applied_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci`); createErr != nil {
		return fmt.Errorf(
			"migrate Petclinic MySQL database: create registry: %w",
			createErr,
		)
	}
	for _, specification := range []resumableMigration{
		{version: schemaVersion, name: "create Petclinic schema", steps: (Schema{}).Steps()},
		{version: seedVersion, name: "load canonical Petclinic data", steps: (Schema{}).SeedSteps()},
	} {
		if applyErr := migration.apply(
			ctx,
			connection,
			specification,
		); applyErr != nil {
			return applyErr
		}
	}
	return nil
}

func (migration *Migration) acquireLock(ctx context.Context, connection *sql.Conn) error {
	var acquired sql.NullInt64
	if err := connection.QueryRowContext(
		ctx,
		`SELECT GET_LOCK(?, ?)`,
		migrationLockName,
		migrationLockWaitSeconds,
	).Scan(&acquired); err != nil {
		return fmt.Errorf("migrate Petclinic MySQL database: acquire lock: %w", err)
	}
	if !acquired.Valid || acquired.Int64 != 1 {
		return errors.New("migrate Petclinic MySQL database: acquire lock: timed out")
	}
	return nil
}

func (migration *Migration) releaseLock(ctx context.Context, connection *sql.Conn) error {
	cleanupContext, cancel := context.WithTimeout(
		context.WithoutCancel(ctx),
		migrationCleanupTimeout,
	)
	defer cancel()
	var released sql.NullInt64
	if err := connection.QueryRowContext(
		cleanupContext,
		`SELECT RELEASE_LOCK(?)`,
		migrationLockName,
	).Scan(&released); err != nil {
		return fmt.Errorf("migrate Petclinic MySQL database: release lock: %w", err)
	}
	if !released.Valid || released.Int64 != 1 {
		return errors.New("migrate Petclinic MySQL database: release lock: lock was not owned")
	}
	return nil
}

func (migration *Migration) apply(
	ctx context.Context,
	connection *sql.Conn,
	specification resumableMigration,
) error {
	checksum := migration.checksum(specification.steps)
	var existing string
	err := connection.QueryRowContext(
		ctx,
		`SELECT checksum FROM spice_schema_history WHERE version = ?`,
		specification.version,
	).Scan(&existing)
	switch {
	case err == nil && existing == checksum:
		return nil
	case err == nil:
		return fmt.Errorf(
			"migrate Petclinic MySQL database: migration %d checksum changed",
			specification.version,
		)
	case !errors.Is(err, sql.ErrNoRows):
		return fmt.Errorf(
			"migrate Petclinic MySQL database: inspect migration %d: %w",
			specification.version,
			err,
		)
	}
	for index, statement := range specification.steps {
		if strings.TrimSpace(statement) == "" {
			return fmt.Errorf(
				"migrate Petclinic MySQL database: migration %d step %d is empty",
				specification.version,
				index+1,
			)
		}
		if _, err := connection.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf(
				"migrate Petclinic MySQL database: migration %d step %d: %w",
				specification.version,
				index+1,
				err,
			)
		}
	}
	if _, err := connection.ExecContext(
		ctx,
		`INSERT INTO spice_schema_history (version, name, checksum) VALUES (?, ?, ?)`,
		specification.version,
		specification.name,
		checksum,
	); err != nil {
		return fmt.Errorf(
			"migrate Petclinic MySQL database: record migration %d: %w",
			specification.version,
			err,
		)
	}
	return nil
}

func (migration *Migration) checksum(steps []string) string {
	digest := sha256.Sum256([]byte(strings.Join(steps, "\x00")))
	return hex.EncodeToString(digest[:])
}
