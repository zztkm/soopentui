// Command opentui-static clones OpenTUI under _build/, patches it for static
// linkage, and builds libopentui.a.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	patchRel      = "patches/opentui-static-linkage.patch"
	buildDirRel   = "_build"
	opentuiRel    = "_build/opentui"
	opentuiGitURL = "https://github.com/anomalyco/opentui.git"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	force := flag.Bool("force", false, "delete _build/opentui and re-clone before patch/build")
	opentuiFlag := flag.String("opentui", opentuiRel, "path to the OpenTUI checkout (relative to repo root or absolute)")
	skipPatch := flag.Bool("skip-patch", false, "do not apply the static-linkage patch")
	skipBuild := flag.Bool("skip-build", false, "clone/patch only; do not run zig build")
	optimize := flag.String("optimize", "ReleaseFast", "Zig optimize mode")
	flag.Parse()

	modRoot, err := findModuleRoot()
	if err != nil {
		return err
	}
	workRoot, err := workRoot(modRoot)
	if err != nil {
		return err
	}

	opentuiDir := *opentuiFlag
	if !filepath.IsAbs(opentuiDir) {
		opentuiDir = filepath.Join(workRoot, opentuiDir)
	}
	opentuiDir, err = filepath.Abs(opentuiDir)
	if err != nil {
		return err
	}

	buildDir := filepath.Join(workRoot, buildDirRel)
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", buildDir, err)
	}

	if err := ensureOpenTUI(opentuiDir, *force); err != nil {
		return err
	}

	zigDir := filepath.Join(opentuiDir, "packages", "core", "src", "zig")
	buildZig := filepath.Join(zigDir, "build.zig")
	if _, err := os.Stat(buildZig); err != nil {
		return fmt.Errorf("OpenTUI zig tree not found at %s", zigDir)
	}

	patchPath := filepath.Join(modRoot, patchRel)
	if _, err := os.Stat(patchPath); err != nil {
		return fmt.Errorf("patch not found: %s", patchPath)
	}

	if !*skipPatch {
		if err := applyPatch(opentuiDir, patchPath); err != nil {
			return err
		}
	} else {
		fmt.Println("skip-patch: leaving OpenTUI sources unchanged")
	}

	if *skipBuild {
		fmt.Println("skip-build: clone/patch finished")
		return nil
	}

	libPath, err := buildStatic(zigDir, *optimize)
	if err != nil {
		return err
	}

	info, err := os.Stat(libPath)
	if err != nil {
		return err
	}
	fmt.Printf("OK: static OpenTUI library ready (%s)\n", humanSize(info.Size()))
	fmt.Println(libPath)
	return nil
}

func ensureOpenTUI(opentuiDir string, force bool) error {
	gitDir := filepath.Join(opentuiDir, ".git")
	exists := fileExists(opentuiDir)

	if force && exists {
		fmt.Printf("force: removing %s\n", opentuiDir)
		if err := os.RemoveAll(opentuiDir); err != nil {
			return fmt.Errorf("remove %s: %w", opentuiDir, err)
		}
		exists = false
	}

	if exists {
		if !fileExists(gitDir) {
			return fmt.Errorf("%s exists but is not a git checkout; remove it or pass -force", opentuiDir)
		}
		fmt.Println("clone: already present")
		return nil
	}

	if _, err := exec.LookPath("git"); err != nil {
		return errors.New("git not found in PATH")
	}

	parent := filepath.Dir(opentuiDir)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}

	fmt.Printf("clone: %s -> %s\n", opentuiGitURL, opentuiDir)
	cmd := exec.Command("git", "clone", "--depth", "1", opentuiGitURL, opentuiDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}
	return nil
}

// findModuleRoot locates the soopentui module (go.mod + patches/), preferring
// the source tree that contains this command so `go run github.com/zztkm/soopentui/cmd/...` works.
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
	// cmd/opentui-static/main.go -> module root
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	if !isModuleRoot(root) {
		return "", fmt.Errorf("not a module root: %s", root)
	}
	return root, nil
}

