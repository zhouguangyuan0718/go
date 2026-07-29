// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#ifndef GOALLC_STATEPOINTS_H
#define GOALLC_STATEPOINTS_H

#include "llvm/Support/Error.h"

namespace llvm {

class Module;
class TargetMachine;

namespace goallc {

// Rewrites calls in Go ABI functions to statepoints using GoALLC's value
// liveness and relocation policy.
Error rewriteStatepoints(Module &M, TargetMachine &TM);

} // namespace goallc
} // namespace llvm

#endif // GOALLC_STATEPOINTS_H
