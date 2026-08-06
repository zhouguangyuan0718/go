// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testdir_test

import (
	"bytes"
	"cmd/internal/quoted"
	"encoding/json"
	"fmt"
	"go/build/constraint"
	"internal/testenv"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

const llvmTestToolexecEnv = "GOALLC_TEST_TOOLEXEC"

type llvmTestSet struct {
	Whitelist         map[string]string            `json:"whitelist"`
	Blacklist         map[string]string            `json:"blacklist"`
	PlatformBlacklist map[string]map[string]string `json:"platform_blacklist,omitempty"`
}

type llvmTestPolicy struct {
	Codegen llvmTestSet `json:"codegen"`
	Runtime llvmTestSet `json:"runtime"`
}

// runLLVMTests reuses the existing GOROOT/test inputs. The whitelist contains
// tests that must pass today; the blacklist classifies known unsupported tests.
// Whitelist entries take precedence over broad blacklist patterns, so enabling
// another test is a one-line policy change plus its LLVM IR checks.
func runLLVMTests(t *testing.T, common testCommon) {
	t.Run("LLVM", func(t *testing.T) {
		switch runtime.GOOS + "/" + runtime.GOARCH {
		case "darwin/arm64", "linux/amd64", "linux/arm64":
		default:
			t.Skipf("LLVM GoObj is not configured for %s/%s", runtime.GOOS, runtime.GOARCH)
		}
		configureLLVMTestToolchain(t)

		policy := readLLVMTestPolicy(t, common.gorootTestDir)
		platform := runtime.GOOS + "/" + runtime.GOARCH
		if err := applyLLVMPlatformBlacklist("codegen", platform, &policy.Codegen); err != nil {
			t.Fatal(err)
		}
		if err := applyLLVMPlatformBlacklist("runtime", platform, &policy.Runtime); err != nil {
			t.Fatal(err)
		}
		codegenCandidates := llvmTestCandidates(t, common.gorootTestDir, []string{"codegen"}, "asmcheck")
		runtimeCandidates := llvmTestCandidates(t, common.gorootTestDir, dirs, "run")
		for name := range llvmTestCandidates(t, common.gorootTestDir, dirs, "runoutput") {
			runtimeCandidates[name] = true
		}
		validateLLVMTestSet(t, common.gorootTestDir, "codegen", codegenCandidates, policy.Codegen, true)
		validateLLVMTestSet(t, common.gorootTestDir, "runtime", runtimeCandidates, policy.Runtime, false)

		t.Logf("LLVM codegen whitelist: %d/%d files", len(policy.Codegen.Whitelist), len(codegenCandidates))
		t.Logf("LLVM runtime whitelist: %d/%d files", len(policy.Runtime.Whitelist), len(runtimeCandidates))

		t.Run("codegen", func(t *testing.T) {
			names := sortedLLVMWhitelist(policy.Codegen.Whitelist)
			for _, name := range names {
				t.Run(name, func(t *testing.T) {
					runLLVMCodegenTest(t, common.gorootTestDir, name)
				})
			}
		})

		t.Run("abi-differential", func(t *testing.T) {
			runLLVMABIDifferentialTest(t, common.gorootTestDir)
		})

		t.Run("alloca-statepoint", func(t *testing.T) {
			runLLVMAllocaStatepointTest(t, common.gorootTestDir)
		})

		t.Run("caller-state", func(t *testing.T) {
			runLLVMCallerStateTest(t, common.gorootTestDir)
		})

		t.Run("writebarrier-helpers", runLLVMWriteBarrierHelperTest)

		t.Run("compile-only-regressions", func(t *testing.T) {
			for _, name := range []string{
				"cmp.go",
				"typeparam/issue47684c.go",
			} {
				t.Run(name, func(t *testing.T) {
					runLLVMCompileOnlyRegression(t, common.gorootTestDir, name)
				})
			}
		})

		t.Run("runtime", func(t *testing.T) {
			names := sortedLLVMWhitelist(policy.Runtime.Whitelist)
			for _, name := range names {
				t.Run(name, func(t *testing.T) {
					dir, file := path.Split(name)
					tc := test{
						testCommon: common,
						T:          t,
						dir:        strings.TrimSuffix(dir, "/"),
						goFile:     file,
						llvm:       true,
					}
					if err := tc.run(); err != nil {
						t.Fatal(err)
					}
				})
			}
		})

		t.Run("writebarrier-ir", runLLVMWriteBarrierIRTests)
	})
}

func runLLVMCompileOnlyRegression(t *testing.T, gorootTestDir, name string) {
	t.Helper()
	toolexec := llvmToolexec(t, "")
	exe := filepath.Join(t.TempDir(), "test.exe")
	cmd := exec.Command(goTool, "build",
		"-gcflags=all="+os.Getenv("GO_GCFLAGS"),
		"-gcflags=-enablellvm",
		"-ldflags=-w",
		"-toolexec="+toolexec,
		"-o", exe,
		filepath.Join(gorootTestDir, filepath.FromSlash(name)),
	)
	cmd.Env = append(os.Environ(), "GOENV=off", "GOFLAGS=")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("LLVM compile-only regression %s failed: %v\n%s", name, err, out)
	}
}

