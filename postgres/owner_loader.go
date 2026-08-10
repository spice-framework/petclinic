package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/spice-framework/petclinic/model"
	"github.com/spice-framework/petclinic/owner"
)

// OwnerLoader reconstructs complete owner aggregates from one SQL querier.
type OwnerLoader struct {
	querier ownerQuerier
}

// NewOwnerLoader constructs an aggregate loader for a database or transaction.
func NewOwnerLoader(querier ownerQuerier) *OwnerLoader {
	return &OwnerLoader{querier: querier}
}

// LoadOwner reconstructs one owner aggregate by identity.
func (loader *OwnerLoader) LoadOwner(
	ctx context.Context,
	id model.ID,
) (owner.Owner, bool, error) {
	var value owner.Owner
	err := loader.querier.QueryRowContext(ctx, findOwnerSQL, id).Scan(
		&value.ID,
		&value.FirstName,
		&value.LastName,
		&value.Address,
		&value.City,
		&value.Telephone,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return owner.Owner{}, false, nil
	}
	if err != nil {
		return owner.Owner{}, false, fmt.Errorf("query owner row: %w", err)
	}
	pets, err := loader.loadPets(ctx, id)
	if err != nil {
		return owner.Owner{}, false, err
	}
	value.Pets = pets
	return value, true, nil
}

func (loader *OwnerLoader) loadPets(
	ctx context.Context,
	ownerID model.ID,
) (values []owner.Pet, err error) {
	rows, err := loader.querier.QueryContext(ctx, findPetsSQL, ownerID)
	if err != nil {
		return nil, fmt.Errorf("query owner pets: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	for rows.Next() {
		var value owner.Pet
		if err := rows.Scan(
			&value.ID,
			&value.Name,
			&value.BirthDate,
			&value.Type.ID,
			&value.Type.Name,
		); err != nil {
			return nil, fmt.Errorf("scan owner pet: %w", err)
		}
		visits, err := loader.loadVisits(ctx, value.ID)
		if err != nil {
			return nil, err
		}
		value.Visits = visits
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read owner pets: %w", err)
	}
	if values == nil {
		values = []owner.Pet{}
	}
	return values, nil
}

func (loader *OwnerLoader) loadVisits(
	ctx context.Context,
	petID model.ID,
) (values []owner.Visit, err error) {
	rows, err := loader.querier.QueryContext(ctx, findVisitsSQL, petID)
	if err != nil {
		return nil, fmt.Errorf("query pet visits: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	for rows.Next() {
		var value owner.Visit
		if err := rows.Scan(&value.ID, &value.Date, &value.Description); err != nil {
			return nil, fmt.Errorf("scan pet visit: %w", err)
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read pet visits: %w", err)
	}
	if values == nil {
		values = []owner.Visit{}
	}
	return values, nil
}
