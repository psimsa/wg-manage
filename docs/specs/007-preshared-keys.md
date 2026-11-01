# Spec 007 — Preshared key generation on bootstrap (issue #16)

Closes #16

## Problem (issue #16)

Users who rely on preshared keys must manually generate one (`wg genpsk`),
learn the YAML key (`presharedKey`), and edit the file for every new
network. Requested: a `-presharedKey`-style option on `bootstrap`, either
a single PSK for the whole network or a unique PSK per peer. The issue also
asks for easier defaults for `persistentKeepalive` and non-killswitch
`allowedIps` — those are covered by `settings.clientDefaults` /
`serverDefaults` in spec 001.

## Proposal

Two parts: per-peer PSK support in the model, and bootstrap flags.

### 1. Per-peer preshared keys (model change)

WireGuard's `PresharedKey` is actually a **per-peer** (`[Peer]` section)
option; the current model models it only globally on `Configuration` and
`GetPeer` writes the global one into every `[Peer]` block. Extend the
model:

```go
type PeerSection struct {
    // ... existing ...
    PresharedKey *string `yaml:"presharedKey,omitempty"` // per-peer PSK
}
```

Rendering precedence in `GetPeer`: peer's own `presharedKey` if set, else
the global `Configuration.PresharedKey` (backward compatible — existing
configs with only the global key render exactly as before).

### 2. CLI surface

- `bootstrap -preshared` (bool, default false): generate **one** PSK and
  set it as the global `presharedKey` (current behavior of `init
  -preshared`).
- `bootstrap -preshared=per-peer` (string enum: `none|single|per-peer`,
  default `none`): generate a **unique PSK for every peer** and store it
  in each peer's `peerSection.presharedKey`; no global key.
- New command `wg-manage psk` (short `psk`):
  - `psk -single` — set/replace the global preshared key in the file.
  - `psk -per-peer` — generate a unique PSK per peer (replacing global).
  - `psk -remove` — clear all PSKs.
  - `psk` (no args) — print which peers have PSKs (masked) without
    revealing them in logs beyond what's already in the file.
- Same flags on `init` for consistency (`init -preshared` already exists;
  extend it to the same enum).

### 3. TUI

Peer edit form (spec 003) gets a "preshared key" field with actions:
generate-new / clear / (leave as inherited global).

### Acceptance criteria

- [ ] `bootstrap -preshared=single` writes a global `presharedKey`;
      generated configs contain `PresharedKey=` in every `[Peer]` block
      (unchanged from today's global behavior).
- [ ] `bootstrap -preshared=per-peer` writes a distinct `presharedKey`
      under each peer's `peerSection`; generated configs carry each peer's
      own PSK in its `[Peer]` blocks.
- [ ] `psk -single|per-peer|remove` mutate the file as described.
- [ ] Old configs with only a global key render identically to before.
- [ ] Unit tests: precedence (peer > global), generation uniqueness
      (all per-peer PSKs distinct), and that keys are valid base64 32-byte
      wg keys (via `wgtypes.ParseKey`).
