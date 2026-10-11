# Fragmented Keys — Conformance (stub)

What every implementation in this composite must agree on to be considered
conforming (see [`ALGORITHM.md`](ALGORITHM.md) for the full algorithm):

- **Hash function** — MD5, applied to the composite string exactly as
  specified (key name, group id, then tags).
- **Separator characters** — `_` between key name and group id; `:` between
  every subsequent field. No other separators.
- **Field order** — key name, group id, then tags in the order supplied, each
  tag serialized as `t<name>` followed by `v<version>`. Order is significant.
- **Encoding** — the MD5 digest rendered as a hex string. Exact casing and
  numeric rendering of versions (int vs float) are **not yet pinned**.
- **Invalidation semantics** — bumping any constituent tag's version must
  yield a different composite key; invalidation must not delete entries.
- **Handler contract** — `get`/`set` plus a bulk multi-get, with tags grouped
  per backend so one round trip resolves a composite key.

> **TODO: cross-language test vectors.** None exist yet — no implementation
> in this repo has been verified byte-for-byte against another. Do not claim
> cross-language conformance until a shared vector set lands here.
