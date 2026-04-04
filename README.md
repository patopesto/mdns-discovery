# mDNS Discovery

![Version](https://img.shields.io/gitlab/v/tag/patopest%2Fmdns-discovery?style=for-the-badge&label=Latest)
![License](https://img.shields.io/badge/license-MIT-blue?style=for-the-badge)
![Platforms](https://img.shields.io/badge/platforms-macOS%20%7C%20Windows%20%7C%20Linux-blue?style=for-the-badge)

> A TUI for discovering mDNS/Zeroconf/Bonjour services and devices on your local network. Built with Go and the beautiful [charm.sh](https://charm.sh/) libraries.

![Demo](assets/demo.gif)

## 🌟 Features

- **Real-time Discovery**: Automatically discovers mDNS services on your network
- **Filtering & Sorting**: Search services and sort by any column (Name, Service, Domain, IP, Port, etc.)
- **Interface Management**: Toggle network interfaces on/off dynamically
- **Service Details**: View complete service information including TXT records
- **Beautiful TUI**: Built with Bubble Tea for a polished terminal experience
- **Cross-Platform**: Works on macOS, Linux, and Windows

---

## 📦 Installation

### Download Binary

Precompiled binaries are available on the [Releases](https://gitlab.com/patopest/mdns-discovery/-/releases) page.

### Homebrew (macOS/Linux)

```bash
brew install patopesto/tap/mdns-discovery
```

### Build from Source

```bash
# Clone the repository
git clone https://gitlab.com/patopest/mdns-discovery.git
cd mdns-discovery

# Build
go build

# Or install directly
go install
```

## ⌨️ Usage

Launch the application:

```bash
mdns-discovery
```

### Command-Line Options

```
Flags:
  -d, --domain strings    Domain(s) to use (default: local)
  -i, --interface strings Use specified interface(s), e.g., '-i eth0,wlan0' (default: all interfaces)
  -v, --version           Version for mdns-discovery
  -h, --help              Help for mdns-discovery
```

### Environment Variables

All flags can also be set via environment variables with the `MDNS_` prefix:

```bash
export MDNS_INTERFACE=eth0,wlan0
export MDNS_DOMAIN=local
mdns-discovery
```

### Keyboard Shortcuts

#### General

| Key | Action |
|-----|--------|
| `?` | Toggle help |
| `s` | Open settings (interface selection) |
| `q` / `ctrl+c` | Quit |

#### Navigation

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |

#### Filtering & Sorting

| Key | Action |
|-----|--------|
| `/` | Focus filter input |
| `esc` | Clear filter / close modal |
| `enter` / `space` | View service details |
| `1` | Sort by hostname |
| `2` | Sort by service |
| `3` | Sort by domain |
| `4` | Sort by hostname |
| `5` | Sort by IP |
| `6` | Sort by port |

---

## 🚀 Development

### Running Locally

```bash
go run . [flags]
```

### Building

```bash
go build
```

---

## References

- [mDNS Wikipedia](https://en.wikipedia.org/wiki/Multicast_DNS)
- [DNS Service Discovery](https://github.com/libp2p/specs/blob/master/discovery/mdns.md)
- [Multicast DNS RFC](https://datatracker.ietf.org/doc/html/rfc6762)
- [DNS-Based Service Discovery RFC](https://datatracker.ietf.org/doc/html/rfc6763)

## License

MIT License - see [LICENSE](./LICENSE) for details
