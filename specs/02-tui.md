# 02 — TUI (`tui` command)

Status: **proposed**
Affects: new `cmd/tui/` package, new `internal/tui` (or top-level `tui`) package, `main.go`, `models/*` (view helpers only).

## 1. Problem

Every meaningful change today requires opening `config.yaml` in a text
editor: adding a peer with routing, setting the external interface, editing
`peerOverrides`, or fixing an endpoint. The YAML is correct but verbose, and
mistakes (wrong indentation, a stray `eth0`) only surface at `generate` time.

## 2. Goal

A `tui` command that lets the user manage the whole `config.yaml` from the
terminal, so manual YAML editing is not needed for any common operation.
Key flows it must cover (from the user's request):

- Add a peer (server or client) — including setting the external interface
  used for routing when adding a server.
- Remove a peer.
- Edit per-peer overrides (`peerOverrides`).
- Configure the `settings` section (pools, external interface, defaults).
- Generate configs from within the TUI (one keystroke, same as `generate`).
- Save the YAML (explicit; never auto-write on every keystroke).

## 3. New command

```
wg-manage tui | t            # opens TUI on ./config.yaml (default)
wg-manage tui -config foo.yaml
```

- Registered in `main.go` alongside the existing commands
  (`ShortCommand() "t"`, `LongCommand() "tui"`).
- The TUI is the only command that **does not** use the stdlib `flag`-based
  one-shot pattern; it runs an event loop and only writes the file on save.
- All other commands remain non-interactive and unchanged.

## 4. Library decision (open — recommend Bubble Tea)

| | **bubbletea** + **lipgloss** + **bubbles** | **tview** |
| --- | --- | --- |
| Model | Elm architecture (Model/Update/View) | Widget/callback |
| Dependencies | 3 small Charm modules, already widely cached | 1 larger lib, pulls in `tcell` |
| Forms/inputs | `bubbles/textinput`, `bubbles/list`, `bubbles/table` built-in | Rich built-in forms |
| Styling | `lipgloss`, modern look | Mature, older look |
| Maintenance | Very active | Active but slower |
| Risk | More glue code for forms | Easier forms, heavier |

**Recommendation: Bubble Tea.** It is the de-facto modern Go TUI, has first-
class list/table/textinput components (exactly what we need: a peer list, a
settings form, an overrides editor), and stays close to the repo's "small,
std-only" spirit by adding only well-scoped modules. Final call is an open
decision (see README), but the spec is written against Bubble Tea.

Dependency footprint to add to `go.mod`:
- `github.com/charmbracelet/bubbletea`
- `github.com/charmbracelet/bubbles`
- `github.com/charmbracelet/lipgloss`

## 5. Screens & navigation

```
┌─ wg-manage — home-vpn ─────────────────────────────┐
│ [Peers]  Settings  Generate  Help  Quit             │
│                                                     │
│   Name          Role     Address        Endpoint   │
│ ▸ Server       server   10.0.2.1/32    1.2.3.4:51820│
│   My Laptop    client   10.0.2.100/32              │
│   My Phone     client   10.0.2.101/32              │
│                                                     │
│  [a] add  [d] delete  [e] edit  [o] overrides       │
│  [s] settings  [g] generate  [q] quit               │
└────────────────────────────────────────────────────┘
```

### 5.1 Peer list (main screen)
- `bubbles/table` listing peers: Name, Role, Address[0], Endpoint.
- Role column shows the **effective** role (`EffectiveRole()`), with an
  indicator (`*`) when it is inferred (Role=="").
- Keys:
  - `a` — add peer (opens Add flow)
  - `e` / Enter — edit selected peer
  - `d` — delete selected peer (confirm)
  - `o` — edit `peerOverrides` for selected peer
  - `s` — settings screen
  - `g` — generate (runs the same logic as `generate` command, then shows
    a status toast with the output dir)
  - `q` / Ctrl-C — quit (confirm if unsaved changes)
- A "dirty" indicator (`*`) in the header when there are unsaved edits.
- `Ctrl-S` — save (writes YAML via `models.SaveYaml`).

### 5.2 Add / Edit peer form
Fields (server vs client fields shown conditionally on role):
- `name` (text input, required)
- `role` (server | client | auto/infer) — toggles which fields show
- `address` (text; prefilled with auto-assigned value from pool when adding;
  editable)
- `endpoint` (text; only meaningful for server; shown/required for servers)
- `listenPort` (text; for servers)
- `persistentKeepalive` (text)
- `allowedIps` (text, comma-separated)
- `routing` (bool; server only) — when on, exposes:
  - `externalInterface` (text; **per-peer**, prefilled from
    `settings.externalInterface` as a default but editable per peer, since
    routing depends on the underlying OS of each peer)
  - live preview of the derived `PostUp`/`PostDown` rules using this peer's
    interface (generated via `models.RoutingRules(peer)` — the same helper
    `generate` uses), shown at the bottom of the form
  - if editing an old peer that has verbatim `PostUp`/`PostDown` strings
    baked in: enabling `routing` shows a one-time confirm to replace those
    with derived rules (see [01 §8, Q4.5](01-data-model-and-settings.md));
    disabling `routing` leaves the verbatim block untouched
- Save / Cancel.

On add, if the user keeps the prefilled address it is recorded as
pool-assigned (so a later `remove` releases it). If the user types a custom
address it is used as-is and added to the pool's `used` set to avoid
collisions.

### 5.3 Per-peer overrides editor
For the selected peer, a list of `peerOverrides` entries:

```
Overrides for: My Laptop
  [publicKey]  AllowedIPs=192.168.0.0/24      [edit] [del]
  [+ add line for a target peer]
```

- "Target peer" chosen from the list of other peers (by public key).
- Each override is a free-text line (e.g. `AllowedIPs=192.168.0.0/24`).
- Add / edit / delete lines; saved into `PeerOverrides[targetPubKey]`.
- This fully replaces manual YAML editing of the `peerOverrides` map.

### 5.4 Settings screen
- `name` (text)
- `externalInterface` (text) — **default only**: pre-fills the per-peer
  field when adding a new server. A hint clarifies each peer overrides this
  with its own value (routing is per-peer).
- Pools editor: per role (`server`, `client`) → `cidr`, `start`, `end`.
  - Live validation: invalid CIDR, `start > end`, range outside CIDR shown
    inline in red.
  - "Available" preview: shows count of free addresses per pool given the
    currently assigned peers.
- Defaults editor: per role → `listenPort`, `persistentKeepalive`.

### 5.5 Generate flow
- Reuses the existing `generate` logic (extracted into a shared helper so
  both the `generate` command and the TUI call it — see implementation plan).
- Prompts for output dir (default `./output`) and `png` toggle, then runs.
- Shows a result toast: `Generated 3 configs + QR codes → ./output`.

## 6. Data flow

```
tui.Run(configFile)
  → models.LoadYaml(configFile)            // same loader as CLI
  → build a mutable in-memory *models.Configuration
  → event loop edits the in-memory config ONLY
  → on Ctrl-S: models.SaveYaml(cfg, configFile)   // same saver as CLI
  → on `g`: generate.RunFromConfig(cfg, outDir, png)  // shared helper
```

- The TUI never reads/writes YAML directly; it always goes through
  `models.LoadYaml` / `GetYaml` / `SaveYaml` so the on-disk format stays
  identical to the CLI's.
- IP assignment inside the TUI uses the same `pool.Manager` as `add`, built
  from `cfg.Settings`.

## 7. Saving & safety

- Single source of truth: an in-memory `*models.Configuration`. Edits mutate
  it; the file is touched only on explicit save.
- On quit with unsaved changes: confirm prompt (`y/n`).
- Save is atomic-ish: write to `configFile + ".tmp"` then rename, so a crash
  mid-write does not corrupt the existing file. (This also improves the
  existing `SaveYaml`, which uses `os.Create` directly — see implementation
  plan, but kept as a separate, optional hardening step.)
- Keys are regenerated only via the existing `recreate` command/flow; the
  TUI does not silently rotate keys.

## 8. Constraints

- Must build on the same OS targets as today (linux, darwin, windows). The
  Windows terminal handling for Bubble Tea is supported by the library.
- No goroutines touching the file system except the save action.
- The TUI package must not import `cmd/*` (to avoid import cycles with
  `main`); it imports only `models`, `pool`, and a shared `generate`
  helper placed outside `cmd`.

## 9. Open decisions

- **9.1** Bubble Tea vs tview — **resolved: Bubble Tea.**
- **9.2** Whether the TUI should auto-save on quit if dirty (recommend:
  always confirm).
- **9.3** Whether `generate`-from-TUI should run synchronously with a
  spinner (recommend: yes, it's fast).
