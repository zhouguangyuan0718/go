#!/usr/bin/env python3

import argparse
import json
import re
import subprocess
import sys
import tempfile


BEGIN = 1195461697
ALLOCA_TAG = 1095520067
END = 1095519299
WORD_BITS = 64


def fail(message):
    print(f"alloca pointer-map check failed: {message}", file=sys.stderr)
    raise SystemExit(1)


def run(command, *, input_text=None):
    result = subprocess.run(
        command, input=input_text, capture_output=True, text=True
    )
    if result.returncode != 0:
        fail(
            f"command failed ({' '.join(command)}):\n"
            f"{result.stdout}{result.stderr}"
        )
    return result.stdout + result.stderr


def function_body(ir, name):
    match = re.search(
        rf"define goabiinternal [^@]+@{name}\b.*?^}}",
        ir,
        re.MULTILINE | re.DOTALL,
    )
    if not match:
        fail(f"missing function {name}")
    return match.group(0)


def deopt_bundles(function):
    return re.findall(r'"deopt"\((.*?)\)', function)


def parse_i64(token, description):
    match = re.fullmatch(r"i64 (-?[0-9]+)", token)
    if not match:
        fail(f"{description} is not an i64 constant: {token}")
    return int(match.group(1))


def parse_protocol(bundle):
    tokens = [token.strip() for token in bundle.split(",")]
    if len(tokens) < 5:
        fail(f"truncated protocol: {bundle}")
    protocol_length = parse_i64(tokens[-1], "trailing protocol length")
    protocol_start = len(tokens) - protocol_length - 1
    if protocol_start < 0:
        fail(f"protocol length exceeds deopt bundle: {bundle}")
    if parse_i64(tokens[protocol_start], "begin magic") != BEGIN:
        fail(f"wrong begin magic: {bundle}")
    if parse_i64(tokens[protocol_start + 1], "protocol length") != protocol_length:
        fail(f"protocol length copies disagree: {bundle}")
    if parse_i64(tokens[-2], "end magic") != END:
        fail(f"wrong end magic: {bundle}")

    record_count = parse_i64(tokens[protocol_start + 2], "record count")
    cursor = protocol_start + 3
    records = []
    for record_index in range(record_count):
        if cursor + 10 > len(tokens) - 1:
            fail(f"truncated record {record_index}: {bundle}")
        tag = parse_i64(tokens[cursor], "record tag")
        if tag != ALLOCA_TAG:
            fail(f"wrong record tag {record_index}: {bundle}")
        record_length = parse_i64(tokens[cursor + 1], "record length")
        base = tokens[cursor + 2]
        if not re.fullmatch(r"ptr %[A-Za-z0-9_.]+", base):
            fail(f"record {record_index} has non-alloca base: {base}")
        byte_offset = parse_i64(tokens[cursor + 3], "byte offset")
        byte_size = parse_i64(tokens[cursor + 4], "byte size")
        alignment = parse_i64(tokens[cursor + 5], "alignment")
        pointer_size = parse_i64(tokens[cursor + 6], "pointer size")
        bit_count = parse_i64(tokens[cursor + 7], "bit count")
        word_bits = parse_i64(tokens[cursor + 8], "word width")
        word_count = parse_i64(tokens[cursor + 9], "word count")
        if record_length != 10 + word_count:
            fail(f"record {record_index} length is inconsistent: {bundle}")
        if cursor + record_length > len(tokens) - 2:
            fail(f"record {record_index} overruns protocol: {bundle}")
        words = [
            parse_i64(tokens[cursor + 10 + word], "bitmap word")
            & ((1 << WORD_BITS) - 1)
            for word in range(word_count)
        ]
        records.append(
            (
                base.removeprefix("ptr %"),
                byte_offset,
                byte_size,
                alignment,
                pointer_size,
                bit_count,
                word_bits,
                tuple(words),
            )
        )
        cursor += record_length
    if cursor != len(tokens) - 2:
        fail(f"records do not cover protocol payload: {bundle}")
    return tokens[:protocol_start], protocol_length, records


