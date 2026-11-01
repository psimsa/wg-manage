# Spec 002 — Automatic IP address management (IPAM)

## Problem

Today the user must pick each peer's address by hand (`add -ip ...`) and
keep track of which addresses are free. With a declared pool this can — and
should — be automatic.

## Proposal

When `settings.ipPool` (and optionally `settings.subnets`) is present,
`wg-manage` allocates addresses automatically:

- **Servers** (peers with `endpoint`) are allocated from the `servers`
  subnet if defined, else from the whole pool.
- **Clients** (no `endpoint`) are allocated from the `clients` subnet if
  defined, else from the whole pool.
- Allocation picks the **first free address** in the subnet (lowest address
  that is not used by any peer's `address` list), skipping the network and
  broadcast addresses for IPv4. The allocated address is stored on the peer
  as `address: [<ip>/32]` (or `/128` for IPv6) and mirrored into
  `allowedIps`, exactly as `add -ip` does today.
- `remove` implicitly frees the address (nothing to do — addresses are
  derived from the peer list — but a `list`/`status` view should show free
  addresses; see spec 003).

### New/changed CLI surface

- `add -name X` (no `-ip`): allocate automatically. `-ip` remains as an
  explicit override (must be inside the pool; error otherwise).
- `add -server -endpoint host:port`: mark the new peer as a server; address
  comes from the servers subnet, `listenPort` defaults from
  `settings.serverDefaults.listenPort`.
- New command `wg-manage pool` (short `p`): prints the pool, subnets, used
  addresses (with owning peer names) and free addresses.
- `bootstrap`: instead of hardcoding `10.0.2.1/32`, `10.0.2.2/32`,
  `10.0.2.3/32`, it declares `ipPool: 10.0.2.0/24` in `settings` (plus
  `servers`/`clients` subnets) and allocates `10.0.2.1/32` to the server and
  `10.0.2.2/32`, `10.0.2.3/32` to the clients via the same IPAM path.
- `remove`: unchanged interface; freed addresses show up in `pool`.

### Allocation algorithm (pure function, unit-testable)

```
AllocateAddress(pool Prefix, subnet *Prefix, used []Prefix) (netip.Addr, error)
```

1. Determine range = subnet if given else pool.
2. Build the set of used addresses from all peers' `address` entries
   (parse each CIDR, take its IP).
3. Iterate addresses in the range in ascending order; skip network/broadcast
   for IPv4 (for IPv6 skip only the subnet-router anycast address, i.e. the
   first address, conservatively).
4. Return the first address not in the used set. Error if the range is full.

### Model changes

New file `models/ipam.go` with the pure functions above plus helpers:

- `UsedAddresses(cfg *Configuration) (map[netip.Addr]string, error)` —
  address → peer name.
- `AllocateForPeer(cfg *Configuration, peer *Peer) error` — fills
  `peer.Address`/`peer.AllowedIps` based on whether `peer.Endpoint != nil`
  and `settings.subnets`.
- `ValidatePool(settings *Settings) error` — called from `LoadYaml`;
  validates CIDRs and subnet containment.

`Peer.Address`/`AllowedIps` stay `[]string` — no change to the wire format.

### Edge cases

- Peer with multiple addresses: IPAM manages the first one; extra addresses
  are the user's responsibility (TUI can still edit them).
- Explicit `-ip` outside the pool: allowed only with `-force`; otherwise
  error "address X is outside pool Y".
- Pool exhaustion: clear error naming the subnet that is full.
- Overlapping peer addresses after manual YAML edits: `LoadYaml` warns
  (stderr) but does not fail; `pool` command reports the conflict.

### Acceptance criteria

- [ ] `bootstrap` produces the same three addresses as before, but via IPAM,
      and writes `settings.ipPool` + subnets.
- [ ] `add -name X` on a pool-config assigns the lowest free client address.
- [ ] `add -server -endpoint ...` assigns from the servers subnet and sets
      `listenPort` from `serverDefaults`.
- [ ] `remove` frees the address; a subsequent `add` reuses it.
- [ ] `pool` command lists used/free addresses per subnet.
- [ ] Old configs without `settings` behave exactly as today (`-ip`
      required or address stays empty, as now).
- [ ] Unit tests: allocation order, exhaustion, IPv4 network/broadcast
      skip, IPv6, subnet containment, explicit-ip validation.
