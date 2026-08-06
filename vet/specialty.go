// @import { Module } from "github.com/spice-framework/spice/annotation/modulith"

// Package vet owns Petclinic veterinarians and specialties.
// @Module(allowedDependencies=["github.com/spice-framework/petclinic/model", "github.com/spice-framework/petclinic/presentation"])
package vet

import "github.com/spice-framework/petclinic/model"

// Specialty identifies an area of veterinary expertise.
type Specialty struct {
	model.NamedEntity
}
