# wg-manage functionality

## Overview

`wg-manage` is a CLI tool for managing WireGuard configuration files from a single YAML source. It can bootstrap a starter topology, add/remove peers, generate per-peer WireGuard configs (optionally with QR codes), and reformat or regenerate keys.

## Configuration model

- All configuration lives in a YAML file (default: `config.yaml`).
- The YAML schema is defined in `models/models.go` and includes WireGuard and `wg-quick` options.
- Optional fields use pointers to distinguish between omitted vs. explicit values.
- `peerOverrides` allows adding extra lines for a specific peer in another peer's `[Peer]` section, keyed by the target public key.

## Commands

All commands are invoked as `wg-manage <command> [flags]`. The short aliases are shown below.

### `bootstrap` (`b`)
Creates a starter configuration with one server and two clients.

- **Inputs**:
  - `-endpoint` (default: `some.server.somewhere:51820`)
  - `-persistent` (default: `false`)
  - `-output` (default: `config.yaml`)
- **Behavior**:
  - Generates keys for three peers.
  - Creates a server peer with NAT/PostUp/PostDown rules and a `10.0.2.1/32` address.
  - Creates two client peers with `10.0.2.2/32` and `10.0.2.3/32` addresses.
  - Writes the resulting YAML to the output file.
- **Example**:
  ```bash
  wg-manage bootstrap -endpoint "my.example.com:51820" -output config.yaml
  ```

### `generate` (`g`)
Generates WireGuard configuration files from the YAML configuration.

- **Inputs**:
  - `-config` (default: `config.yaml`)
  - `-output` (default: `./output`)
  - `-png` (default: `true`)
- **Behavior**:
  - For each peer in the YAML, writes a `<PeerName>.conf` file to the output directory.
  - Includes `[Interface]` block for the peer and `[Peer]` blocks for other peers.
  - When `-png=true`, generates a QR code PNG for each `.conf` file.
- **Example**:
  ```bash
  wg-manage generate -config config.yaml -output ./out -png=false
  ```

### `add` (`a`)
Adds a new peer to the YAML configuration.

- **Inputs**:
  - `-name` (default: `peer-1`)
  - `-ip` (default: empty)
  - `-endpoint` (default: empty)
  - `-persistent` (default: `false`)
  - `-add-routing` (default: empty string)
  - `-config` (default: `config.yaml`)
- **Behavior**:
  - Generates a new keypair for the peer.
  - Sets allowed IPs to the provided `-ip` address (if present).
  - When `-endpoint` is supplied, extracts the port into `ListenPort`.
  - When `-add-routing` is supplied, appends PostUp/PostDown NAT rules.
  - Appends the peer to the YAML file and writes it back.
- **Example**:
  ```bash
  wg-manage add -name "Office" -ip 10.0.2.10/32 -config config.yaml
  ```

### `remove` (`r`)
Removes a peer from the YAML configuration by name or public key.

- **Inputs**:
  - `-name` (optional)
  - `-key` (optional)
  - `-config` (default: `config.yaml`)
- **Behavior**:
  - Removes the first peer matching the provided name or public key.
  - Writes the updated YAML if a match was found.
- **Example**:
  ```bash
  wg-manage remove -name "Office" -config config.yaml
  ```

### `init` (`i`)
Creates a new YAML configuration with sample data.

- **Inputs**:
  - `-peers` (default: `2`)
  - `-output` (default: `config.yaml`)
  - `-simple` (default: `false`)
  - `-preshared` (default: `true`)
- **Behavior**:
  - When `-simple=true`, creates minimal peer entries with placeholders.
  - When `-simple=false`, creates filled sample data for all WireGuard options.
  - Generates a preshared key when `-preshared=true`.
- **Example**:
  ```bash
  wg-manage init -peers 3 -output config.yaml
  ```

### `format` (`f`)
Prints a normalized YAML representation of the config to stdout.

- **Inputs**:
  - `-input` (default: `config.yaml`)
- **Behavior**:
  - Loads the YAML and writes it back to stdout with sorted peers.
- **Example**:
  ```bash
  wg-manage format -input config.yaml
  ```

### `recreate` (`rc`)
Regenerates keys for all peers in the YAML and prints the updated configuration.

- **Inputs**:
  - `-config` (default: `config.yaml`)
- **Behavior**:
  - Generates new keypairs for all peers.
  - Updates peer public keys and replaces any references in `peerOverrides`.
  - Prints updated YAML to stdout (does not write back to file).
- **Example**:
  ```bash
  wg-manage recreate -config config.yaml > new-config.yaml
  ```

## Expected behaviors

- Commands read and write the YAML file specified by `-config`/`-output`.
- `generate` creates a `.conf` and (optionally) `.png` for each peer.
- `format` and `recreate` write output to stdout.
- Errors result in non-zero exit codes or logged fatal errors.
