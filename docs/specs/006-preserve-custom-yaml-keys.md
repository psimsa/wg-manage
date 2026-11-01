# Spec 006 — Preserve user-added YAML keys on peers (issue #34)

Closes #34

## Problem (issue #34)

Users add their own `key: value` pairs to peer entries in `config.yaml`
(description, location, primary use, owner, ...) to drive *other* tooling
(e.g. generating an HTML overview from the same file). `generate` leaves
them alone, but `add` / `remove` (and any other command that round-trips
the file through `LoadYaml` + `SaveYaml`) silently **delete** them, because
`yaml.Unmarshal` into the `Peer` struct drops unknown fields and
`yaml.Marshal` only writes known ones.

## Proposal

Round-tripping the config must preserve unknown keys at every level:
peers, the top-level configuration, and (from spec 001) `settings`.

### Approach

Use `yaml.v3`'s `yaml.Node` to capture unknown fields, or keep a
side-channel map per struct. Recommended: add an unexported/ignored field
to each model struct that captures the raw node, and re-emit unknown keys
on marshal via a custom `MarshalYAML`/`UnmarshalYAML` on `Peer` and
`Configuration`:

```go
type Peer struct {
    // ... existing fields ...
    Extra map[string]interface{} `yaml:",inline"`
}
```

`yaml.v2` (currently used) supports `,inline` maps for this. If inline-map
round-tripping proves lossy (key order, nested unknown structures), switch
the models to `yaml.v3` (`gopkg.in/yaml.v3` is already an indirect
dependency in go.mod) which handles `yaml.Node` cleanly. Prefer whichever
keeps the diff smallest while passing the golden tests below; do not mix
v2 and v3 on the same struct.

### Semantics

- Unknown keys are preserved **verbatim in value**; their position in the
  file may move (inline maps marshal after known fields) — acceptable.
- Known keys keep their current names/behavior; a user key that collides
  with a known key name is treated as the known key (document this).
- `format` also preserves unknown keys (it uses the same
  load/marshal path).
- Deep-merging on `add`/`remove`: the peer being added gets only known
  fields (+ its own extras if constructed with any); all *other* peers'
  extras are untouched because we no longer rebuild them from scratch —
  we mutate the slice in place, which the current code already does.

### Acceptance criteria

- [ ] A `config.yaml` with custom keys on peers (e.g. `location: Berlin`,
      `owner: alice`) survives `add`, `remove`, `recreate`, `bootstrap`-less
      flows, `format`, and TUI saves byte-value-identical for the custom keys.
- [ ] Golden tests: fixture YAML in, command out, fixture keys present.
- [ ] Unknown keys at top level (next to `peers:`/`presharedKey:`) and in
      `settings` are preserved too.
- [ ] No change to generated `.conf` output.
