# 04 — Implementation plan

Status: **proposed**
Read with [01–03](README.md). Phased so each phase ships and builds independently.

## Guiding principles

- Smallest correct change per phase; each phase leaves `main` compiling and
  the existing CLI green.
- Backward compatibility is a hard constraint (see
  [01 §7](01-data-model-and-settings.md)): a config without `settings` and
  without `role` must load and generate exactly as today.
- New dependencies only where justified: `pool` uses stdlib `net/netip`
  (no dep); the TUI adds the three Charm modules (justified, see
  [02 §4](02-tui.md)); visualization adds none.
- No code comments per repo style; specs carry the rationale.

## Phase 0 — Refactor for sharing (no behavior change)

Goal: extract shared logic so the new TUI and `view` can reuse it without
importing `cmd/*` (which would cycle with `main`).

- `models`:
  - Add `Peer.EffectiveRole() string` (server/client inference from
    `Endpoint`; pure function, no I/O).
  - Add `models.ShouldIncludePeer(parent, candidate Peer) bool` encoding the
    exact skip rule currently inline in `cmd/generate/generate.go`.
- `generate`:
  - Move the generate core out of `cmd/generate/generate.go` into a new
    package `generate` (top-level, not under `cmd`) exposing
    `Run(cfg models.Configuration, outputDir string, png bool) error`.
  - `cmd/generate/generate.go` becomes a thin flag-parsing wrapper calling
    `generate.Run(models.LoadYaml(*configFile), *outputDir, *png)`.
- `SaveYaml` hardening (small, optional in this phase): write to temp then
  rename. Keep `os.Create` path as fallback only if rename fails across
  filesystems.
- Add the first tests (see Testing) to lock current behavior before
  changing anything.

Exit criteria: `go build ./...`, `go vet ./...`, all cross-builds (see Build matrix), and
existing CLI smoke (bootstrap→generate) produce byte-identical output to
pre-refactor. No new flags yet.

## Phase 1 — `settings` & `role` (data model) + `pool` package

Spec: [01](01-data-model-and-settings.md).

- `models/models.go`: add `Settings`, `Pool`, `Defaults` structs; add
  `Settings *Settings` to `Configuration`; add `Role string` to `Peer`.
  Pointer + `omitempty` so old files round-trip cleanly.
- `models/peerFactories.go`: `NewPeer`/`SamplePeer` accept/set role; update
  `SamplePeer` dummy.
- New package `pool`:
  - `Manager`, `New(s *models.Settings)`, `Assign(role)`, `Release(addr)`,
    `Available(role)`, `Validate()`.
  - Uses `net/netip`; skips network/broadcast; honors `start`/`end` when set.
  - Builds the `used` set from all peers' `Address[0]` on construction.
- `cmd/bootstrap`: write a default `settings` block; assign server `.1`,
  clients `.100/.101` via `pool`; set explicit `role` on peers.
- `models`: add `Peer.Routing`, `Peer.ExternalInterface`; add
  `models.RoutingRules(peer) (postUp, postDown []string)` deriving the
  iptables/MASQUERADE rules from the per-peer fields (the rule set currently
  inlined in `add`/`bootstrap`). This is the shared helper.
- `cmd/add`: add `-role` (default infer), `-routing` (bool), `-interface`
  flags; keep `-add-routing` as a deprecated alias; auto-assign from pool
  when `-ip` omitted and pool exists; store `routing`+`externalInterface` on
  the peer (no verbatim `PostUp`/`PostDown`); write `role`.
- `cmd/bootstrap`: set `routing: true` + `externalInterface: eth0` on the
  server peer (instead of verbatim rules); assign via pool; set `role`.
- `generate`: when `peer.Routing == true`, emit `RoutingRules(peer)`;
  when `peer.Routing == false` but the peer already has verbatim
  `PostUp`/`PostDown`, emit them unchanged (back-compat).
- `cmd/remove`: `pool.Release` the removed peer's address (no-op without
  pool).
- `cmd/initialize`: `-with-settings` flag; `SamplePeer` role.
- `format`/`recreate`: no logic change (they re-marshal); verify
  `format` round-trips a settings-bearing file.

Exit criteria: old config (no settings/role) → `generate` output unchanged.
New `bootstrap` output includes `settings` + `role` and assigns from pools.
`go test ./pool/...` green.

## Phase 2 — TUI

Spec: [02](02-tui.md). Depends on Phase 1 (pool, role, settings).

- Add `go.mod` deps: `bubbletea`, `bubbles`, `lipgloss`.
- New package `tui` (top-level; imports `models`, `pool`, `generate` only).
  - `tui.Run(configFile string) error`.
  - Screens: peer list, add/edit form (role-aware, routing + external
    interface), overrides editor, settings editor, generate flow.
  - In-memory `*models.Configuration`; save only on Ctrl-S via
    `models.SaveYaml`; confirm-on-quit if dirty.
- New `cmd/tui/tui.go`: thin wrapper registering `tui`/`t`, parsing
  `-config`.
- `main.go`: register `tui` command.

