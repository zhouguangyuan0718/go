#!/usr/bin/env python3

import argparse
import json
import subprocess
import sys


FUNCTION = "aarch64_pointer_and_code_live"
POINTER_SIZE = 8


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

    pcsp = only(
        metadata["pc_tables"],
        lambda item: item["kind"] == "pcsp",
        "PCSP tables",
    )
    frame_sizes = {
        item["value"] for item in pcsp["ranges"] if item["value"] > 0
    }
    if len(frame_sizes) != 1:
        fail(f"unexpected framed PCSP values: {sorted(frame_sizes)}")
    frame_size = frame_sizes.pop()

    locals_size = metadata["info"]["locals"]
    if locals_size != frame_size - POINTER_SIZE:
        fail(
            f"FuncInfo.locals={locals_size}, want frame size {frame_size} "
            f"minus LR"
        )

    locals_maps = only(
        metadata["funcdata"],
        lambda item: item["kind"] == "locals_pointer_maps",
        "FUNCDATA_LocalsPointerMaps tables",
    )["stack_map"]
    want_bits = frame_size // POINTER_SIZE - 2
    if locals_maps["num_bits"] != want_bits:
        fail(
            f"locals bitmap has {locals_maps['num_bits']} bits, want "
            f"{want_bits} after excluding LR and FP"
        )
    if locals_maps["count"] != 2:
        fail(f"locals stack-map count is {locals_maps['count']}, want 2")

    ordinary_bits = set(locals_maps["bitmaps"][0]["set_bits"] or [])
    morestack_bits = set(locals_maps["bitmaps"][1]["set_bits"] or [])
    want_ordinary_bits = {0, 1}
    if ordinary_bits != want_ordinary_bits:
        fail(
            "ordinary statepoint should conservatively contain the live data "
            f"and code pointers at the first two GC locals, got "
            f"{sorted(ordinary_bits)}, want {sorted(want_ordinary_bits)}"
        )
    if any(bit < 0 or bit >= want_bits for bit in ordinary_bits):
        fail(f"ordinary roots escape the GC locals range: {ordinary_bits}")
    if morestack_bits:
        fail(f"morestack statepoint has GC live pointers: {morestack_bits}")

    print(
        f"{FUNCTION}: frame={frame_size} locals={locals_size} "
        f"bitmap-bits={want_bits} roots={sorted(ordinary_bits)}"
    )


if __name__ == "__main__":
    try:
        main()
    except (KeyError, RuntimeError, subprocess.CalledProcessError, json.JSONDecodeError) as error:
        print(f"check-aarch64-frame: {error}", file=sys.stderr)
        sys.exit(1)
