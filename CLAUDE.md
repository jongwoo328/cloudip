# CLAUDE.md

## Build

- **Version injection**: Binary version is set via ldflags at build time: `go build -o build/cloudip --ldflags '-X cloudip/cmd.Version=0.7.0'`
- **Release builds**: `make build` uses GoReleaser for cross-platform compilation
