package main

import (
	"crypto/sha256"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

type stylePolicy struct {
	SchemaVersion             int                      `json:"schemaVersion"`
	Profile                   string                   `json:"profile"`
	AllowedBoundaryFiles      []string                 `json:"allowedBoundaryFiles"`
	PackageVariableExceptions []styleVariableException `json:"packageVariableExceptions"`
}

type styleVariableException struct {
	Glob   string `json:"glob"`
	Symbol string `json:"symbol"`
	Type   string `json:"type"`
	Issue  string `json:"issue"`
}

func TestParseTotalCoverage(t *testing.T) {
	t.Parallel()
	percentage, err := parseTotalCoverage("file.go:1:\tValue\t100.0%\ntotal:\t(statements)\t89.3%\n")
	if err != nil || percentage != 89.3 {
		t.Fatalf("parseTotalCoverage() = %v, %v", percentage, err)
	}
	for _, invalid := range []string{"", "total: no-percentage", "total: invalid%"} {
		if _, err := parseTotalCoverage(invalid); err == nil {
			t.Errorf("parseTotalCoverage(%q) error = nil", invalid)
		}
	}
}

func TestCheckIdentityRejectsLegacyPathAndReplace(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		goMod   string
		extra   map[string]string
		wantErr string
	}{
		{
			name:  "canonical",
			goMod: "module " + modulePath + "\n\ngo 1.26.0\n",
		},
		{
			name:  "canonical CRLF checkout",
			goMod: "module " + modulePath + "\r\n\r\ngo 1.26.0\r\n",
		},
		{
			name:    "replace",
			goMod:   "module " + modulePath + "\n\ngo 1.26.0\n\nreplace example.com/a => ../a\n",
			wantErr: "replace directives",
		},
		{
			name:    "legacy source",
			goMod:   "module " + modulePath + "\n\ngo 1.26.0\n",
			extra:   map[string]string{"application.go": "package app\n// " + legacyModulePath + "\n"},
			wantErr: "legacy monorepo",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeTestFile(t, root, "go.mod", test.goMod)
			for name, content := range test.extra {
				writeTestFile(t, root, name, content)
			}
			err := checkIdentity(root)
			if test.wantErr == "" && err != nil {
				t.Fatalf("checkIdentity() error = %v", err)
			}
			if test.wantErr != "" && (err == nil || !strings.Contains(err.Error(), test.wantErr)) {
				t.Fatalf("checkIdentity() error = %v, want %q", err, test.wantErr)
			}
		})
	}
}

func TestBatchesAndDigestEquality(t *testing.T) {
	t.Parallel()
	values := []string{"a", "b", "c", "d", "e"}
	got := batches(values, 2)
	want := [][]string{{"a", "b"}, {"c", "d"}, {"e"}}
	if !slices.EqualFunc(got, want, slices.Equal) {
		t.Fatalf("batches() = %#v, want %#v", got, want)
	}
	digest := sha256.Sum256([]byte("value"))
	if !equalDigests(
		map[string][sha256.Size]byte{"file": digest},
		map[string][sha256.Size]byte{"file": digest},
	) {
		t.Fatal("equalDigests() = false for identical maps")
	}
	if equalDigests(
		map[string][sha256.Size]byte{"file": digest},
		map[string][sha256.Size]byte{"other": digest},
	) {
		t.Fatal("equalDigests() = true for different paths")
	}
}

func TestStylePolicyOwnsOnlyExactInteropBoundaries(t *testing.T) {
	t.Parallel()
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(root, ".spice", "style.json"))
	if err != nil {
		t.Fatal(err)
	}
	var policy stylePolicy
	if err := json.Unmarshal(content, &policy); err != nil {
		t.Fatal(err)
	}
	if policy.SchemaVersion != 1 || policy.Profile != "java-structured" {
		t.Fatalf("style policy identity = %d/%q", policy.SchemaVersion, policy.Profile)
	}
	for _, required := range []string{
		"mysql/owner_sql.go",
		"owner/errors.go",
		"postgres/owner_sql.go",
		"postgres/schema.go",
	} {
		if !slices.Contains(policy.AllowedBoundaryFiles, required) {
			t.Errorf("style boundary %q is missing", required)
		}
	}
	want := map[string]string{
		"presentation/assets.go/staticFiles":   "embed.FS/PETCLINIC-STYLE-001",
		"presentation/catalog.go/messageFiles": "embed.FS/PETCLINIC-STYLE-001",
		"presentation/renderer.go/templates":   "embed.FS/PETCLINIC-STYLE-001",
		"owner/errors.go/ErrOwnerNotFound":     "error/PETCLINIC-STYLE-002",
		"owner/errors.go/ErrPetNotFound":       "error/PETCLINIC-STYLE-002",
	}
	if len(policy.PackageVariableExceptions) != len(want) {
		t.Fatalf(
			"package-variable exception count = %d, want %d",
			len(policy.PackageVariableExceptions),
			len(want),
		)
	}
	for _, exception := range policy.PackageVariableExceptions {
		key := exception.Glob + "/" + exception.Symbol
		if got, exists := want[key]; !exists || got != exception.Type+"/"+exception.Issue {
			t.Errorf("unexpected package-variable exception %#v", exception)
		}
	}
}

func writeTestFile(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
