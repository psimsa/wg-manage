# Spec 005 — ListenPort for servers (issue #59)

**Status: implemented** (branch `vibe/revive-wg-manage-f50cd7`, PR #65).
Closes #59

## Problem (issue #59)

Every time a server is created, the user must manually add
`ListenPort=<port>` to the generated WireGuard config (or edit the YAML).
Automated scripts become harder to read and manage.

## Proposal

Make the listen port a first-class, automatic part of server creation:

1. **`bootstrap -port <port>`** (new flag, default `51820`): sets
   `listenPort` on the server peer and derives the endpoint port when the
   endpoint has no explicit port.
2. **`add -server -endpoint host:port`** (from spec 002): sets `listenPort`
   from the endpoint's port, or from `settings.serverDefaults.listenPort`
   (spec 001) when the endpoint omits the port or the flag is given
   explicitly via a new `add -port <port>` flag.
3. **TUI** (spec 003): the peer edit form exposes `listenPort` directly.
4. Precedence: explicit flag > endpoint port > `settings.serverDefaults.listenPort`
   > no ListenPort line (current wg default: random port).

## Implementation notes

- `models.InterfaceSection.ListenPort` already exists (`*int`) and
  `generate` already renders it — only the *setting* of the value is
  missing. Changes are confined to `cmd/add`, `cmd/bootstrap`, and the
  defaults plumbing from spec 001.
- Validation: port must be 1–65535; endpoint parsing should no longer
  panic on a missing `:` (current code does `strings.Split(*endpoint, ":")[1]`
  which would index out of range — fix while here, guarded by tests).

## Acceptance criteria

- [ ] `wg-manage bootstrap -endpoint vpn.example.com -port 443` produces a
      server peer with `listenPort: 443` and endpoint `vpn.example.com:443`.
- [ ] `wg-manage add -server -endpoint vpn.example.com:51820 -name S` sets
      `listenPort: 51820` automatically.
- [ ] `settings.serverDefaults.listenPort` is used when no flag/port is
      given.
- [ ] Endpoint without port no longer crashes; falls back per precedence.
- [ ] Unit tests for the precedence chain and endpoint parsing.
