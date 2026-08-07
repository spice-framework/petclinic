package presentation

import (
	"net/http"

	"github.com/spice-framework/spice/i18n"
	"github.com/spice-framework/spice/view"
)

// @import { Bean, Configuration, Fallback } from "github.com/spice-framework/spice/annotation/core"

// Configuration owns Petclinic's presentation infrastructure factories.
// @Configuration
type Configuration struct{}

// Catalog supplies the immutable localized-message catalog.
// @Bean
func (*Configuration) Catalog() (*i18n.Catalog, error) {
	return NewCatalog()
}

// Mux supplies the application's fallback HTTP request multiplexer.
// @Bean
// @Fallback
func (*Configuration) Mux() (*http.ServeMux, error) {
	return NewMux()
}

// Renderer supplies the immutable view renderer.
// @Bean
func (*Configuration) Renderer(
	catalog *i18n.Catalog,
) (*view.Renderer, error) {
	return NewRenderer(catalog)
}

// Server supplies the lifecycle-managed HTTP server.
// @Bean
func (*Configuration) Server(
	settings ServerSettings,
	handler *http.ServeMux,
) (*Server, error) {
	return NewServer(settings, handler)
}
