// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testdir_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

type llvmABIDocument struct {
	Members []struct {
		GoObject *struct {
			Symbols []llvmABISymbol `json:"symbols"`
		} `json:"go_object"`
	} `json:"members"`
}

type llvmABISymbol struct {
	Name     string `json:"name"`
	ABI      uint16 `json:"abi"`
	Function *struct {
		Info *struct {
			Args   uint32 `json:"args"`
			Locals uint32 `json:"locals"`
		} `json:"info"`
		PCData []struct {
			Kind   string `json:"kind"`
			Ranges []struct {
				Value int32 `json:"value"`
			} `json:"ranges"`
		} `json:"pcdata"`
		FuncData []struct {
			Kind     string `json:"kind"`
			StackMap *struct {
				Count   int32 `json:"count"`
				NumBits int32 `json:"num_bits"`
				Bitmaps []struct {
					SetBits []int `json:"set_bits"`
				} `json:"bitmaps"`
			} `json:"stack_map"`
		} `json:"funcdata"`
		StackMapQueries []struct {
			StackMapIndex int32  `json:"stack_map_index"`
			DecodeError   string `json:"decode_error"`
		} `json:"stack_map_queries"`
	} `json:"function"`
}

type llvmABICase struct {
	name            string
	args            uint32
	pointerBits     []int
	nativeArgsMaps  [][]int
	goallcArgsMaps  [][]int
	nativeStackMaps []int32
	goallcStackMaps []int32
}

