# gokrazy-cmdgroup

Run multiple instances of the same command with different arguments in [gokrazy](https://gokrazy.org).

Gokrazy's `PackageConfig` is keyed by package path, so you can only have one configuration entry per package. `cmdgroup` works around this by managing multiple instances of a command, with optional automatic restarting.

## Usage

```sh
gok add github.com/tho/gokrazy-cmdgroup
```

Configure it in your gokrazy config. The first positional argument is the command to run, and `--` separates per-instance arguments:

```json
"PackageConfig": {
    "github.com/tho/gokrazy-cmdgroup": {
        "CommandLineFlags": [
            "command", "-flag0",
            "--", "arg1", "-flag1.1", "-flag1.2",
            "--", "arg2", "-flag2.1"
        ]
    }
}
```

This runs two `command` instances:
1. `command -flag0 arg1 -flag1.1 -flag1.2`
2. `command -flag0 arg2 -flag2.1`

### Flags

| Flag | Description |
|------|-------------|
| `-watch` | Restart instances on exit: `none` (default), `all`, or comma-separated indices (e.g. `0,1`) |

## Example: Tailscale

Run both `tailscale up` and `tailscale serve` on a single gokrazy instance:

```json
{
    "Hostname": "tailscale",
    "Packages": [
        "tailscale.com/cmd/tailscale",
        "tailscale.com/cmd/tailscaled",
        "github.com/tho/gokrazy-cmdgroup"
    ],
    "PackageConfig": {
        "github.com/tho/gokrazy-cmdgroup": {
            "CommandLineFlags": [
                "tailscale",
                "--", "up", "--auth-key=file:/etc/tailscale/auth_key",
                "--", "serve", "text:hello"
            ],
            "ExtraFilePaths": {
                "/etc/tailscale/auth_key": "tailscale.auth_key"
            }
        }
    }
}
```

This runs two `tailscale` instances:

1. `tailscale up --auth-key=file:/etc/tailscale/auth_key`
2. `tailscale serve text:hello`
