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