func runLLVMABIDifferentialTest(t *testing.T, gorootTestDir string) {
	t.Helper()
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Skip("exact ABI differential expectations are currently qualified on darwin/arm64")
	}

	dir := t.TempDir()
	source := filepath.Join(gorootTestDir, "abi", "llvm_args_results.go")
	nativeObject := filepath.Join(dir, "native.o")
	llvmArchive := filepath.Join(dir, "goallc.a")
	llvmIR := llvmArchive + ".ll"
	goallcObject := filepath.Join(dir, "goallc.o")

	nativeAssembly := runLLVMABICommand(t, nil, goTool, "tool", "compile",
		"-S", "-l", "-p=main", "-importcfg="+stdlibImportcfgFile(),
		"-o", nativeObject, source)
	if !bytes.Contains(nativeAssembly, []byte("main.mixedABI")) ||
		!bytes.Contains(nativeAssembly, []byte("main.bothOverflow")) {
		t.Fatalf("native -S output does not contain representative ABI functions")
	}

	runLLVMABICommand(t, nil, goTool, "tool", "compile",
		"-l", "-p=main", "-importcfg="+stdlibImportcfgFile(),
		"-enablellvm", "-llvmironly", "-o", llvmArchive, source)
	ir, err := os.ReadFile(llvmIR)
	if err != nil {
		t.Fatal(err)
	}
	for _, needle := range [][]byte{
		[]byte("define goabiinternal"),
		[]byte(`"go_results_tuple"`),
		[]byte(`gc "goallc"`),
		[]byte(`"go-stack-growth-statepoint"`),
	} {
		if !bytes.Contains(ir, needle) {
			t.Fatalf("GoALLC IR does not contain %q", needle)
		}
	}

	opt := llvmToolPath(t, "opt", "GOALLC_OPT")
	runLLVMABICommand(t, ir, opt, "-passes=verify", "-disable-output")

	llc := llvmToolPath(t, "llc", "GOALLC_LLC")
	plugin := llvmABIPassPlugin(t, llc)
	rewrittenIR := runLLVMABICommand(t, nil, llc,
		"-load-pass-plugin="+plugin,
		"-goallc-pass-plugin-emit-ir",
		"-filetype=null", "-o", "-", llvmIR)
	for _, needle := range [][]byte{
		[]byte(`"gc-live"(`),
		[]byte("@llvm.experimental.gc.relocate"),
		[]byte("extractvalue %main.stackAggregate"),
	} {
		if !bytes.Contains(rewrittenIR, needle) {
			t.Fatalf("rewritten GoALLC IR does not contain %q", needle)
		}
	}
	for _, line := range bytes.Split(rewrittenIR, []byte("\n")) {
		const gcLive = `"gc-live"(`
		start := bytes.Index(line, []byte(gcLive))
		if start < 0 {
			continue
		}
		operands := line[start+len(gcLive):]
		if end := bytes.IndexByte(operands, ')'); end >= 0 {
			operands = operands[:end]
		}
		if bytes.Contains(operands, []byte("{")) ||
			bytes.Contains(operands, []byte("[")) ||
			bytes.Contains(operands, []byte("<")) ||
			bytes.Contains(operands, []byte("%main.stackAggregate")) {
			t.Fatalf("aggregate survived in gc-live: %s", line)
		}
	}
	checkLLVMABIStatepointTupleAttrs(t, rewrittenIR,
		"overflowResults", "stackAggregateResult", "bothOverflow")
	runLLVMABICommand(t, rewrittenIR, opt,
		"-load-pass-plugin="+plugin, "-passes=verify", "-disable-output", "-")

	goallcAssembly := runLLVMABICommand(t, nil, llc,
		"-load-pass-plugin="+plugin, "-filetype=asm", llvmIR, "-o", "-")
	for _, name := range []string{"main.mixedABI", "main.growPointer", "main.bothOverflow"} {
		if !bytes.Contains(goallcAssembly, []byte(name)) {
			t.Fatalf("GoALLC assembly does not contain %s", name)
		}
	}
	checkLLVMABIAssembly(t, nativeAssembly, goallcAssembly)

	runLLVMABICommand(t, nil, llc,
		"-load-pass-plugin="+plugin, "-filetype=obj", llvmIR, "-o", goallcObject)

	native := readLLVMABIObject(t, nativeObject)
	goallc := readLLVMABIObject(t, goallcObject)
	cases := []llvmABICase{
		{
			name: "checkpoint", args: 8, pointerBits: []int{0},
			nativeArgsMaps:  [][]int{{0}, nil},
			goallcArgsMaps:  [][]int{{0}, nil, nil},
			nativeStackMaps: []int32{-1, 0, 1, -1},
			goallcStackMaps: []int32{-1, 1, 0, 2},
		},
		{
			name: "mixedABI", args: 152, pointerBits: []int{2, 4, 18},
			nativeArgsMaps:  [][]int{{2, 4, 18}, nil},
			goallcArgsMaps:  [][]int{{2, 4, 18}, {2}},
			nativeStackMaps: []int32{-1, 0, -1},
			goallcStackMaps: []int32{-1, 1, 0},
		},
		{
			name: "growPointer", args: 16, pointerBits: []int{0},
			nativeArgsMaps:  [][]int{{0}, nil},
			goallcArgsMaps:  [][]int{{0}, nil},
			nativeStackMaps: []int32{-1, 0, -1},
			goallcStackMaps: []int32{-1, 1, 0},
		},
		{
			name: "overflowResults", args: 48, pointerBits: []int{2, 3, 4, 5},
			nativeArgsMaps:  [][]int{{2, 3, 4, 5}, nil},
			goallcArgsMaps:  [][]int{{2, 3, 4, 5}, nil},
			nativeStackMaps: []int32{-1, 0, -1},
			goallcStackMaps: []int32{-1, 1, 0},
		},
		{
			name: "initializedStackResult", args: 16, pointerBits: []int{1},
			nativeArgsMaps:  [][]int{{1}, nil},
			goallcArgsMaps:  [][]int{{1}, nil},
			nativeStackMaps: []int32{-1, 0, -1},
			goallcStackMaps: []int32{-1, 1, 0},
		},
		{
			name: "stackAggregateResult", args: 40, pointerBits: []int{4},
			nativeArgsMaps:  [][]int{{4}, nil},
			goallcArgsMaps:  [][]int{{4}, nil},
			nativeStackMaps: []int32{-1, 0, -1},
			goallcStackMaps: []int32{-1, 1, 0},
		},
		{
			name: "bothOverflow", args: 168, pointerBits: []int{2, 6, 20},
			nativeArgsMaps:  [][]int{{2, 6, 20}, nil},
			goallcArgsMaps:  [][]int{{2, 6, 20}, {2}, nil},
			nativeStackMaps: []int32{-1, 0, 1, -1},
			goallcStackMaps: []int32{-1, 1, 0, 2},
		},
		{
			name: "requireAggregate", args: 40, pointerBits: []int{4},
			nativeArgsMaps:  [][]int{{4}, nil},
			goallcArgsMaps:  [][]int{{4}, nil, nil},
			nativeStackMaps: []int32{-1, 0, 1, -1},
			goallcStackMaps: []int32{-1, 1, 2, 0},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nativeSymbol := findLLVMABISymbol(t, native, "main."+tc.name)
			goallcSymbol := findLLVMABISymbol(t, goallc, "main."+tc.name)
			checkLLVMABISymbol(t, "native", nativeSymbol, tc)
			checkLLVMABISymbol(t, "GoALLC", goallcSymbol, tc)

			if nativeSymbol.Function.Info.Args != goallcSymbol.Function.Info.Args {
				t.Fatalf("FuncInfo args differ: native=%d GoALLC=%d",
					nativeSymbol.Function.Info.Args, goallcSymbol.Function.Info.Args)
			}
			nativeBits := llvmABIEntryPointerBits(t, nativeSymbol)
			goallcBits := llvmABIEntryPointerBits(t, goallcSymbol)
			if !reflect.DeepEqual(nativeBits, goallcBits) {
				t.Fatalf("entry ArgsPointerMaps differ: native=%v GoALLC=%v",
					nativeBits, goallcBits)
			}
			checkLLVMABIStackMaps(t, "native", nativeSymbol,
				tc.nativeArgsMaps, tc.nativeStackMaps)
			checkLLVMABIStackMaps(t, "GoALLC", goallcSymbol,
				tc.goallcArgsMaps, tc.goallcStackMaps)
		})
	}

	t.Run("source-args-pointer-maps", func(t *testing.T) {
		runLLVMABIArgsPointerMapSourceTest(t, gorootTestDir, llc, opt, plugin)
	})
	t.Run("machine-args-pointer-maps", func(t *testing.T) {
		runLLVMABIArgsPointerMapMachineTest(t, filepath.Dir(gorootTestDir), llc, plugin)
	})
}

