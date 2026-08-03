//go:build compiler_bootstrap

package ssa

import "cmd/compile/internal/types"

func LLVMCompile(f *Func) {}

func InitModule(pkg *types.Pkg) {
}

func LowerGoObjData() {
}

func FinalizeGoObjContentHashes() {
}

func Output(fileName string) error {
	return nil
}
