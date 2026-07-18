// Package build translates a Solod package and links it against libopentui.a.
package build

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const (
	ModulePath = "github.com/zztkm/soopentui"
	includeRel = "include"
	patchRel   = "patches/opentui-static-linkage.patch"
)

// Options configures a package build.
type Options struct {
	// PackageDir is the Solod package to translate (absolute or relative to cwd).
	PackageDir string
	// Out is the output binary path. Empty uses WorkRoot/<basename of PackageDir>.
	Out string
	// SkipLib skips building libopentui.a when missing.
	SkipLib bool
	// Run runs the binary after a successful build.
	Run bool
	// WorkRoot is where _build/ (libopentui.a) is written. Empty uses the process cwd.
	WorkRoot string
}

// Build translates PackageDir with so and links the result with libopentui.a.
func Build(opts Options) error {
	pkgDir := opts.PackageDir
	if pkgDir == "" {
		pkgDir = "."
	}
	pkgDir, err := filepath.Abs(pkgDir)
	if err != nil {
		return err
	}
	if _, err := os.Stat(pkgDir); err != nil {
		return fmt.Errorf("package dir: %w", err)
	}

	workRoot := opts.WorkRoot
	if workRoot == "" {
		workRoot, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	workRoot, err = filepath.Abs(workRoot)
	if err != nil {
		return err
	}

	out := opts.Out
	if out == "" {
		name := filepath.Base(pkgDir)
		if name == "" || name == "." || name == string(filepath.Separator) {
			name = "app"
		}
		out = filepath.Join(workRoot, name)
	}
	if !filepath.IsAbs(out) {
		out = filepath.Join(workRoot, out)
	}

	modRoot, err := SoopentuiDir(pkgDir)
	if err != nil {
		return err
	}

	libPath, err := opentuiStaticLibPath(workRoot)
	if err != nil {
		return err
	}
	if !fileExists(libPath) {
		if opts.SkipLib {
			return fmt.Errorf("OpenTUI static library not found: %s", libPath)
		}
		fmt.Println("building static OpenTUI...")
		staticCmd := filepath.Join(modRoot, "cmd", "opentui-static")
		if err := runCmd(workRoot, "go", "run", staticCmd); err != nil {
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

	includeDir, err := includeDir(pkgDir, modRoot)
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "soopentui-build-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	fmt.Println("translating So -> C...")
	if err := runCmd(pkgDir, "so", "translate", "-o", tmpDir, "."); err != nil {
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

	if opts.Run {
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

// SoopentuiDir resolves the soopentui module directory via `go list -m`.
func SoopentuiDir(fromDir string) (string, error) {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", ModulePath)
	cmd.Dir = fromDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if root, ferr := FindModuleRoot(); ferr == nil {
			return root, nil
		}
		return "", fmt.Errorf("go list -m %s: %w\n%s", ModulePath, err, strings.TrimSpace(stderr.String()))
	}
	dir := strings.TrimSpace(stdout.String())
	if dir == "" {
		return "", fmt.Errorf("go list -m %s: empty Dir", ModulePath)
	}
	return dir, nil
}

// FindModuleRoot locates the soopentui module (go.mod + patches/), preferring
// the source tree that contains this package so `go run .../cmd/...` works.
func FindModuleRoot() (string, error) {
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

// WorkRootForModule returns where _build/ should be written for a soopentui checkout.
// Module cache is often read-only, so use cwd then.
func WorkRootForModule(modRoot string) (string, error) {
	if InModuleCache(modRoot) {
		return os.Getwd()
	}
	return modRoot, nil
}

// InModuleCache reports whether path is under GOMODCACHE.
func InModuleCache(path string) bool {
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

func includeDir(pkgDir, fallbackModRoot string) (string, error) {
	dir, err := SoopentuiDir(pkgDir)
	if err != nil {
		include := filepath.Join(fallbackModRoot, includeRel)
		if fileExists(filepath.Join(include, "opentui.h")) {
			return include, nil
		}
		return "", err
	}
	include := filepath.Join(dir, includeRel)
	if !fileExists(filepath.Join(include, "opentui.h")) {
		return "", fmt.Errorf("opentui.h not found in %s", include)
	}
	return include, nil
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

func moduleRootFromCaller() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("runtime.Caller failed")
	}
	// internal/build/build.go -> module root
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	if !isModuleRoot(root) {
		return "", fmt.Errorf("not a module root: %s", root)
	}
	return root, nil
}

func isModuleRoot(dir string) bool {
	return fileExists(filepath.Join(dir, "go.mod")) && fileExists(filepath.Join(dir, patchRel))
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
