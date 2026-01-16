# nethopper

A CLI tool for sharing internet access from a "free" network to a restricted network using [sing-box](https://github.com/SagerNet/sing-box) as the underlying proxy engine.

## Overview

nethopper enables users in a restricted network to access the internet through a bridge device that has connections to both networks. It uses MTProto (Telegram protocol) for client connections, making it compatible with Telegram proxy clients.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           NETWORK TOPOLOGY                               │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌──────────────────┐         ┌──────────────────┐         ┌──────────┐ │
│  │  Telegram App    │         │     Server       │         │ Freenet  │ │
│  │  (Mobile/Desktop)│         │   (Linux VM)     │         │   Node   │ │
│  └────────┬─────────┘         └────────┬─────────┘         └────┬─────┘ │
│           │                            │                        │       │
│           │ MTProto                    │ Reverse               │       │
│           │ :8443                      │ Tunnel                 │       │
│           ▼                            │ :2083                  │       │
│  ┌────────────────────────────────────┐│                        │       │
│  │       RESTRICTED NETWORK           ││◄───────────────────────┤       │
│  │                                    ││  restricted_interface  │       │
│  │  Users connect to Server:8443     ││                        │       │
│  │  using Telegram MTProto proxy     ││                        │       │
│  └────────────────────────────────────┘│                        │       │
│                                        │                        │       │
│  ┌────────────────────────────────────┐│                        │       │
│  │        FREE INTERNET               ││                        │       │
│  │                                    ││                        │       │
│  │  ◄─────────────────────────────────┴┼────────────────────────┤       │
│  │                                     │    free_interface      │       │
│  │  Traffic exits to real internet    │                        │       │
│  └────────────────────────────────────┘                        │       │
│                                                                  │       │
└─────────────────────────────────────────────────────────────────────────┘

DATA FLOW:
1. Telegram App → MTProto (:8443) → Server
2. Server → VMess (internal) → Reverse Tunnel (:2083) → Freenet Node
3. Freenet Node → free_interface → Internet
```

## Component Architecture

```
┌───────────────────────────────────────────────────────────────────┐
│                    COMPONENT ARCHITECTURE                          │
├───────────────────────────────────────────────────────────────────┤
│                                                                    │
│  SERVER (Linux)                    FREENET NODE (Any OS)          │
│  ═══════════════                   ══════════════════════          │
│                                                                    │
│  ┌──────────────┐                  ┌──────────────┐               │
│  │   MTProto    │ ◄── Telegram     │   VMess      │               │
│  │   Inbound    │     Clients      │   Inbound    │               │
│  │   :8443      │                  │   :8081      │               │
│  └──────┬───────┘                  └──────┬───────┘               │
│         │                                 │                        │
│         ▼                                 ▼                        │
│  ┌──────────────┐                  ┌──────────────┐               │
│  │   VMess      │                  │    Direct    │               │
│  │   Outbound   │ ────────────────►│   Outbound   │───► Internet  │
│  │   :8081      │  Reverse Tunnel  │  (free_if)   │               │
│  └──────────────┘      :2083       └──────────────┘               │
│                                                                    │
│  ┌──────────────┐                  ┌──────────────┐               │
│  │   Reverse    │ ◄───────────────►│   Reverse    │               │
│  │   Endpoint   │                  │   Endpoint   │               │
│  │   (listen)   │                  │   (connect)  │               │
│  └──────────────┘                  └──────────────┘               │
│                                           │                        │
│                                           ▼                        │
│                                    ┌──────────────┐               │
│                                    │    Direct    │               │
│                                    │   Outbound   │               │
│                                    │ (restricted) │               │
│                                    └──────────────┘               │
│                                                                    │
└───────────────────────────────────────────────────────────────────┘
```

## Requirements

### Server Mode (Linux only)
- Linux operating system (Debian, Ubuntu, RHEL, CentOS, Fedora)
- Root privileges (for systemd service management)
- Network connectivity within the restricted network

### Freenet Node Mode (Windows/macOS/Linux)
- Two network interfaces:
  - One connected to free/unrestricted internet
  - One connected to the restricted network
- Both interfaces must be active simultaneously

## Installation

Download the pre-built binary for your platform from the [releases page](https://github.com/nethopper/nethopper/releases).

Or build from source:

```bash
git clone https://github.com/nethopper/nethopper.git
cd nethopper
make build
```

## Quick Start

### Step 1: Set Up the Server (in restricted network)

```bash
# On a Linux server in the restricted network
sudo ./nethopper server install

# This will:
# - Extract sing-box binary
# - Generate configuration with random ports/secrets
# - Create systemd service
# - Add firewall rules (if ufw/firewalld detected)
# - Display a connection string for the freenet node
```

Copy the displayed connection string (starts with `nh://...`).

### Step 2: Set Up the Freenet Node

```bash
# On a device with two network interfaces
./nethopper freenet configure --connection "nh://..."

# Or run interactively:
./nethopper freenet configure
# Follow the prompts to select interfaces
```

### Step 3: Run the Freenet Node

```bash
./nethopper freenet run
```

### Step 4: Configure Telegram Clients

On devices in the restricted network, configure Telegram to use the MTProto proxy:
- Server: `<server-ip>`
- Port: `8443` (or the MTProto port shown in server status)
- Secret: `<displayed in server status>`

## Usage

### Interactive Mode

Simply run `nethopper` without arguments to enter the interactive menu:

```bash
./nethopper
```

### CLI Mode

#### Server Commands

```bash
# Install server (auto-select ports)
sudo nethopper server install

# Install with specific ports
sudo nethopper server install --port 2083 --mtproto-port 8443 --vmess-port 8081

# Modify configuration
sudo nethopper server configure --fallback example.com

# View status
sudo nethopper server status
sudo nethopper server status --json
sudo nethopper server status --logs

# Get connection string
sudo nethopper server connection-string

# Uninstall
sudo nethopper server uninstall
```

#### Freenet Commands

```bash
# Configure with connection string (recommended)
nethopper freenet configure --connection "nh://..."

# Configure manually
nethopper freenet configure \
  --server-ip 192.168.1.100 \
  --server-port 2083 \
  --vmess-port 8081 \
  --vmess-uuid "103b0aae-3384-4d23-9f5b-2d15be377a23" \
  --free-iface eth0 \
  --restricted-iface wlan0

# Run
nethopper freenet run

# View status
nethopper freenet status
nethopper freenet status --ping

# Uninstall
nethopper freenet uninstall
```

### Global Flags

```bash
--config <path>      Override config path
--verbose, -v        Verbose output
--no-color           Disable colored output
--non-interactive    Disable interactive prompts
```

## Configuration Reference

### Server Configuration Fields

| Field | Description | Default |
|-------|-------------|---------|
| `listen_port` | Reverse tunnel endpoint port | Auto-selected |
| `mtproto_port` | MTProto proxy port for Telegram clients | Auto-selected |
| `vmess_port` | Internal VMess communication port | Auto-selected |
| `mtproto_secret` | MTProto authentication secret (32 hex chars) | Generated |
| `vmess_uuid` | VMess authentication UUID | Generated |
| `fallback_host` | TLS camouflage domain | storage.googleapis.com |
| `multiplex_per_connection` | Streams per multiplexed connection | 50 |
| `log_level` | Logging verbosity | warn |

### Freenet Configuration Fields

| Field | Description | Required |
|-------|-------------|----------|
| `server_ip` | IP address of the server | Yes |
| `server_port` | Server's reverse endpoint port | Yes |
| `vmess_port` | Server's VMess port | Yes |
| `vmess_uuid` | VMess UUID (must match server) | Yes |
| `free_interface` | Network interface with free internet | Yes |
| `restricted_interface` | Network interface on restricted network | Yes |
| `max_connections` | Maximum multiplexed connections | 30 |
| `log_level` | Logging verbosity | warn |

## Connection String Format

The connection string encodes server configuration for easy transfer:

```
nh://<base64-encoded-json>

JSON structure:
{
  "v": 1,                    // Version
  "s": "192.168.1.100",      // Server IP
  "p": 2083,                 // Reverse endpoint port
  "vp": 8081,                // VMess port
  "u": "103b0aae-..."        // VMess UUID
}
```

## File Locations

### Server Mode (Linux)
- Binary: `/usr/local/bin/sing-box`
- Config: `/etc/nethopper/server.json`
- Service: `/etc/systemd/system/nethopper.service`

### Freenet Mode
| Platform | Binary | Config |
|----------|--------|--------|
| Linux | `~/.local/bin/sing-box` | `~/.config/nethopper/freenet.json` |
| macOS | `~/Library/Application Support/nethopper/bin/sing-box` | `~/Library/Application Support/nethopper/freenet.json` |
| Windows | `%APPDATA%\nethopper\bin\sing-box.exe` | `%APPDATA%\nethopper\freenet.json` |

## Troubleshooting

### Server Issues

**Service fails to start:**
```bash
# Check service status
sudo systemctl status nethopper

# View logs
sudo journalctl -u nethopper -f
```

**Port already in use:**
```bash
# Check which process is using the port
sudo ss -tlnp | grep <port>

# Reconfigure with different port
sudo nethopper server configure --port <new-port>
```

### Freenet Issues

**"Interface not found" error:**
```bash
# List available interfaces
ip link show  # Linux
ifconfig      # macOS
ipconfig      # Windows
```

**Cannot connect to server:**
1. Verify server is running: `sudo nethopper server status`
2. Check firewall allows the ports
3. Test connectivity: `nethopper freenet status --ping`

### Common Issues

**Configuration mismatch:**
Ensure the connection string was copied correctly, or manually verify:
- Server port matches `server_port` in freenet config
- VMess port matches `vmess_port` in freenet config
- VMess UUID matches exactly on both sides

## Building from Source

### Prerequisites
- Go 1.21 or later
- Make

### Build Commands

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Format code
make fmt
```

## Development: Building sing-box Locally

> **Note:** This section is for **development and testing only**. Production builds use CI/CD pipelines that automatically build and embed sing-box binaries for all platforms.

nethopper embeds the sing-box binary using Go's `//go:embed` directive. For local development, you can build sing-box from source using the provided scripts.

### Prerequisites
- [sing-box source code](https://github.com/SagerNet/sing-box) cloned locally
- Go 1.24+ (matching sing-box requirements)

### Build sing-box for Current Platform (Fastest)

```bash
# Default: expects sing-box source at ../sing-box-new
make build-singbox-local

# Or specify custom path
SINGBOX_SOURCE=/path/to/sing-box make build-singbox-local
```

This builds a single binary for your current OS/architecture (~59MB) and places it in `internal/binary/embedded/`.

### Build sing-box for All Platforms (Cross-compile)

```bash
make build-singbox
```

This cross-compiles sing-box for all supported platforms:
- `linux/amd64`, `linux/arm64`
- `darwin/amd64`, `darwin/arm64`
- `windows/amd64`

### Full Local Build

```bash
# Build sing-box for all platforms + build nethopper
make local

# Or for current platform only (faster)
make build-singbox-local && make build
```

### Verify Embedded Binary

After building, the nethopper binary will include the embedded sing-box:

```bash
./nethopper version
# Shows nethopper version

./nethopper freenet status
# Will show if sing-box binary is available
```

### Clean Embedded Binaries

```bash
make clean
# Removes nethopper binary and all embedded sing-box binaries
```

## License

MIT License - see LICENSE file for details.

## Acknowledgments

- [sing-box](https://github.com/SagerNet/sing-box) - The universal proxy platform
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [promptui](https://github.com/manifoldco/promptui) - Interactive prompts
