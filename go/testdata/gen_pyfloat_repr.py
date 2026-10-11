#!/usr/bin/env python3
"""Regenerate testdata/pyfloat_repr.txt — the CPython float-repr corpus.

The Go pyfloat package must reproduce CPython's repr(float) byte-for-byte,
because tag versions are interpolated into the raw key string before it is
MD5-hashed. One differing digit means the two runtimes compute different
cache keys and silently stop sharing entries.

Usage:
    python testdata/gen_pyfloat_repr.py > testdata/pyfloat_repr.txt

The seed is fixed so the corpus is stable across regenerations; regenerate
only when you want to widen coverage, and commit the result.
"""

import random
import struct

SEED = 1337


def values():
    random.seed(SEED)

    # The drift sequence StandardTag actually produces: a millisecond epoch
    # seed followed by repeated +0.1 increments. Binary float error makes these
    # spellings non-obvious (1748000000000.2002, 1748000000000.901, ...), which
    # is exactly why this corpus exists.
    v = 1748000000000.0
    for _ in range(400):
        yield v
        v += 0.1

    # Constant-tag versions and other hand-set values.
    yield from (1.0, 2.0, 3.5, 0.0, -1.0, 1.1, 5.0)

    # Random epoch-shaped seeds with a few increments each.
    for _ in range(400):
        b = random.uniform(1.0e12, 1.8e12)
        yield b
        for _ in range(random.randint(0, 6)):
            b += 0.1
            yield b

    # The notation switch points. CPython renders positionally while the
    # decimal exponent is in [-4, 16) and switches to scientific outside it.
    yield from (
        1e15,
        9.999999999999998e15,
        1e16,
        1.0000000000000002e16,
        1e17,
        1e-4,
        9.99e-5,
        1e-5,
        1.5e-5,
        1e-320,          # subnormal
        5e-324,          # smallest positive subnormal
        1.7976931348623157e308,  # DBL_MAX
    )

    # Arbitrary bit patterns, to catch anything the structured cases miss.
    for _ in range(400):
        yield struct.unpack("<d", struct.pack("<Q", random.getrandbits(64)))[0]


def main():
    out = []
    for x in values():
        # NaN and the infinities are covered by dedicated Go unit tests; they
        # have no meaningful ParseFloat round trip in this corpus format.
        if x != x or x in (float("inf"), float("-inf")):
            continue
        out.append(repr(x))
    print("\n".join(out))


if __name__ == "__main__":
    main()
