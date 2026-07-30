#!/usr/bin/env python3

import argparse
import json
import subprocess
import sys


FUNCTION = "x86_byval_stack_pointer"


def fail(message):
    raise RuntimeError(message)


def only(items, predicate, description):
    matches = [item for item in items if predicate(item)]
    if len(matches) != 1:
        fail(f"found {len(matches)} {description}, want 1")
    return matches[0]


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--objview", required=True)
    parser.add_argument("--object", required=True)
    args = parser.parse_args()

    result = subprocess.run(
        [args.objview, "-json", args.object],
        check=True,
        stdout=subprocess.PIPE,
        text=True,
    )
    document = json.loads(result.stdout)
    go_objects = [
        member["go_object"]
        for member in document["members"]
        if member.get("go_object") is not None
    ]
    obj = only(go_objects, lambda item: True, "Go objects")
    function = only(
        obj["symbols"],
        lambda item: item["name"] == FUNCTION,
        f"{FUNCTION} symbols",
    )
    metadata = function.get("function")
    if metadata is None:
        fail(f"{FUNCTION} has no function metadata")

    # Nine register homes plus the complete byval stack object occupy ten Go
    # ABI words. The AsmPrinter must use the typed pointee, not opaque ptr, when
    # it computes FuncInfo.args.
    if metadata["info"]["args"] != 80:
        fail(f"FuncInfo.args={metadata['info']['args']}, want 80")

    args_maps = only(
        metadata["funcdata"],
        lambda item: item["kind"] == "args_pointer_maps",
        "FUNCDATA_ArgsPointerMaps tables",
    )["stack_map"]
    if args_maps["count"] != 1 or args_maps["num_bits"] != 10:
        fail(f"unexpected args stack-map dimensions: {args_maps}")
    entry_args = set(args_maps["bitmaps"][0]["set_bits"] or [])
    if entry_args != {0}:
        fail(f"entry args bits={sorted(entry_args)}, want [0]")

    locals_maps = only(
        metadata["funcdata"],
        lambda item: item["kind"] == "locals_pointer_maps",
        "FUNCDATA_LocalsPointerMaps tables",
    )["stack_map"]
    if locals_maps["count"] != 1:
        fail(f"locals stack-map count={locals_maps['count']}, want 1")
    locals_bits = set(locals_maps["bitmaps"][0]["set_bits"] or [])
    if locals_bits:
        fail(f"byval entry pointer leaked into locals: {sorted(locals_bits)}")

    stack_index = only(
        metadata["pcdata"],
        lambda item: item["kind"] == "stack_map_index" and item["index"] == 1,
        "PCDATA_StackMapIndex tables",
    )
    values = [item["value"] for item in stack_index["ranges"]]
    if values != [-1, 0]:
        fail(f"unexpected stack-map range values: {values}")
    queries = metadata["stack_map_queries"]
    if len(queries) != 1 or queries[0]["stack_map_index"] != 0:
        fail(f"unexpected morestack call query: {queries}")

    # X86 R_CALL relocations name the displacement byte, while the PCDATA
    # transition begins at the CALL opcode one byte earlier.
    if stack_index["ranges"][-1]["start"] != queries[0]["call_offset"] - 1:
        fail("morestack stack-map range does not start at the CALL opcode")

    print(
        f"{FUNCTION}: args=80 entry-args={sorted(entry_args)} "
        f"locals={sorted(locals_bits)} PCDATA={values}"
    )


if __name__ == "__main__":
    try:
        main()
    except (
        KeyError,
        RuntimeError,
        subprocess.CalledProcessError,
        json.JSONDecodeError,
    ) as error:
        print(f"check-x86-byval-entry-args: {error}", file=sys.stderr)
        sys.exit(1)
