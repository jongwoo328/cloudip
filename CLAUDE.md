# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

- **Build**: `make build` (uses goreleaser for snapshot builds)
- **Direct Go build**: `go build -o dist --ldflags '-X cloudip/cmd.Version=0.6.0'`
- **Test**: `go test ./...` (tests available in util/ package)
- **Tidy dependencies**: `go mod tidy`
- **Run locally**: `go run main.go [IP_ADDRESSES]`

## Project Architecture

This is a Go CLI tool that identifies which cloud provider (AWS, GCP, Azure) owns given IP addresses.

### Core Architecture
- **Entry Point**: `main.go` initializes app directory and executes CLI commands
- **CLI Framework**: Uses Cobra for command-line interface (`cmd/` package)
- **IP Checking Logic**: Modular cloud provider detection in `ip/` package
- **Provider Interface**: Common interface and base implementation in `ip/provider.go`
- **Provider Modules**: Each cloud provider has its own subdirectory (`ip/aws/`, `ip/gcp/`, `ip/azure/`)
- **Utilities**: Common utilities in `util/` package including CIDR tree implementation

### Key Components
- `cmd/root.go`: Main CLI command definition with flags (format, header, delimiter, verbose)
- `ip/check.go`: Core IP checking orchestrator that iterates through providers
- `ip/provider.go`: Defines CloudProvider interface and BaseProvider implementation
- `ip/{provider}/provider.go`: Each provider implements CloudProvider interface using BaseProvider
- `ip/common.go`: Provider order configuration
- `common/`: Shared types and constants (CheckIpResult, Result structs, provider constants)
- `util/cidr_tree.go`: CIDR range matching data structure for efficient IP lookups

### Data Flow
1. CLI parses IP addresses and flags
2. `ip.CheckIp()` processes each IP sequentially
3. For each IP, `checkCloudIp()` checks providers in order (AWS, GCP, Azure)
4. Each provider initializes its IP ranges using BaseProvider and performs CIDR matching
5. Results formatted according to output flags (text/table/json)

### Provider Implementation Pattern
Each cloud provider follows the CloudProvider interface:
- `Initialize() error`: Downloads/loads IP ranges and builds CIDR trees
- `CheckIP(ip string) (bool, error)`: Checks if IP belongs to provider using CIDR tree matching
- `GetName() string`: Returns provider name
- Uses composition with BaseProvider for common functionality (CIDR trees, initialization logic)
- Implements DataManager interface for provider-specific data handling

## Module Information
- **Go Version**: 1.23.2  
- **Key Dependencies**: cobra (CLI), tablewriter (output formatting), goquery (HTML parsing)
- **Build Tool**: GoReleaser for cross-platform releases