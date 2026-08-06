package modules_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
)

const toolPackage = "github.com/spice-framework/toolchain/cmd/spice"

type moduleDocument struct {
	Schema             string       `json:"schema"`
	Modules            []moduleInfo `json:"modules"`
	Cycles             [][]string   `json:"cycles"`
	UnassignedPackages []string     `json:"unassigned_packages"`
}

type moduleInfo struct {
	ID string `json:"id"`
}

func TestApplicationTargetsHaveCompleteModuleGraphs(t *testing.T) {
	tests := []struct {
		name     string
		pkg      string
		expected []string
	}{
		{
			name: "memory",
			pkg:  ".",
			expected: []string{
				"github.com/spice-framework/petclinic",
				"github.com/spice-framework/petclinic/memory",
				"github.com/spice-framework/petclinic/model",
				"github.com/spice-framework/petclinic/owner",
				"github.com/spice-framework/petclinic/presentation",
				"github.com/spice-framework/petclinic/system",
				"github.com/spice-framework/petclinic/vet",
			},
		},
		{
			name: "mysql",
			pkg:  "./cmd/mysql",
			expected: []string{
				"github.com/spice-framework/petclinic/cmd/mysql",
				"github.com/spice-framework/petclinic/model",
				"github.com/spice-framework/petclinic/mysql",
				"github.com/spice-framework/petclinic/owner",
				"github.com/spice-framework/petclinic/presentation",
				"github.com/spice-framework/petclinic/system",
				"github.com/spice-framework/petclinic/vet",
			},
		},
		{
			name: "postgres",
			pkg:  "./cmd/postgres",
			expected: []string{
				"github.com/spice-framework/petclinic/cmd/postgres",
				"github.com/spice-framework/petclinic/model",
				"github.com/spice-framework/petclinic/owner",
				"github.com/spice-framework/petclinic/postgres",
				"github.com/spice-framework/petclinic/presentation",
				"github.com/spice-framework/petclinic/system",
				"github.com/spice-framework/petclinic/vet",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			document := loadModules(t, test.pkg)
			if document.Schema != "spice.modules/v1" {
				t.Fatalf("schema = %q, want spice.modules/v1", document.Schema)
			}
			if len(document.Cycles) != 0 {
				t.Fatalf("module cycles = %v, want none", document.Cycles)
			}
			if len(document.UnassignedPackages) != 0 {
				t.Fatalf("unassigned packages = %v, want none", document.UnassignedPackages)
			}

			actual := make([]string, 0, len(document.Modules))
			for _, module := range document.Modules {
				actual = append(actual, module.ID)
			}
			slices.Sort(actual)
			slices.Sort(test.expected)
			if !slices.Equal(actual, test.expected) {
				t.Fatalf("module IDs = %v, want %v", actual, test.expected)
			}
		})
	}
}

func loadModules(t *testing.T, pkg string) moduleDocument {
	t.Helper()

	repositoryRoot := filepath.Clean(filepath.Join("..", ".."))
	command := exec.Command(
		"go",
		"tool",
		toolPackage,
		"modules",
		"--format=json",
		pkg,
	)
	command.Dir = repositoryRoot
	command.Env = append(
		os.Environ(),
		"GOENV=off",
		"GOFLAGS=-mod=vendor",
		"GOPROXY=off",
		"GOSUMDB=off",
		"GOTOOLCHAIN=local",
		"GOWORK=off",
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("document modules for %s: %v\n%s", pkg, err, output)
	}

	var document moduleDocument
	if err := json.Unmarshal(output, &document); err != nil {
		t.Fatalf("decode module document for %s: %v\n%s", pkg, err, output)
	}
	return document
}