func runLLVMABIArgsPointerMapSourceTest(t *testing.T, gorootTestDir, llc, opt, plugin string) {
	t.Helper()
	dir := t.TempDir()
	source := filepath.Join(gorootTestDir, "abi", "llvm_args_pointer_maps.go")
	nativeObject := filepath.Join(dir, "native.o")
	goallcArchive := filepath.Join(dir, "goallc.a")
	goallcIR := goallcArchive + ".ll"
	goallcObject := filepath.Join(dir, "goallc.o")

	nativeAssembly := runLLVMABICommand(t, nil, goTool, "tool", "compile",
		"-S", "-l", "-p=p", "-o", nativeObject, source)
	for _, pattern := range []string{
		`(?s)TEXT\s+p\.initializedPointerResult.*?CALL\s+p\.safepoint\(SB\).*?MOVD\s+R0,\s+p\.result\(FP\)`,
		`(?s)TEXT\s+p\.liveScalarStackArgument.*?CALL\s+p\.safepoint\(SB\).*?MOVD\s+p\.pointer\(FP\),\s+R0`,
		`(?s)TEXT\s+p\.liveAggregateStackArgument.*?CALL\s+p\.safepoint\(SB\).*?MOVD\s+p\.value\(FP\),\s+R0.*?MOVD\s+p\.value\+16\(FP\),\s+R1`,
	} {
		if !regexp.MustCompile(pattern).Match(nativeAssembly) {
			t.Fatalf("native assembly does not match %q", pattern)
		}
	}

	runLLVMABICommand(t, nil, goTool, "tool", "compile",
		"-l", "-p=p", "-enablellvm", "-llvmironly", "-o", goallcArchive, source)
	runLLVMABICommand(t, nil, opt, "-passes=verify", "-disable-output", goallcIR)
	rewrittenIR := runLLVMABICommand(t, nil, llc,
		"-load-pass-plugin="+plugin, "-goallc-pass-plugin-emit-ir",
		"-filetype=null", "-o", "-", goallcIR)
	for _, pattern := range []string{
		`(?s)define goabiinternal ptr @p\.liveScalarStackArgument.*?"gc-live"\(ptr %pointer\).*?gc\.relocate`,
		`(?s)define goabiinternal \{ ptr, ptr \} @p\.liveAggregateStackArgument.*?"gc-live"\(ptr %value\.leaf\.2, ptr %value\.leaf\.0\).*?gc\.relocate`,
	} {
		if !regexp.MustCompile(pattern).Match(rewrittenIR) {
			t.Fatalf("rewritten source IR does not match %q", pattern)
		}
	}
	runLLVMABICommand(t, rewrittenIR, opt, "-load-pass-plugin="+plugin,
		"-passes=verify", "-disable-output", "-")

	machineIR := runLLVMABICommand(t, nil, llc,
		"-load-pass-plugin="+plugin, "-stop-after=finalize-isel",
		"-o", "-", goallcIR)
	for _, pattern := range []string{
		`(?s)name:\s+p\.liveScalarStackArgument.*?fixedStack:.*?isImmutable:\s+false.*?stack:\s+\[\].*?STATEPOINT[^\n]*%fixed-stack\.0.*?LDRXui\s+%fixed-stack\.0`,
		`(?s)name:\s+p\.liveAggregateStackArgument.*?fixedStack:.*?stack:\s+\[\].*?STATEPOINT[^\n]*%fixed-stack\.2[^\n]*%fixed-stack\.0.*?LDRXui\s+%fixed-stack\.[02].*?LDRXui\s+%fixed-stack\.[02]`,
	} {
		if !regexp.MustCompile(pattern).Match(machineIR) {
			t.Fatalf("source MIR does not match %q", pattern)
		}
	}
	runLLVMABICommand(t, nil, llc, "-load-pass-plugin="+plugin,
		"-filetype=obj", goallcIR, "-o", goallcObject)

	native := readLLVMABIObject(t, nativeObject)
	goallc := readLLVMABIObject(t, goallcObject)
	type sourceCase struct {
		name          string
		args          uint32
		entryBits     []int
		nativeLocals  uint32
		goallcLocals  uint32
		nativeArgs    [][]int
		goallcArgs    [][]int
		nativeMaps    [][]int
		goallcMaps    [][]int
		nativePCData  []int32
		goallcPCData  []int32
		nativeQueries []int32
		goallcQueries []int32
	}
	cases := []sourceCase{
		{
			name: "initializedPointerResult", args: 16, entryBits: []int{1},
			nativeLocals: 8, goallcLocals: 24,
			nativeArgs: [][]int{{1}, nil}, goallcArgs: [][]int{{1}, nil},
			nativeMaps: [][]int{nil, nil}, goallcMaps: [][]int{nil, {1}},
			nativePCData: []int32{-1, 0, -1}, goallcPCData: []int32{-1, 1, 0},
			nativeQueries: []int32{0, -1}, goallcQueries: []int32{1, 0},
		},
		{
			name: "partiallyInitializedAggregateResult", args: 40, entryBits: []int{3, 4},
			nativeLocals: 8, goallcLocals: 24,
			nativeArgs: [][]int{{3, 4}, nil}, goallcArgs: [][]int{{3, 4}, nil},
			nativeMaps: [][]int{nil, nil}, goallcMaps: [][]int{nil, {0, 1}},
			nativePCData: []int32{-1, 0, -1}, goallcPCData: []int32{-1, 1, 0},
			nativeQueries: []int32{0, 0, -1}, goallcQueries: []int32{1, 1, 0},
		},
		{
			name: "liveScalarStackArgument", args: 136, entryBits: []int{0},
			nativeLocals: 8, goallcLocals: 8,
			nativeArgs: [][]int{{0}, nil}, goallcArgs: [][]int{{0}},
			nativeMaps: [][]int{nil, nil}, goallcMaps: [][]int{nil},
			nativePCData: []int32{-1, 0, -1}, goallcPCData: []int32{-1, 0},
			nativeQueries: []int32{0, -1}, goallcQueries: []int32{0, 0},
		},
		{
			name: "liveAggregateStackArgument", args: 136, entryBits: []int{0, 2},
			nativeLocals: 8, goallcLocals: 8,
			nativeArgs: [][]int{{0, 2}, nil}, goallcArgs: [][]int{{0, 2}},
			nativeMaps: [][]int{nil, nil}, goallcMaps: [][]int{nil},
			nativePCData: []int32{-1, 0, -1}, goallcPCData: []int32{-1, 0},
			nativeQueries: []int32{0, -1}, goallcQueries: []int32{0, 0},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nativeSymbol := findLLVMABISymbol(t, native, "p."+tc.name)
			goallcSymbol := findLLVMABISymbol(t, goallc, "p."+tc.name)
			base := llvmABICase{name: tc.name, args: tc.args, pointerBits: tc.entryBits}
			checkLLVMABISymbol(t, "native", nativeSymbol, base)
			checkLLVMABISymbol(t, "GoALLC", goallcSymbol, base)
			checkLLVMABISourceStackMaps(t, "native", nativeSymbol,
				tc.nativeLocals, tc.nativeArgs, tc.nativeMaps, tc.nativePCData, tc.nativeQueries)
			checkLLVMABISourceStackMaps(t, "GoALLC", goallcSymbol,
				tc.goallcLocals, tc.goallcArgs, tc.goallcMaps, tc.goallcPCData, tc.goallcQueries)
		})
	}
}

