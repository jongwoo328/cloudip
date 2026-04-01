# CLAUDE.md

## Build

- **Version injection**: Binary version is set via ldflags at build time: `go build -o build/cloudip --ldflags '-X cloudip/cmd.Version=0.7.4'`
- **Release builds**: `make build` uses GoReleaser for cross-platform compilation

## Architecture

- **Provider initialization**: Each provider's data loading and metadata setup runs inside `BaseProvider.Initialize()`, not in `init()`. New providers must follow this pattern — `checkCloudIp()` calls `Initialize()` before each IP check, which guarantees setup.
