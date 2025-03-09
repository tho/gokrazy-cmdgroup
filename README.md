# gokrazy-cmdgroup

A tool to run a group of instances of the same command in [gokrazy](https://gokrazy.org).

## Problem

Gokrazy's `PackageConfig` is a map keyed by package path, which means it is only
possible to have one configuration entry per package. This becomes a limitation
when the same package or command needs to be run multiple times with different
arguments, like running both `tailscale up` and `tailscale serve`.

## Solution

`gokrazy-cmdgroup` solves this by providing a single command that can manage and
supervise a group of instances of the same command.

Features:

- Run multiple instances of the same command with different arguments
- Optional automatic restarting of instances when they exit (`-watch`)
- Process group management for clean shutdowns

## Usage

Add the cmdgroup package to your gokrazy configuration:

```sh
gok add github.com/tho/gokrazy-cmdgroup/cmd/cmdgroup
```

or

```json
"Packages": [
    "github.com/tho/gokrazy-cmdgroup/cmd/cmdgroup"
]
```

Configure it with command line flags:

```json
"PackageConfig": {
    "github.com/tho/gokrazy-cmdgroup/cmd/cmdgroup": {
        "CommandLineFlags": [
            "-name", "command",
            "--",
            "arg1",
            "-flag1",
            "--",
            "arg2",
	    "-arg2"
        ],
        "ExtraFilePaths": {
            "/etc/tailscale/auth_key": "tailscale.auth_key"
        }
    }
}
```

### Command Line Flags

- `-name`: The command to run
- `-watch`: Control which instances to restart on exit
  - `none`: Don't restart any instance (default)
  - `all`: Restart all instances when they exit
  - `0,1`: Restart instances 0 and 1
- Separate arguments for each instance with `--`

## Example: Tailscale Up and Serve

Run both `tailscale up` and `tailscale serve`:

```json
{
    "Hostname": "ts",
    "Packages": [
        "tailscale.com/cmd/tailscale",
        "tailscale.com/cmd/tailscaled",
        "github.com/tho/gokrazy-cmdgroup/cmd/cmdgroup"
    ],
    "PackageConfig": {
        "github.com/tho/gokrazy-cmdgroup/cmd/cmdgroup": {
            "CommandLineFlags": [
                "-name", "/user/tailscale",
                "--",
                "up",
                "--ssh",
                "--auth-key=file:/etc/tailscale/auth_key",
                "--",
                "serve",
                "text:hello"
            ],
            "ExtraFilePaths": {
                "/etc/tailscale/auth_key": "tailscale.auth_key"
            }
        }
    }
}
```

This runs two instances:
1. `tailscale up --ssh --auth-key=file:/etc/tailscale/auth_key`
2. `tailscale serve text:hello`