def expect_records(ir, function_name, expected, expected_prefixes=None):
    function = function_body(ir, function_name)
    bundles = deopt_bundles(function)
    if len(bundles) != len(expected):
        fail(
            f"{function_name} has {len(bundles)} deopt bundles, "
            f"want {len(expected)}"
        )
    if expected_prefixes is None:
        expected_prefixes = [[] for _ in expected]
    for bundle, want_records, want_prefix in zip(
        bundles, expected, expected_prefixes
    ):
        got_prefix, _, got_records = parse_protocol(bundle)
        if got_prefix != want_prefix:
            fail(
                f"{function_name} ordinary deopt prefix={got_prefix}, "
                f"want {want_prefix}"
            )
        if got_records != want_records:
            fail(
                f"{function_name} records={got_records}, "
                f"want {want_records}"
            )
    return function


def record(base, size, bits, words):
    return (base, 0, size, 8, 8, bits, WORD_BITS, tuple(words))


def check_rewritten_ir(ir):
    statepoints = re.findall(r"@llvm\.experimental\.gc\.statepoint", ir)
    # Thirty-four calls plus the intrinsic declaration.
    if len(statepoints) != 35:
        fail(f"found {len(statepoints) - 1} statepoints, want 34")
    if len(re.findall(r'"deopt"\(', ir)) != 31:
        fail("pointer allocas do not have thirty-one deopt records")
    if "llvm.statepoint.fixed_stack_home" in ir:
        fail("obsolete fixed-home metadata remains in rewritten IR")

    if re.search(r"\.gc\.leaf\..*\.init\.addr", ir):
        fail("ordinary StackObject received plugin entry initialization")

    one = [record("slot", 8, 1, [1])]
    pointer_slot = expect_records(ir, "pointer_slot", [one], [["i64 7"]])
    nested = expect_records(
        ir,
        "nested_whole_aggregate",
        [[record("slot", 48, 6, [0x29])]],
    )
    if "llvm.memset.inline" in pointer_slot:
        fail("complete scalar pointer store received a duplicate GC zero")
    if "llvm.memset.inline" in nested:
        fail("complete aggregate pointer store received a duplicate GC zero")
    expect_records(ir, "alloca_call_skip", [one])
    multiple_calls = expect_records(ir, "alloca_multiple_calls", [one, one])
    if "llvm.memset.inline" in multiple_calls:
        fail("pointer store before multiple calls received a duplicate GC zero")
    aggregate_home = function_body(
        ir, "argument_aggregate_home_address_across_calls"
    )
    if "llvm.memset.inline" in aggregate_home:
        fail("aggregate argument store received a duplicate GC zero")
    partial = expect_records(
        ir,
        "alloca_partial_initialization",
        [[record("slot", 16, 2, [0x3])]] * 2,
    )
    lifetime_start = partial.find("call void @llvm.lifetime.start")
    entry_zero = partial.find("call void @llvm.memset.inline")
    first_call = partial.find("@llvm.experimental.gc.statepoint")
    if (
        min(lifetime_start, entry_zero, first_call) < 0
        or not lifetime_start < entry_zero < first_call
    ):
        fail("partially initialized alloca was not zeroed after lifetime start")
    expect_records(ir, "alloca_loop", [one])
    direct_gep = expect_records(
        ir,
        "alloca_gep_address_across_call",
        [[record("slot", 16, 2, [0x2])]],
    )
    if (
        "%result = load ptr, ptr %field" not in direct_gep
        or "%field.remat" in direct_gep
        or "%field.relocated.merge" in direct_gep
    ):
        fail("direct derived alloca address entered relocation SSA")
    direct_address = expect_records(
        ir, "alloca_direct_address_across_calls", [one] * 3
    )
    gep_address = expect_records(
        ir,
        "alloca_gep_value_across_calls",
        [[record("slot", 16, 2, [0x2])]] * 3,
    )
    pointer_free_address = function_body(
        ir, "alloca_pointer_free_address_across_calls"
    )
    if (
        '"deopt"(' in pointer_free_address
        or '"gc-live"(' in pointer_free_address
        or len(
            re.findall(
                r"%slot\.address[0-9]* = getelementptr inbounds i8, ptr %slot, i64 0",
                pointer_free_address,
            )
        )
        != 2
    ):
        fail("pointer-free alloca address changed the content-ptrmap protocol")
    expect_records(ir, "alloca_marker_free_at_safepoint", [one])
    expect_records(
        ir,
        "alloca_high_bitmap_word",
        [[record("slot", 512, 64, [1 << 63])]],
    )
    expect_records(
        ir,
        "alloca_multiple_records",
        [[record("left", 8, 1, [1])]],
    )
    expect_records(ir, "alloca_select_same_base", [one])

    escaped = expect_records(
        ir, "alloca_address_passed_to_callee", [one]
    )
    expect_records(ir, "alloca_nocapture_writable", [one])
    expect_records(
        ir, "alloca_escaped_before_unknown_write", [one, one]
    )
    expect_records(
        ir, "alloca_readonly_and_readnone", [one, one]
    )

    selected = function_body(ir, "alloca_select_same_base")
    if '"gc-live"(ptr %selected' not in selected or "%selected.relocated" not in selected:
        fail("alloca-address select is not represented as an ordinary live pointer")

    if (
        "%slot.address = getelementptr inbounds i8, ptr %slot, i64 0"
        not in direct_address
        or '"gc-live"(ptr %slot)' not in direct_address
        or '"gc-live"(ptr %slot.address' in direct_address
        or len(
            re.findall(
                r"%slot\.address[0-9]* = getelementptr inbounds i8, ptr %slot, i64 0",
                direct_address,
            )
        )
        != 2
    ):
        fail("direct alloca address is not rebuilt at each first-class use")
    if (
        '"gc-live"(ptr %slot)' not in gep_address
        or '"gc-live"(ptr %field' in gep_address
        or len(
            re.findall(
                r"%field\.remat[0-9]* = getelementptr inbounds %pointer_field, ptr %slot",
                gep_address,
            )
        )
        != 2
    ):
        fail("derived alloca address is not rebuilt at each first-class use")
    for name, function in (
        ("direct", direct_address),
        ("derived", gep_address),
        ("pointer-free", pointer_free_address),
    ):
        if ".address.relocated.merge" in function or ".remat.relocated.merge" in function:
            fail(f"{name} alloca address is merged across safepoints")

    if '"gc-live"(ptr %slot)' not in escaped or "%slot.relocated" not in escaped:
        fail("callee-writable alloca is not an explicit rematerialized root")
    statepoint_end = escaped.index("@llvm.experimental.gc.statepoint")
    if re.search(r"store ptr .*ptr %slot", escaped[statepoint_end:]):
        fail("callee-writable alloca is stored after the call")
    source_store = escaped.find("store ptr %pointer, ptr %slot")
    if source_store < 0 or "store ptr null, ptr %slot" in escaped:
        fail("callee-writable StackObject source initialization changed")

    locals_only = function_body(ir, "alloca_multiple_calls")
    if "store ptr null" in locals_only:
        fail("direct-only alloca received plugin initialization")
    relocates = re.findall(
        r"= call coldcc ptr @llvm\.experimental\.gc\.relocate", ir
    )
    # Twenty-six active pointer-containing alloca roots retain their storage
    # identity and two ordinary scalar roots are relocated separately.
    # Pointer-free and derived stack addresses are rebuilt at their uses, so
    # they need neither an alloca relocate nor their own ptrmap root.
    if len(relocates) != 28:
        fail(f"found {len(relocates)} relocates, want 28")
    marker_free = function_body(ir, "alloca_marker_free_at_safepoint")
    if '"gc-live"(ptr %pointer' not in marker_free:
        fail("ordinary scalar SSA pointer is missing from gc-live")
    if "%pointer.relocated" not in marker_free:
        fail("ordinary scalar SSA pointer was not relocated")


