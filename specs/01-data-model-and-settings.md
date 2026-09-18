# 01 — Data model & `settings` section

Status: **proposed**
Affects: `models/models.go`, `models/peerFunctions.go`, `models/peerFactories.go`, `cmd/add/add.go`, `cmd/bootstrap/bootstrap.go`, every command that loads/saves YAML.

## 1. Problem

Today every IP address is typed by hand (`-ip` on `add`, hardcoded `10.0.2.x`
in `bootstrap`). There is no notion of "this is a server" vs "this is a
client" beyond the presence of an `Endpoint`, and there is no project-level
place to store cross-cutting settings such as the address pool or the external
interface used by routing rules.

## 2. Goal

Introduce a `settings` section in the YAML and a light-weight `role` on each
peer so that:

- Addresses are assigned automatically from declared pools, role-aware and
  collision-free.
- Project settings (pool, external interface, defaults) live in the file.
- Existing files without `settings` / without `role` keep loading and
  generating unchanged.

## 3. Proposed schema

### 3.1 New top-level `settings`

```yaml
settings:
  # Human-friendly project name; used as TUI title and as output subdir hint.
  name: home-vpn

  # IP assignment pools. Each pool is a CIDR. Reserved/burned addresses
  # (network, broadcast, gateway) are skipped automatically.
  # At least `server` and `client` are recognized roles.
  pools:
    server:
      cidr: 10.0.2.0/24
      # Optional explicit range to assign from inside the cidr.
      # Defaults to the whole cidr minus reserved addresses.
      start: 10.0.2.1
      end: 10.0.2.10
    client:
      cidr: 10.0.2.0/24
      start: 10.0.2.100
      end: 10.0.2.200

  # Default external interface for routing rules. DEFAULT ONLY:
  # it pre-fills the per-peer externalInterface when adding a new server via
  # the TUI or `add`. The value that actually drives generation lives on each
  # peer (peer.externalInterface), because routing is per-peer and depends on
  # the underlying OS of that peer.
  externalInterface: eth0

  # Defaults applied to a newly created peer of a given role (merged into the
  # peer on creation, overridable later in the TUI).
  defaults:
    server:
      listenPort: 51820
      persistentKeepalive: 21
    client:
      persistentKeepalive: 21
```

### 3.2 New `role` and per-peer routing fields on `Peer`

```yaml
peers:
  - name: Server
    role: server          # NEW. Optional; "" means "infer".
    routing: true                 # NEW. Whether to emit PostUp/PostDown rules.
    externalInterface: eth0       # NEW. Per-peer interface for the MASQUERADE rule.
    interfaceSection: { ... }
    peerSection:
      endpoint: 1.2.3.4:51820
      allowedIps: [0.0.0.0/0]
  - name: My Laptop
    role: client
    # routing/externalInterface omitted → no routing rules generated.
    ...
```

`routing` (bool, default false) and `externalInterface` (string, optional)
replace today's `-add-routing <iface>` flag, which currently bakes the
interface into opaque `PostUp`/`PostDown` strings. Storing them as
structured fields means:

- the interface is editable per peer in the TUI (no string surgery);
- the `PostUp`/`PostDown` blocks are *generated* from the fields rather
  than stored verbatim, so changing the interface later updates the rules;
- a server on `eth0` and a server on `wlan0` keep their own interfaces.

