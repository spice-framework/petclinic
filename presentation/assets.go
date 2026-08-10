package presentation

import (
	"context"
	"embed"
	"fmt"
	"net/http"

	"github.com/spice-framework/spice/resource"
)

//go:embed static/*
var staticFiles embed.FS

// Assets owns Petclinic's immutable embedded static-file handler.
type Assets struct {
	handler http.Handler
}

// NewAssets constructs the bounded static-file handler.
func NewAssets() (*Assets, error) {
	loader, err := resource.NewLoader(resource.Mount{
		Name: "petclinic",
		FS:   staticFiles,
	})
	if err != nil {
		return nil, fmt.Errorf("construct Petclinic resource loader: %w", err)
	}
	staticRoot, err := loader.Resolve("spice://petclinic/static")
	if err != nil {
		return nil, fmt.Errorf("resolve Petclinic static resource: %w", err)
	}
	source, err := staticRoot.Sub(context.Background())
	if err != nil {
		return nil, fmt.Errorf("construct Petclinic static filesystem: %w", err)
	}
	files := http.FileServer(http.FS(source))
	handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Cache-Control", "public, max-age=3600")
		writer.Header().Set("X-Content-Type-Options", "nosniff")
		files.ServeHTTP(writer, request)
	})
	return &Assets{handler: handler}, nil
}

// Handler returns the immutable static-file handler.
func (assets *Assets) Handler() http.Handler {
	if assets == nil {
		return nil
	}
	return assets.handler
}