def check_objview(objview, object_path):
    document = json.loads(run([objview, "-format=json", object_path]))
    symbols = document["members"][0]["go_object"]["symbols"]

    def function(name):
        matches = [symbol for symbol in symbols if symbol.get("name") == name]
        if len(matches) != 1 or not matches[0].get("function"):
            fail(f"objview did not find one function symbol {name}")
        return matches[0]["function"]

    def funcdata_kinds(name):
        return [entry["kind"] for entry in function(name)["funcdata"]]

    argument_home = function("alloca_multiple_calls")
    argument_home_data = {
        entry["kind"]: entry for entry in argument_home["funcdata"]
    }
    if "stack_objects" in argument_home_data:
        fail("all-callsite-live argument home unexpectedly emitted StackObjects")
    argument_maps = argument_home_data["args_pointer_maps"]["stack_map"]
    if not any(bitmap.get("set_bits") for bitmap in argument_maps["bitmaps"]):
        fail("active argument home did not expand ArgsPointerMaps")
    argument_locals = argument_home_data["locals_pointer_maps"]["stack_map"]
    if any(bitmap.get("set_bits") for bitmap in argument_locals["bitmaps"]):
        fail("active argument home polluted LocalsPointerMaps")

    active_only_kinds = funcdata_kinds("alloca_address_passed_to_callee")
    if "stack_objects" in active_only_kinds:
        fail("all-callsite-live alloca unexpectedly emitted StackObjects")
    active_only = function("alloca_address_passed_to_callee")
    active_only_data = {
        entry["kind"]: entry for entry in active_only["funcdata"]
    }
    active_only_maps = active_only_data["args_pointer_maps"]["stack_map"]
    if not any(bitmap.get("set_bits") for bitmap in active_only_maps["bitmaps"]):
        fail("matching argument-home gc-live did not expand ArgsPointerMaps")

    argument_stack_object = function("argument_home_address_across_calls")
    argument_stack_object_data = {
        entry["kind"]: entry for entry in argument_stack_object["funcdata"]
    }
    argument_objects = argument_stack_object_data.get("stack_objects", {}).get(
        "stack_objects", []
    )
    if len(argument_objects) != 1:
        fail(f"argument home StackObjects={argument_objects}, want one object")
    argument_object = argument_objects[0]
    if (
        argument_object["offset"] != 0
        or argument_object["size"] != 8
        or argument_object["ptr_bytes"] != 8
    ):
        fail(f"argument home has malformed argp-relative object: {argument_object}")
    argument_object_maps = argument_stack_object_data["args_pointer_maps"][
        "stack_map"
    ]
    if not any(
        bitmap.get("set_bits") for bitmap in argument_object_maps["bitmaps"]
    ):
        fail("active argument StackObject layout did not expand ArgsPointerMaps")
    argument_object_locals = argument_stack_object_data["locals_pointer_maps"][
        "stack_map"
    ]
    if any(
        bitmap.get("set_bits") for bitmap in argument_object_locals["bitmaps"]
    ):
        fail("argument StackObject polluted LocalsPointerMaps")

    aggregate_stack_object = function(
        "argument_aggregate_home_address_across_calls"
    )
    aggregate_stack_object_data = {
        entry["kind"]: entry for entry in aggregate_stack_object["funcdata"]
    }
    aggregate_objects = aggregate_stack_object_data.get(
        "stack_objects", {}
    ).get("stack_objects", [])
    if len(aggregate_objects) != 1:
        fail(f"aggregate argument StackObjects={aggregate_objects}, want one object")
    aggregate_object = aggregate_objects[0]
    if (
        aggregate_object["offset"] != 0
        or aggregate_object["size"] != 48
        or aggregate_object["ptr_bytes"] != 48
        or aggregate_object["gcdata"]["name"]
        != "runtime.gcbits.2900000000000000"
    ):
        fail(
            "aggregate argument has malformed argp-relative object: "
            f"{aggregate_object}"
        )
    aggregate_maps = aggregate_stack_object_data["args_pointer_maps"][
        "stack_map"
    ]["bitmaps"]
    if not any(bitmap.get("set_bits") == [0, 3, 5] for bitmap in aggregate_maps):
        fail("aggregate argument home did not preserve split pointer pieces")

    stack_object_kinds = funcdata_kinds("alloca_direct_address_across_calls")
    if "stack_objects" not in stack_object_kinds:
        fail("unmatched alloca layout did not emit StackObjects")
    stack_object = function("alloca_direct_address_across_calls")
    stack_object_data = {
        entry["kind"]: entry for entry in stack_object["funcdata"]
    }
    stack_object_maps = stack_object_data["locals_pointer_maps"]["stack_map"]
    if not any(bitmap.get("set_bits") for bitmap in stack_object_maps["bitmaps"]):
        fail("active StackObject layout did not expand LocalsPointerMaps")

    pointer_free = function("alloca_pointer_free_address_across_calls")
    pointer_free_data = {
        entry["kind"]: entry for entry in pointer_free["funcdata"]
    }
    if "stack_objects" in pointer_free_data:
        fail("pointer-free address root unexpectedly emitted StackObjects")
    pointer_free_maps = pointer_free_data["locals_pointer_maps"]["stack_map"]
    if any(bitmap.get("set_bits") for bitmap in pointer_free_maps["bitmaps"]):
        fail("pointer-free address root polluted LocalsPointerMaps")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--llc", required=True)
    parser.add_argument("--opt", required=True)
    parser.add_argument("--plugin", required=True)
    parser.add_argument("--input", required=True)
    parser.add_argument("--objview")
    parser.add_argument("--mtriple")
    args = parser.parse_args()
    llc = [args.llc]
    if args.mtriple:
        llc.append(f"-mtriple={args.mtriple}")
        if args.mtriple.startswith("aarch64"):
            llc.append("-frame-pointer=all")

    rewritten = run(
        [
            *llc,
            f"-load-pass-plugin={args.plugin}",
            "-goallc-pass-plugin-emit-ir",
            "-filetype=null",
            "-o",
            "-",
            args.input,
        ]
    )
    check_rewritten_ir(rewritten)
    run(
        [
            args.opt,
            f"-load-pass-plugin={args.plugin}",
            "-passes=verify",
            "-disable-output",
            "-",
        ],
        input_text=rewritten,
    )

    with tempfile.TemporaryDirectory() as directory:
        optimized = run(
            [
                args.opt,
                "-passes=default<O2>,verify",
                "-S",
                "-o",
                "-",
                args.input,
            ]
        )
        optimized_rewritten = run(
            [
                *llc,
                f"-load-pass-plugin={args.plugin}",
                "-goallc-pass-plugin-emit-ir",
                "-filetype=null",
                "-o",
                "-",
                "-",
            ],
            input_text=optimized,
        )
        if len(re.findall(r'"deopt"\(', optimized_rewritten)) != 19:
            fail("default<O2> did not preserve stack-object records")
        if "alloca %nested" in function_body(
            optimized_rewritten, "nested_whole_aggregate"
        ):
            fail("fixed-home classification prevented SROA")

        machine_ir = run(
            [
                *llc,
                f"-load-pass-plugin={args.plugin}",
                "-verify-machineinstrs",
                "-stop-after=finalize-isel",
                "-o",
                "-",
                args.input,
            ]
        )
        statepoint_lines = re.findall(r"(?m)^.*STATEPOINT.*$", machine_ir)
        if len(statepoint_lines) != 34:
            fail(f"MIR has {len(statepoint_lines)} statepoints, want 34")
        alloca_statepoints = [
            line for line in statepoint_lines if str(BEGIN) in line
        ]
        if len(alloca_statepoints) != 31:
            fail("MIR does not contain thirty-one alloca statepoints")
        for statepoint in alloca_statepoints:
            if str(ALLOCA_TAG) not in statepoint or not re.search(
                r"%(?:fixed-)?stack\.[0-9]+", statepoint
            ):
                fail(f"MIR alloca record is malformed: {statepoint}")

        output = f"{directory}/alloca-pointer-roots.goobj"
        run(
            [
                *llc,
                f"-load-pass-plugin={args.plugin}",
                "-verify-machineinstrs",
                "-filetype=obj",
                "-o",
                output,
                args.input,
            ]
        )
        if args.objview:
            check_objview(args.objview, output)
        optimized_output = f"{directory}/alloca-pointer-roots-o2.goobj"
        run(
            [
                *llc,
                f"-load-pass-plugin={args.plugin}",
                "-verify-machineinstrs",
                "-filetype=obj",
                "-o",
                optimized_output,
                "-",
            ],
            input_text=optimized,
        )


if __name__ == "__main__":
    main()
