// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !compiler_bootstrap

package ssa

import (
	"cmd/compile/internal/ir"
	"cmd/compile/internal/types"
	"cmd/internal/obj"
	"internal/buildcfg"
	"slices"
	"strings"

	"github.com/goallc/go-llvm"
)

// llvmCPUProfile is generated registry data, not a mutable per-function state.
// predicate is used for dispatch and source tests; capabilities only for ISA
// legality. A feature floor supplies capabilities without supplying predicates.
type llvmCPUProfile struct {
	name, arch, field, runtimeGuard string
	predicate, capabilities         uint64
}

func llvmCPUProfileByName(name string) *llvmCPUProfile {
	for i := range llvmCPUProfiles {
		if llvmCPUProfiles[i].name == name {
			return &llvmCPUProfiles[i]
		}
	}
	return nil
}

func llvmCPUProfileByRuntimeGuard(arch, symbol string) string {
	if symbol != "" {
		for _, p := range llvmCPUProfiles {
			if p.arch == arch && p.runtimeGuard == symbol {
				return p.name
			}
		}
	}
	return ""
}

func llvmBaselineCPUCapabilities(arch string, amd64Level int, arm64LSE bool) uint64 {
	switch arch {
	case "amd64":
		switch {
		case amd64Level >= 4:
			return llvmCPUBaselineV4
		case amd64Level >= 3:
			return llvmCPUBaselineV3
		case amd64Level >= 2:
			return llvmCPUBaselineV2
		}
	case "arm64":
		if arm64LSE {
			return llvmCPUBaselineLSE
		}
	}
	return 0
}

func llvmCPUProfileCoveredByBaseline(arch, profile string) bool {
	if profile == "" {
		return true
	}
	p := llvmCPUProfileByName(profile)
	caps := llvmBaselineCPUCapabilities(arch, buildcfg.GOAMD64, buildcfg.GOARM64.LSE)
	return p != nil && p.arch == arch && caps&p.capabilities == p.capabilities
}

func llvmFunctionName(f *Func) string {
	name := f.Name
	if f.OwnAux != nil && f.OwnAux.Fn != nil {
		name = f.OwnAux.Fn.Name
	}
	return name
}

func llvmMidwaySIMDFeatureFloor(f *Func) (string, bool) {
	switch name := llvmFunctionName(f); {
	case strings.Contains(name, "@simd512"):
		return goCPUProfileX86AVX512, true
	case strings.Contains(name, "@simd256"):
		return goCPUProfileX86AVX2, true
	case strings.Contains(name, "@simd128"):
		return goCPUProfileX86AVX, true
	case strings.Contains(name, "@simd0"):
		return "", true
	}
	return "", false
}

// llvmSIMDFeatureFloor returns the function-wide target features established
// by Go's SSA CPU-feature analysis in the entry block or by a Midway variant's
// selected width. The generic Vec256 ABI itself needs AVX, while @simd256 is
// reached only after the portable dispatcher has observed HasAVX2; keep those
// two contracts distinct. This is a precondition, not a new runtime dispatch
// request: the shared early CPU-feature pass consumes the attribute and adds
// the target feature.
func llvmSIMDFeatureFloor(f *Func) string {
	if f.Config.arch != "amd64" {
		return ""
	}
	if floor, midway := llvmMidwaySIMDFeatureFloor(f); midway {
		return floor
	}
	features := CPUNone
	if f.Entry != nil {
		features = f.Entry.CPUfeatures
	}
	switch {
	case features.hasFeature(CPUavx512):
		return goCPUProfileX86AVX512
	case features.hasFeature(CPUavx2):
		return goCPUProfileX86AVX2
	case features.hasFeature(CPUavx):
		return goCPUProfileX86AVX
	}
	return ""
}

func llvmCPUProfileCoveredByFloor(required, floor string) bool {
	return llvmCPUProfileSupplies(floor, required)
}

// llvmWideVectorTypeWidth returns the widest fixed SIMD vector nested directly
// in a Go ABI value. Pointers deliberately stop the walk: their referents do
// not cross the call boundary in vector registers.
func llvmWideVectorTypeWidth(t *types.Type) int64 {
	if t == nil {
		return 0
	}
	if t.IsSIMD() {
		return t.Size()
	}
	switch {
	case t.IsArray():
		return llvmWideVectorTypeWidth(t.Elem())
	case t.IsStruct():
		var width int64
		for _, field := range t.Fields() {
			if candidate := llvmWideVectorTypeWidth(field.Type); candidate > width {
				width = candidate
			}
		}
		return width
	}
	return 0
}

