package owner

import "github.com/spice-framework/petclinic/presentation"

// OwnerDetailsModel renders one complete owner aggregate.
type OwnerDetailsModel struct {
	presentation.Page
	Owner Owner
}
