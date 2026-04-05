# CLAUDE.md

## Build

- **Version injection**: Binary version is set via ldflags at build time: `go build -o build/cloudip --ldflags '-X cloudip/cmd.Version=0.8.0'`
- **Release builds**: `make build` uses GoReleaser for cross-platform compilation
- **Version bump**: Update version string in these locations: `CLAUDE.md`, `README.md`, `docs/README_ko.md`

## Architecture

- **Provider initialization**: Providers have no `init()` — all setup (metadata, data loading, CIDR tree) happens lazily via `Initialize()`. Metadata is prepared in `EnsureDataFile()`, data loading in `loadFunc`. `checkCloudIp()` calls `Initialize()` before each IP check, which guarantees setup. New providers must follow this pattern.
