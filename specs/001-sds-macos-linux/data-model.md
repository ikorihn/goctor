# Data Model: Development Environment Checker

## Core Entities

### Tool Definition
Represents a development tool with its requirements and detection logic.

**Fields**:
- `ID` (string): Unique identifier for the tool (e.g., "go", "docker")
- `Name` (string): Human-readable display name (e.g., "Go", "Docker")
- `Rationale` (string): Explanation of why this tool is required
- `RequiredVersion` (string): Semantic version constraint (e.g., ">=1.22 <1.25")
- `CheckCommand` ([]string): Command and arguments to execute for version detection
- `VersionRegex` (string): Regular expression with named capture group to extract version
- `Links` (map[string]string): Official URLs (homepage, download, docs, github)
- `TimeoutSeconds` (int): Maximum time to wait for command execution

**Validation Rules**:
- ID must be lowercase alphanumeric with hyphens
- RequiredVersion must be valid semver constraint syntax
- VersionRegex must contain named capture group `(?P<ver>...)`
- CheckCommand must be non-empty array
- TimeoutSeconds must be positive integer

### Manifest
Configuration defining the complete set of required tools for an environment.

**Fields**:
- `Meta` (ManifestMeta): Metadata about the manifest
- `Defaults` (ManifestDefaults): Default values for tool definitions
- `Tools` ([]ToolDefinition): List of required tools

**Validation Rules**:
- Tools list must be non-empty
- Tool IDs must be unique within manifest
- Meta version must be supported version number

### ManifestMeta
Metadata section of a manifest file.

**Fields**:
- `Version` (int): Manifest schema version
- `Name` (string): Human-readable manifest name
- `Language` (string): Localization language code (e.g., "ja", "en")

### ManifestDefaults
Default values applied to tool definitions when not specified.

**Fields**:
- `TimeoutSeconds` (int): Default command timeout
- `RegexKey` (string): Default named capture group name for version extraction

### CheckResult
The outcome of verifying a single tool installation.

**Fields**:
- `ToolID` (string): References Tool Definition ID
- `ToolName` (string): Human-readable tool name
- `Status` (CheckStatus): Current status (OK, Missing, Outdated, Error)
- `RequiredVersion` (string): Version constraint from manifest
- `InstalledVersion` (string): Detected version (null if not found)
- `Rationale` (string): Why this tool is required
- `Links` (map[string]string): Help links for installation/upgrade
- `Errors` ([]string): Any error messages during checking
- `CheckDuration` (time.Duration): Time taken to perform the check

**State Transitions**:
- Initial → Missing (tool command not found)
- Initial → Error (command execution failed)
- Initial → Outdated (version found but doesn't meet requirements)
- Initial → OK (version found and meets requirements)

### CheckStatus
Enumeration of possible tool verification states.

**Values**:
- `StatusOK`: Tool is installed and meets version requirements
- `StatusMissing`: Tool is not installed or command not found
- `StatusOutdated`: Tool is installed but version doesn't meet requirements
- `StatusError`: Error occurred during tool detection

### EnvironmentReport
Comprehensive summary of all tool checks for a development environment.

**Fields**:
- `SchemaVersion` (int): Output format version
- `Platform` (PlatformInfo): Information about the target platform
- `Summary` (CheckSummary): Aggregate statistics
- `ManifestSource` (string): Path or URL of manifest used
- `Items` ([]CheckResult): Individual tool check results
- `GeneratedAt` (time.Time): When the report was created

### PlatformInfo
Information about the platform where checks were performed.

**Fields**:
- `OS` (string): Operating system (darwin, linux)
- `Architecture` (string): CPU architecture (amd64, arm64)
- `Hostname` (string): Machine hostname for identification

### CheckSummary
Statistical summary of tool verification results.

**Fields**:
- `Total` (int): Total number of tools checked
- `OK` (int): Number of tools meeting requirements
- `Missing` (int): Number of tools not found
- `Outdated` (int): Number of tools with incompatible versions
- `Errors` (int): Number of tools with check errors

## Relationships

```
Manifest (1) → (many) ToolDefinition
ToolDefinition (1) → (1) CheckResult
EnvironmentReport (1) → (many) CheckResult
EnvironmentReport (1) → (1) PlatformInfo
EnvironmentReport (1) → (1) CheckSummary
```

## Data Flow

1. **Manifest Loading**: Parse YAML → Manifest with ToolDefinitions
2. **Tool Checking**: For each ToolDefinition → execute CheckCommand → CheckResult
3. **Report Generation**: Aggregate CheckResults → EnvironmentReport
4. **Output Formatting**: EnvironmentReport → Human/JSON output

## Storage Considerations

- No persistent storage required
- Manifest files stored as YAML (local files or HTTP URLs)
- Runtime data kept in memory during execution
- Output written to stdout/stderr
- No caching of tool check results between runs