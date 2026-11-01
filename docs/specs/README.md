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

Specs derived from filed GitHub feature requests:

- [005 — ListenPort for servers](./005-listen-port.md) — closes #59:
  automatic `ListenPort` when creating servers (`bootstrap -port`,
  `add -server -endpoint host:port`, `settings.serverDefaults.listenPort`).
- [006 — Preserve custom YAML keys](./006-preserve-custom-yaml-keys.md) —
  closes #34: user-added `key: value` pairs on peers/top level must survive
  `add`/`remove`/`format`/TUI round-trips (inline-map capture, golden tests).
- [007 — Preshared keys on bootstrap](./007-preshared-keys.md) — closes #16:
  per-peer PSK support in the model plus `bootstrap -preshared=single|per-peer`
  and a `psk` command; global PSK behavior stays backward compatible.

## Implementation order

1. **006** — preserve-unknown-keys is a prerequisite for everything else:
   once the model round-trips arbitrary YAML, new sections (001) and fields
   (007) can never silently destroy user data.
2. **001 + 002** — extend the data model and CLI commands (`add`, `bootstrap`,
   `generate`, `remove`, `init`) to understand `settings` and the IP pool.
   This makes the YAML schema the single source of truth and is fully testable
   without a UI.
3. **005 + 007** — small, self-contained CLI/model features (listen port
   automation, PSK generation) on top of the new schema.
4. **003** — the TUI on top of the same model; it only reads/writes YAML via
   the existing `models` package plus the IPAM logic from 002.
5. **004** — visualization, usable both as a CLI command and as a TUI tab.

Cross-cutting: unit tests for all pure logic (parsing, allocation, rendering),
backward compatibility with existing `config.yaml` files (no `settings`
section), and README updates once features land.
