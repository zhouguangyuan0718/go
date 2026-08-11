// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testdir_test

import (
	"bytes"
	stdcontext "context"
	"encoding/json"
	"fmt"
	"internal/testenv"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
)

const llvmStdlibPolicyEnv = "GOALLC_RUN_LLVM_STDLIB"

type llvmStdlibTestSet struct {
	Whitelist map[string]string `json:"whitelist"`
	Blacklist map[string]string `json:"blacklist"`
}

type llvmStdlibPolicy struct {
	EntryPackage llvmStdlibTestSet `json:"entry_package"`
}

type llvmStdlibClass uint8

const (
	llvmStdlibUnclassified llvmStdlibClass = iota
	llvmStdlibWhite
	llvmStdlibBlack
)

func readLLVMStdlibPolicy(t *testing.T) llvmStdlibPolicy {
	t.Helper()
	filename := filepath.Join(testenv.GOROOT(t), "test", "llvm_stdlib_packages.json")
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var policy llvmStdlibPolicy
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&policy); err != nil {
		t.Fatalf("parse llvm_stdlib_packages.json: %v", err)
	}
	return policy
}

func llvmStdlibPackages(t *testing.T) map[string]bool {
	t.Helper()
	cmd := testenv.Command(t, llvmStdlibGoTool(t), "list", "std")
	cmd.Env = append(os.Environ(), "GOENV=off", "GOFLAGS=", "GOROOT="+testenv.GOROOT(t))
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("list standard library packages: %v\n%s", err, out)
	}
	packages := make(map[string]bool)
	for _, name := range strings.Fields(string(out)) {
		packages[name] = true
	}
	return packages
}

func llvmStdlibGoTool(t *testing.T) string {
	t.Helper()
	if goTool == "" {
		goTool = testenv.GoToolPath(t)
	}
	return goTool
}

func classifyLLVMStdlibPackage(set llvmStdlibTestSet, name string) llvmStdlibClass {
	// An exact whitelist entry deliberately wins over the catch-all blacklist.
	if _, ok := set.Whitelist[name]; ok {
		return llvmStdlibWhite
	}
	if _, ok := set.Blacklist[name]; ok {
		return llvmStdlibBlack
	}
	if _, ok := set.Blacklist["*"]; ok {
		return llvmStdlibBlack
	}
	return llvmStdlibUnclassified
}

func validateLLVMStdlibPolicy(t *testing.T, packages map[string]bool, set llvmStdlibTestSet) {
	t.Helper()
	failed := false
	if len(set.Whitelist) == 0 {
		t.Error("LLVM standard library whitelist is empty")
		failed = true
	}
	for name, reason := range set.Whitelist {
		if name == "*" || strings.ContainsAny(name, "*?[\\") {
			t.Errorf("LLVM standard library whitelist entry %q is not an exact package", name)
			failed = true
		}
		if !packages[name] {
			t.Errorf("LLVM standard library whitelist entry %q is not a standard library package", name)
			failed = true
		}
		if strings.TrimSpace(reason) == "" {
			t.Errorf("LLVM standard library whitelist entry %q has no reason", name)
			failed = true
		}
		if _, ok := set.Blacklist[name]; ok {
			t.Errorf("LLVM standard library package %q appears in both exact lists", name)
			failed = true
		}
	}
	for name, reason := range set.Blacklist {
		if name != "*" && strings.ContainsAny(name, "*?[\\") {
			t.Errorf("LLVM standard library blacklist entry %q is neither an exact package nor the catch-all", name)
			failed = true
		}
		if name != "*" && !packages[name] {
			t.Errorf("LLVM standard library blacklist entry %q is not a standard library package", name)
			failed = true
		}
		if strings.TrimSpace(reason) == "" {
			t.Errorf("LLVM standard library blacklist entry %q has no reason", name)
			failed = true
		}
	}
	for name := range packages {
		if classifyLLVMStdlibPackage(set, name) == llvmStdlibUnclassified {
			t.Errorf("standard library package %q is not classified as white or black", name)
			failed = true
		}
	}
	if failed {
		t.FailNow()
	}
}