// llvmWideVectorCallWidth reports wide vectors that the Go ABI assigns to
// registers for this call. Stack-only byval arguments and result homes are not
// LLVM register-ABI carriers and therefore need no caller target feature.
func llvmWideVectorCallWidth(aux *AuxCall) int64 {
	var width int64
	for i := int64(0); i < aux.NArgs(); i++ {
		if len(aux.RegsOfArg(i)) == 0 {
			continue
		}
		if candidate := llvmWideVectorTypeWidth(aux.TypeOfArg(i)); candidate > width {
			width = candidate
		}
	}
	for i := int64(0); i < aux.NResults(); i++ {
		if len(aux.RegsOfResult(i)) == 0 {
			continue
		}
		if candidate := llvmWideVectorTypeWidth(aux.TypeOfResult(i)); candidate > width {
			width = candidate
		}
	}
	return width
}

func llvmWideVectorCPUProfile(width int64) string {
	switch {
	case width > 32:
		return goCPUProfileX86AVX512
	case width > 16:
		return goCPUProfileX86AVX
	}
	return ""
}

func llvmCPUProfileSupplies(profile, required string) bool {
	if required == "" {
		return true
	}
	p, r := llvmCPUProfileByName(profile), llvmCPUProfileByName(required)
	return p != nil && r != nil && p.arch == r.arch && p.capabilities&r.capabilities == r.capabilities
}

// llvmX86CPUFeatureGuard recognizes which successor of an ordinary
// internal/cpu.X86 feature test has the feature enabled. This mirrors the
// source shape accepted by the native cpufeatures pass without changing that
// pass or making fixed-width archsimd participate in Midway rewriting.
func llvmX86CPUFeatureGuard(b *Block) (profile string, enabled int) {
	if b.Kind != BlockIf || b.Controls[0] == nil || b.Controls[1] != nil || len(b.Succs) != 2 {
		return "", -1
	}
	condition := b.Controls[0]
	taken := 0
	if condition.Op == OpNot {
		if len(condition.Args) != 1 {
			return "", -1
		}
		taken = 1
		condition = condition.Args[0]
	}
	profile = llvmX86CPUFeatureProfile(llvmX86CPUFeatureField(condition))
	if profile == "" {
		return "", -1
	}
	return profile, taken
}

// llvmCPUFeatureGuardProfiles finds effective predicates that protect every
// path to v. Keep the nearest single dominating guard when possible, avoiding
// extra clones for redundant outer conditions. A successor block alone is not
// enough: another incoming edge could bypass the enabled guard edge.
func llvmCPUFeatureGuardProfiles(f *Func, v *Value, required string) []string {
	sdom := f.Sdom()
	for b := sdom.Parent(v.Block); b != nil; b = sdom.Parent(b) {
		profile, taken := llvmX86CPUFeatureGuard(b)
		if taken < 0 || !llvmCPUProfileSupplies(profile, required) {
			continue
		}
		enabled := b.Succs[taken].Block()
		if enabled == f.Entry || !sdom.IsAncestorEq(enabled, v.Block) {
			continue
		}
		dominates := true
		for _, pred := range enabled.Preds {
			if pred.Block() == b && pred.Index() == taken {
				continue
			}
			// Backedges do not provide a new entry to the guarded region.
			if !sdom.IsAncestorEq(enabled, pred.Block()) {
				dominates = false
				break
			}
		}
		if dominates {
			return []string{profile}
		}
	}

	// Cut each backwards path at an enabled edge supplying the requirement.
	// Reaching entry means some path is unguarded. Visited blocks also bound
	// traversal through loops, whose first entry must still cross the cut.
	seen := make(map[*Block]bool)
	profiles := make(map[string]bool)
	work := []*Block{v.Block}
	for len(work) != 0 {
		b := work[len(work)-1]
		work = work[:len(work)-1]
		if b == f.Entry {
			return nil
		}
		if seen[b] {
			continue
		}
		seen[b] = true
		for _, pred := range b.Preds {
			profile, taken := llvmX86CPUFeatureGuard(pred.Block())
			if taken == pred.Index() && llvmCPUProfileSupplies(profile, required) {
				profiles[profile] = true
				continue
			}
			work = append(work, pred.Block())
		}
	}
	var guards []string
	for profile := range profiles {
		guards = append(guards, profile)
	}
	slices.Sort(guards)
	return guards
}

func llvmCallAux(v *Value) *AuxCall {
	switch v.Op {
	case OpStaticCall, OpStaticLECall, OpTailLECall,
		OpClosureCall, OpClosureLECall,
		OpInterCall, OpInterLECall, OpTailLECallInter:
		return auxToCall(v.Aux)
	}
	return nil
}

