# Spec 004 — Network visualization (optional)

## Problem

With many peers it is hard to see at a glance who connects to whom, which
addresses are used, and which peers are servers vs clients.

## Proposal

A `viz` command (`wg-manage viz`) that renders the current configuration as
a graph. Default output: **ASCII diagram to stdout** (zero new heavy deps,
works over SSH). Optional formats via flags: Graphviz DOT (`-format dot`,
pipe to `dot -Tsvg`) and Mermaid (`-format mermaid`, paste into any Mermaid
renderer). The same renderer is embedded as the **Network** tab in the TUI
(spec 003) in ASCII mode.

### ASCII rendering

```
WireGuard network  (pool 10.0.2.0/24, servers 10.0.2.0/26, clients 10.0.2.64/26)

  ┌─ Servers ─────────────────────────────┐
  │ Server  10.0.2.1/32  vpn.example.com:51820 │
  └───────────────┬────────────────────────┘
                  │ peers
   ┌──────────────┼──────────────┐
   ▼              ▼              ▼
 My Laptop     My Phone     (free: 10.0.2.4 – 10.0.2.63, 10.0.2.65 – ...)
 10.0.2.2/32   10.0.2.3/32
```

Rules:

- Servers (peers with `endpoint`) in a top box; clients below; edges from
  each client to each server it references (in practice all servers, since
  `generate` emits all other peers as `[Peer]` blocks — the diagram reflects
  the generated configs: every non-server peer gets a `[Peer]` block for
  every server, and servers get blocks for each other).
- Each node shows name, address(es), and key fingerprint (first 8 chars of
  the public key) for identification.
- Free addresses from the pool are listed at the bottom (from the IPAM
  helpers in spec 002).
- Peer overrides are marked on the edge (e.g. `+2 override lines`).

### DOT / Mermaid

Same graph, emitted as `digraph` / `graph TD` with server/client node
shapes and edge labels.

### Implementation

New package `viz/` with a pure function
`Render(cfg *models.Configuration, format string) (string, error)` —
trivially unit-testable (golden tests on small configs). The command is
thin: load YAML, call `Render`, print.

### Acceptance criteria

- [ ] `wg-manage viz` prints an ASCII diagram of a bootstrapped config.
- [ ] `-format dot` / `-format mermaid` emit valid graph syntax (checked
      against `dot` in CI if available, otherwise golden files).
- [ ] Diagram reflects overrides and free-pool info.
- [ ] TUI Network tab shows the same ASCII view.
- [ ] Golden-file unit tests.
