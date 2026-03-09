# Nethopper

A tool for sharing internet access from a free network to a restricted network using [Xray-core](https://github.com/XTLS/Xray-core) reverse proxy.

## Overview

Nethopper enables users in a restricted network to access the internet through a bridge device that has connections to both networks. It uses a VLESS reverse tunnel between server and client, and exposes a SOCKS5 proxy for end users.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         NETWORK TOPOLOGY                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   App/Browser                  nhserver                  nhclient   │
│   (SOCKS5 client)             (Linux VM)              (Bridge PC)   │
│        │                          │                        │        │
│        │ SOCKS5                   │ VLESS Reverse          │        │
│        │ :1080                    │ Tunnel :2083           │        │
│        ▼                          │                        │        │
│  ┌────────────────────────────────┤                        │        │
│  │     RESTRICTED NETWORK         │◄───────────────────────┤        │
│  │                                │   restricted interface │        │
│  │  Users connect to Server:1080  │                        │        │
│  │  via SOCKS5 proxy              │                        │        │
│  └────────────────────────────────┤                        │        │
│                                   │                        │        │
│  ┌────────────────────────────────┤                        │        │
│  │     FREE INTERNET              │                        │        │
│  │                                ◄────────────────────────┤        │
│  │  Traffic exits via free        │     free interface     │        │
│  │  interface to real internet    │                        │        │
│  └────────────────────────────────┘                        │        │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘

Data flow:
1. App → SOCKS5 (:1080) → nhserver
2. nhserver (Portal) → VLESS reverse tunnel (:2083) → nhclient (Bridge)
3. nhclient → free interface → Internet
```

## Components

| Component | Binary | Platforms | Description |
|-----------|--------|-----------|-------------|
| Server | `nhserver` | Linux | Xray portal + SOCKS5 inbound, runs as systemd service |
| Client | `nhclient` | Linux, macOS, Windows | Xray bridge, connects to server, routes via free interface |

Xray-core is downloaded automatically on first install. Set `NETHOPPER_XRAY_PATH` to use a local binary instead:

```bash
sudo NETHOPPER_XRAY_PATH=/path/to/xray nhserver install
NETHOPPER_XRAY_PATH=/path/to/xray nhclient install
```

On the server, the binary is copied to `/usr/local/bin/xray` so the systemd service can access it regardless of where the original is located.

## Requirements

### Server (nhserver)
- Linux with systemd
- Root privileges
- Network connectivity within the restricted network

### Client (nhclient)
- Two network interfaces:
  - **Free**: connected to free/unrestricted internet (traffic exits here)
  - **Restricted**: connected to the restricted network (tunnel to server goes through here)
- On Linux, `sudo` is needed once during install to set `CAP_NET_RAW` on the xray binary (required for interface binding)

## Usage Modes

Both `nhserver` and `nhclient` can be used in two ways:

- **Interactive TUI**: Run without arguments (`sudo nhserver` or `nhclient`) to launch a menu-driven interface that guides you through each step.
- **CLI**: Use subcommands directly (e.g., `sudo nhserver install`, `nhclient configure ...`) for scripting or quick usage.

## Quick Start

### 1. Install and Set Up the Server

```bash
sudo nhserver install
```

Or via TUI: run `sudo nhserver` and select **Install**.

This downloads Xray, generates config with a random UUID, creates a systemd service, and configures the firewall.

### 2. Get the Connection String

```bash
sudo nhserver status
```

Or via TUI: run `sudo nhserver` and select **Status**.

Copy the connection string (`nh://...`).

### 3. Install the Client

```bash
nhclient install
```

Or via TUI: run `nhclient` and select **Install**.

Downloads the Xray binary and sets `CAP_NET_RAW` capability (prompts for sudo on Linux).

### 4. Configure the Client

```bash
nhclient configure -c "nh://..." -f <free-interface> -r <restricted-interface>
```

Or via TUI: run `nhclient` and select **Configure** — it will prompt for the connection string and guide you through interface selection.

### 5. Run the Client

```bash
nhclient run
```

Or via TUI: run `nhclient` and select **Run**.

### 6. Use the Proxy

Configure apps to use the SOCKS5 proxy at `<server-ip>:1080`.

## CLI Reference

### Server Commands

```bash
sudo nhserver                      # Launch interactive TUI
sudo nhserver install              # Download xray, create service, configure firewall
sudo nhserver configure            # Update ports interactively
sudo nhserver configure --socks-port 8080 --tunnel-port 3000
sudo nhserver status               # Show status and connection string
sudo nhserver uninstall --force    # Remove everything
```

### Client Commands

```bash
nhclient                                                  # Launch interactive TUI
nhclient install                                          # Download xray + set capabilities
nhclient configure -c "nh://..." -f wlan0 -r eth0        # Configure with connection string and interfaces
nhclient run                                              # Start bridge (foreground)
nhclient status                                           # Show config and status
nhclient uninstall --force                                # Remove config and binary
```

## Connection String

The `nh://` connection string encodes server details for easy sharing:

```json
{
  "v": 1,
  "s": "192.168.1.100",
  "p": 2083,
  "sp": 1080,
  "u": "uuid-here"
}
```

## File Locations

### Server (Linux, root)
| File | Path |
|------|------|
| Xray binary | `/usr/local/bin/xray` |
| Server config | `/etc/nethopper/server.json` |
| Xray config | `/etc/nethopper/xray.json` |
| Systemd service | `/etc/systemd/system/nethopper.service` |

### Client (user-level)
| Platform | Binary | Config |
|----------|--------|--------|
| Linux | `~/.local/bin/xray` | `~/.config/nethopper/` |
| macOS | `~/.local/bin/xray` | `~/Library/Application Support/nethopper/` |
| Windows | `%APPDATA%\nethopper\bin\xray.exe` | `%APPDATA%\nethopper\` |

## Building from Source

Requires Go 1.24+.

```bash
make build          # Build both nhserver and nhclient
make build-server   # Build nhserver only
make build-client   # Build nhclient only
make build-all      # Cross-compile for all platforms
make test           # Run tests
```

## License

MIT
