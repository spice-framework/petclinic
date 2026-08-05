// Command qualitygate runs Petclinic's repository-owned cross-platform checks.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	requiredGoVersion = "go1.26.5"
	modulePath        = "github.com/spice-framework/petclinic"
	legacyModulePath  = "github.com/spice-framework/spice/examples/" + "petclinic"
	minimumCoverage   = 85.0
	spiceTool         = "github.com/spice-framework/toolchain/cmd/spice"
)

var output = log.New(os.Stdout, "", 0)

func main() {
	os.Exit(execute()) // Entrypoint exception: propagate verification failure.
}

func execute() int {
	mode := flag.String(
		"mode",
		"verify",
		"verification mode: check, fmt, lint, security, smoke, test, or verify",
	)
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	root, err := repositoryRoot()
	if err == nil {
		err = run(ctx, root, *mode)
	}
	if err != nil {
		output.Printf("quality gate failed: %v", err)
		return 1
	}
	return 0
}

type verificationStep struct {
	name string
	run  func() error
}

func run(ctx context.Context, root, mode string) error {
	if err := checkGoVersion(); err != nil {
		return err
	}
	switch mode {
	case "check":
		return runSequential([]verificationStep{
			{"repository identity", func() error { return checkIdentity(root) }},
			{"formatting", func() error { return format(ctx, root, false) }},
			{"module and vendor", func() error { return checkModule(ctx, root) }},
			{"go vet", func() error { return command(ctx, root, nil, "go", "vet", "./...") }},
		})
	case "fmt":
		return format(ctx, root, true)
	case "lint":
		return runSequential([]verificationStep{
			{"lint and nil safety", func() error { return lint(ctx, root) }},
		})
	case "security":
		return runSequential([]verificationStep{
			{"security", func() error { return security(ctx, root) }},
		})
	case "smoke":
		return runSequential([]verificationStep{
			{"Spice targets", func() error { return spiceTargets(ctx, root) }},
		})
	case "test":
		return runSequential([]verificationStep{
			{"shuffled and race tests", func() error { return tests(ctx, root) }},
			{"business coverage", func() error { return coverage(ctx, root) }},
		})
	case "verify":
		return runSequential([]verificationStep{
			{"repository identity", func() error { return checkIdentity(root) }},
			{"formatting", func() error { return format(ctx, root, false) }},
			{"module and vendor", func() error { return checkModule(ctx, root) }},
			{"go vet", func() error { return command(ctx, root, nil, "go", "vet", "./...") }},
			{"lint and nil safety", func() error { return lint(ctx, root) }},
			{"security", func() error { return security(ctx, root) }},
			{"shuffled and race tests", func() error { return tests(ctx, root) }},
			{"business coverage", func() error { return coverage(ctx, root) }},
			{"offline vendor", func() error { return offline(ctx, root) }},
			{"Spice targets", func() error { return spiceTargets(ctx, root) }},
		})
	default:
		return fmt.Errorf("unknown mode %q", mode)
	}
}

func runSequential(steps []verificationStep) error {
	for _, step := range steps {
		started := time.Now()
		output.Printf("==> %s", step.name)
		if err := step.run(); err != nil {
			return fmt.Errorf(
				"%s (%s): %w",
				step.name,
				time.Since(started).Round(time.Millisecond),
				err,
			)
		}
		output.Printf(
			"<== %s passed in %s",
			step.name,
			time.Since(started).Round(time.Millisecond),
		)
	}
	output.Print("==> all verification passed")
	return nil
}

func checkGoVersion() error {
	if runtime.Version() != requiredGoVersion {
		return fmt.Errorf(
			"go version is %s; require exactly %s",
			runtime.Version(),
			requiredGoVersion,
		)
	}
	return nil
}

func checkIdentity(root string) error {
	opened, err := os.OpenRoot(root)
	if err != nil {
		return fmt.Errorf("open repository root: %w", err)
	}
	defer func() {
		if closeErr := opened.Close(); closeErr != nil {
			output.Printf("warning: close repository root: %v", closeErr)
		}
	}()
	goMod, err := opened.ReadFile("go.mod")
	if err != nil {
		return fmt.Errorf("read go.mod: %w", err)
	}
	content := string(goMod)
	if !hasModuleDirective(content, modulePath) {
		return fmt.Errorf("go.mod does not declare canonical module %s", modulePath)
	}
	if strings.Contains(content, "\nreplace ") || strings.Contains(content, "\nreplace (") {
		return errors.New("committed Petclinic go.mod must not contain replace directives")
	}
	return fs.WalkDir(opened.FS(), ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() && path != root && slices.Contains(
			[]string{".git", "tools", "vendor"},
			entry.Name(),
		) {
			return filepath.SkipDir
		}
		if entry.IsDir() || !identityFile(entry.Name()) {
			return nil
		}
		data, readErr := opened.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read identity file %q: %w", path, readErr)
		}
		if bytes.Contains(data, []byte(legacyModulePath)) {
			return fmt.Errorf("legacy monorepo module path remains in %s", path)
		}
		return nil
	})
}

