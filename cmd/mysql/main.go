package main

import (
	"os"

	spiceapp "github.com/spice-framework/petclinic/internal/spicegen/mysql"
	_ "github.com/spice-framework/petclinic/mysql"
	_ "github.com/spice-framework/petclinic/owner"
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
