// @import { Module } from "github.com/spice-framework/spice/annotation/modulith"

// Package main assembles the PostgreSQL Petclinic application.
// @Module(allowedDependencies=["github.com/spice-framework/petclinic/model", "github.com/spice-framework/petclinic/owner", "github.com/spice-framework/petclinic/postgres", "github.com/spice-framework/petclinic/presentation", "github.com/spice-framework/petclinic/system", "github.com/spice-framework/petclinic/vet"])
package main

import (
	"os"

	spiceapp "github.com/spice-framework/petclinic/internal/spicegen/postgres"
	_ "github.com/spice-framework/petclinic/model"
	_ "github.com/spice-framework/petclinic/owner"
	_ "github.com/spice-framework/petclinic/postgres"
	_ "github.com/spice-framework/petclinic/presentation"
	_ "github.com/spice-framework/petclinic/system"
	_ "github.com/spice-framework/petclinic/vet"
)

// @import { Application } from "github.com/spice-framework/spice/annotation/core"
// @import { Enable } from "github.com/spice-framework/spice/annotation/management"
// @import { Logging } from "github.com/spice-framework/spice/annotation/observability"

// @Application
// @Enable(expose=["health", "liveness", "readiness", "info", "metrics", "configprops", "modules"], access="loopback")
// @Logging
func main() {
	// This package-main boundary alone owns the process status.
	os.Exit(spiceapp.Main(os.Args[1:]))
}