func readLLVMTestPolicy(t *testing.T, gorootTestDir string) llvmTestPolicy {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(gorootTestDir, "llvm_tests.json"))
	if err != nil {
		t.Fatal(err)
	}
	var policy llvmTestPolicy
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&policy); err != nil {
		t.Fatalf("parse llvm_tests.json: %v", err)
	}
	return policy
}

func applyLLVMPlatformBlacklist(name, platform string, set *llvmTestSet) error {
	for target, entries := range set.PlatformBlacklist {
		if strings.TrimSpace(target) == "" {
			return fmt.Errorf("LLVM %s platform blacklist has an empty platform", name)
		}
		for filename, reason := range entries {
			if strings.TrimSpace(reason) == "" {
				return fmt.Errorf("LLVM %s platform blacklist entry %q for %s has no reason", name, filename, target)
			}
			if _, ok := set.Whitelist[filename]; !ok {
				return fmt.Errorf("LLVM %s platform blacklist entry %q for %s is not in the common whitelist", name, filename, target)
			}
		}
	}

	for filename := range set.PlatformBlacklist[platform] {
		delete(set.Whitelist, filename)
	}
	return nil
}

func TestApplyLLVMPlatformBlacklist(t *testing.T) {
	set := llvmTestSet{
		Whitelist: map[string]string{
			"common.go": "common",
			"linux.go":  "linux",
			"darwin.go": "darwin",
		},
		PlatformBlacklist: map[string]map[string]string{
			"linux/amd64":  {"linux.go": "linux limitation"},
			"darwin/arm64": {"darwin.go": "darwin limitation"},
		},
	}
	if err := applyLLVMPlatformBlacklist("runtime", "linux/amd64", &set); err != nil {
		t.Fatal(err)
	}
	if _, ok := set.Whitelist["linux.go"]; ok {
		t.Fatal("current-platform exclusion remained in the effective whitelist")
	}
	for _, filename := range []string{"common.go", "darwin.go"} {
		if _, ok := set.Whitelist[filename]; !ok {
			t.Errorf("applyLLVMPlatformBlacklist removed %q for another platform", filename)
		}
	}

	tests := []struct {
		name string
		set  llvmTestSet
		want string
	}{
		{
			name: "empty platform",
			set: llvmTestSet{
				Whitelist:         map[string]string{"test.go": "test"},
				PlatformBlacklist: map[string]map[string]string{"": {"test.go": "reason"}},
			},
			want: "empty platform",
		},
		{
			name: "empty reason",
			set: llvmTestSet{
				Whitelist:         map[string]string{"test.go": "test"},
				PlatformBlacklist: map[string]map[string]string{"linux/amd64": {"test.go": " "}},
			},
			want: "has no reason",
		},
		{
			name: "not in common whitelist",
			set: llvmTestSet{
				Whitelist:         map[string]string{"test.go": "test"},
				PlatformBlacklist: map[string]map[string]string{"linux/amd64": {"missing.go": "reason"}},
			},
			want: "is not in the common whitelist",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := applyLLVMPlatformBlacklist("runtime", "linux/amd64", &tc.set)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("applyLLVMPlatformBlacklist error = %v, want substring %q", err, tc.want)
			}
		})
	}
}

