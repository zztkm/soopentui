// Command hello-tui builds examples/hello-tui with a statically linked OpenTUI.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const (
	modulePath = "github.com/zztkm/soopentui"
	exampleRel = "examples/hello-tui"
	includeRel = "include"
	patchRel   = "patches/opentui-static-linkage.patch"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	outFlag := flag.String("o", "", "output binary (default: examples/hello-tui/hello-tui or ./hello-tui)")
	skipLib := flag.Bool("skip-lib", false, "do not build libopentui.a if missing")
	runApp := flag.Bool("run", false, "run the binary after a successful build")
	flag.Parse()

	modRoot, err := findModuleRoot()
	if err != nil {
		return err
	}
	workRoot, err := workRoot(modRoot)
	if err != nil {
		return err
	}

	exampleDir := filepath.Join(modRoot, exampleRel)
	if _, err := os.Stat(filepath.Join(exampleDir, "main.go")); err != nil {
		return fmt.Errorf("example not found: %s", exampleDir)
	}

	out := *outFlag
	if out == "" {
		if inModuleCache(modRoot) {
			out = filepath.Join(workRoot, "hello-tui")
		} else {
			out = filepath.Join(exampleDir, "hello-tui")
		}
	}
	if !filepath.IsAbs(out) {
		out = filepath.Join(workRoot, out)
	}

	libPath, err := opentuiStaticLibPath(workRoot)
	if err != nil {
		return err
	}
	if !fileExists(libPath) {
		if *skipLib {
			return fmt.Errorf("OpenTUI static library not found: %s", libPath)
		}
		fmt.Println("building static OpenTUI...")
		if err := runCmd(workRoot, "go", "run", filepath.Join(modRoot, "cmd", "opentui-static")); err != nil {
			return err
		}
		if !fileExists(libPath) {
			return fmt.Errorf("expected %s after opentui-static build", libPath)
		}
	}

	if _, err := exec.LookPath("so"); err != nil {
		return errors.New("so not found in PATH (install with: go install solod.dev/cmd/so@main)")
	}
	if _, err := exec.LookPath("zig"); err != nil {
		return errors.New("zig not found in PATH")
	}

	includeDir, err := soopentuiIncludeDir(exampleDir, modRoot)
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "hello-tui-build-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	fmt.Println("translating So -> C...")
	if err := runCmd(exampleDir, "so", "translate", "-o", tmpDir, "."); err != nil {
		return fmt.Errorf("so translate: %w", err)
	}

	cFiles, err := findCFiles(tmpDir)
	if err != nil {
		return err
	}
	if len(cFiles) == 0 {
		return fmt.Errorf("no generated .c files in %s", tmpDir)
	}

	fmt.Printf("linking %d C files + libopentui.a -> %s\n", len(cFiles), out)
	if err := link(tmpDir, includeDir, libPath, cFiles, out); err != nil {
		return err
	}

	info, err := os.Stat(out)
	if err != nil {
		return err
	}
	fmt.Printf("OK: built %s (%s)\n", out, humanSize(info.Size()))

	if *runApp {
		fmt.Println("running...")
		cmd := exec.Command(out)
		cmd.Dir = workRoot
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return nil
}

// soopentuiIncludeDir resolves include/ via `go list -m` (works with replace and module cache).
func soopentuiIncludeDir(exampleDir, fallbackModRoot string) (string, error) {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", modulePath)
	cmd.Dir = exampleDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		includeDir := filepath.Join(fallbackModRoot, includeRel)
		if fileExists(filepath.Join(includeDir, "opentui.h")) {
			return includeDir, nil
		}
		return "", fmt.Errorf("go list -m %s: %w\n%s", modulePath, err, strings.TrimSpace(stderr.String()))
	}
	dir := strings.TrimSpace(stdout.String())
	if dir == "" {
		return "", fmt.Errorf("go list -m %s: empty Dir", modulePath)
	}
	includeDir := filepath.Join(dir, includeRel)
	if !fileExists(filepath.Join(includeDir, "opentui.h")) {
		return "", fmt.Errorf("opentui.h not found in %s", includeDir)
	}
	return includeDir, nil
}