func runLLVMABIArgsPointerMapMachineTest(t *testing.T, goroot, llc, plugin string) {
	t.Helper()
	object := filepath.Join(t.TempDir(), "args-pointer-maps.o")
	source := filepath.Join(goroot, "src", "cmd", "internal", "testdir",
		"testdata", "llvm_args_pointer_maps.mir")
	runLLVMABICommand(t, nil, llc,
		"-load-pass-plugin="+plugin,
		"-mtriple=aarch64-apple-darwin-goobj",
		"-start-after=prolog-epilog",
		"-filetype=obj", source, "-o", object)

	symbol := findLLVMABISymbol(t, readLLVMABIObject(t, object), "p.argResult")
	if symbol.Function == nil || symbol.Function.Info == nil {
		t.Fatal("machine ArgsPointerMaps symbol has no FuncInfo")
	}
	if got := symbol.Function.Info.Args; got != 16 {
		t.Fatalf("machine ArgsPointerMaps FuncInfo args=%d, want 16", got)
	}
	if got, want := llvmABIStackMapBitmaps(t, symbol, "args_pointer_maps"),
		[][]int{{1}, {0}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("machine ArgsPointerMaps=%v, want %v", got, want)
	}
	if got, want := llvmABIStackMapBitmaps(t, symbol, "locals_pointer_maps"),
		[][]int{nil, nil}; !reflect.DeepEqual(got, want) {
		t.Fatalf("machine LocalsPointerMaps=%v, want %v", got, want)
	}
	if got, want := llvmABIStackMapRanges(symbol), []int32{-1, 0, 1}; !reflect.DeepEqual(got, want) {
		t.Fatalf("machine PCDATA_StackMapIndex=%v, want %v", got, want)
	}
	var queryIndexes []int32
	for _, query := range symbol.Function.StackMapQueries {
		if query.DecodeError != "" {
			t.Fatalf("machine stack-map query failed: %s", query.DecodeError)
		}
		queryIndexes = append(queryIndexes, query.StackMapIndex)
	}
	if got, want := queryIndexes, []int32{0, 1}; !reflect.DeepEqual(got, want) {
		t.Fatalf("machine stack-map query indexes=%v, want %v", got, want)
	}
	t.Logf("machine ArgsPointerMaps=%v LocalsPointerMaps=%v PCDATA=%v",
		llvmABIStackMapBitmaps(t, symbol, "args_pointer_maps"),
		llvmABIStackMapBitmaps(t, symbol, "locals_pointer_maps"),
		llvmABIStackMapRanges(symbol))
}