func llvmTestCandidates(t *testing.T, gorootTestDir string, scanDirs []string, wantAction string) map[string]bool {
	t.Helper()
	candidates := make(map[string]bool)
	for _, dir := range scanDirs {
		entries, err := os.ReadDir(filepath.Join(gorootTestDir, dir))
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") || !strings.HasSuffix(entry.Name(), ".go") {
				continue
			}
			name := path.Join(dir, entry.Name())
			if llvmTestAction(t, filepath.Join(gorootTestDir, filepath.FromSlash(name))) == wantAction {
				candidates[name] = true
			}
		}
	}
	return candidates
}

func llvmTestAction(t *testing.T, filename string) string {
	t.Helper()
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	for src := string(data); src != ""; {
		var line string
		line, src, _ = strings.Cut(src, "\n")
		if constraint.IsGoBuild(line) || constraint.IsPlusBuild(line) || strings.TrimSpace(line) == "" {
			continue
		}
		action := strings.TrimSpace(strings.TrimPrefix(line, "//"))
		fields, err := splitQuoted(action)
		if err != nil {
			t.Fatalf("%s: invalid test recipe: %v", filename, err)
		}
		if len(fields) == 0 {
			return ""
		}
		return fields[0]
	}
	return ""
}

func validateLLVMTestSet(t *testing.T, gorootTestDir, name string, candidates map[string]bool, set llvmTestSet, requireChecks bool) {
	t.Helper()
	failed := false
	for filename := range set.Whitelist {
		if !candidates[filename] {
			t.Errorf("LLVM %s whitelist entry %q is not a %s test", name, filename, name)
			failed = true
		}
		if requireChecks {
			source, err := os.ReadFile(filepath.Join(gorootTestDir, filepath.FromSlash(filename)))
			if err != nil {
				t.Errorf("read LLVM %s whitelist entry %q: %v", name, filename, err)
				failed = true
			} else if !bytes.Contains(source, []byte("// LLVM")) {
				t.Errorf("LLVM %s whitelist entry %q has no FileCheck directives", name, filename)
				failed = true
			}
		}
	}

	for pattern := range set.Blacklist {
		matched := false
		for filename := range candidates {
			if llvmPathMatch(t, pattern, filename) {
				matched = true
				break
			}
		}
		if !matched {
			t.Errorf("LLVM %s blacklist pattern %q matches no tests", name, pattern)
			failed = true
		}
	}

	for filename := range candidates {
		if _, ok := set.Whitelist[filename]; ok {
			continue
		}
		classified := false
		for pattern := range set.Blacklist {
			if llvmPathMatch(t, pattern, filename) {
				classified = true
				break
			}
		}
		if !classified {
			t.Errorf("LLVM %s test %q is in neither whitelist nor blacklist", name, filename)
			failed = true
		}
	}
	if failed {
		t.FailNow()
	}
}

func llvmPathMatch(t *testing.T, pattern, filename string) bool {
	t.Helper()
	matched, err := path.Match(pattern, filename)
	if err != nil {
		t.Fatalf("invalid LLVM test blacklist pattern %q: %v", pattern, err)
	}
	if matched || strings.Contains(pattern, "/") {
		return matched
	}
	matched, err = path.Match(pattern, path.Base(filename))
	if err != nil {
		t.Fatalf("invalid LLVM test blacklist pattern %q: %v", pattern, err)
	}
	return matched
}