Resolution order for the interface used in a peer's MASQUERADE rule:
1. `peer.externalInterface` (per-peer, authoritative);
2. `settings.externalInterface` (default, used only to pre-fill #1 on add);
3. legacy `eth0` (only if neither is set, to preserve today's behavior).

The `wg0` references in the rules are left as-is (they are the wireguard
interface name, a separate concern; see Open Question 4.4).

- `role` is one of `server`, `client`, or empty.
- When empty, role is **inferred**: a peer with a non-empty `Endpoint` is a
  `server`, otherwise a `client`. This preserves current semantics so old
  files keep working with no edits.
- Role drives: which pool an address is drawn from, whether `listenPort` /
  routing are *relevant* (only servers typically enable `routing`), and where
  the peer is rendered in the TUI. The routing rules themselves are gated by
  `peer.routing` and use `peer.externalInterface` (per-peer, not global).

### 3.3 IP address representation

- A peer's address stays in the existing field
  `InterfaceSectionWgQuick.Address` (already a `[]string`). Pool-assigned
  addresses are written **with the pool CIDR mask** (e.g. `10.0.2.101/32`
  for a client in a `/24` pool — note: assignment today uses `/32`; see
  Open Question 4.1 about mask).
- The pool tracks which addresses are *in use* by scanning all peers'
  `Address[0]` on load, so assignment is stateless and survives manual edits.

## 4. Go struct changes (`models/models.go`)

```go
type Configuration struct {
    PresharedKey *string   `yaml:"presharedKey,omitempty"`
    Settings     *Settings `yaml:"settings,omitempty"`   // NEW
    Peers        []Peer    `yaml:"peers"`
}

type Settings struct {
    Name               string            `yaml:"name,omitempty"`
    Pools              map[string]Pool   `yaml:"pools,omitempty"` // keyed by role
    // ExternalInterface is a DEFAULT ONLY — pre-fills Peer.ExternalInterface
    // on add. Generation uses the per-peer field; this does not drive output.
    ExternalInterface  string            `yaml:"externalInterface,omitempty"`
    Defaults           map[string]Defaults `yaml:"defaults,omitempty"`
}

type Pool struct {
    CIDR  string `yaml:"cidr"`
    Start string `yaml:"start,omitempty"`
    End   string `yaml:"end,omitempty"`
}

type Defaults struct {
    ListenPort          *int `yaml:"listenPort,omitempty"`
    PersistentKeepalive *int `yaml:"persistentKeepalive,omitempty"`
}

type Peer struct {
    Name        string `yaml:"name"`
    Role        string `yaml:"role,omitempty"`   // NEW: server|client|""
    Description *string `yaml:"description,omitempty"`
    PeerOverrides map[string][]string `yaml:"peerOverrides,omitempty"`
    // NEW: per-peer routing. routing toggles PostUp/PostDown emission;
    // externalInterface is the OS interface used in the MASQUERADE rule.
    // Both empty on old configs → no routing rules (unchanged behavior).
    Routing            bool   `yaml:"routing,omitempty"`
    ExternalInterface  string `yaml:"externalInterface,omitempty"`
    InterfaceSection        `yaml:"interfaceSection"`
    InterfaceSectionWgQuick `yaml:"interfaceSectionWgQuick,omitempty"`
    PeerSection             `yaml:"peerSection"`
    PeerSectionWgQuick      `yaml:"peerSectionWgQuick,omitempty"`
}
```

`*Settings` (pointer) keeps `omitempty` meaningful: a file with no
`settings:` block deserializes to `nil`, and we branch on that for backward
compatibility.

## 5. New `pool` package

A new `internal/pool` (or top-level `pool`) package owns address math:

```
type Manager struct { settings *models.Settings; used map[string]struct{} }

func New(s *models.Settings) *Manager
func (m *Manager) Assign(role string) (string, error)      // next free in role's pool, with mask
func (m *Manager) Release(addr string)                     // mark free (used by remove)
func (m *Manager) Available(role string) (int, []string)   // count + sample for TUI display
func (m *Manager) Validate() []error                        // overlaps, bad cidr, start>end
```

Implementation uses stdlib `net/netip` (already in the module graph via
wireguard deps) — no new dependency. Reserved-address skipping:
network/broadcast for IPv4; configurable gateway skip later.

## 6. Behavior changes per command

### `bootstrap` (`cmd/bootstrap/bootstrap.go`)
- When creating the config, also write a default `settings` block with
  `server`/`client` pools (`10.0.2.0/24`, server `.1`-`.10`, client `.100`-
  `.200`), `externalInterface: eth0`.
- Assign `.1` to the server, `.100`/`.101` to the two clients via the pool
  (instead of the hardcoded `10.0.2.1/2/3`).
- `role` is set explicitly (`server`/`client`) on the created peers.
- Existing flag behavior preserved; `bootstrap` output is still a ready file.

### `add` (`cmd/add/add.go`)
- Add `-role` flag (`server`|`client`; default infer from `-endpoint`).
- If `-ip` omitted **and** `settings.pools[role]` exists → auto-assign from
  pool (new behavior).
- If `-ip` omitted and no pool → current behavior (no address set), with a
  one-line warning suggesting to define a pool.
- Routing becomes structured and per-peer:
  - Add `-routing` (bool) replacing the old `-add-routing <iface>` string
    flag. `-add-routing` is kept as a deprecated alias: if passed non-empty,
    it sets `routing=true` and `externalInterface=<value>` (back-compat).
  - Add `-interface` flag (string). When `-routing` is set, the per-peer
    `externalInterface` is resolved as: `-interface` flag → else
    `settings.externalInterface` → else `eth0` (legacy default).
  - The value is stored on the peer (`peer.Routing=true`,
    `peer.ExternalInterface=<iface>`); the `PostUp`/`PostDown` blocks are
    **no longer written verbatim by `add`** — generation derives them from
    these fields (see §6 generate). This is the key change that makes the
    interface editable later without string surgery.
- `role` written to the new peer.

### `bootstrap` (`cmd/bootstrap/bootstrap.go`) — routing
- The server peer keeps `routing: true` and `externalInterface: eth0`
  as explicit per-peer fields (instead of the verbatim `PostUp`/`PostDown`
  it writes today). Generation derives the rules from these fields.

### `remove` (`cmd/remove/remove.go`)
- After removing a peer, call `pool.Release` on its address (no-op if no
  pool configured). The address becomes available for reuse.

### `generate` / `format` / `recreate`
- `format` / `recreate` unchanged in output logic (they re-marshal; `format`
  will now also print `settings` and the new peer fields).
- `generate` gains one new responsibility: **derive `PostUp`/`PostDown` for a
  peer from `peer.Routing` + `peer.ExternalInterface`** instead of emitting
  whatever `PostUp`/`PostDown` strings the peer happens to carry. The rules
  are the same set `add`/`bootstrap` write today (forward + NAT + MASQUERADE
  on `wg0` and the external interface), produced by a shared helper
  (`models.RoutingRules(peer) (postUp, postDown []string)`) so `add`,
  `bootstrap`, and `generate` never disagree.
- Backward-compat path: if `peer.Routing == false` **and** the peer already
  has non-empty `PostUp`/`PostDown` (an old config), `generate` emits them
  verbatim as today — so existing files keep their hand-written rules. The
  derived path only takes over once a peer has `routing: true` (i.e. created
  via the new `add`/`bootstrap`/TUI).

### `init` (`cmd/initialize/initialize.go`)
- New `-with-settings` (bool, default true) to also emit a `settings` block.
  `SamplePeer` gains a `role` field for the dummy template.

## 7. Backward compatibility — strict rules

1. `LoadYaml` on a file with no `settings:` → `cfg.Settings == nil`.
   All pool-based code paths must no-op when `Settings == nil` or
   `Pools[role]` missing.
2. `LoadYaml` on a file with no `role` on peers → `Role == ""`; inference
   runs on demand (`models.Peer.EffectiveRole()`), never rewriting the file
   unless the user edits in the TUI.
3. `GetYaml` must not emit `settings: {}` for a nil settings (pointer +
   `omitempty` ensures this). It must not emit `role: ""`.
4. No existing test fixtures or example YAML in the repo today, so no
   fixture breakage — but any future fixtures must keep working without
   `settings`.
5. Routing: an old peer with no `routing` field and no `externalInterface`
   but with existing `PostUp`/`PostDown` strings must keep generating those
   strings verbatim (see §6 generate). The derived-rules path only applies
   when `routing: true` is present. `Routing` is a plain bool with `omitempty`,
   so `false` (the default) is never serialized — old files stay clean.
6. `externalInterface` resolves per-peer first; `settings.externalInterface`
   never overrides an explicit per-peer value during generation.

## 8. Open questions

- **4.1 Mask on assigned addresses.** Keep current `/32` per address, or
  assign with the pool's mask (e.g. `/24`)? WireGuard commonly uses `/32`
  per peer; we default to `/32` to match current behavior and avoid route
  surprises.
- **4.2 Pool shared vs split.** Should server and client pools be allowed
  to overlap (same CIDR, different ranges) as in the example, or must they
  be disjoint? Default: allow overlap; ranges just must not collide. The
  `used` set is global so a server `.5` won't be handed to a client.
- **4.3 Renaming `AllowedIps` default for clients.** `bootstrap` currently
  sets a client's own `AllowedIps` to its own address. With pools we keep
  that pattern (auto-assign writes the same value into both `Address[0]`
  and `AllowedIps[0]` for clients).
- **4.4 WireGuard interface name (`wg0`).** The routing rules reference `wg0`
  (the wireguard interface) in addition to the external interface. Should
  `wg0` also be a per-peer field (or a setting)? Default: keep `wg0` as a
  constant for now — it matches `bootstrap`/`add` today and the wg-quick
  interface name rarely varies per peer. Promote to a field later if needed.
- **4.5 Migrating existing verbatim rules.** When a user loads an old config
  whose server has hand-written `PostUp`/`PostDown` (with `eth0` baked in)
  and enables `routing: true` in the TUI, do we (a) replace the verbatim
  block with derived rules using the per-peer `externalInterface`, or
  (b) keep both? Recommend (a) with a one-time confirm prompt in the TUI;
  this is the migration that finally makes routing editable. CLI path
  stays non-destructive (only acts when `routing: true`).
