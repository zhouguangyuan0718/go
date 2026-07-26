//go:build llvm23 || (!byollvm && !llvm14 && !llvm15 && !llvm16 && !llvm17 && !llvm18 && !llvm19 && !llvm20 && !llvm21 && !llvm22)

package llvm

/*
#include "llvm-c/Core.h"
*/
import "C"

func (v Value) IsABranchInst() (rv Value) {
	rv.C = C.LLVMIsAUncondBrInst(v.C)
	if rv.C == nil {
		rv.C = C.LLVMIsACondBrInst(v.C)
	}
	return
}

func (v Value) IsAUncondBrInst() (rv Value) {
	rv.C = C.LLVMIsAUncondBrInst(v.C)
	return
}

func (v Value) IsACondBrInst() (rv Value) {
	rv.C = C.LLVMIsACondBrInst(v.C)
	return
}
