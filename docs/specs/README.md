# Specs

Specification documents for the wg-manage revival. Each spec is self-contained
and can be implemented independently, though they are designed to compose.

## Goals (from the revival request)

1. **No manual YAML editing** — a TUI should cover adding peers, routing/NAT
   setup (including the egress interface), and per-peer overrides.
2. **Automatic IP assignment** — an IP pool declared in the config; addresses
   are allocated automatically when adding nodes. Separate subpools for
   servers (peers with an endpoint) and clients (peers without).
3. **Network visualization** (optional) — render the configured network.

All project-level concerns (pool, defaults) live in a new top-level `settings`
section of the YAML file.

## Documents

- [001 — Settings section](./001-settings-section.md) — new top-level
  `settings` YAML section: IP pool, subpools, server defaults, generate
  defaults. Backward compatible with existing config files.
- [002 — Automatic IP address management (IPAM)](./002-ipam.md) — automatic
  address allocation from the pool/subpools on `add`/`bootstrap`, freeing on
  `remove`, validation.
- [003 — TUI](./003-tui.md) — interactive terminal UI (`wg-manage tui`) for
  peer list, peer editing (incl. routing helper and overrides), settings
  editing, and generating configs. No manual YAML editing needed.
- [004 — Network visualization](./004-network-visualization.md) — optional
  `wg-manage viz` command rendering the network as ASCII / Graphviz DOT /
  Mermaid.

## Implementation order

1. **001 + 002** — extend the data model and CLI commands (`add`, `bootstrap`,
   `generate`, `remove`, `init`) to understand `settings` and the IP pool.
   This makes the YAML schema the single source of truth and is fully testable
   without a UI.
2. **003** — the TUI on top of the same model; it only reads/writes YAML via
   the existing `models` package plus the IPAM logic from 002.
3. **004** — visualization, usable both as a CLI command and as a TUI tab.

Cross-cutting: unit tests for all pure logic (parsing, allocation, rendering),
backward compatibility with existing `config.yaml` files (no `settings`
section), and README updates once features land.