func checkLLVMABIStatepointTupleAttrs(t *testing.T, ir []byte, callees ...string) {
	t.Helper()
	for _, callee := range callees {
		call := regexp.MustCompile(`(?m)^.*@llvm\.experimental\.gc\.statepoint.*@main\.` +
			regexp.QuoteMeta(callee) + `.*#([0-9]+)$`).FindSubmatch(ir)
		if len(call) != 2 {
			t.Fatalf("rewritten IR has no attributed statepoint call to main.%s", callee)
		}
		attr := regexp.MustCompile(`(?m)^attributes #` + string(call[1]) +
			` = \{[^\n]*"go_results_tuple"`).Find(ir)
		if attr == nil {
			t.Fatalf("statepoint call to main.%s lost go_results_tuple", callee)
		}
	}
}

func checkLLVMABIAssembly(t *testing.T, native, goallc []byte) {
	t.Helper()

	// The native listing is the ABI reference for frame/argument size. These
	// signatures deliberately exercise input-only, result-only, and combined
	// overflow layouts.
	for _, pattern := range []string{
		`TEXT\s+main\.mixedABI\(SB\), ABIInternal, \$[0-9]+-152`,
		`TEXT\s+main\.overflowResults\(SB\), ABIInternal, \$[0-9]+-48`,
		`TEXT\s+main\.stackAggregateResult\(SB\), ABIInternal, \$[0-9]+-40`,
		`TEXT\s+main\.bothOverflow\(SB\), ABIInternal, \$[0-9]+-168`,
		`(?s)CALL\s+main\.overflowResults\(SB\).*?MOVD\s+R15,.*?LDP\s+8\(RSP\)`,
		`(?s)CALL\s+main\.stackAggregateResult\(SB\).*?LDP\s+8\(RSP\).*?MOVD\s+R15,`,
		`(?s)CALL\s+main\.bothOverflow\(SB\).*?MOVD\s+R15,.*?LDP\s+32\(RSP\)`,
	} {
		if !regexp.MustCompile(pattern).Match(native) {
			t.Fatalf("native assembly does not match %q", pattern)
		}
	}

	// GoALLC may choose different virtual/temporary registers, but the Go ABI
	// stack-result offsets and the register-stack-register result order must
	// agree with the native listing.
	for _, pattern := range []string{
		`(?m)^main\.mixedABI:`,
		`(?m)^main\.overflowResults:`,
		`(?m)^main\.stackAggregateResult:`,
		`(?m)^main\.bothOverflow:`,
		`(?s)\bbl\s+main\.overflowResults.*?\bldp\s+x[0-9]+, x[0-9]+, \[sp, #8\]`,
		`(?s)\bbl\s+main\.overflowResults.*?\b(?:mov|stp)\s+[^\n]*x15`,
		`(?s)\bbl\s+main\.stackAggregateResult.*?\bldp\s+x[0-9]+, x[0-9]+, \[sp, #8\].*?\bmov\s+x[0-9]+, x15`,
		`(?s)\bbl\s+main\.bothOverflow.*?\bldp\s+x[0-9]+, x[0-9]+, \[sp, #32\]`,
		`(?s)\bbl\s+main\.bothOverflow.*?\b(?:mov|stp)\s+[^\n]*x15`,
	} {
		if !regexp.MustCompile(pattern).Match(goallc) {
			t.Fatalf("GoALLC assembly does not match %q\n%s", pattern,
				llvmABIAssemblyExcerpt(goallc, "bl\tmain.overflowResults"))
		}
	}
}

