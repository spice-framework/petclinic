package mysql

import spicestarter "github.com/spice-framework/spice/annotation/sdk/starter"

// Manifest returns MySQL starter compatibility and review metadata.
func Manifest() spicestarter.Manifest {
	return spicestarter.Must(spicestarter.Spec{
		Schema:    spicestarter.Schema,
		ID:        "github.com/spice-framework/starter-mysql",
		Version:   "0.1.0-dev",
		Module:    "github.com/spice-framework/starter-mysql",
		SpiceAPI:  spicestarter.APIVersion,
		MinimumGo: "1.26",
		License:   "Apache-2.0",
		Review:    "docs/dependency-review.md",
		Activation: spicestarter.Activation{
			Mode: spicestarter.ActivationExplicitConstructor,
			EntryPoints: []spicestarter.EntryPoint{
				{
					Package: "github.com/spice-framework/starter-mysql",
					Symbol:  "Open",
				},
			},
		},
		Capabilities: []string{
			"data.mysql",
			"data.sql",
		},
		Dependencies: []spicestarter.Dependency{
			{
				Module:  "github.com/go-sql-driver/mysql",
				Version: "v1.10.0",
				License: "MPL-2.0",
			},
		},
	})
}