func sortedLLVMWhitelist(entries map[string]string) []string {
	names := make([]string, 0, len(entries))
	for name := range entries {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func runLLVMCodegenTest(t *testing.T, gorootTestDir, name string) {
	t.Helper()
	source := filepath.Join(gorootTestDir, filepath.FromSlash(name))
	src, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	header, _, _ := strings.Cut(string(src), "\npackage")
	if ok, why := shouldTest(header, goos, goarch); !ok {
		t.Skip(why)
	}

	archive := filepath.Join(t.TempDir(), "codegen.a")
	cmd := exec.Command(goTool, "tool", "compile",
		"-p=codegen",
		"-importcfg="+stdlibImportcfgFile(),
		"-enablellvm",
		"-llvmironly",
		"-c=16",
		"-o", archive,
		source,
	)
	cmd.Env = append(os.Environ(), "GOENV=off", "GOFLAGS=")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("LLVM compilation failed: %v\n%s", err, out)
	}
	irBytes, err := os.ReadFile(archive + ".ll")
	if err != nil {
		t.Fatal(err)
	}
	opt := llvmToolPath(t, "opt", "GOALLC_OPT")
	cmd = exec.Command(opt, "-passes=verify", "-disable-output")
	cmd.Stdin = bytes.NewReader(irBytes)
	cmd.Env = append(os.Environ(), "GOENV=off", "GOFLAGS=")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("LLVM verifier failed: %v\n%s", err, out)
	}
	fileCheck := llvmToolPath(t, "FileCheck", "GOALLC_FILECHECK")
	cmd = exec.Command(fileCheck, "--check-prefixes="+llvmFileCheckPrefixes("LLVM", src), source)
	cmd.Stdin = bytes.NewReader(irBytes)
	cmd.Env = append(os.Environ(), "GOENV=off", "GOFLAGS=")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("FileCheck failed: %v\n%s", err, out)
	}

	if !bytes.Contains(src, []byte("// LLVM-OPT")) {
		return
	}
	cmd = exec.Command(opt, "-passes=default<O2>", "-S")
	cmd.Stdin = bytes.NewReader(irBytes)
	cmd.Env = append(os.Environ(), "GOENV=off", "GOFLAGS=")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	optimizedIR, err := cmd.Output()
	if err != nil {
		t.Fatalf("LLVM optimization failed: %v\n%s", err, stderr.Bytes())
	}
	cmd = exec.Command(fileCheck, "--check-prefixes="+llvmFileCheckPrefixes("LLVM-OPT", src), source)
	cmd.Stdin = bytes.NewReader(optimizedIR)
	cmd.Env = append(os.Environ(), "GOENV=off", "GOFLAGS=")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("optimized LLVM FileCheck failed: %v\n%s", err, out)
	}
}

func llvmFileCheckPrefixes(base string, source []byte) string {
	prefixes := []string{base}
	var architecturePrefix string
	switch runtime.GOARCH {
	case "amd64":
		architecturePrefix = base + "-AMD64"
	case "arm64":
		architecturePrefix = base + "-ARM64"
	}
	if architecturePrefix != "" &&
		(bytes.Contains(source, []byte(architecturePrefix+":")) ||
			bytes.Contains(source, []byte(architecturePrefix+"-"))) {
		prefixes = append(prefixes, architecturePrefix)
	}
	return strings.Join(prefixes, ",")
}

func runLLVMWriteBarrierIRTests(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "ZeroWithPointers",
			source: "package p\nfunc zero(dst *[2]*int) { *dst = [2]*int{} }\n",
			want:   []string{"call goabiinternal void @runtime.wbZero(ptr", "@llvm.memset.inline", `"gc-leaf-function"`},
		},
		{
			name:   "MoveWithPointers",
			source: "package p\nfunc move(dst, src *[2]*int) { *dst = *src }\n",
			want:   []string{"call goabiinternal void @runtime.wbMove(ptr", "@llvm.memmove", `"gc-leaf-function"`},
		},
		{
			name:   "DeletePointer",
			source: "package p\nfunc delete(dst **int) { *dst = nil }\n",
			want:   []string{"@llvm.go.gc.write.barrier"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			source := filepath.Join(dir, "fail.go")
			if err := os.WriteFile(source, []byte(tc.source), 0o666); err != nil {
				t.Fatal(err)
			}
			archive := filepath.Join(dir, "writebarrier.a")
			cmd := exec.Command(goTool, "tool", "compile",
				"-p=p",
				"-importcfg="+stdlibImportcfgFile(),
				"-enablellvm",
				"-llvmironly",
				"-o", archive,
				source,
			)
			cmd.Env = append(os.Environ(), "GOENV=off", "GOFLAGS=")
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("LLVM compilation failed: %v\n%s", err, out)
			}
			ir, err := os.ReadFile(archive + ".ll")
			if err != nil {
				t.Fatal(err)
			}
			want := append(tc.want, `"go-async-unsafe"`)
			for _, want := range want {
				if !bytes.Contains(ir, []byte(want)) {
					t.Fatalf("LLVM write-barrier IR does not contain %s\n%s", want, ir)
				}
			}
			if bytes.Contains(ir, []byte(".rawarg = ptrtoint")) {
				t.Fatalf("LLVM write-barrier IR still coerces pointer arguments to raw uintptr\n%s", ir)
			}
			if bytes.Contains(ir, []byte("llvm.go.gc.unsafe.point")) {
				t.Fatalf("LLVM write-barrier IR still contains unsafe-point marker calls\n%s", ir)
			}
		})
	}
}