func llvmABIAssemblyExcerpt(assembly []byte, needle string) []byte {
	start := bytes.Index(assembly, []byte(needle))
	if start < 0 {
		return assembly
	}
	end := start + 1200
	if end > len(assembly) {
		end = len(assembly)
	}
	return assembly[start:end]
}

func runLLVMABICommand(t *testing.T, stdin []byte, name string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), "GOENV=off", "GOFLAGS=")
	cmd.Stdin = bytes.NewReader(stdin)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s failed: %v\n%s", name, strings.Join(args, " "), err, out)
	}
	return out
}

func llvmABIPassPlugin(t *testing.T, llc string) string {
	t.Helper()
	if configured := os.Getenv("GOALLC_PASS_PLUGIN"); configured != "" {
		return configured
	}
	suffix := ".so"
	if runtime.GOOS == "darwin" {
		suffix = ".dylib"
	}
	root := filepath.Dir(filepath.Dir(llc))
	for _, libdir := range []string{"lib", "lib64"} {
		path := filepath.Join(root, libdir, "GoALLCStatepoints"+suffix)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}
	t.Fatalf("GoALLC pass plugin is unavailable next to %s; set GOALLC_PASS_PLUGIN", llc)
	return ""
}

func readLLVMABIObject(t *testing.T, object string) llvmABIDocument {
	t.Helper()
	out := runLLVMABICommand(t, nil, goTool, "tool", "objview", "-json", object)
	var document llvmABIDocument
	if err := json.Unmarshal(out, &document); err != nil {
		t.Fatalf("decode objview output for %s: %v", object, err)
	}
	return document
}