func hasModuleDirective(content, expected string) bool {
	for line := range strings.SplitSeq(content, "\n") {
		if strings.TrimSpace(line) == "module "+expected {
			return true
		}
	}
	return false
}

func identityFile(name string) bool {
	switch filepath.Ext(name) {
	case ".go", ".json", ".md", ".mod", ".sum", ".yaml", ".yml":
		return true
	default:
		return false
	}
}

func format(ctx context.Context, root string, write bool) error {
	files, err := goFiles(root)
	if err != nil {
		return err
	}
	goimports, err := toolPath(ctx, root, "goimports")
	if err != nil {
		return err
	}
	gofumpt, err := toolPath(ctx, root, "gofumpt")
	if err != nil {
		return err
	}
	if write {
		if err := fileBatches(ctx, root, goimports, "-w", files); err != nil {
			return err
		}
		return fileBatches(ctx, root, gofumpt, "-w", files)
	}
	if err := checkFormatted(ctx, root, goimports, files); err != nil {
		return fmt.Errorf("goimports: %w", err)
	}
	if err := checkFormatted(ctx, root, gofumpt, files); err != nil {
		return fmt.Errorf("gofumpt: %w", err)
	}
	return nil
}

func goFiles(root string) ([]string, error) {
	var result []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() && path != root && slices.Contains(
			[]string{".git", "tools", "vendor"},
			entry.Name(),
		) {
			return filepath.SkipDir
		}
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".go" {
			result = append(result, path)
		}
		return nil
	})
	slices.Sort(result)
	return result, err
}

func checkFormatted(
	ctx context.Context,
	root, executable string,
	files []string,
) error {
	var changed []string
	for _, batch := range batches(files, 64) {
		stdout, err := capture(ctx, root, nil, executable, append([]string{"-l"}, batch...)...)
		if err != nil {
			return err
		}
		changed = append(changed, strings.Fields(stdout)...)
	}
	if len(changed) != 0 {
		return fmt.Errorf("files require formatting: %s", strings.Join(changed, ", "))
	}
	return nil
}

func fileBatches(
	ctx context.Context,
	root, executable, option string,
	files []string,
) error {
	for _, batch := range batches(files, 64) {
		if err := command(
			ctx,
			root,
			nil,
			executable,
			append([]string{option}, batch...)...,
		); err != nil {
			return err
		}
	}
	return nil
}

func batches(values []string, size int) [][]string {
	var result [][]string
	for len(values) != 0 {
		length := min(size, len(values))
		result = append(result, values[:length])
		values = values[length:]
	}
	return result
}

func checkModule(ctx context.Context, root string) error {
	if err := command(ctx, root, nil, "go", "mod", "tidy", "-diff"); err != nil {
		return err
	}
	if err := command(ctx, root, nil, "go", "-C", "tools", "mod", "tidy", "-diff"); err != nil {
		return err
	}
	temporary, err := os.MkdirTemp("", "spice-petclinic-vendor-")
	if err != nil {
		return fmt.Errorf("create vendor comparison directory: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(temporary); removeErr != nil {
			output.Printf("warning: remove vendor comparison directory: %v", removeErr)
		}
	}()
	candidate := filepath.Join(temporary, "vendor")
	if commandErr := command(ctx, root, nil, "go", "mod", "vendor", "-o", candidate); commandErr != nil {
		return commandErr
	}
	current, err := treeDigests(filepath.Join(root, "vendor"))
	if err != nil {
		return err
	}
	expected, err := treeDigests(candidate)
	if err != nil {
		return err
	}
	if !equalDigests(current, expected) {
		return errors.New("vendor differs from a fresh go mod vendor result")
	}
	return nil
}

