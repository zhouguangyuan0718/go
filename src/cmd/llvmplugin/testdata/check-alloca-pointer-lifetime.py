#!/usr/bin/env python3

import argparse
import json
import re
import subprocess
import sys
import tempfile


ALLOCA_TAG = 1095520067


def fail(message):
    print(f"alloca pointer lifetime check failed: {message}", file=sys.stderr)
    raise SystemExit(1)


def function_body(ir, name):
    match = re.search(
        rf"define goabiinternal [^@]+@{name}\b.*?^}}",
        ir,
        re.MULTILINE | re.DOTALL,
    )
    if not match:
        fail(f"missing function {name}")
    return match.group(0)


def check_goobj(llc, plugin, objview, input_path):
    with tempfile.TemporaryDirectory() as directory:
        object_path = f"{directory}/lifetime.goobj"
        result = subprocess.run(
            [
                llc,
                f"-load-pass-plugin={plugin}",
                "-verify-machineinstrs",
                "-filetype=obj",
                "-o",
                object_path,
                input_path,
            ],
            capture_output=True,
            text=True,
        )
        if result.returncode != 0:
            fail(f"GoObj emission failed:\n{result.stdout}{result.stderr}")
        result = subprocess.run(
            [objview, "-format=json", object_path],
            capture_output=True,
            text=True,
        )
        if result.returncode != 0:
            fail(f"objview failed:\n{result.stdout}{result.stderr}")
        document = json.loads(result.stdout)

    symbols = document["members"][0]["go_object"]["symbols"]
    matches = [
        symbol
        for symbol in symbols
        if symbol.get("name") == "stack_object_alloca_with_lifetime"
    ]
    if len(matches) != 1 or not matches[0].get("function"):
        fail("missing stack-object lifetime function in GoObj")
    function = matches[0]["function"]
    funcdata = {
        entry["kind"]: entry
        for entry in function.get("funcdata", [])
        if entry.get("kind")
    }
    stack_objects = funcdata.get("stack_objects", {}).get("stack_objects", [])
    if len(stack_objects) != 1:
        fail(f"function-wide StackObjects={stack_objects}, want one object")
    locals_maps = funcdata.get("locals_pointer_maps", {}).get("stack_map", {})
    bitmaps = locals_maps.get("bitmaps", [])
    if any(bitmap.get("set_bits") for bitmap in bitmaps):
        fail(f"StackObject lifetime polluted LocalsPointerMaps: {bitmaps}")
    queries = [
        query["stack_map_index"]
        for query in function.get("stack_map_queries", [])[:3]
    ]
    if queries != [0, 0, 0]:
        fail(f"StackObject callsite stack-map indices={queries}, want [0, 0, 0]")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--llc", required=True)
    parser.add_argument("--plugin", required=True)
    parser.add_argument("--input", required=True)
    parser.add_argument("--objview")
    args = parser.parse_args()

    result = subprocess.run(
        [
            args.llc,
            f"-load-pass-plugin={args.plugin}",
            "-goallc-pass-plugin-emit-ir",
            "-filetype=null",
            "-o",
            "-",
            args.input,
        ],
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        fail(f"llc failed:\n{result.stdout}{result.stderr}")
    ir = result.stdout + result.stderr

    locals_body = function_body(ir, "locals_pointer_alloca_with_lifetime")
    if locals_body.count("@llvm.experimental.gc.statepoint") != 3:
        fail("locals-only function does not contain three statepoints")
    if locals_body.count('"deopt"(') != 1 or f"i64 {ALLOCA_TAG}" not in locals_body:
        fail("direct-only lifetime is not present at exactly one statepoint")
    first_call = locals_body.find("@llvm.experimental.gc.statepoint")
    initialize = locals_body.find("store ptr null")
    active = locals_body.find(f"i64 {ALLOCA_TAG}")
    if min(first_call, initialize, active) < 0 or not first_call < initialize < active:
        fail("direct-only initialization is not inside its active region")
    if locals_body.count("call void @llvm.lifetime.start") != 1:
        fail("direct-only lifetime start was not preserved")
    if locals_body.count('"gc-live"(ptr %slot)') != 1 or locals_body.count(
        "%slot.relocated"
    ) != 1:
        fail("direct-only alloca is not one explicit rematerialized gc-live root")
    if "!llvm.stackcoloring.no_merge" in locals_body:
        fail("direct-only alloca unnecessarily disables stack coloring")

    stack_body = function_body(ir, "stack_object_alloca_with_lifetime")
    if stack_body.count("@llvm.experimental.gc.statepoint") != 3:
        fail("stack-object function does not contain three statepoints")
    if stack_body.count('"deopt"(') != 3 or stack_body.count(
        f"i64 {ALLOCA_TAG}"
    ) != 3:
        fail("stack object is not present at every statepoint")
    if stack_body.count("call void @llvm.lifetime.start") != 1:
        fail("stack-object lifetime start was not preserved")
    if '"gc-live"(ptr %slot)' in stack_body or "%slot.relocated" in stack_body:
        fail("call-only stack-object address entered caller live-out")
    if "alloca ptr, align 8, !llvm.stackcoloring.no_merge" not in stack_body:
        fail("stack object does not preserve its frame identity across stack coloring")
    source_initialize = stack_body.find("store ptr null")
    first_call = stack_body.find("@llvm.experimental.gc.statepoint")
    lifetime_start = stack_body.find("call void @llvm.lifetime.start")
    if min(first_call, lifetime_start, source_initialize) < 0:
        fail("stack object initialization sequence is incomplete")
    if stack_body.count("store ptr null") != 1:
        fail("ordinary stack object was unnecessarily initialized at entry")
    if not first_call < lifetime_start < source_initialize:
        fail("stack object source initialization is misordered")

    loop_body = function_body(ir, "loop_reinitialized_pointer_alloca")
    if loop_body.count("@llvm.experimental.gc.statepoint") != 2:
        fail("loop lifetime function does not contain two statepoints")
    if loop_body.count('"deopt"(') != 1 or loop_body.count(
        f"i64 {ALLOCA_TAG}"
    ) != 1:
        fail("loop lifetime is not active only inside the VarDef iteration")
    if loop_body.count('"gc-live"(ptr %slot)') != 1:
        fail("loop alloca is not one explicit active root")
    if loop_body.count("call void @llvm.lifetime.start") != 1:
        fail("loop lifetime start was not preserved")

    phi_body = function_body(ir, "phi_edge_pointer_alloca")
    if phi_body.count("@llvm.experimental.gc.statepoint") != 1:
        fail("PHI-edge lifetime function does not contain one statepoint")
    if phi_body.count('"deopt"(') != 1 or phi_body.count(
        f"i64 {ALLOCA_TAG}"
    ) != 1:
        fail("PHI-edge stack object layout is missing")
    if (
        "%slot.address = getelementptr inbounds i8, ptr %slot, i64 0"
        not in phi_body
    ):
        fail("PHI incoming alloca was not canonicalized after lifetime.start")
    if (
        "phi ptr [ %slot.address, %initialize ], [ %other, %external ]"
        not in phi_body
    ):
        fail("PHI incoming edge did not use the canonical alloca address")
    if '"gc-live"(ptr %selected)' not in phi_body:
        fail("merged PHI address is not an independent scalar root")
    if '"gc-live"(ptr %slot)' in phi_body:
        fail("PHI incoming alloca incorrectly entered merge-block live-out")
    if "alloca ptr, align 8, !llvm.stackcoloring.no_merge" not in phi_body:
        fail("PHI-reachable stack object did not preserve its frame identity")

    aggregate_body = function_body(ir, "hoisted_aggregate_pointer_alloca")
    if aggregate_body.count("@llvm.experimental.gc.statepoint") != 3:
        fail("hoisted aggregate lifetime function does not contain three statepoints")
    if aggregate_body.count(f"i64 {ALLOCA_TAG}") != 3:
        fail("promoted alloca layout is not available at every statepoint")
    if aggregate_body.count("call void @llvm.lifetime.start") != 1:
        fail("promoted alloca does not have one whole-function lifetime start")
    if "call void @llvm.lifetime.end" in aggregate_body:
        fail("promoted alloca retained a lifetime end")
    lifetime_start = aggregate_body.find("call void @llvm.lifetime.start")
    entry_initialize = aggregate_body.find("call void @llvm.memset.inline")
    first_call = aggregate_body.find("@llvm.experimental.gc.statepoint")
    source_initialize = aggregate_body.find("store ptr null")
    if min(lifetime_start, entry_initialize, first_call, source_initialize) < 0:
        fail("promoted alloca initialization sequence is incomplete")
    if re.search(r"call void @llvm\.memset\.p\d", aggregate_body):
        fail("promoted alloca initialization may not use hosted memset")
    if not lifetime_start < entry_initialize < first_call < source_initialize:
        fail("promoted alloca was not initialized before its widened live range")
    if "%slice.cap.leaf.0 = extractvalue" not in aggregate_body:
        fail("hoisted aggregate stack address was not scalarized")
    if "%slice.cap.leaf.0.relocated" not in aggregate_body:
        fail("hoisted aggregate stack address was not relocated")
    statepoints = [
        line
        for line in aggregate_body.splitlines()
        if "@llvm.experimental.gc.statepoint" in line
    ]
    if '"gc-live"(ptr %slice.cap.leaf.0)' not in statepoints[0]:
        fail("hoisted stack address is not live before its eventual use")
    if '"gc-live"' in statepoints[-1]:
        fail("promoted storage remained GC-live after its final use")
    if args.objview:
        check_goobj(args.llc, args.plugin, args.objview, args.input)


if __name__ == "__main__":
    main()
