// Command build translates a Solod package and links it against libopentui.a.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/zztkm/soopentui/internal/build"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	outFlag := flag.String("o", "", "output binary (default: ./<package-basename>)")
	skipLib := flag.Bool("skip-lib", false, "do not build libopentui.a if missing")
	runApp := flag.Bool("run", false, "run the binary after a successful build")
	flag.Parse()

	pkg := "."
	if args := flag.Args(); len(args) > 0 {
		pkg = args[0]
	}
	if len(flag.Args()) > 1 {
		return fmt.Errorf("usage: build [flags] [package-dir]")
	}

	return build.Build(build.Options{
		PackageDir: pkg,
		Out:        *outFlag,
		SkipLib:    *skipLib,
		Run:        *runApp,
	})
}