func findLLVMABISymbol(t *testing.T, document llvmABIDocument, name string) llvmABISymbol {
	t.Helper()
	for _, member := range document.Members {
		if member.GoObject == nil {
			continue
		}
		for _, symbol := range member.GoObject.Symbols {
			if symbol.Name == name {
				return symbol
			}
		}
	}
	t.Fatalf("objview output has no symbol %s", name)
	return llvmABISymbol{}
}

func checkLLVMABISymbol(t *testing.T, backend string, symbol llvmABISymbol, tc llvmABICase) {
	t.Helper()
	if symbol.ABI != 1 {
		t.Fatalf("%s symbol ABI=%d, want ABIInternal (1)", backend, symbol.ABI)
	}
	if symbol.Function == nil || symbol.Function.Info == nil {
		t.Fatalf("%s symbol has no FuncInfo", backend)
	}
	if symbol.Function.Info.Args != tc.args {
		t.Fatalf("%s FuncInfo args=%d, want %d", backend, symbol.Function.Info.Args, tc.args)
	}
	if got := llvmABIEntryPointerBits(t, symbol); !reflect.DeepEqual(got, tc.pointerBits) {
		t.Fatalf("%s entry pointer bits=%v, want %v", backend, got, tc.pointerBits)
	}

	haveStackMapIndex := false
	for _, pcdata := range symbol.Function.PCData {
		if pcdata.Kind == "stack_map_index" && len(pcdata.Ranges) != 0 {
			haveStackMapIndex = true
		}
	}
	if !haveStackMapIndex {
		t.Fatalf("%s has no decoded PCDATA_StackMapIndex", backend)
	}
	for _, query := range symbol.Function.StackMapQueries {
		if query.DecodeError != "" {
			t.Fatalf("%s stack-map query failed: %s", backend, query.DecodeError)
		}
	}
}

