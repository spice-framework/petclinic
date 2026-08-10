package memory

// @import { Bean, Configuration } from "github.com/spice-framework/spice/annotation/core"

// Configuration owns the in-memory target's application object factories.
// @Configuration
type Configuration struct{}

// PetclinicDatabase supplies the deterministic reference data set.
// @Bean
func (*Configuration) PetclinicDatabase() (*Database, error) {
	return NewSeed().Database()
}
