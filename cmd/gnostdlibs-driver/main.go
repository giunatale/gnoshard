// gnostdlibs-driver answers the one go/packages query x/gnovm makes, at
// InitGenesis, for the directory of the Gno standard library. x/gnovm locates
// the library by asking the Go toolchain where the gno module's stdlibs
// package is, which a node image without a Go toolchain cannot answer. With
// GOPACKAGESDRIVER pointing at this program and GNO_STDLIBS_DIR at a copy of
// the library, the query is answered from the environment instead.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

const (
	stdlibsPkgPath = "github.com/gnolang/gno/gnovm/stdlibs"
	dirEnvVar      = "GNO_STDLIBS_DIR"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gnostdlibs-driver:", err)
		os.Exit(1)
	}
}

func run() error {
	dir := os.Getenv(dirEnvVar)
	if dir == "" {
		return fmt.Errorf("%s is not set", dirEnvVar)
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(dir, "stdlibs.go")); err != nil {
		return fmt.Errorf("%s does not hold the Gno standard library: %w", dir, err)
	}

	// The request on stdin carries the load mode and environment; the
	// patterns are the arguments. Only the stdlibs package is served.
	var req packages.DriverRequest
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		return fmt.Errorf("reading the driver request: %w", err)
	}
	for _, pattern := range os.Args[1:] {
		if pattern != stdlibsPkgPath {
			return fmt.Errorf("package %q is not served; only %s is", pattern, stdlibsPkgPath)
		}
	}

	goFiles, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return err
	}
	resp := packages.DriverResponse{
		Roots: []string{stdlibsPkgPath},
		Packages: []*packages.Package{{
			ID:              stdlibsPkgPath,
			Name:            "stdlibs",
			PkgPath:         stdlibsPkgPath,
			Dir:             dir,
			GoFiles:         goFiles,
			CompiledGoFiles: goFiles,
		}},
	}
	return json.NewEncoder(os.Stdout).Encode(resp)
}
