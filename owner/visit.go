package owner

import (
	"strings"
	"time"

	"github.com/spice-framework/petclinic/model"
	"github.com/spice-framework/spice/validation"
)

// Visit records veterinary care for one pet.
type Visit struct {
	model.BaseEntity
	Date        time.Time
	Description string
}

// Validate returns stable visit constraints.
func (visit Visit) Validate() (validation.Errors, error) {
	var violations []validation.Violation
	if visit.Date.IsZero() {
		violations = append(violations, validation.Violation{
			Field:   "date",
			Code:    "required",
			Message: "must be provided",
		})
	}
	if strings.TrimSpace(visit.Description) == "" {
		violations = append(violations, validation.Violation{
			Field:   "description",
			Code:    "required",
			Message: "must not be blank",
		})
	}
	return validation.New(violations...)
}