func treeDigests(root string) (map[string][sha256.Size]byte, error) {
	opened, err := os.OpenRoot(root)
	if err != nil {
		return nil, fmt.Errorf("open digest root %q: %w", root, err)
	}
	defer func() {
		if closeErr := opened.Close(); closeErr != nil {
			output.Printf("warning: close digest root %q: %v", root, closeErr)
		}
	}()
	result := make(map[string][sha256.Size]byte)
	err = fs.WalkDir(opened.FS(), ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("vendor contains symbolic link %q", path)
		}
		if entry.IsDir() {
			return nil
		}
		content, readErr := opened.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		result[filepath.ToSlash(path)] = sha256.Sum256(content)
		return nil
	})
	return result, err
}

func equalDigests(left, right map[string][sha256.Size]byte) bool {
	if len(left) != len(right) {
		return false
	}
	for path, digest := range left {
		if right[path] != digest {
			return false
		}
	}
	return true
}

func lint(ctx context.Context, root string) error {
	golangci, err := toolPath(ctx, root, "golangci-lint")
	if err != nil {
		return err
	}
	if commandErr := command(ctx, root, nil, golangci, "run", "--timeout=10m"); commandErr != nil {
		return commandErr
	}
	nilaway, err := toolPath(ctx, root, "nilaway")
	if err != nil {
		return err
	}
	return command(ctx, root, nil, nilaway, "-include-pkgs="+modulePath, "./...")
}

func security(ctx context.Context, root string) error {
	gosec, err := toolPath(ctx, root, "gosec")
	if err != nil {
		return err
	}
	if commandErr := command(ctx, root, nil, gosec, "-quiet", "-exclude-generated", "./..."); commandErr != nil {
		return commandErr
	}
	govulncheck, err := toolPath(ctx, root, "govulncheck")
	if err != nil {
		return err
	}
	return command(ctx, root, nil, govulncheck, "./...")
}

func tests(ctx context.Context, root string) error {
	if err := command(
		ctx,
		root,
		nil,
		"go",
		"test",
		"-shuffle=on",
		"-count=1",
		"./...",
	); err != nil {
		return err
	}
	return command(ctx, root, nil, "go", "test", "-race", "-count=1", "./...")
}

func coverage(ctx context.Context, root string) (returnErr error) {
	profile, err := os.CreateTemp("", "spice-petclinic-coverage-*.out")
	if err != nil {
		return fmt.Errorf("create coverage profile: %w", err)
	}
	profilePath := profile.Name()
	if closeErr := profile.Close(); closeErr != nil {
		return fmt.Errorf("close coverage profile: %w", closeErr)
	}
	defer func() {
		returnErr = errors.Join(returnErr, os.Remove(profilePath))
	}()
	packages := []string{"memory", "model", "owner", "presentation", "system", "vet"}
	covered := make([]string, len(packages))
	tests := make([]string, len(packages))
	for index, name := range packages {
		covered[index] = modulePath + "/" + name
		tests[index] = "./" + name
	}
	arguments := []string{
		"test",
		"-covermode=atomic",
		"-coverpkg=" + strings.Join(covered, ","),
		"-coverprofile=" + profilePath,
	}
	arguments = append(arguments, tests...)
	if commandErr := command(ctx, root, nil, "go", arguments...); commandErr != nil {
		return commandErr
	}
	stdout, err := capture(ctx, root, nil, "go", "tool", "cover", "-func="+profilePath)
	if err != nil {
		return err
	}
	percentage, err := parseTotalCoverage(stdout)
	if err != nil {
		return err
	}
	output.Printf("business coverage %.1f%% (minimum %.1f%%)", percentage, minimumCoverage)
	if percentage < minimumCoverage {
		return fmt.Errorf("business coverage %.1f%% is below %.1f%%", percentage, minimumCoverage)
	}
	return nil
}

func parseTotalCoverage(report string) (float64, error) {
	lines := strings.Split(strings.TrimSpace(report), "\n")
	if len(lines) == 0 {
		return 0, errors.New("coverage report is empty")
	}
	fields := strings.Fields(lines[len(lines)-1])
	if len(fields) == 0 || !strings.HasSuffix(fields[len(fields)-1], "%") {
		return 0, errors.New("coverage report has no total percentage")
	}
	value := strings.TrimSuffix(fields[len(fields)-1], "%")
	percentage, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("parse coverage percentage %q: %w", value, err)
	}
	return percentage, nil
}

func offline(ctx context.Context, root string) error {
	environment := map[string]string{"GOFLAGS": "-mod=vendor"}
	if err := command(ctx, root, environment, "go", "test", "-count=1", "./..."); err != nil {
		return err
	}
	return command(ctx, root, environment, "go", "build", "-trimpath", "./...")
}

