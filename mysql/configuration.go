package mysql

import "github.com/spice-framework/spice/lifecycle"

// @import { Bean, Configuration } from "github.com/spice-framework/spice/annotation/core"

// Configuration owns the MySQL target's infrastructure factories.
// @Configuration
type Configuration struct{}

// Database supplies the lifecycle-managed MySQL connection pool.
// @Bean
func (*Configuration) Database(
	settings Settings,
) (*Database, lifecycle.Cleanup, error) {
	return NewDatabase(settings)
}
