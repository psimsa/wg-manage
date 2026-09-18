# 03 — Network visualization (optional)

Status: **proposed / optional**
Affects: new `internal/view` (or top-level `view`) package, new `view` command (optional).

## 1. Problem & goal

The user listed visualization as optional. The goal is a quick, readable
picture of the configured network: which peers are servers (have endpoints),
which are clients, who points at whom, and what addresses are in use —
without leaving the terminal or installing heavy tooling.

## 2. Scope

This is explicitly **optional** and should be implemented last (after the TUI
and pool work). It must not block the core deliverables.

- **ASCII render in the terminal** (the chosen approach). Zero new
  dependencies. Printed by a `view` command and also reachable as a
  tab/panel inside the TUI.

No Graphviz/`dot` export, no interactive graph, no embedded image
rendering, no web server. (A dot exporter was considered and dropped —
keeping the dependency-free footprint.)

## 3. `view` command

```
wg-manage view | v                 # ASCII to stdout
wg-manage view -config foo.yaml
```

- Registered in `main.go` (`ShortCommand() "v"`, `LongCommand() "view"`).
- Non-interactive, like the other existing commands (uses `flag`).

## 4. Model of the graph

Build a small graph from the loaded `Configuration`:

- **Nodes** = peers. Each node has: name, effective role, address, endpoint
  (servers), public key (truncated for labels).
- **Edges** = "config references": peer A's generated `[Peer]` section
  references peer B when B appears as a peer in A's config. Concretely, the
  `generate` command already computes this: for each peer, it writes a
  `[Peer]` section for every *other* peer that is either a server (has
  endpoint) or — in the general case — any peer whose section should appear.
  We reuse the **exact same skip rule** as `generate`:

  `generate.go` skips writing peer2 into peer's file when:
  `peer2.PublicKey == peer.PublicKey || (peer2.Endpoint == nil && peer.Endpoint == nil)`

  So an edge `A → B` exists iff B would appear in A's generated config. We
  factor that skip rule into a shared helper
  (`models.ShouldIncludePeer(parent, candidate) bool`) so the view and
  generate never disagree.

- **Roles:**
  - `server` nodes (effective role server) are drawn with a distinct glyph
    and annotated with their endpoint.
  - `client` nodes are drawn plainly.
- Address shown under each node; pool utilization (used/total) printed as a
  summary footer when a `settings` pool exists.

## 5. ASCII output

A directed text layout. Example:

```
home-vpn  (pool: server 10.0.2.0/24 1-10, client 10.0.2.0/24 100-200; used 3/..)

  ┌─[ Server ]──────────────────┐
  │ 10.0.2.1/32  1.2.3.4:51820  │
  └─────────────┬───────────────┘
                │
      ┌─────────┴─────────┐
      ▼                   ▼
 ┌─[ My Laptop ]─┐   ┌─[ My Phone ]─┐
 │ 10.0.2.100/32 │   │ 10.0.2.101/32 │
 └───────────────┘   └──────────────┘
```

Implementation notes:
- Pure `fmt`/`strings`; no dependency. Layout is a simple tiered layout:
  servers on top, clients below, edges as ASCII connectors.
- For > a handful of nodes this won't be pretty; that's acceptable for an
  optional feature. If there are many clients, list them in columns and
  fan out connectors; if it overflows the terminal width, fall back to an
  indented adjacency list:

```
Server (10.0.2.1/32 @ 1.2.3.4:51820)
  -> My Laptop   (10.0.2.100/32)
  -> My Phone    (10.0.2.101/32)
My Laptop (10.0.2.100/32)
  -> Server      (10.0.2.1/32 @ 1.2.3.4:51820)
```

The adjacency-list form is the safe fallback and also the simplest first
implementation; the box-and-arrow form is a polish step.

## 6. TUI integration (optional)

If the TUI exists, add a `view` panel (key `v` from the peer list) that
renders the ASCII form in a scrollable viewport. No separate command needed
for this; reuses the ASCII renderer. This is lower priority than the core
TUI flows.

## 7. Open decisions

- **7.1** Edge semantics: "who references whom in generated config" (chosen)
  vs "who can reach whom" (requires routing analysis, out of scope).