func spiceTargets(ctx context.Context, root string) error {
	// #nosec G101 -- these fixed port-1 URLs contain public, non-connecting
	// smoke-test identities; they are not production credentials or secrets.
	environment := map[string]string{
		"GOFLAGS":                                 "-mod=vendor",
		"SPICE_PETCLINIC_POSTGRES_URL":            "postgres://petclinic:petclinic@127.0.0.1:1/petclinic?sslmode=disable",
		"SPICE_PETCLINIC_POSTGRES_ALLOW_INSECURE": "true",
		"SPICE_PETCLINIC_MYSQL_URL":               "mysql://petclinic:petclinic@127.0.0.1:1/petclinic?tls=disable",
		"SPICE_PETCLINIC_MYSQL_ALLOW_INSECURE":    "true",
	}
	commands := [][]string{
		{"verify", "."},
		{"verify", "./cmd/postgres"},
		{"verify", "./cmd/mysql"},
		{"generate", "--check", "--target", "Petclinic", "."},
		{"generate", "--check", "--target", "Postgres", "./cmd/postgres"},
		{"generate", "--check", "--target", "Mysql", "./cmd/mysql"},
		{"run", "--target", "Petclinic", ".", "--", "-check"},
		{"run", "--target", "Postgres", "./cmd/postgres", "--", "-check"},
		{"run", "--target", "Mysql", "./cmd/mysql", "--", "-check"},
	}
	for _, arguments := range commands {
		if err := command(
			ctx,
			root,
			environment,
			"go",
			append([]string{"tool", spiceTool}, arguments...)...,
		); err != nil {
			return err
		}
	}
	return nil
}

func toolPath(ctx context.Context, root, name string) (string, error) {
	stdout, err := capture(ctx, root, nil, "go", "tool", "-C", "tools", "-n", name)
	if err != nil {
		return "", err
	}
	path := strings.TrimSpace(stdout)
	if path == "" {
		return "", fmt.Errorf("resolve tool %q: empty path", name)
	}
	return path, nil
}

func repositoryRoot() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	for {
		opened, openErr := os.OpenRoot(current)
		if openErr != nil {
			return "", fmt.Errorf("open repository candidate %q: %w", current, openErr)
		}
		content, readErr := opened.ReadFile("go.mod")
		closeErr := opened.Close()
		if closeErr != nil {
			return "", fmt.Errorf("close repository candidate %q: %w", current, closeErr)
		}
		if readErr == nil && bytes.Contains(content, []byte("module "+modulePath)) {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", errors.New("find Petclinic repository root: go.mod not found")
		}
		current = parent
	}
}

func command(
	ctx context.Context,
	directory string,
	environment map[string]string,
	executable string,
	arguments ...string,
) error {
	// #nosec G204,G702 -- executable and arguments are fixed repository-owned values.
	cmd := exec.CommandContext(ctx, executable, arguments...)
	cmd.Dir = directory
	cmd.Env = mergedEnvironment(environment)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", executable, strings.Join(arguments, " "), err)
	}
	return nil
}

func capture(
	ctx context.Context,
	directory string,
	environment map[string]string,
	executable string,
	arguments ...string,
) (string, error) {
	// #nosec G204,G702 -- executable and arguments are fixed repository-owned values.
	cmd := exec.CommandContext(ctx, executable, arguments...)
	cmd.Dir = directory
	cmd.Env = mergedEnvironment(environment)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf(
			"%s %s: %w\n%s",
			executable,
			strings.Join(arguments, " "),
			err,
			strings.TrimSpace(stderr.String()),
		)
	}
	return stdout.String(), nil
}

func mergedEnvironment(overrides map[string]string) []string {
	values := map[string]string{
		"GOWORK":      "off",
		"GOPROXY":     "off",
		"GOFLAGS":     "",
		"GOTOOLCHAIN": "local",
	}
	maps.Copy(values, overrides)
	result := make([]string, 0, len(os.Environ())+len(values))
	for _, entry := range os.Environ() {
		key, _, found := strings.Cut(entry, "=")
		if !found {
			continue
		}
		if _, replaced := values[strings.ToUpper(key)]; !replaced {
			result = append(result, entry)
		}
	}
	for key, value := range values {
		result = append(result, key+"="+value)
	}
	slices.Sort(result)
	return result
}
