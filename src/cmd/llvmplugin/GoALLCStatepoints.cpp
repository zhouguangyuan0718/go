// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "GoALLCStatepoints.h"
#include "llvm/ADT/MapVector.h"
#include "llvm/ADT/SetVector.h"
#include "llvm/ADT/SmallSet.h"
#include "llvm/ADT/StringRef.h"
#include "llvm/IR/CFG.h"
#include "llvm/IR/CallingConv.h"
#include "llvm/IR/Constants.h"
#include "llvm/IR/Dominators.h"
#include "llvm/IR/GCStrategy.h"
#include "llvm/IR/IRBuilder.h"
#include "llvm/IR/InstIterator.h"
#include "llvm/IR/Instructions.h"
#include "llvm/IR/Module.h"
#include "llvm/IR/Statepoint.h"
#include "llvm/IR/Verifier.h"
#include "llvm/Support/Error.h"
#include "llvm/Transforms/Utils/Local.h"

#include <optional>

using namespace llvm;

namespace {

constexpr StringLiteral GoALLCGCName = "goallc";
constexpr StringLiteral GoALLCLeafAttr = "goallc-gc-leaf";
constexpr StringLiteral GCLeafAttr = "gc-leaf-function";

// This strategy exists for statepoint verification and lowering. GoALLC owns
// statepoint insertion, so UseRS4GC deliberately remains false.
class GoALLCGCStrategy final : public GCStrategy {
public:
  GoALLCGCStrategy() {
    UseStatepoints = true;
    UseRS4GC = false;
    NeededSafePoints = false;
    // The GoALLC metadata printer consumes the Machine StackMaps records at
    // AsmPrinter finalization and hands their raw locations to the GoObj
    // writer. It does not use the legacy gc.root metadata model.
    UsesMetadata = true;
  }