func isModuleRoot(dir string) bool {
	return fileExists(filepath.Join(dir, "go.mod")) && fileExists(filepath.Join(dir, patchRel))
}

// workRoot is where _build/ is written. Module cache is often read-only, so use cwd then.
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

func applyPatch(opentuiDir, patchPath string) error {
	if err := gitQuiet(opentuiDir, "apply", "--check", "--reverse", patchPath); err == nil {
		fmt.Println("patch: already applied")
		return nil
	}

	if out, err := gitOutput(opentuiDir, "apply", "--check", patchPath); err != nil {
		return fmt.Errorf("patch does not apply cleanly to %s: %w\n%s\nre-run with -force to re-clone, then retry", opentuiDir, err, strings.TrimSpace(out))
	}
	if out, err := gitOutput(opentuiDir, "apply", patchPath); err != nil {
		return fmt.Errorf("git apply failed: %w\n%s", err, strings.TrimSpace(out))
	}
	fmt.Println("patch: applied", filepath.Base(patchPath))
	return nil
}

func buildStatic(zigDir, optimize string) (string, error) {
	if _, err := exec.LookPath("zig"); err != nil {
		return "", errors.New("zig not found in PATH (OpenTUI requires Zig " + readZigVersionHint(zigDir) + ")")
	}

	sdk, err := macosSDK()
	if err != nil {
		return "", err
	}

	arch, osName, err := opentuiPlatform()
	if err != nil {
		return "", err
	}
	outDir := filepath.Join(zigDir, "lib", arch+"-"+osName+"-static")
	outLib := filepath.Join(outDir, "libopentui.a")

	args := []string{
		"build",
		"-Doptimize=" + optimize,
		"-Dlinkage=static",
	}
	if sdk != "" {
		args = append(args, "-Dmacos-sdk="+sdk)
	}

	cmd := exec.Command("zig", args...)
	cmd.Dir = zigDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = buildEnv(sdk)

	fmt.Printf("zig: %s\n", strings.TrimSpace(zigVersion()))
	if sdk != "" {
		fmt.Printf("sdk: %s\n", sdk)
	}
	fmt.Printf("outdir: %s\n", outDir)
	fmt.Printf("running: zig %s\n", strings.Join(args, " "))

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("zig build failed: %w", err)
	}
	if !fileExists(outLib) {
		return "", fmt.Errorf("expected %s after build", outLib)
	}
	return outLib, nil
}

func buildEnv(sdk string) []string {
	env := os.Environ()
	if runtime.GOOS == "darwin" {
		// Zig 0.15.2 build-runner link fails against newer Xcode SDKs unless
		// DEVELOPER_DIR=/dev/null. See https://codeberg.org/ziglang/zig/issues/31658
		if os.Getenv("OPENTUI_KEEP_DEVELOPER_DIR") == "" {
			env = setEnv(env, "DEVELOPER_DIR", "/dev/null")
		}
	}
	if sdk != "" {
		env = setEnv(env, "SDKROOT", sdk)
	}
	return env
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

func macosSDK() (string, error) {
	if runtime.GOOS != "darwin" {
		return "", nil
	}
	if sdk := os.Getenv("SDKROOT"); sdk != "" {
		return sdk, nil
	}
	out, err := exec.Command("xcrun", "--show-sdk-path").Output()
	if err != nil {
		return "", fmt.Errorf("macOS SDK not found (set SDKROOT or install Xcode CLT): %w", err)
	}
	return strings.TrimSpace(string(out)), nil
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

func gitQuiet(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.Run()
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func zigVersion() string {
	out, err := exec.Command("zig", "version").Output()
	if err != nil {
		return "unknown"
	}
	return string(out)
}

func readZigVersionHint(zigDir string) string {
	p := filepath.Join(zigDir, "..", "..", "..", "..", "..", ".zig-version")
	p = filepath.Clean(p)
	b, err := os.ReadFile(p)
	if err != nil {
		return "0.15.2"
	}
	return strings.TrimSpace(string(b))
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
