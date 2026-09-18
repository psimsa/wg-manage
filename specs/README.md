# wg-manage revival — spec & roadmap

This directory contains the design specs for reviving `wg-manage` with three
additions on top of the existing CLI:

1. **TUI** — edit the whole configuration interactively, no manual YAML editing.
2. **Automatic IP assignment** — define an address pool; new nodes get an
   address automatically, with separate sub-pools for servers and clients.
3. **Network visualization** (optional) — render the configured network.

All three rest on a new **`settings`** section in the YAML that holds
project-level configuration (primarily the IP pool definition).

## Documents

| File | Topic |
| --- | --- |
| [01-data-model-and-settings.md](01-data-model-and-settings.md) | YAML schema changes, the `settings` section, IP pools, peer role, backward compatibility |
| [02-tui.md](02-tui.md) | The `tui` command, screens, flows, library decision |
| [03-network-visualization.md](03-network-visualization.md) | Optional network rendering |
| [04-implementation-plan.md](04-implementation-plan.md) | Phasing, affected files, testing approach, risks |

## Goals

- **No manual YAML editing for common operations.** Adding a server/client,
  setting the external interface for routing, and editing per-peer overrides
  should all be doable from the TUI.
- **Automatic, collision-free IP assignment.** The user declares pools once;
  the tool assigns the next free address per role when a node is added.
- **Backward compatible.** Existing `config.yaml` files without a `settings`
  section and without explicit roles must continue to load and generate
  correctly. New behavior only activates when the relevant settings are present
  (or when the user opts in via the TUI).
- **Non-interactive CLI still works.** The existing `bootstrap`, `add`, `remove`,
  `generate`, `init`, `format`, `recreate` commands keep working. The TUI is an
  additional command that operates on the same YAML and the same `models`
  package.

## Decisions (resolved)

- **TUI library: Bubble Tea** (`bubbletea` + `bubbles` + `lipgloss`) — see [02-tui.md](02-tui.md).
- **Explicit `role` field on `Peer`** (with endpoint-based inference for old configs) — see [01-data-model-and-settings.md](01-data-model-and-settings.md).
- **Visualization: ASCII only** (no Graphviz/dot export) — see [03-network-visualization.md](03-network-visualization.md).
- **`go.mod` bumped to the latest Go** (not pinned to 1.18) — see [04-implementation-plan.md](04-implementation-plan.md).

## Out of scope

- Running/applying configs to live WireGuard interfaces (the tool generates
  files; applying them with `wg-quick` stays a manual/OS step).
- Key management beyond generation (no importing existing keys, no rotation
  scheduling).
- Multi-server mesh routing auto-generation (servers can still be configured
  manually; the pool just assigns addresses).
