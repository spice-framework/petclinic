// @import { Bean } from "github.com/spice-framework/spice/annotation/core"
// @import { Module } from "github.com/spice-framework/spice/annotation/modulith"

// Package memory provides the zero-network Petclinic persistence profile.
// @Module(allowedDependencies=["github.com/spice-framework/petclinic/model", "github.com/spice-framework/petclinic/owner", "github.com/spice-framework/petclinic/vet"])
package memory

import (
	"errors"
	"sync"

	"github.com/spice-framework/petclinic/model"
	"github.com/spice-framework/petclinic/owner"
	"github.com/spice-framework/petclinic/vet"
)

// Database is one concurrency-safe in-process Petclinic data set.
type Database struct {
	mu          sync.RWMutex
	owners      map[model.ID]owner.Owner
	petTypes    []owner.PetType
	vets        []vet.Vet
	nextOwnerID model.ID
	nextPetID   model.ID
	nextVisitID model.ID
}

// NewDatabase validates and defensively copies initial data.
func NewDatabase(
	owners []owner.Owner,
	petTypes []owner.PetType,
	vets []vet.Vet,
) (*Database, error) {
	petTypesRepository := &PetTypeRepository{}
	vetsRepository := &VetRepository{}
	database := &Database{
		owners:   make(map[model.ID]owner.Owner, len(owners)),
		petTypes: petTypesRepository.clonePetTypes(petTypes),
		vets:     vetsRepository.cloneVets(vets),
	}
	if err := database.loadOwners(owners); err != nil {
		return nil, err
	}
	if err := petTypesRepository.validatePetTypes(database.petTypes); err != nil {
		return nil, err
	}
	if err := vetsRepository.validateVets(database.vets); err != nil {
		return nil, err
	}
	return database, nil
}

func (database *Database) loadOwners(values []owner.Owner) error {
	for _, value := range values {
		if !value.ID.Valid() {
			return errors.New(
				"construct memory database: seeded owner ID must be positive",
			)
		}
		if _, duplicate := database.owners[value.ID]; duplicate {
			return errors.New(
				"construct memory database: seeded owner IDs must be unique",
			)
		}
		if err := database.observeOwnerIDs(value); err != nil {
			return err
		}
		database.owners[value.ID] = value.Clone()
	}
	return nil
}

func (database *Database) observeOwnerIDs(value owner.Owner) error {
	database.nextOwnerID = database.maxID(database.nextOwnerID, value.ID)
	seenPets := make(map[model.ID]struct{}, len(value.Pets))
	for _, pet := range value.Pets {
		if !pet.ID.Valid() {
			return errors.New(
				"construct memory database: seeded pet ID must be positive",
			)
		}
		if _, duplicate := seenPets[pet.ID]; duplicate {
			return errors.New(
				"construct memory database: owner pet IDs must be unique",
			)
		}
		seenPets[pet.ID] = struct{}{}
		database.nextPetID = database.maxID(database.nextPetID, pet.ID)
		for _, visit := range pet.Visits {
			if !visit.ID.Valid() {
				return errors.New(
					"construct memory database: seeded visit ID must be positive",
				)
			}
			database.nextVisitID = database.maxID(
				database.nextVisitID,
				visit.ID,
			)
		}
	}
	return nil
}

func (database *Database) maxID(left, right model.ID) model.ID {
	if right > left {
		return right
	}
	return left
}
