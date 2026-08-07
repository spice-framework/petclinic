package presentation

import (
	"embed"

	"github.com/spice-framework/spice/i18n"
)

//go:embed messages/*.properties
var messageFiles embed.FS

// NewCatalog constructs the immutable Petclinic message catalog.
func NewCatalog() (*i18n.Catalog, error) {
	return i18n.ParseProperties(messageFiles, "messages/*.properties", "en")
}