type llvmCPURequirementKind uint8

const (
	llvmCPUGenerated llvmCPURequirementKind = iota
	llvmCPUCompilerOp
	llvmCPUWideCall
)

// Retain why the entry assumption exists instead of treating every floor as
// an ABI property. Entry SSA facts also include unconditional block effects.
type llvmCPUFeatureFloor struct {
	profile string
	source  string // "entry-ssa", "midway", or "wide-call"
	call    ID     // the call which raised the floor, for "wide-call"
}

type llvmCPUFeaturePlan struct {
	floor        llvmCPUFeatureFloor
	requirements map[ID]string
	guards       map[ID]string
	profiles     []string
}

func llvmCPURequirement(v *Value, arch string) (string, llvmCPURequirementKind) {
	if info, ok := goALLCSIMDInfo(v.Op); ok {
		return info.archInfo(arch).cpuProfile, llvmCPUGenerated
	}
	if name := llvmCPUOpProfiles[v.Op]; name != "" {
		if p := llvmCPUProfileByName(name); p.arch == arch {
			return name, llvmCPUCompilerOp
		}
	}
	if arch == "amd64" {
		if aux := llvmCallAux(v); aux != nil {
			return llvmWideVectorCPUProfile(llvmWideVectorCallWidth(aux)), llvmCPUWideCall
		}
	}
	return "", llvmCPUGenerated
}

// llvmPlanCPUFeatures is the one planning boundary for generated SIMD,
// compiler scalar/atomic intrinsics and wide register-ABI calls. Emission
// consumes the completed plan; it does not discover or mutate feature policy.
func llvmPlanCPUFeatures(f *Func) *llvmCPUFeaturePlan {
	plan := &llvmCPUFeaturePlan{
		floor:        llvmCPUFeatureFloor{profile: llvmSIMDFeatureFloor(f), source: "entry-ssa"},
		requirements: make(map[ID]string),
		guards:       make(map[ID]string),
	}
	if _, midway := llvmMidwaySIMDFeatureFloor(f); f.Config.arch == "amd64" && midway {
		plan.floor.source = "midway"
	}
	type guardKey struct {
		block    *Block
		required string
	}
	// Operations in one block share the same protection. In particular, do
	// not repeat a CFG walk for each vector operation in a large block.
	guardCache := make(map[guardKey][]string)
	findGuards := func(v *Value, required string) []string {
		key := guardKey{v.Block, required}
		if guards, ok := guardCache[key]; ok {
			return guards
		}
		guards := llvmCPUFeatureGuardProfiles(f, v, required)
		guardCache[key] = guards
		return guards
	}
	type requirement struct {
		value   *Value
		profile string
		kind    llvmCPURequirementKind
		guards  []string
	}
	var pending []requirement
	for _, b := range f.Blocks {
		for _, v := range b.Values {
			profile, kind := llvmCPURequirement(v, f.Config.arch)
			if profile == "" || llvmCPUProfileCoveredByBaseline(f.Config.arch, profile) {
				continue
			}
			r := requirement{value: v, profile: profile, kind: kind}
			if kind == llvmCPUWideCall {
				if llvmCPUProfileCoveredByFloor(profile, plan.floor.profile) {
					continue
				}
				r.guards = findGuards(v, profile)
			}
			pending = append(pending, r)
		}
	}
	// Preserve native fixed-width calling semantics: an unguarded wide call
	// raises the function floor, whereas a guarded call requests FMV. Finish
	// this step before handling any operation, independent of block order.
	for _, r := range pending {
		if r.kind == llvmCPUWideCall && len(r.guards) == 0 && !llvmCPUProfileCoveredByFloor(r.profile, plan.floor.profile) {
			plan.floor = llvmCPUFeatureFloor{profile: r.profile, source: "wide-call", call: r.value.ID}
		}
	}
	selected := make(map[string]bool)
	for _, r := range pending {
		if r.kind != llvmCPUCompilerOp && llvmCPUProfileCoveredByFloor(r.profile, plan.floor.profile) {
			continue
		}
		guards := r.guards
		if r.kind == llvmCPUGenerated {
			guards = findGuards(r.value, r.profile)
		}
		if len(guards) == 0 {
			// Compiler-generated guards use the operation's own predicate.
			// An unguarded SIMD operation also requests it so the baseline
			// verifier continues to reject the surviving requirement.
			guards = []string{r.profile}
		}
		plan.requirements[r.value.ID] = r.profile
		for _, guard := range guards {
			selected[guard] = true
		}
	}
	for _, name := range llvmCPURequestOrder {
		if selected[name] {
			plan.profiles = append(plan.profiles, name)
		}
	}
	for _, b := range f.Blocks {
		for _, v := range b.Values {
			name, selective := llvmCPUFeatureGuardValue(v, f.Config.arch)
			if name != "" && (!selective || selected[name]) {
				plan.guards[v.ID] = name
			}
		}
	}
	return plan
}