func link(tmpDir, includeDir, libPath string, cFiles []string, out string) error {
	args := []string{
		"cc", "-O2",
		"-I" + tmpDir,
		"-I" + includeDir,
		"-DSO_PANIC_MODE=SO_PANIC_EXIT",
	}
	args = append(args, cFiles...)
	args = append(args, libPath)

	env := os.Environ()
	switch runtime.GOOS {
	case "darwin":
		sdk, err := macosSDK()
		if err != nil {
			return err
		}
		args = append(args,
			"--sysroot", sdk,
			"-F"+filepath.Join(sdk, "System", "Library", "Frameworks"),
			"-lc++", "-lpthread",
			"-framework", "CoreFoundation",
			"-framework", "CoreAudio",
			"-framework", "AudioToolbox",
			"-o", out,
		)
		if os.Getenv("OPENTUI_KEEP_DEVELOPER_DIR") == "" {
			env = setEnv(env, "DEVELOPER_DIR", "/dev/null")
		}
		env = setEnv(env, "SDKROOT", sdk)
	case "linux":
		args = append(args, "-lc++", "-ldl", "-lpthread", "-lm", "-o", out)
	default:
		return fmt.Errorf("unsupported GOOS %s", runtime.GOOS)
	}

	cmd := exec.Command("zig", args...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("zig cc: %w", err)
	}
	return nil
}

func opentuiStaticLibPath(workRoot string) (string, error) {
	arch, osName, err := opentuiPlatform()
	if err != nil {
		return "", err
	}
	return filepath.Join(
		workRoot, "_build", "opentui", "packages", "core", "src", "zig", "lib",
		arch+"-"+osName+"-static", "libopentui.a",
	), nil
}

func opentuiPlatform() (arch, osName string, err error) {
	switch runtime.GOARCH {
	case "arm64":
		arch = "aarch64"
	case "amd64":
		arch = "x86_64"
	default:
		return "", "", fmt.Errorf("unsupported GOARCH %s", runtime.GOARCH)
	}
	switch runtime.GOOS {
	case "darwin":
		osName = "macos"
	case "linux":
		osName = "linux"
	default:
		return "", "", fmt.Errorf("unsupported GOOS %s", runtime.GOOS)
	}
	return arch, osName, nil
}

func findCFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".c") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func findModuleRoot() (string, error) {
	if root, err := moduleRootFromCaller(); err == nil {
		return root, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		if isModuleRoot(dir) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("soopentui module root not found (go.mod + patches/)")
		}
		dir = parent
	}
}

func moduleRootFromCaller() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	if !isModuleRoot(root) {
		return "", fmt.Errorf("not a module root: %s", root)
	}
	return root, nil
}

func isModuleRoot(dir string) bool {
	return fileExists(filepath.Join(dir, "go.mod")) && fileExists(filepath.Join(dir, patchRel))
}

func workRoot(modRoot string) (string, error) {
	if inModuleCache(modRoot) {
		return os.Getwd()
	}
	return modRoot, nil
}

func inModuleCache(path string) bool {
	cache := os.Getenv("GOMODCACHE")
	if cache == "" {
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return false
			}
			gopath = filepath.Join(home, "go")
		}
		cache = filepath.Join(filepath.SplitList(gopath)[0], "pkg", "mod")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	cacheAbs, err := filepath.Abs(cache)
	if err != nil {
		return false
	}
	sep := string(os.PathSeparator)
	return abs == cacheAbs || strings.HasPrefix(abs, cacheAbs+sep)
}

func macosSDK() (string, error) {
	if sdk := os.Getenv("SDKROOT"); sdk != "" {
		return sdk, nil
	}
	out, err := exec.Command("xcrun", "--show-sdk-path").Output()
	if err != nil {
		return "", fmt.Errorf("macOS SDK not found (set SDKROOT or install Xcode CLT): %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func runCmd(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func setEnv(env []string, key, value string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env)+1)
	found := false
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			out = append(out, prefix+value)
			found = true
			continue
		}
		out = append(out, e)
	}
	if !found {
		out = append(out, prefix+value)
	}
	return out
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func humanSize(n int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
	)
	switch {
	case n >= mb:
		return fmt.Sprintf("%.1fM", float64(n)/mb)
	case n >= kb:
		return fmt.Sprintf("%.1fK", float64(n)/kb)
	default:
		return fmt.Sprintf("%dB", n)
	}
}
