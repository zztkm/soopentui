// Command quit-tui builds examples/quit-tui with a statically linked OpenTUI.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zztkm/soopentui/internal/build"
)

const exampleRel = "examples/quit-tui"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	outFlag := flag.String("o", "", "output binary (default: examples/quit-tui/quit-tui or ./quit-tui)")
	skipLib := flag.Bool("skip-lib", false, "do not build libopentui.a if missing")
	runApp := flag.Bool("run", false, "run the binary after a successful build")
	flag.Parse()

	modRoot, err := build.FindModuleRoot()
	if err != nil {
		return err
	}
	workRoot, err := build.WorkRootForModule(modRoot)
	if err != nil {
		return err
	}

	exampleDir := filepath.Join(modRoot, exampleRel)
	if _, err := os.Stat(filepath.Join(exampleDir, "main.go")); err != nil {
		return fmt.Errorf("example not found: %s", exampleDir)
	}

	out := *outFlag
	if out == "" {
		if build.InModuleCache(modRoot) {
			out = filepath.Join(workRoot, "quit-tui")
		} else {
			out = filepath.Join(exampleDir, "quit-tui")
		}
	}

	return build.Build(build.Options{
		PackageDir: exampleDir,
		Out:        out,
		SkipLib:    *skipLib,
		Run:        *runApp,
		WorkRoot:   workRoot,
	})
}