func (t test) runLLVMCase(tempDir string, flags, args []string, runcmd runCmd) error {
	out, err := t.runLLVMProgram(tempDir, "test.exe", t.goFileName(), flags, args, runcmd)
	if err != nil {
		return err
	}
	return t.checkExpectedOutput(out)
}

func (t test) runLLVMRunoutputCase(tempDir string, args []string, runcmd runCmd) error {
	out, err := t.runLLVMProgram(tempDir, "generator.exe", t.goFileName(), nil, args, runcmd)
	if err != nil {
		return err
	}
	generated := filepath.Join(tempDir, "tmp__.go")
	if err := os.WriteFile(generated, out, 0o666); err != nil {
		return err
	}
	out, err = t.runLLVMProgram(tempDir, "generated.exe", generated, nil, nil, runcmd)
	if err != nil {
		return err
	}
	return t.checkExpectedOutput(out)
}

func (t test) runLLVMProgram(tempDir, executable, source string, flags, args []string, runcmd runCmd) ([]byte, error) {
	if goos != runtime.GOOS || goarch != runtime.GOARCH {
		t.Skip("LLVM execution tests do not support cross compilation")
	}
	toolexec := llvmToolexec(t.T, "default<O2>")
	exe := filepath.Join(tempDir, executable)
	// GoObj DWARF emission is not implemented by the LLVM backend yet. Disable
	// debug data explicitly so runtime qualification measures compile, link, and
	// execution support instead of failing in the linker's DWARF writer.
	cmd := []string{goTool, "build", t.goGcflags(), "-ldflags=-w", "-toolexec=" + toolexec, "-o", exe}
	cmd = append(cmd, flags...)
	cmd = append(cmd, source)
	if _, err := runcmd(cmd...); err != nil {
		return nil, err
	}
	out, err := runcmd(append([]string{exe}, args...)...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func llvmToolexec(t *testing.T, optPasses string) string {
	t.Helper()
	wrapper := llvmToolexecPath(t)

	llc := llvmToolPath(t, "llc", "GOALLC_LLC")
	plugin := llvmPassPluginPath(t)
	args := []string{
		wrapper,
		"-llc=" + llc,
		"-pass-plugin=" + plugin,
		"-llvm-package=command-line-arguments",
	}
	if optPasses != "" {
		opt := llvmToolPath(t, "opt", "GOALLC_OPT")
		args = append(args, "-opt="+opt, "-opt-passes="+optPasses)
	}
	value, err := quoted.Join(args)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

type llvmTestToolchain struct {
	root      string
	llc       string
	opt       string
	fileCheck string
	plugin    string
	wrapper   string
}

func configureLLVMTestToolchain(t *testing.T) llvmTestToolchain {
	t.Helper()
	root, configured := llvmTestPayloadRoot(t)
	if root == "" {
		t.Skip("LLVM payload is unavailable; set GOALLC_LLVM_DIR or build Go with -llvm-dir")
	}

	llvmConfig := llvmPayloadExecutable(t, root, "llvm-config")
	version := llvmCommandOutput(t, llvmConfig, "--version")
	if version != "23" && !strings.HasPrefix(version, "23.") {
		t.Fatalf("selected LLVM payload %q has version %q, want LLVM 23", root, version)
	}
	prefix := llvmCommandOutput(t, llvmConfig, "--prefix")
	if !sameLLVMTestPath(prefix, root) {
		t.Fatalf("selected LLVM payload prefix mismatch: root %q, llvm-config --prefix %q", root, prefix)
	}

	tools := llvmTestToolchain{
		root:      root,
		llc:       llvmPayloadTool(t, root, "llc", "GOALLC_LLC"),
		opt:       llvmPayloadTool(t, root, "opt", "GOALLC_OPT"),
		fileCheck: llvmPayloadTool(t, root, "FileCheck", "GOALLC_FILECHECK"),
		plugin:    llvmPayloadPlugin(t, root),
		wrapper:   buildLLVMTestToolexec(t),
	}

	// Freeze every consumer to the validated payload. In particular, do not let
	// an individual subtest silently fall back to a stale GOROOT/llvm tree.
	t.Setenv("GOALLC_LLVM_DIR", tools.root)
	t.Setenv("GOALLC_LLC", tools.llc)
	t.Setenv("GOALLC_OPT", tools.opt)
	t.Setenv("GOALLC_FILECHECK", tools.fileCheck)
	t.Setenv("GOALLC_PASS_PLUGIN", tools.plugin)
	t.Setenv(llvmTestToolexecEnv, tools.wrapper)
	t.Logf("LLVM test toolchain: go=%s wrapper=%s payload=%s llc=%s opt=%s FileCheck=%s plugin=%s runtime-pipeline=default<O2>",
		goTool, tools.wrapper, tools.root, tools.llc, tools.opt, tools.fileCheck, tools.plugin)
	if configured != "" {
		t.Logf("LLVM payload selected by %s", configured)
	}
	return tools
}

func llvmTestPayloadRoot(t *testing.T) (root, configured string) {
	t.Helper()
	if value := strings.TrimSpace(os.Getenv("GOALLC_LLVM_DIR")); value != "" {
		return llvmAbsolutePath(t, value, "GOALLC_LLVM_DIR"), "GOALLC_LLVM_DIR"
	}
	if value := strings.TrimSpace(os.Getenv("GOALLC_LLC")); value != "" {
		llc := llvmRegularFile(t, value, "GOALLC_LLC", true)
		return filepath.Dir(filepath.Dir(llc)), "GOALLC_LLC"
	}

	goroot := testenv.GOROOT(t)
	payloadConfig := filepath.Join(goroot, "pkg", "goallc-llvm-payload")
	if data, err := os.ReadFile(payloadConfig); err == nil {
		value := strings.TrimSpace(string(data))
		if value == "" {
			t.Fatalf("LLVM payload configuration %s is empty", payloadConfig)
		}
		return llvmAbsolutePath(t, value, payloadConfig), payloadConfig
	} else if !os.IsNotExist(err) {
		t.Fatalf("reading LLVM payload configuration %s: %v", payloadConfig, err)
	}

	legacy := filepath.Join(goroot, "llvm")
	if _, err := os.Stat(filepath.Join(legacy, "bin", "llvm-config")); err == nil {
		return legacy, "GOROOT/llvm"
	} else if !os.IsNotExist(err) {
		t.Fatalf("checking legacy LLVM payload %s: %v", legacy, err)
	}
	return "", ""
}

func llvmPayloadTool(t *testing.T, root, name, envName string) string {
	t.Helper()
	expected := llvmPayloadExecutable(t, root, name)
	if value := strings.TrimSpace(os.Getenv(envName)); value != "" {
		configured := llvmRegularFile(t, value, envName, true)
		if !sameLLVMTestFile(configured, expected) {
			t.Fatalf("%s=%q does not belong to selected LLVM payload %q (expected %q)", envName, configured, root, expected)
		}
	}
	return expected
}

func llvmPayloadExecutable(t *testing.T, root, name string) string {
	t.Helper()
	return llvmRegularFile(t, filepath.Join(root, "bin", name), "LLVM payload", true)
}

func llvmPayloadPlugin(t *testing.T, root string) string {
	t.Helper()
	name := "GoALLCStatepoints.so"
	if runtime.GOOS == "darwin" {
		name = "GoALLCStatepoints.dylib"
	}
	expected := llvmRegularFile(t, filepath.Join(root, "lib", name), "LLVM payload", false)
	if value := strings.TrimSpace(os.Getenv("GOALLC_PASS_PLUGIN")); value != "" {
		configured := llvmRegularFile(t, value, "GOALLC_PASS_PLUGIN", false)
		if !sameLLVMTestFile(configured, expected) {
			t.Fatalf("GOALLC_PASS_PLUGIN=%q does not belong to selected LLVM payload %q (expected %q)", configured, root, expected)
		}
	}
	return expected
}

func llvmToolexecPath(t *testing.T) string {
	t.Helper()
	if value := strings.TrimSpace(os.Getenv(llvmTestToolexecEnv)); value != "" {
		return llvmRegularFile(t, value, llvmTestToolexecEnv, true)
	}
	out, err := exec.Command(goTool, "tool", "-n", "llvmtoolexec").CombinedOutput()
	if err != nil {
		t.Fatalf("resolving llvmtoolexec from selected Go tool %q: %v\n%s", goTool, err, out)
	}
	return llvmRegularFile(t, strings.TrimSpace(string(out)), "go tool llvmtoolexec", true)
}

func buildLLVMTestToolexec(t *testing.T) string {
	t.Helper()
	wrapper := filepath.Join(t.TempDir(), "llvmtoolexec")
	cmd := exec.Command(goTool, "build", "-o", wrapper, "cmd/llvmtoolexec")
	cmd.Dir = testenv.GOROOT(t)
	cmd.Env = append(os.Environ(), "GOENV=off", "GOFLAGS=")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building llvmtoolexec from the selected Go checkout: %v\n%s", err, out)
	}
	return llvmRegularFile(t, wrapper, "fresh llvmtoolexec", true)
}

func llvmCommandOutput(t *testing.T, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), "GOENV=off", "GOFLAGS=")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running %s %s: %v\n%s", name, strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

func llvmAbsolutePath(t *testing.T, value, source string) string {
	t.Helper()
	path, err := filepath.Abs(value)
	if err != nil {
		t.Fatalf("resolving %s path %q: %v", source, value, err)
	}
	return filepath.Clean(path)
}

func llvmRegularFile(t *testing.T, value, source string, executable bool) string {
	t.Helper()
	path := llvmAbsolutePath(t, value, source)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("checking %s file %q: %v", source, path, err)
	}
	if !info.Mode().IsRegular() {
		t.Fatalf("%s file %q is not a regular file", source, path)
	}
	if executable && info.Mode()&0o111 == 0 {
		t.Fatalf("%s file %q is not executable", source, path)
	}
	return path
}

func sameLLVMTestPath(a, b string) bool {
	a, errA := filepath.EvalSymlinks(filepath.Clean(a))
	b, errB := filepath.EvalSymlinks(filepath.Clean(b))
	return errA == nil && errB == nil && a == b
}

func sameLLVMTestFile(a, b string) bool {
	aInfo, aErr := os.Stat(a)
	bInfo, bErr := os.Stat(b)
	return aErr == nil && bErr == nil && os.SameFile(aInfo, bInfo)
}

func llvmToolPath(t *testing.T, name, envName string) string {
	t.Helper()
	root, _ := llvmTestPayloadRoot(t)
	if root == "" {
		t.Skipf("%s is unavailable; set GOALLC_LLVM_DIR or build Go with -llvm-dir", name)
	}
	return llvmPayloadTool(t, root, name, envName)
}

func llvmPassPluginPath(t *testing.T) string {
	t.Helper()
	root, _ := llvmTestPayloadRoot(t)
	if root == "" {
		t.Skip("GoALLC pass plugin is unavailable; set GOALLC_LLVM_DIR or build Go with -llvm-dir")
	}
	return llvmPayloadPlugin(t, root)
}
