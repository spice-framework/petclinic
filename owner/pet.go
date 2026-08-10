package owner

import (
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/spice-framework/petclinic/model"
	"github.com/spice-framework/spice/validation"
)

// Pet is an animal owned by one Petclinic owner.
type Pet struct {
	model.NamedEntity
	BirthDate time.Time
	Type      PetType
	Visits    []Visit
}

// Clone returns a deep value copy.
func (pet *Pet) Clone() Pet {
	if pet == nil {
		return Pet{}
	}
	result := *pet
	result.Visits = slices.Clone(pet.Visits)
	return result
}

// Validate returns stable Petclinic pet constraints.
func (pet *Pet) Validate(today time.Time) (validation.Errors, error) {
	if pet == nil {
		return validation.New(validation.Violation{
			Code:    "required",
			Message: "pet must be provided",
		})
	}
	var violations []validation.Violation
	if strings.TrimSpace(pet.Name) == "" {
		violations = append(violations, validation.Violation{
			Field:   "name",
			Code:    "required",
			Message: "must not be blank",
		})
	}
	if pet.BirthDate.IsZero() {
		violations = append(violations, validation.Violation{
			Field:   "birthDate",
			Code:    "required",
			Message: "must be provided",
		})
	} else if pet.BirthDate.After(today) {
		violations = append(violations, validation.Violation{
			Field:   "birthDate",
			Code:    "future",
			Message: "must not be in the future",
		})
	}
	if !pet.Type.ID.Valid() || strings.TrimSpace(pet.Type.Name) == "" {
		violations = append(violations, validation.Violation{
			Field:   "type",
			Code:    "required",
			Message: "must identify a known pet type",
		})
	}
	return validation.New(violations...)
}

// AddVisit adds one new visit and keeps visits ordered by date then identity.
func (pet *Pet) AddVisit(visit Visit) error {
	if pet == nil {
		return errors.New("add visit: pet is nil")
	}
	pet.Visits = append(pet.Visits, visit)
	slices.SortStableFunc(pet.Visits, pet.compareVisits)
	return nil
}

func (pet *Pet) compareVisits(left, right Visit) int {
	if compared := left.Date.Compare(right.Date); compared != 0 {
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