func TestLLVMStdlibPolicy(t *testing.T) {
	packages := llvmStdlibPackages(t)
	validateLLVMStdlibPolicy(t, packages, readLLVMStdlibPolicy(t).EntryPackage)
}

func TestClassifyLLVMStdlibPackage(t *testing.T) {
	set := llvmStdlibTestSet{
		Whitelist: map[string]string{"cmp": "qualified"},
		Blacklist: map[string]string{"*": "not yet qualified", "sort": "known failure"},
	}
	for _, test := range []struct {
		name string
		want llvmStdlibClass
	}{
		{"cmp", llvmStdlibWhite},
		{"sort", llvmStdlibBlack},
		{"bytes", llvmStdlibBlack},
	} {
		if got := classifyLLVMStdlibPackage(set, test.name); got != test.want {
			t.Errorf("classifyLLVMStdlibPackage(%q) = %v, want %v", test.name, got, test.want)
		}
	}
}

func TestLLVMStdlib(t *testing.T) {
	if os.Getenv(llvmStdlibPolicyEnv) != "1" {
		t.Skipf("set %s=1 to run the LLVM standard library package policy", llvmStdlibPolicyEnv)
	}
	testenv.MustHaveGoBuild(t)
	platform := runtime.GOOS + "/" + runtime.GOARCH
	switch platform {
	case "darwin/arm64", "linux/amd64", "linux/arm64":
	default:
		t.Skipf("LLVM GoObj is not configured for %s", platform)
	}

	packages := llvmStdlibPackages(t)
	set := readLLVMStdlibPolicy(t).EntryPackage
	validateLLVMStdlibPolicy(t, packages, set)
	configureLLVMTestToolchain(t)
	toolexec := llvmToolexec(t, "default<O2>")

	whitelist := make([]string, 0, len(set.Whitelist))
	for name := range set.Whitelist {
		whitelist = append(whitelist, name)
	}
	sort.Strings(whitelist)
	t.Logf("LLVM standard library entry-package policy: %d white, %d black (%d packages)", len(whitelist), len(packages)-len(whitelist), len(packages))

	knownBlacklist := make([]string, 0, len(set.Blacklist)-1)
	for name := range set.Blacklist {
		if name != "*" {
			knownBlacklist = append(knownBlacklist, name)
		}
	}
	sort.Strings(knownBlacklist)
	for _, name := range knownBlacklist {
		t.Logf("LLVM stdlib blacklist result: NOT RUN package=%q reason=%q", name, set.Blacklist[name])
	}
	t.Logf("LLVM stdlib blacklist result: NOT RUN remaining=%d reason=%q", len(packages)-len(whitelist)-len(knownBlacklist), set.Blacklist["*"])

	// The target package, gcflags, and toolexec pipeline are part of cmd/go's
	// action IDs, so one isolated cache can safely serve every package in this
	// policy run. The toolchain, payload, and pass plugin remain fixed for the
	// lifetime of the test process.
	cache := t.TempDir()
	for _, name := range whitelist {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := stdcontext.WithTimeout(stdcontext.Background(), 5*time.Minute)
			defer cancel()
			cmd := testenv.CommandContext(t, ctx, llvmStdlibGoTool(t),
				"test",
				"-count=1",
				"-timeout=2m",
				"-toolexec="+toolexec,
				fmt.Sprintf("-gcflags=%s=-enablellvm -llvmironly", name),
				name,
			)
			cmd.Env = append(os.Environ(),
				"GOENV=off",
				"GOFLAGS=",
				"GOROOT="+testenv.GOROOT(t),
				"GOCACHE="+cache,
			)
			out, err := cmd.CombinedOutput()
			if err != nil {
				if ctx.Err() != nil {
					t.Fatalf("LLVM stdlib whitelist result: TIMEOUT package=%q: %v\n%s", name, ctx.Err(), out)
				}
				t.Fatalf("LLVM stdlib whitelist result: FAIL package=%q: %v\n%s", name, err, out)
			}
			t.Logf("LLVM stdlib whitelist result: PASS package=%q", name)
		})
	}
}
