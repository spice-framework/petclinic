// @import { Module } from "github.com/spice-framework/spice/annotation/modulith"

// Package presentation owns Petclinic's HTTP presentation infrastructure.
// @Module
package presentation

import "net/http"

// NewMux returns an instance-owned Go 1.22 pattern-aware HTTP mux with
// immutable embedded assets.
func NewMux() (*http.ServeMux, error) {
	mux := http.NewServeMux()
	assets, err := staticHandler()
	if err != nil {
		return nil, err
	}
	mux.Handle("GET /resources/", http.StripPrefix("/resources/", assets))
	return mux, nil
}
