[![Go](https://github.com/ofcoursedude/wg-manage/actions/workflows/go.yml/badge.svg)](https://github.com/ofcoursedude/wg-manage/actions/workflows/go.yml)
[![wg-manage on snapcraft](https://snapcraft.io/wg-manage/badge.svg)](https://snapcraft.io/wg-manage)

# wg-manage

A command line tool to centrally manage [Wireguard](https://www.wireguard.com/) configuration files - all config options are stored in one YAML file that is then used to generate the config files for each device. It supports all options found in wg config files including wg-quick extensions (e.g. Address, Post/Pre-Up/Down etc.). It also has a quickstart option that bootstraps configs for ready to run network (one server, two clients).

# Motivation

Wireguard is great, but managing it can be pain in the butt. I created this tool to keep in sync two servers (endpoints) and roughly 10 clients across my home, google cloud, azure and my vacation home, ranging from Raspberry Pi Zero to my development laptop.

# Installation

There are a few options to install:

- Download a pre-built binary from the Releases page,
- All commits to `main` are built with Github Actions and artifacts are published, so you can grab an executable for any build you see in the history, or
- Simply git-clone the repo and build from source ([golang](https://golang.org) required)

and, newly

[![Get it from the Snap Store](https://snapcraft.io/static/images/badges/en/snap-store-black.svg)](https://snapcraft.io/wg-manage)

> Note: Though there are darwin binaries available, they are simply auto-built but untested. On Mac, you might be better off building your own.

# Prerequisites

The only real prerequisite is to have a public IP address for your server with an open port. Wireguard official documentation often uses 51820, but you can pick whatevs, of course.

Also, Wireguard or OS-specific topics are not in scope of this guide.

# Quickstart

The easiest way is to run the following two commands:

```
wg-manage bootstrap -endpoint "<public IP or fqdn>:<port>"
wg-manage generate
```

The first command creates a new YAML file with basic configuration (one server and two clients) and the second turns it into distinct config files that can be used with Wireguard's `wg-quick` command.

Once you're done, you can distribute the config files to the devices and fire up your wireguard network.

> **Please note:** The bootstrapped configuration includes PostUp and PostDown command sets for the server configuration, and so will probably fail, or misbehave, on systems without `sysctl` and `iptables` (haven't tried). Client configurations should work on any device (tried Linux and iOS)

---

# Settings

New configurations created by `bootstrap` or `init` include a top-level `settings` section with project-level options. It is optional — older config files without it keep working unchanged.

```yaml
settings:
  ipPool: 10.0.2.0/24            # base CIDR for peer addresses
  subnets:                       # named subpools of ipPool
    servers: { cidr: 10.0.2.0/26 }
    clients: { cidr: 10.0.2.0/24 }
  serverDefaults:                # applied to peers with an endpoint when they are created
    listenPort: 51820
    egressInterface: eth0        # used for NAT/routing rules
  clientDefaults:                # applied to peers without an endpoint
    persistentKeepalive: 25
  generate:                      # defaults for the `generate` command (CLI flags win)
    outputDir: ./output
    png: true
```

Defaults are stamped onto a peer when it is added (`add`, `bootstrap`, `init`); changing `settings` later does not modify existing peers. `generate` uses `settings.generate` for `-output`/`-png` when the flags are not given.

## Automatic IP addresses (IPAM)

When `settings.ipPool` is set, addresses are allocated automatically — no need to pass `-ip`:

```
wg-manage add -name MyPhone            # gets the first free address from the clients range
wg-manage add -server -endpoint vpn.example.com:51821 -name Srv2   # from the servers range
wg-manage pool                         # show pool, subnets, used and free addresses
```

Servers (peers with `-endpoint`) are allocated from `subnets.servers` when defined, clients from `subnets.clients`, otherwise both from the whole pool; the lowest free address is picked and network/broadcast addresses are skipped. Removing a peer frees its address. An explicit `-ip` is validated against the pool (`-force` overrides).

---

# Adding and removing configurations

Aside from the `bootstrap` and `generate` commands, there are 3 more commands currently available:

### `add`

Used to add a new computer (either client or server) to the configuration.

### `remove`

Used to remove an existing computer from the configuration by either name or public key.

> **Note:** Once you remove the computer, its public and private keys will be lost and so reversing this operation is equal to adding a new computer.

After adding or removing a computer, it is necessary to run the `generate` command again and re-distribute new config files.

### `init`

Running this command will create a config file with all available options populated with dummy data. It can be helpful if you want to explore what options you have available.

### `format`

Reformats the input config yaml file and outputs to stdout.

### `recreate`

This command recreates all keys (private and public) in the yaml file. For safety reasons outputs to stdout.

### `pool`

Shows the configured IP pool, subnets, which addresses are used by which peer, and how many are free. Addresses are allocated automatically when `settings.ipPool` is set (see above), so `wg-manage add -name X` needs no `-ip`.

### `psk`

Manages preshared keys: `wg-manage psk -mode status` (default) shows the current state, `-mode single` sets one global `presharedKey`, `-mode per-peer` gives every peer its own `presharedKey` under `peerSection`, `-mode remove` clears all. `bootstrap -preshared single|per-peer` and `init -preshared none|single|per-peer` do the same at creation time. A peer-level key overrides the global one in generated `[Peer]` blocks.

### `tui`

Interactive terminal UI — no manual YAML editing needed:

```
wg-manage tui -config config.yaml
```

Tabs: **Peers** (list with server/client badges; `a` add — address auto-allocated from the pool, `e` edit all fields, `d` delete, `r` toggle NAT/routing on a server using `settings.serverDefaults.egressInterface`, `o` edit per-peer override lines by target peer *name*), **Settings** (view pool/subnets/defaults; `a` adds a default `settings` section to an existing config that lacks one), plus `g` generate, `s` save, `q` quit (asks to save unsaved changes). All edits go through the same model/IPAM code as the CLI commands.

### `viz`

Visualizes the network described by the config file:

```
wg-manage viz                          # ASCII diagram to stdout
wg-manage viz -format dot | dot -Tsvg > network.svg
wg-manage viz -format mermaid          # paste into any Mermaid renderer
```

Shows servers (peers with an endpoint) on top, clients below, key fingerprints, per-peer override counts, and free addresses from the pool.

### Server listen port

Servers get `ListenPort` automatically: `bootstrap -port 443` (also rewrites the endpoint port), `add -server -endpoint host:port` (port taken from the endpoint), or `settings.serverDefaults.listenPort`. An explicit `add -port` wins over both.

---

# Other information

## config files

The available options correspond to options available in config files for the [wg](https://git.zx2c4.com/wireguard-tools/about/src/man/wg.8) and [wg-quick](https://git.zx2c4.com/wireguard-tools/about/src/man/wg-quick.8) tools as of 19-Mar-2021. Options specific for the wg-quick tool are in separate property, so you know what to remove from the yaml file if you want to use only the wg command.

## overrides

Information generated for the `[peer]` section of each configuration is primarily controlled by that peer's record in the yaml config file. There is however an ability to add extra lines to a `[peer]` section not defined by that given peer. It is a list of strings specified by peer's public key. It might be helpful for example if you want to have extra addresses or routing rules.
For example:

```yaml
- name: My Phone
  peerOverrides:
    im/jE5i7pvvODrlGbRaZT35C+NnrRfeFYR4IwAqNUkk=:
      - AllowedIPs=192.168.0.0/24
```

will result in an extra line for peer with public key `im/jE5i7pvvODrlGbRaZT35C+NnrRfeFYR4IwAqNUkk=` (assuming that's your Server) in configuration file for `My Phone` which won't appear in other configuration files for this peer. So for `My Phone` the peer section will be:

```
[Peer]
# Server
PublicKey=im/jE5i7pvvODrlGbRaZT35C+NnrRfeFYR4IwAqNUkk=
AllowedIPs=10.0.2.3/32
Endpoint=<endpoint info>
AllowedIPs=192.168.0.0/24
```

while for `My Laptop` it will be:

```
[Peer]
# Server
PublicKey=im/jE5i7pvvODrlGbRaZT35C+NnrRfeFYR4IwAqNUkk=
AllowedIPs=10.0.2.2/32
Endpoint=<endpoint info>
```
