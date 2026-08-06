// @import { Module } from "github.com/spice-framework/spice/annotation/modulith"

// Package model contains the shared Petclinic domain building blocks.
// @Module
package model

// BaseEntity gives persistent Petclinic entities a stable identity. A zero ID
// identifies an entity that has not been persisted.
type BaseEntity struct {
	ID ID `json:"id"`
}

// New reports whether the entity has not yet been persisted.
func (entity BaseEntity) New() bool {
	return entity.ID == 0
}
