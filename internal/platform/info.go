package platform

import (
	"os"
	"runtime"
	"strings"
)

// PlatformInfo contains information about the current platform
type PlatformInfo struct {
	OS           string `json:"os"`
	Architecture string `json:"arch"`
	Hostname     string `json:"hostname,omitempty"`
}

// CheckSummary provides statistical summary (duplicate here for package independence)
type CheckSummary struct {
	Total    int `json:"total"`
	OK       int `json:"ok"`
	Missing  int `json:"missing"`
	Outdated int `json:"outdated"`
	Errors   int `json:"errors"`
}

// DetectPlatform detects the current platform information
func DetectPlatform() PlatformInfo {
	platform := PlatformInfo{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
	}

	// Try to get hostname, but don't fail if we can't
	if hostname, err := os.Hostname(); err == nil {
		platform.Hostname = hostname
	}

	return platform
}

// IsSupported returns true if the platform is supported
func (pi *PlatformInfo) IsSupported() bool {
	supportedOS := map[string]bool{
		"darwin": true,
		"linux":  true,
	}

	supportedArch := map[string]bool{
		"amd64": true,
		"arm64": true,
		"386":   true,
	}

	return supportedOS[pi.OS] && supportedArch[pi.Architecture]
}

// String returns a human-readable representation of the platform
func (pi *PlatformInfo) String() string {
	if pi.Hostname != "" {
		return pi.OS + "/" + pi.Architecture + " (" + pi.Hostname + ")"
	}
	return pi.OS + "/" + pi.Architecture
}

// GetPlatformSpecificCommands returns platform-specific variations of commands
func (pi *PlatformInfo) GetPlatformSpecificCommands(baseCommand []string) []string {
	// For most tools, commands are the same across platforms
	// This could be extended for platform-specific tool variations
	return baseCommand
}

// ShouldUseShell returns whether commands should be executed through shell
func (pi *PlatformInfo) ShouldUseShell() bool {
	// We explicitly don't use shell for security reasons
	// Direct command execution is preferred
	return false
}

// GetPathSeparator returns the path separator for this platform
func (pi *PlatformInfo) GetPathSeparator() string {
	if pi.OS == "windows" {
		return "\\"
	}
	return "/"
}

// GetExecutableExtension returns the executable extension for this platform
func (pi *PlatformInfo) GetExecutableExtension() string {
	if pi.OS == "windows" {
		return ".exe"
	}
	return ""
}

// IsMacOS returns true if the platform is macOS
func (pi *PlatformInfo) IsMacOS() bool {
	return pi.OS == "darwin"
}

// IsLinux returns true if the platform is Linux
func (pi *PlatformInfo) IsLinux() bool {
	return pi.OS == "linux"
}

// IsARM returns true if the architecture is ARM-based
func (pi *PlatformInfo) IsARM() bool {
	return strings.HasPrefix(pi.Architecture, "arm")
}

// IsIntel returns true if the architecture is Intel/AMD x86-based
func (pi *PlatformInfo) IsIntel() bool {
	return pi.Architecture == "amd64" || pi.Architecture == "386"
}

// GetPackageManager returns the likely package manager for this platform
func (pi *PlatformInfo) GetPackageManager() string {
	switch pi.OS {
	case "darwin":
		return "homebrew"
	case "linux":
		// This is a simplified detection - in reality we'd check for specific distros
		return "apt" // Default to apt for Linux
	default:
		return "unknown"
	}
}

// GetShellCommands returns platform-specific shell commands
func (pi *PlatformInfo) GetShellCommands() map[string]string {
	commands := map[string]string{
		"which": "which",
		"ls":    "ls",
		"cat":   "cat",
	}

	// Platform-specific overrides could go here
	// For now, Unix commands work on both macOS and Linux

	return commands
}

// Validate performs validation of platform information
func (pi *PlatformInfo) Validate() error {
	supportedOS := map[string]bool{
		"darwin": true,
		"linux":  true,
	}

	if !supportedOS[pi.OS] {
		return &PlatformError{
			Message: "unsupported operating system: " + pi.OS,
			Code:    ErrorUnsupportedOS,
		}
	}

	supportedArch := map[string]bool{
		"amd64": true,
		"arm64": true,
		"386":   true,
	}

	if !supportedArch[pi.Architecture] {
		return &PlatformError{
			Message: "unsupported architecture: " + pi.Architecture,
			Code:    ErrorUnsupportedArch,
		}
	}

	return nil
}

// PlatformError represents platform-specific errors
type PlatformError struct {
	Message string
	Code    ErrorCode
}

func (pe *PlatformError) Error() string {
	return pe.Message
}

// ErrorCode represents platform error types
type ErrorCode int

const (
	ErrorUnsupportedOS ErrorCode = iota
	ErrorUnsupportedArch
	ErrorDetectionFailed
)

// DetectDistribution attempts to detect the Linux distribution
func (pi *PlatformInfo) DetectDistribution() string {
	if !pi.IsLinux() {
		return ""
	}

	// Try to read /etc/os-release
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		content := string(data)

		// Look for ID= line
		for _, line := range strings.Split(content, "\n") {
			if strings.HasPrefix(line, "ID=") {
				id := strings.TrimPrefix(line, "ID=")
				id = strings.Trim(id, "\"")
				return id
			}
		}
	}

	// Fallback checks for common distributions
	if _, err := os.Stat("/etc/debian_version"); err == nil {
		return "debian"
	}

	if _, err := os.Stat("/etc/redhat-release"); err == nil {
		return "redhat"
	}

	return "unknown"
}

// GetPreferredPackageManager returns the preferred package manager for the detected platform
func (pi *PlatformInfo) GetPreferredPackageManager() string {
	if pi.IsMacOS() {
		return "brew"
	}

	if pi.IsLinux() {
		distro := pi.DetectDistribution()
		switch distro {
		case "ubuntu", "debian":
			return "apt"
		case "fedora", "centos", "rhel":
			return "yum"
		case "arch":
			return "pacman"
		default:
			return "apt" // Default fallback
		}
	}

	return "unknown"
}