Exit criteria: `wg-manage tui` opens, can add a server with routing + custom
external interface, edit overrides, edit pools, save, and generate — all
without touching YAML by hand. Builds on linux/darwin/windows. Manual
terminal test; add a non-TTY unit test for the model update logic (the Elm
model is pure and testable without a terminal).

## Phase 3 — Visualization (optional)

Spec: [03](03-network-visualization.md). Depends on Phase 1 (`EffectiveRole`,
`ShouldIncludePeer`); independent of the TUI but can integrate with it.

- New package `view`:
  - `view.ASCII(cfg models.Configuration) string` (adjacency-list first;
    box layout as polish). ASCII only — no dot export.
- New `cmd/view/view.go`: `v`/`view`, flags `-config`.
- `main.go`: register `view`.
- (Optional) TUI `v` panel reusing `view.ASCII`.

Exit criteria: `wg-manage view` prints a correct ASCII adjacency view for the
config. Edges match `generate`'s skip rule exactly (tested via
`models.ShouldIncludePeer`).

## Testing approach

The repo currently has no tests. Introduce table-driven tests alongside
the work (do not retrofit a test framework; use stdlib `testing`):

- `models`:
  - YAML round-trip: old file (no settings/role/routing) loads and
    re-marshals identically; new file (with settings/role/routing)
    round-trips.
  - `EffectiveRole` inference cases.
  - `ShouldIncludePeer` mirrors the generate skip rule (golden cases taken
    from a real bootstrap→generate run).
  - `RoutingRules`: derived rules for `routing:true` + given
    `externalInterface` equal the rule set `add`/`bootstrap` write today;
    back-compat — a peer with `routing:false` + verbatim `PostUp` keeps
    those strings through `generate` (golden test).
- `pool`:
  - Assign sequential, skip used, skip reserved, honor start/end, exhaust
    → error, custom-address added to used, Release reclaims.
  - `Validate` catches bad CIDR, start>end, range outside CIDR.
- `generate` (shared): golden output for a fixed config (with and without
  settings) so refactor and future changes are regression-guarded. Use a
  fixed key pair (inject keys, don't generate randomly in the test) so
  output is deterministic.
- `view`: edges equal `ShouldIncludePeer` for a few topologies; ASCII does
  not panic on empty config / single peer / many clients.
- TUI: test the pure `Update`/model transitions (add peer, toggle routing,
  mark dirty, save) without a real terminal.

## Build matrix & CI

Extend `.github/workflows/go.yml`:
- Add a `go test ./...` step (the workflow currently only builds).
- Add a **Windows arm64** build target alongside the existing ones. The
  full matrix becomes:

  | GOOS | GOARCH | output path |
  | --- | --- | --- |
  | windows | amd64 | `out/windows-amd64/wg-manage.exe` |
  | windows | arm64 | `out/windows-arm64/wg-manage.exe` |
  | linux | amd64 | `out/linux-amd64/wg-manage` |
  | linux | arm64 | `out/linux-arm64/wg-manage` |
  | darwin | amd64 | `out/darwin/wg-manage` |
  | darwin | arm64 | `out/darwin-arm64/wg-manage` |

  (darwin/arm64 is included as a no-cost addition since Apple Silicon is now
  the common Mac; the current workflow builds only generic `darwin`.)
- Upload artifacts for the new windows-arm64 (and darwin-arm64) targets.

## Risks & mitigations

| Risk | Mitigation |
| --- | --- |
| Backward-compat break on old configs | Phase 0 golden tests lock current output before any change; pointer+omitempty on new fields; nil-settings no-op everywhere. |
| Pool collision when users hand-edit addresses | `used` set rebuilt from peers on every load; `Validate` surfaces overlaps; TUI shows conflicts. |
| TUI import cycle with `cmd/*` / `main` | Shared `generate` lives top-level; `tui` imports only `models`,`pool`,`generate`,`view`. |
| New deps bloat binary / slow CI | Charm modules are small; cross-builds already pass. Keep deps to the three modules. |
| Windows TUI terminal quirks | Bubble Tea supports Windows ConHost/WT; test the windows cross-build and a manual run if feasible. Keep a non-TTY fallback path (`tui` prints "not a terminal" and exits cleanly if `isatty` false). |
| `net/netip` availability | `net/netip` is 1.18+. We bump `go.mod` to the **latest** Go (not 1.18) — CI already uses 1.27 and our local build uses 1.23; the bump is noted in the commit. |

## Suggested commit shape

One commit per phase (or a few logical commits per phase), all on
`vibe/revive-wg-manage-c3a210`. Each commit compiles and passes tests. PR
opened as draft after Phase 1 (core value, no TUI yet) or after Phase 2
depending on preference — open decision when we start implementing.

## Decisions (resolved)

1. TUI library — **Bubble Tea** ([02 §9](02-tui.md)).
2. `go.mod` Go version — **bump to the latest Go** (not 1.18); enables
   `net/netip` and matches CI's toolchain.
3. Visualization — **ship Phase 3, ASCII only** (no Graphviz/dot) — see
   [03](03-network-visualization.md).
4. Cross-builds — **add native Windows arm64** alongside the existing
   windows/amd64, linux/amd64, linux/arm64, darwin targets (see Build matrix).
