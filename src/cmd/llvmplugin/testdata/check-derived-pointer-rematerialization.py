#!/usr/bin/env python3

import argparse
import re
import subprocess
import sys


def fail(message):
    print(
        f"derived-pointer rematerialization check failed: {message}",
        file=sys.stderr,
    )
    raise SystemExit(1)


def function_body(ir, name):
    body = re.search(
        rf"define goabiinternal .*? @{name}\(.*?^}}",
        ir,
        re.MULTILINE | re.DOTALL,
    )
    if not body:
        fail(f"missing rewritten function {name}")
    return body.group(0)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--llc", required=True)
    parser.add_argument("--plugin", required=True)
    parser.add_argument("--input", required=True)
    args = parser.parse_args()

    rewrite = subprocess.run(
        [
            args.llc,
            f"-load-pass-plugin={args.plugin}",
            "-goallc-pass-plugin-emit-ir",
            "-verify-machineinstrs",
            "-filetype=null",
            "-o",
            "-",
            args.input,
        ],
        capture_output=True,
        text=True,
    )
    if rewrite.returncode != 0:
        fail(f"llc failed:\n{rewrite.stdout}{rewrite.stderr}")
    ir = rewrite.stdout + rewrite.stderr

    hoisted = function_body(ir, "hoisted_null_offset")
    if not re.search(r'"gc-live"\(ptr %base\)', hoisted):
        fail("hoisted null-offset statepoint does not keep exactly the base live")
    if '"gc-live"(ptr %derived)' in hoisted:
        fail("hoisted null-offset statepoint still keeps the derived pointer live")
    if not re.search(r"%base\.relocated.*?= call.*?gc\.relocate", hoisted):
        fail("hoisted null-offset function does not relocate the base")
    if not re.search(
        r"%derived\.remat.*?= getelementptr i8, ptr %base\.relocated.*?, i64 96",
        hoisted,
    ):
        fail("hoisted null-offset function does not rebuild the derived pointer")

    chain = function_body(ir, "derived_chain")
    if not re.search(r'"gc-live"\(ptr %base\)', chain):
        fail("derived chain statepoint does not keep exactly the base live")
    if '"gc-live"(ptr %field)' in chain or '"gc-live"(ptr %element)' in chain:
        fail("derived chain records an interior pointer as a GC root")
    if not re.search(
        r"%field\.remat.*?= getelementptr i8, ptr %base\.relocated.*?, i64 16",
        chain,
    ):
        fail("derived chain does not rebuild its first GEP")
    if not re.search(
        r"%element\.remat.*?= getelementptr i8, ptr %field\.remat.*?, i64 8",
        chain,
    ):
        fail("derived chain does not rebuild its second GEP")

    conditional = function_body(ir, "conditional_derived")
    if not re.search(r'"gc-live"\(ptr %base\)', conditional):
        fail("conditional derived statepoint does not keep exactly the base live")
    if '"gc-live"(ptr %derived)' in conditional:
        fail("conditional derived statepoint records the interior pointer")
    if not re.search(
        r"%derived\.relocated\.merge.*?= phi ptr "
        r"\[ %derived\.remat, %call \], \[ %derived, %skip \]",
        conditional,
    ):
        fail("conditional derived paths do not merge rebuilt and original values")

    vector = function_body(ir, "derived_vector")
    if not re.search(r'"gc-live"\(<2 x ptr> %base\)', vector):
        fail("derived vector statepoint does not keep exactly the vector base live")
    if '"gc-live"(<2 x ptr> %derived)' in vector:
        fail("derived vector still records its interior-pointer lanes")
    if not re.search(r"%base\.relocated.*?= call.*?gc\.relocate\.v2p", vector):
        fail("derived vector does not relocate the vector base")
    if not re.search(
        r"%derived\.remat.*?= getelementptr i8, <2 x ptr> "
        r"%base\.relocated.*?<2 x i64> <i64 16, i64 32>",
        vector,
    ):
        fail("derived vector is not rebuilt from the relocated vector base")

    vector_from_scalar = function_body(ir, "derived_vector_from_scalar")
    if not re.search(r'"gc-live"\(ptr %base\)', vector_from_scalar):
        fail("vector-from-scalar statepoint does not keep the scalar base live")
    if '"gc-live"(<2 x ptr> %derived)' in vector_from_scalar:
        fail("vector-from-scalar still records its interior-pointer lanes")
    if not re.search(
        r"%derived\.remat.*?= getelementptr i8, ptr %base\.relocated.*?"
        r"<2 x i64> <i64 16, i64 32>",
        vector_from_scalar,
    ):
        fail("vector-from-scalar is not rebuilt from the relocated scalar base")


if __name__ == "__main__":
    main()