func llvmX86CPUFeatureProfile(field string) string {
	for _, p := range llvmCPUProfiles {
		if p.arch == "amd64" && p.field == field {
			return p.name
		}
	}
	return ""
}

// The accepted source shapes are unchanged: user archsimd X86 field loads
// are marked only for selected predicates; compiler-generated scalar/atomic
// guard loads retain their unconditional marker behavior.
func llvmCPUFeatureGuardValue(v *Value, arch string) (profile string, selective bool) {
	if v.Op == OpHasCPUFeature && arch == "amd64" {
		if sym, ok := v.Aux.(*obj.LSym); ok {
			return llvmCPUProfileByRuntimeGuard(arch, sym.Name), false
		}
	}
	if v.Op == OpLoad {
		if arch == "amd64" {
			return llvmX86CPUFeatureProfile(llvmX86CPUFeatureField(v)), true
		}
		if arch == "arm64" && len(v.Args) != 0 && v.Args[0].Op == OpAddr {
			if sym, ok := v.Args[0].Aux.(*obj.LSym); ok {
				return llvmCPUProfileByRuntimeGuard(arch, sym.Name), false
			}
		}
	}
	return "", false
}

// llvmX86CPUFeatureField recognizes the ordinary internal/cpu.X86 field load
// used by archsimd feature checks. Keep this LLVM-only matching separate from
// the native SSA CPU-feature analysis.
func llvmX86CPUFeatureField(v *Value) string {
	if v.Op != OpLoad || len(v.Args) == 0 {
		return ""
	}
	offPtr := v.Args[0]
	if offPtr.Op != OpOffPtr || len(offPtr.Args) == 0 {
		return ""
	}
	addr := offPtr.Args[0]
	if addr.Op != OpAddr || len(addr.Args) == 0 || addr.Args[0].Op != OpSB {
		return ""
	}
	sym, ok := addr.Aux.(*obj.LSym)
	if !ok || sym.Name != "internal/cpu.X86" {
		return ""
	}
	t := addr.Type
	if !t.IsPtr() {
		v.Fatalf("The symbol %s is not a pointer, found %v instead", sym.Name, t)
	}
	t = t.Elem()
	if !t.IsStruct() {
		v.Fatalf("The referent of symbol %s is not a struct, found %v instead", sym.Name, t)
	}
	for _, field := range t.Fields() {
		if offPtr.AuxInt == field.Offset && field.Sym != nil {
			return field.Sym.Name
		}
	}
	return ""
}

// A few lowering unit tests construct contexts directly. They get the same
// planner, once, rather than a second partial feature-discovery path.
func (lfc *LLVMFuncContext) cpuFeaturePlan() *llvmCPUFeaturePlan {
	if lfc.CPUFeatures == nil {
		lfc.CPUFeatures = llvmPlanCPUFeatures(lfc.F)
	}
	return lfc.CPUFeatures
}

func (lfc *LLVMFuncContext) requireCPUFeature(v *Value, instruction llvm.Value) {
	if profile := lfc.cpuFeaturePlan().requirements[v.ID]; profile != "" {
		instruction.SetMetadata(GlobalCtxt.MDKindID(goCPURequiresMD), GlobalCtxt.MDNode([]llvm.Metadata{GlobalCtxt.MDString(profile)}))
	}
}

func (lfc *LLVMFuncContext) markCPUFeatureGuard(v *Value, load llvm.Value) {
	if profile := lfc.cpuFeaturePlan().guards[v.ID]; profile != "" {
		load.SetMetadata(GlobalCtxt.MDKindID(goCPUGuardMD), GlobalCtxt.MDNode([]llvm.Metadata{GlobalCtxt.MDString(profile)}))
	}
}

func (lfc *LLVMFuncContext) finishCPUFeatures() {
	if profiles := lfc.cpuFeaturePlan().profiles; len(profiles) != 0 {
		// Keep the canonical imported GoObj LSym identity alive for the early
		// plugin, which materializes the load before normal optimization.
		llvmGoDataRef(ir.Syms.GoALLCCPUFeatures)
		lfc.LF.AddTargetDependentFunctionAttr(goCPUMultiversionAttr, strings.Join(profiles, ","))
	}
}
