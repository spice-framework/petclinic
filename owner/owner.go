// @import { Module } from "github.com/spice-framework/spice/annotation/modulith"

// Package owner owns Petclinic owners, pets, visits, and pet types.
// @Module(allowedDependencies=["github.com/spice-framework/petclinic/model", "github.com/spice-framework/petclinic/presentation"])
package owner

import (
	"errors"
	"slices"
	"strings"

	"github.com/spice-framework/petclinic/model"
	"github.com/spice-framework/spice/validation"
)

// Owner is a person responsible for zero or more pets.
type Owner struct {
	model.Person
	Address   string
	City      string
	Telephone string
	Pets      []Pet
}

// Clone returns a deep value copy suitable for repository boundaries.
func (owner *Owner) Clone() Owner {
	if owner == nil {
		return Owner{}
	}
	result := *owner
	result.Pets = make([]Pet, len(owner.Pets))
	for index := range owner.Pets {
		result.Pets[index] = owner.Pets[index].Clone()
	}
	return result
}

// Validate returns the Petclinic owner form constraints in stable field order.
func (owner *Owner) Validate() (validation.Errors, error) {
	if owner == nil {
		return validation.New(validation.Violation{
			Code:    "required",
			Message: "owner must be provided",
		})
	}
	personErrors, err := owner.Person.Validate()
	if err != nil {
		return validation.Errors{}, err
	}
	violations := personErrors.All()
	violations = owner.appendRequired(
		violations,
		"address",
		owner.Address,
	)
	violations = owner.appendRequired(violations, "city", owner.City)
	if !owner.validTelephone(owner.Telephone) {
		violations = append(violations, validation.Violation{
			Field:   "telephone",
			Code:    "telephone.invalid",
			Message: "must contain exactly 10 digits",
		})
	}
	return validation.New(violations...)
}

// PetByID returns a defensive copy of one persisted pet.
func (owner *Owner) PetByID(id model.ID) (Pet, bool) {
	if owner == nil {
		return Pet{}, false
	}
	for _, pet := range owner.Pets {
		if pet.ID == id && id.Valid() {
			return pet.Clone(), true
		}
	}
	return Pet{}, false
}

// PetByName returns a case-insensitive defensive copy. New pets can be omitted
// while validating an edit against already-persisted siblings.
func (owner *Owner) PetByName(
	name string,
	ignoreNew bool,
) (Pet, bool) {
	if owner == nil {
		return Pet{}, false
	}
	for _, pet := range owner.Pets {
		if strings.EqualFold(pet.Name, name) &&
			(!ignoreNew || !pet.New()) {
			return pet.Clone(), true
		}
	}
	return Pet{}, false
}

// AddPet adds one new pet and rejects nil receivers or persisted duplicates.
func (owner *Owner) AddPet(pet Pet) error {
	if owner == nil {
		return errors.New("add pet: owner is nil")
	}
	if !pet.New() {
		return errors.New("add pet: pet is already persisted")
	}
	owner.Pets = append(owner.Pets, pet.Clone())
	slices.SortStableFunc(owner.Pets, owner.comparePets)
	return nil
}

// AddVisit adds a visit to an existing pet.
func (owner *Owner) AddVisit(
	petID model.ID,
	visit Visit,
) error {
	if owner == nil {
		return errors.New("add visit: owner is nil")
	}
	for index := range owner.Pets {
		if owner.Pets[index].ID == petID && petID.Valid() {
			return owner.Pets[index].AddVisit(visit)
		}
	}
	return errors.New("add visit: pet was not found")
}

func (owner *Owner) appendRequired(
	violations []validation.Violation,
	field string,
	value string,
) []validation.Violation {
	if strings.TrimSpace(value) != "" {
		return violations
	}
	return append(violations, validation.Violation{
		Field:   field,
		Code:    "required",
		Message: "must not be blank",
	})
}

func (owner *Owner) validTelephone(value string) bool {
	if len(value) != 10 {
		return false
	}
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}

func (owner *Owner) comparePets(left, right Pet) int {
	if compared := strings.Compare(
		strings.ToLower(left.Name),
		strings.ToLower(right.Name),
	); compared != 0 {
		return compared
	}
	switch {
	case left.ID < right.ID:
		return -1
	case left.ID > right.ID:
		return 1
	default:
		return 0
	}
}