  // Opaque pointers do not carry enough information to classify Go pointers.
  // The GoALLC pass performs value-level classification and remains
  // conservative for the first implementation.
  std::optional<bool> isGCManagedPointer(const Type *) const override {
    return std::nullopt;
  }
};

static GCRegistry::Add<GoALLCGCStrategy>
    GoALLCGCRegistration(GoALLCGCName,
                         "GoALLC value-level statepoint strategy");

using ValueSet = SetVector<Value *>;

struct LivenessData {
  MapVector<BasicBlock *, ValueSet> Kill;
  MapVector<BasicBlock *, ValueSet> Gen;
  MapVector<BasicBlock *, ValueSet> LiveIn;
  MapVector<BasicBlock *, ValueSet> LiveOut;
};

struct SafepointRecord {
  CallInst *Call;
  uint64_t ID;
  ValueSet Live;
};

bool isGoCallingConv(CallingConv::ID CC) {
  return CC == CallingConv::GoABIInternal || CC == CallingConv::GoABI0;
}

bool containsPointer(Type *Ty) {
  if (Ty->isPointerTy())
    return true;
  if (auto *ST = dyn_cast<StructType>(Ty))
    return llvm::any_of(ST->elements(), containsPointer);
  if (auto *AT = dyn_cast<ArrayType>(Ty))
    return containsPointer(AT->getElementType());
  if (auto *VT = dyn_cast<VectorType>(Ty))
    return containsPointer(VT->getElementType());
  return false;
}

bool isTrackedValue(const Value *V) {
  return !isa<Constant>(V) && containsPointer(V->getType());
}

bool isLeafCall(const CallBase &Call) {
  if (Call.hasFnAttr(GCLeafAttr))
    return true;
  if (const Function *Callee = Call.getCalledFunction())
    return Callee->isIntrinsic() || Callee->hasFnAttribute(GCLeafAttr) ||
           Callee->hasFnAttribute(GoALLCLeafAttr);
  return false;
}

uint64_t stableStatepointID(StringRef FunctionName, uint64_t CallOrdinal) {
  uint64_t Hash = 14695981039346656037ULL;
  auto Mix = [&](uint8_t Byte) {
    Hash ^= Byte;
    Hash *= 1099511628211ULL;
  };
  for (char Byte : FunctionName)
    Mix(static_cast<uint8_t>(Byte));
  Mix(0);
  for (unsigned I = 0; I != sizeof(CallOrdinal); ++I)
    Mix(static_cast<uint8_t>(CallOrdinal >> (I * 8)));
  return Hash;
}

void scanBackward(BasicBlock::reverse_iterator Begin,
                  BasicBlock::reverse_iterator End, ValueSet &Live) {
  for (Instruction &I : make_range(Begin, End)) {
    Live.remove(&I);
    if (isa<PHINode>(I))
      continue;
    for (Value *Operand : I.operands())
      if (isTrackedValue(Operand))
        Live.insert(Operand);
  }
}

void seedPhiUses(BasicBlock &BB, ValueSet &LiveOut) {
  for (BasicBlock *Succ : successors(&BB)) {
    for (Instruction &I : *Succ) {
      auto *Phi = dyn_cast<PHINode>(&I);
      if (!Phi)
        break;
      Value *Incoming = Phi->getIncomingValueForBlock(&BB);
      if (isTrackedValue(Incoming))
        LiveOut.insert(Incoming);
    }
  }
}

LivenessData computeLiveness(Function &F) {
  LivenessData Data;
  SmallSetVector<BasicBlock *, 32> Worklist;

  for (BasicBlock &BB : F) {
    ValueSet &Kill = Data.Kill[&BB];
    for (Instruction &I : BB)
      if (containsPointer(I.getType()))
        Kill.insert(&I);

    ValueSet &Gen = Data.Gen[&BB];
    scanBackward(BB.rbegin(), BB.rend(), Gen);

    ValueSet &Out = Data.LiveOut[&BB];
    seedPhiUses(BB, Out);

    ValueSet &In = Data.LiveIn[&BB];
    In = Gen;
    In.set_union(Out);
    In.set_subtract(Kill);
    if (!In.empty())
      Worklist.insert_range(predecessors(&BB));
  }

  while (!Worklist.empty()) {
    BasicBlock *BB = Worklist.pop_back_val();
    ValueSet &Out = Data.LiveOut[BB];
    size_t OldSize = Out.size();
    for (BasicBlock *Succ : successors(BB))
      Out.set_union(Data.LiveIn[Succ]);
    if (Out.size() == OldSize)
      continue;

    ValueSet NewIn = Out;
    NewIn.set_union(Data.Gen[BB]);
    NewIn.set_subtract(Data.Kill[BB]);
    if (NewIn.size() != Data.LiveIn[BB].size()) {
      Data.LiveIn[BB] = std::move(NewIn);
      Worklist.insert_range(predecessors(BB));
    }
  }
  return Data;
}

ValueSet liveAtCall(CallInst &Call, LivenessData &Data) {
  ValueSet Live = Data.LiveOut[Call.getParent()];
  scanBackward(Call.getParent()->rbegin(), ++Call.getIterator().getReverse(),
               Live);
  Live.remove(&Call);
  return Live;
}

Error validateSafepoint(const SafepointRecord &Record, DominatorTree &DT) {
  const CallInst &Call = *Record.Call;
  if (Call.isInlineAsm())
    return createStringError(
        std::errc::not_supported,
        "GoALLC statepoints do not support non-leaf inline assembly");
  if (Call.isMustTailCall())
    return createStringError(std::errc::not_supported,
                             "GoALLC statepoints do not support musttail");
  if (Call.getNumOperandBundles() != 0)
    return createStringError(
        std::errc::not_supported,
        "GoALLC statepoints do not yet support call operand bundles");
  for (unsigned I = 0; I != Call.arg_size(); ++I) {
    for (Attribute Attr : Call.getAttributes().getParamAttrs(I)) {
      if (!Attr.hasAttribute(Attribute::Nest))
        return createStringError(
            std::errc::not_supported,
            "GoALLC statepoints only support the nest call parameter "
            "attribute");
    }
  }
  for (Value *V : Record.Live) {
    if (!V->getType()->isPointerTy())
      return createStringError(
          std::errc::not_supported,
          "GoALLC statepoints do not yet support live pointer aggregates");
    for (const Use &U : V->uses()) {
      auto *User = dyn_cast<Instruction>(U.getUser());
      if (!User || User == &Call)
        continue;
      if (User->getParent() == Call.getParent() && !Call.comesBefore(User))
        continue;
      if (DT.dominates(&Call, U) || DT.dominates(User, &Call))
        continue;
      return createStringError(
          std::errc::not_supported,
          "GoALLC statepoints do not yet support conditional relocation PHIs");
    }
  }
  return Error::success();
}

Error rewriteCall(SafepointRecord &Record, DominatorTree &DT) {
  CallInst *Call = Record.Call;

  SmallVector<Value *, 8> CallArgs(Call->args());
  SmallVector<Value *, 8> GCLive(Record.Live.begin(), Record.Live.end());
  FunctionCallee Callee(Call->getFunctionType(), Call->getCalledOperand());

  IRBuilder<> Builder(Call);
  Builder.SetCurrentDebugLocation(Call->getDebugLoc());
  CallInst *Statepoint = Builder.CreateGCStatepointCall(
      Record.ID, 0, Callee, CallArgs, std::nullopt, GCLive, "statepoint_token");
  Statepoint->setCallingConv(Call->getCallingConv());
  for (unsigned I = 0; I != Call->arg_size(); ++I) {
    for (Attribute Attr : Call->getAttributes().getParamAttrs(I))
      Statepoint->addParamAttr(GCStatepointInst::CallArgsBeginPos + I, Attr);
  }

  Instruction *InsertBefore = Call->getNextNode();
  Builder.SetInsertPoint(InsertBefore);
  Builder.SetCurrentDebugLocation(Call->getDebugLoc());

  CallInst *Result = nullptr;
  if (!Call->getType()->isVoidTy() && !Call->use_empty()) {
    Result = Builder.CreateGCResult(Statepoint, Call->getType());
    Result->setAttributes(
        AttributeList::get(Call->getContext(), AttributeList::ReturnIndex,
                           Call->getAttributes().getRetAttrs()));
  }

  SmallVector<CallInst *, 8> Relocates;
  for (auto [Index, V] : llvm::enumerate(GCLive)) {
    CallInst *Relocate = Builder.CreateGCRelocate(
        Statepoint, Index, Index, V->getType(),
        V->hasName() ? V->getName() + ".relocated" : "");
    Relocate->setCallingConv(CallingConv::Cold);
    Relocates.push_back(Relocate);
  }

  if (Result)
    Call->replaceAllUsesWith(Result);
  Call->eraseFromParent();

  for (auto [Index, V] : llvm::enumerate(GCLive))
    replaceDominatedUsesWith(V, Relocates[Index], DT, Relocates[Index]);
  return Error::success();
}

Error rewriteFunction(Function &F) {
  if (F.hasFnAttribute(GoALLCLeafAttr)) {
    F.addFnAttr(GCLeafAttr);
    return Error::success();
  }

  DominatorTree DT(F);
  LivenessData Data = computeLiveness(F);
  SmallVector<SafepointRecord, 8> Records;
  uint64_t CallOrdinal = 0;

  for (Instruction &I : instructions(F)) {
    auto *Call = dyn_cast<CallBase>(&I);
    if (!Call || isa<GCStatepointInst>(Call) || isLeafCall(*Call))
      continue;
    auto *OrdinaryCall = dyn_cast<CallInst>(Call);
    if (!OrdinaryCall)
      return createStringError(
          std::errc::not_supported,
          "GoALLC statepoints do not yet support invoke or callbr");
    Records.push_back({OrdinaryCall,
                       stableStatepointID(F.getName(), CallOrdinal++),
                       liveAtCall(*OrdinaryCall, Data)});
  }
  for (const SafepointRecord &Record : Records)
    if (Error Err = validateSafepoint(Record, DT))
      return Err;
  for (SafepointRecord &Record : llvm::reverse(Records))
    if (Error Err = rewriteCall(Record, DT))
      return Err;
  return Error::success();
}

} // namespace

Error goallc::rewriteStatepoints(Module &M, TargetMachine &) {
  for (Function &F : M) {
    if (F.isDeclaration() || !isGoCallingConv(F.getCallingConv()) ||
        !F.hasGC() || F.getGC() != GoALLCGCName)
      continue;
    if (Error Err = rewriteFunction(F))
      return Err;
  }
  if (verifyModule(M, &errs()))
    return createStringError(std::errc::invalid_argument,
                             "GoALLC statepoint rewrite produced invalid IR");
  return Error::success();
}
