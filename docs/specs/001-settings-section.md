# Spec 001 — `settings` section in the YAML config

**Status: implemented** (branch `vibe/revive-wg-manage-f50cd7`). The section
below describes the design; see `models/settings.go` for the code.

## Problem

Project-level configuration (IP pool, default interface for NAT, output
directory, whether QR codes are generated, ...) currently has no home. Users
must hardcode such values in commands or edit peers one by one.

## Proposal

Add an optional top-level `settings` section to the YAML configuration. It is
purely additive: existing config files without `settings` keep working, and
all settings have sensible defaults.

### Schema

```yaml
settings:
  ipPool: 10.0.2.0/24            # base CIDR from which addresses are allocated
  subnets:                       # optional named subpools; see spec 002
    servers:
      cidr: 10.0.2.0/26          # first quarter of the pool
    clients:
      cidr: 10.0.2.64/26         # second quarter
  serverDefaults:                # applied to peers that have an endpoint
    listenPort: 51820
    egressInterface: eth0        # used by the routing/NAT helper (TUI spec 003)
    postUp: [ ... ]              # optional extra wg-quick lines
    postDown: [ ... ]
  clientDefaults:                # applied to peers without an endpoint
    persistentKeepalive: 25
  generate:                      # defaults for the `generate` command
    outputDir: ./output
    png: true

presharedKey: ...                # unchanged
peers:                           # unchanged
  - name: Server
    ...
```

### Model changes (`models/models.go`)

```go
type Configuration struct {
    Settings     *Settings `yaml:"settings,omitempty"`
    PresharedKey *string   `yaml:"presharedKey,omitempty"`
    Peers        []Peer    `yaml:"peers"`
}

type Settings struct {
    IPPool         string            `yaml:"ipPool,omitempty"`
    Subnets        map[string]Subnet `yaml:"subnets,omitempty"`
    ServerDefaults *PeerDefaults     `yaml:"serverDefaults,omitempty"`
    ClientDefaults *PeerDefaults     `yaml:"clientDefaults,omitempty"`
    Generate       *GenerateSettings `yaml:"generate,omitempty"`
}

type Subnet struct {
    CIDR string `yaml:"cidr"`
}

type PeerDefaults struct {
    ListenPort          *int     `yaml:"listenPort,omitempty"`
    EgressInterface     string   `yaml:"egressInterface,omitempty"`
    PersistentKeepalive *int     `yaml:"persistentKeepalive,omitempty"`
    PostUp              []string `yaml:"postUp,omitempty"`
    PostDown            []string `yaml:"postDown,omitempty"`
}

type GenerateSettings struct {
    OutputDir string `yaml:"outputDir,omitempty"`
    PNG       *bool  `yaml:"png,omitempty"`
}
```

### Semantics

- A peer is a **server** if it has `endpoint` set, a **client** otherwise.
  This matches how `generate` already treats peers when deciding which
  `[Peer]` blocks to include.
- `ipPool` must be a valid IPv4 or IPv6 CIDR. If `subnets` are defined, each
  subnet CIDR must be contained within `ipPool` (validated on load; see
  spec 002 for allocation).
- Defaults are **applied at creation time** (`add`, `bootstrap`, `init`,
  TUI), not at generation time. Once a value is written to a peer it is the
  peer's own value; changing `settings` later does not retroactively change
  existing peers. This keeps `generate` output stable and predictable.
- `settings.generate` provides flag defaults for the `generate` command:
  explicit CLI flags always win over `settings`, which wins over the built-in
  default (`./output`, `png: true`).

### Backward compatibility

- `LoadYaml` on an old file yields `Settings == nil`; every consumer must
  nil-check and fall back to current behavior (hardcoded defaults).
- `init`/`bootstrap` write a `settings` section by default (opt-out flag
  `-settings=false` on `init`), so new files are self-describing.
- `format` round-trips `settings` like any other section.

### Open questions

- Should `settings` be editable only via TUI/CLI, or is manual editing fine?
  Proposal: manual editing remains allowed (it is just YAML), the TUI merely
  makes it unnecessary.
- IPv6 pool support: yes, use `net/netip` (`netip.Prefix`) instead of hand
  -rolled math.

### Acceptance criteria

- [ ] New config files created by `bootstrap`/`init` contain a `settings`
      section with an `ipPool`.
- [ ] Old config files (no `settings`) load and generate identically to
      before this change.
- [ ] `generate` honors `settings.generate.outputDir` / `png` when flags are
      absent.
- [ ] Invalid `ipPool` or subnet-outside-pool fails load with a clear error.
- [ ] Unit tests cover parsing, defaults resolution order, and validation.
