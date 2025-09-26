# goctor

A development environment checker tool that validates required tools and their versions against a YAML manifest.

## Features

- **Environment Checking**: Verify that all required development tools are installed and meet version requirements
- **Tool Listing**: List all tools defined in a manifest file
- **Multiple Output Formats**: Human-readable and JSON output formats
- **Flexible Manifest Sources**: Load manifests from local files or remote URLs
- **Cross-Platform Support**: Works on macOS, Linux, and Windows

## Installation

### Build from Source

```bash
go build -o ./bin/goctor ./cmd/goctor
```

The binary will be created in the `./bin` directory.

## Usage

### Basic Commands

```bash
# Check environment using default manifest (./tools.yaml)
goctor

# Check environment using specific manifest
goctor -f custom-manifest.yaml

# Check environment with JSON output
goctor --json

# List tools defined in manifest
goctor list

# List tools from remote manifest
goctor list -f https://company.com/manifest.yaml

# Show version
goctor -v

# Show help
goctor -h
```

### Commands

- `doctor` (default): Check development environment against manifest
- `list`: List tools defined in manifest

### Flags

- `-f, --manifest PATH_OR_URL`: Manifest file path or URL (default: "./tools.yaml")
- `--json`: Output results in JSON format
- `-h, --help`: Show help information
- `-v, --version`: Show version information

## Manifest Format

The tool uses YAML manifests to define required tools and their versions:

```yaml
meta:
  version: 1
  name: "Project Development Tools"
  language: "en"

defaults:
  timeout_sec: 5
  regex_key: "ver"

tools:
  - id: go
    name: "Go"
    rationale: "Go development toolchain"
    require: ">=1.20"
    check:
      cmd: ["go", "version"]
      regex: "go(?P<ver>\\d+\\.\\d+(\\.\\d+)?)"
    links:
      homepage: "https://go.dev/"
      download: "https://go.dev/dl/"
```

### Manifest Schema

- `meta`: Manifest metadata
  - `version`: Schema version
  - `name`: Manifest name
  - `language`: Language code
- `defaults`: Default settings for all tools
  - `timeout_sec`: Default command timeout
  - `regex_key`: Default regex capture group name
- `tools`: Array of tool definitions
  - `id`: Unique tool identifier
  - `name`: Human-readable tool name
  - `rationale`: Why this tool is required
  - `require`: Version requirement (semver format)
  - `check`: How to check if tool is installed
    - `cmd`: Command to run
    - `regex`: Regex to extract version from output
  - `timeout_sec`: Optional override for command timeout
  - `links`: Helpful links for the tool

## Exit Codes

- `0`: All tools meet requirements
- `1`: One or more tools missing or don't meet version requirements

## Examples

### Check Development Environment

```bash
$ goctor
✓ Go 1.21.3 (>=1.20 required)
✓ Git 2.42.0 (>=2.30 required)
✗ Docker not found (>=24.0 required)
✓ Node.js 18.17.0 (>=18.0 required)

Environment check failed: 1 tool(s) missing or outdated
```

### JSON Output

```bash
$ goctor --json
{
  "platform": {
    "os": "darwin",
    "arch": "arm64"
  },
  "manifest_source": "./tools.yaml",
  "results": [
    {
      "tool_id": "go",
      "status": "ok",
      "installed_version": "1.21.3",
      "required_version": ">=1.20"
    }
  ],
  "summary": {
    "total_tools": 4,
    "passed": 3,
    "failed": 1
  }
}
```

### List Tools

```bash
$ goctor list
Tools defined in ./tools.yaml:

go - Go (>=1.20)
  Go development toolchain for building and testing

git - Git (>=2.30)
  Version control system for source code management

docker - Docker (>=24.0)
  Container runtime for development and deployment

node - Node.js (>=18.0)
  JavaScript runtime for frontend build tools
```

## Development

### Requirements

- Go 1.22 or later
- Standard library only (no external dependencies except YAML parsing)

### Project Structure

```
cmd/goctor/          # Main application entry point
internal/            # Internal packages
├── checker/         # Tool checking logic
├── manifest/        # Manifest loading and parsing
├── output/          # Output formatting
└── platform/        # Platform detection
testdata/           # Test data files
tests/              # Test files
tools.yaml          # Default manifest
```

### Building

```bash
go build -o ./bin/goctor ./cmd/goctor
```

### Testing

```bash
go test ./...
```

## License

[License information would go here]