#!/usr/bin/env python3
"""Regenerate testdata/golden_keys.json from the reference Python implementation.

These vectors are the contract between ports: given the same tag names,
instances, versions, prefix and ordering, every conforming implementation must
produce the same raw string and the same MD5. The Go port's
TestGoldenKeysMatchPython replays this file.

The keys here are produced by *running fragmented-keys-py*, not by
re-implementing its formula, so the file records what the reference actually
does rather than what its documentation says it does.

Usage, from the fragmented-keys-py checkout:

    cd ../fragmented-keys-py
    .venv/bin/python ../fragmented-keys-go/testdata/gen_golden_keys.py \\
        > ../fragmented-keys-go/testdata/golden_keys.json

Versions are pinned explicitly on every tag so the output is deterministic;
nothing here depends on the clock.
"""

import json
import sys

sys.path.insert(0, "src")

from fragmented_keys import (  # noqa: E402
    Configuration,
    ConstantTag,
    MemoryHandler,
    StandardKey,
    StandardTag,
)

# "GNP" is what gnp-backend passes to configure_fragmented_keys(), so these
# vectors exercise the prefix that production actually uses.
GLOBAL_PREFIX = "GNP"


def build_cases(handler):
    """Yield (name, StandardKey, [tag descriptors]) triples.

    Tag descriptors travel with each case so the Go side can rebuild the same
    tags without re-deriving them from the raw string.
    """
    cases = []

    def std(tag, instance, version):
        t = StandardTag(tag, instance, handler=handler)
        t.set_tag_version(version)
        return t, {
            "kind": "standard",
            "tag": tag,
            "instance": instance,
            "version": version,
        }

    def const(tag, instance, version):
        t = ConstantTag(tag, instance, version=version, handler=handler)
        return t, {
            "kind": "constant",
            "tag": tag,
            "instance": instance,
            "version": version,
        }

    # A single standard tag at the constant-default version.
    t1, d1 = std("User", "42", 1.0)
    cases.append(("single_standard", StandardKey("profile", [t1]), [d1]))

    # Two tags, timestamp seed plus a small integer version.
    t1, d1 = std("User", "42", 1748000000000.0)
    t2, d2 = std("City", "chicago", 5.0)
    cases.append(("two_standard", StandardKey("Dashboard", [t1, t2]), [d1, d2]))

    # Order matters: the same two tags reversed must hash differently.
    t1, d1 = std("User", "42", 1748000000000.0)
    t2, d2 = std("City", "chicago", 5.0)
    cases.append(("two_standard_reversed", StandardKey("Dashboard", [t2, t1]), [d2, d1]))

    # Constant tag mixed with a standard one; the constant never hits cache.
    c1, cd1 = const("schema", "v1", 1.0)
    t2, d2 = std("Org", "7", 1748000000000.1)
    cases.append(("constant_plus_standard", StandardKey("org_list", [c1, t2]), [cd1, d2]))

    # group_id namespacing.
    c1, cd1 = const("schema", "v1", 1.0)
    t2, d2 = std("Org", "7", 1748000000000.1)
    k = StandardKey("org_list", [c1, t2], group_id="us")
    cases.append(("group_id", k, [cd1, d2]))

    # Versions carrying binary-float drift from repeated +0.1 increments.
    t1, d1 = std("User", "42", 1748000000000.2002)
    t2, d2 = std("Post", "9", 1748000000000.901)
    cases.append(("drifted_versions", StandardKey("feed", [t1, t2]), [d1, d2]))

    # Empty instance, which is legal and yields "Global_GNP".
    t1, d1 = std("Global", "", 3.5)
    cases.append(("empty_instance", StandardKey("settings", [t1]), [d1]))

    # No tags at all: the raw string is just "{key}_{group_id}".
    cases.append(("no_tags", StandardKey("bare", []), []))

    # Unicode in the instance, to pin UTF-8 handling before the MD5.
    t1, d1 = std("City", "münchen", 2.0)
    cases.append(("unicode_instance", StandardKey("city_page", [t1]), [d1]))

    # A version large enough that CPython switches to scientific notation.
    t1, d1 = std("Epoch", "x", 1e16)
    cases.append(("scientific_version", StandardKey("epoch", [t1]), [d1]))

    # Global cache epoch: the constant dimension prepended to every key by
    # fragmented_keys_django's with_epoch() and fragkeys.WithEpoch(). Two
    # epochs over the same underlying tag must not collide.
    for epoch in (1.0, 2.0):
        e, ed = const("epoch", "global", epoch)
        t, d = std("Page", "careers", 1748000000000.0)
        cases.append(
            (f"epoch_{epoch:g}", StandardKey("content_page", [e, t]), [ed, d])
        )

    # An epoch-bearing key must differ from the same key without one — this is
    # what keeps pre-epoch entries unreachable after the cutover.
    t, d = std("Page", "careers", 1748000000000.0)
    cases.append(("no_epoch_baseline", StandardKey("content_page", [t]), [d]))

    # Many tags, to exercise ordering across a longer list.
    tags, descs = [], []
    for i, (name, inst, ver) in enumerate(
        [
            ("A", "1", 1.0),
            ("B", "2", 2.0),
            ("C", "3", 1748000000000.3003),
            ("D", "4", 0.1),
            ("E", "5", 99.0),
        ]
    ):
        t, d = std(name, inst, ver)
        tags.append(t)
        descs.append(d)
    cases.append(("five_tags", StandardKey("wide", tags), descs))

    return cases


def main():
    handler = MemoryHandler()
    Configuration.reset()
    Configuration.set_default_cache_handler(handler)
    Configuration.set_global_prefix(GLOBAL_PREFIX)

    out = {
        "global_prefix": GLOBAL_PREFIX,
        "source": "fragmented-keys-py",
        "cases": [],
    }
    for name, key, descs in build_cases(handler):
        out["cases"].append(
            {
                "name": name,
                "key": key._key,
                "group_id": key._group_id,
                "tags": descs,
                "raw": key.get_key_str(hash=False),
                "md5": key.get_key_str(hash=True),
            }
        )

    json.dump(out, sys.stdout, indent=2, ensure_ascii=False)
    sys.stdout.write("\n")


if __name__ == "__main__":
    main()
