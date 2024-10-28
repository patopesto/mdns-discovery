# mDNS Discovery

A small TUI app to discover mDNS/Zeroconf/Bonjour services and devices on your network.
Built in go and using [charm.sh](https://charm.sh/) libraries.


## Usage

```bash
./mdns-discovery <flags>
```

View all available configuration options

```bash
./mdns-discovery -h
```



## Development

### Running

```bash
go run . <flags>
```

### Building

```bash
go build
```


## References
- [mDNS wikipedia](https://en.wikipedia.org/wiki/Multicast_DNS) and [go library](https://github.com/hashicorp/mdns)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea): TUI Framework
- [Bubbles](https://github.com/charmbracelet/bubbles): TUI Components for Bubble Tea
- [Lipgloss](https://github.com/charmbracelet/lipgloss): Styling
- [Bubble-table](https://github.com/Evertras/bubble-table): Table componenent for Bubble Tea