func checkLLVMABIStackMaps(t *testing.T, backend string, symbol llvmABISymbol, wantArgs [][]int, wantPCData []int32) {
	t.Helper()
	args := llvmABIArgsPointerBitmaps(t, symbol)
	pcdata := llvmABIStackMapRanges(symbol)
	if !reflect.DeepEqual(args, wantArgs) {
		t.Fatalf("%s ArgsPointerMaps=%v, want %v", backend, args, wantArgs)
	}
	if !reflect.DeepEqual(pcdata, wantPCData) {
		t.Fatalf("%s PCDATA_StackMapIndex=%v, want %v", backend, pcdata, wantPCData)
	}
	t.Logf("%s ArgsPointerMaps=%v PCDATA=%v", backend, args, pcdata)
}

func checkLLVMABISourceStackMaps(t *testing.T, backend string, symbol llvmABISymbol, wantLocals uint32, wantArgs, wantMaps [][]int, wantPCData, wantQueries []int32) {
	t.Helper()
	if got := symbol.Function.Info.Locals; got != wantLocals {
		t.Fatalf("%s FuncInfo locals=%d, want %d", backend, got, wantLocals)
	}
	checkLLVMABIStackMaps(t, backend, symbol, wantArgs, wantPCData)
	if got := llvmABIStackMapBitmaps(t, symbol, "locals_pointer_maps"); !reflect.DeepEqual(got, wantMaps) {
		t.Fatalf("%s LocalsPointerMaps=%v, want %v", backend, got, wantMaps)
	}
	if got := llvmABIStackMapQueryIndexes(symbol); !reflect.DeepEqual(got, wantQueries) {
		t.Fatalf("%s stack-map query indexes=%v, want %v", backend, got, wantQueries)
	}
	t.Logf("%s FuncInfo.locals=%d LocalsPointerMaps=%v stack-map queries=%v",
		backend, wantLocals, wantMaps, wantQueries)
}

func llvmABIEntryPointerBits(t *testing.T, symbol llvmABISymbol) []int {
	t.Helper()
	bitmaps := llvmABIArgsPointerBitmaps(t, symbol)
	return bitmaps[0]
}

func llvmABIArgsPointerBitmaps(t *testing.T, symbol llvmABISymbol) [][]int {
	t.Helper()
	return llvmABIStackMapBitmaps(t, symbol, "args_pointer_maps")
}

func llvmABIStackMapBitmaps(t *testing.T, symbol llvmABISymbol, kind string) [][]int {
	t.Helper()
	for _, data := range symbol.Function.FuncData {
		if data.Kind != kind {
			continue
		}
		if data.StackMap == nil || data.StackMap.Count == 0 ||
			len(data.StackMap.Bitmaps) == 0 {
			t.Fatalf("%s has no decoded %s", symbol.Name, kind)
		}
		bitmaps := make([][]int, 0, len(data.StackMap.Bitmaps))
		for _, bitmap := range data.StackMap.Bitmaps {
			bits := append([]int(nil), bitmap.SetBits...)
			sort.Ints(bits)
			bitmaps = append(bitmaps, bits)
		}
		return bitmaps
	}
	t.Fatalf("%s has no %s", symbol.Name, kind)
	return nil
}

func llvmABIStackMapRanges(symbol llvmABISymbol) []int32 {
	for _, data := range symbol.Function.PCData {
		if data.Kind != "stack_map_index" {
			continue
		}
		values := make([]int32, 0, len(data.Ranges))
		for _, r := range data.Ranges {
			values = append(values, r.Value)
		}
		return values
	}
	return nil
}

func llvmABIStackMapQueryIndexes(symbol llvmABISymbol) []int32 {
	indexes := make([]int32, 0, len(symbol.Function.StackMapQueries))
	for _, query := range symbol.Function.StackMapQueries {
		indexes = append(indexes, query.StackMapIndex)
	}
	return indexes
}
