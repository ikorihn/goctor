package checker

import (
	"context"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/company/doctor-cmd/internal/manifest"
	"github.com/company/doctor-cmd/internal/platform"
	"github.com/company/doctor-cmd/internal/semver"
)

// Checker handles tool detection and version checking
type Checker struct {
	commandTimeout time.Duration
}

// NewChecker creates a new tool checker with default configuration
func NewChecker() *Checker {
	return &Checker{
		commandTimeout: 5 * time.Second,
	}
}

// CheckTool performs a complete check of a tool including detection and version validation
func (c *Checker) CheckTool(tool manifest.ToolDefinition, platformInfo platform.PlatformInfo) CheckResult {
	result := CheckResult{
		ToolID:          tool.ID,
		ToolName:        tool.Name,
		RequiredVersion: tool.RequiredVersion,
		ActualVersion:   "",
		CommandPath:     "",
		Status:          StatusNotFound,
		ErrorMessage:    "",
		Links:           tool.Links,
		Platform:        platformInfo.String(),
	}

	// Check if tool is available and get its path
	commandPath, available, err := c.getToolPath(tool.CheckCommand()[0])
	if err != nil || !available {
		result.Status = StatusNotFound
		if err != nil {
			result.ErrorMessage = err.Error()
		} else {
			result.ErrorMessage = "Command not found"
		}
		return result
	}

	result.CommandPath = commandPath

	// Extract version from command output
	version, err := c.extractVersion(tool)
	if err != nil {
		result.Status = StatusError
		result.ErrorMessage = err.Error()
		return result
	}

	result.ActualVersion = version

	// Parse and validate version against requirements
	if err := c.validateVersion(version, tool.RequiredVersion); err != nil {
		result.Status = StatusOutdated
		result.ErrorMessage = err.Error()
	} else {
		result.Status = StatusOK
	}

	// Determine final status
	result.DetermineStatus()

	return result
}

// getToolPath checks if a command is available and returns its path
func (c *Checker) getToolPath(command string) (string, bool, error) {
	// Use `which` command to check if tool exists and get its path
	ctx, cancel := context.WithTimeout(context.Background(), c.commandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "command", "-v", command)
	output, err := cmd.Output()

	if err != nil {
		// Check if it's a timeout or other error
		if ctx.Err() == context.DeadlineExceeded {
			return "", false, ctx.Err()
		}
		// Command not found is expected for missing tools
		return "", false, nil
	}

	path := strings.TrimSpace(string(output))
	return path, true, nil
}

// extractVersion runs the tool's check command and extracts version using regex
func (c *Checker) extractVersion(tool manifest.ToolDefinition) (string, error) {
	if len(tool.CheckCommand()) == 0 {
		return "", NewCheckError("no check command specified", ErrorTypeConfiguration)
	}

	// Execute the version check command
	output, err := c.runCommand(tool.CheckCommand(), tool.TimeoutSeconds)
	if err != nil {
		return "", NewCheckError("failed to run version command: "+err.Error(), ErrorTypeExecution)
	}

	// Extract version using regex
	version, err := c.parseVersionFromOutput(output, tool.VersionRegex())
	if err != nil {
		return "", NewCheckError("failed to parse version: "+err.Error(), ErrorTypeParsing)
	}

	return version, nil
}

// runCommand executes a command with timeout and returns its output
func (c *Checker) runCommand(command []string, timeoutSec int) (string, error) {
	timeout := c.commandTimeout
	if timeoutSec > 0 {
		timeout = time.Duration(timeoutSec) * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", NewCheckError("command timed out", ErrorTypeTimeout)
		}
		return "", NewCheckError("command failed: "+err.Error(), ErrorTypeExecution)
	}

	return string(output), nil
}

// parseVersionFromOutput extracts version string using regex with named capture groups
func (c *Checker) parseVersionFromOutput(output, regexPattern string) (string, error) {
	if regexPattern == "" {
		return "", NewCheckError("empty regex pattern", ErrorTypeConfiguration)
	}

	// Compile regex
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return "", NewCheckError("invalid regex: "+err.Error(), ErrorTypeConfiguration)
	}

	// Find matches
	matches := regex.FindStringSubmatch(output)
	if matches == nil {
		return "", NewCheckError("no version found in output", ErrorTypeParsing)
	}

	// Get subexp names to find named capture groups
	names := regex.SubexpNames()

	// Look for common capture group names
	versionGroupNames := []string{"ver", "version", "v"}

	for i, name := range names {
		if name != "" && i < len(matches) {
			// Check if this is a version-related capture group
			lowerName := strings.ToLower(name)
			for _, versionName := range versionGroupNames {
				if lowerName == versionName {
					if matches[i] != "" {
						return strings.TrimSpace(matches[i]), nil
					}
				}
			}
		}
	}

	// If no named group found, try the first capture group
	if len(matches) > 1 && matches[1] != "" {
		return strings.TrimSpace(matches[1]), nil
	}

	return "", NewCheckError("no version captured by regex", ErrorTypeParsing)
}

// validateVersion checks if the actual version satisfies the required version constraint
func (c *Checker) validateVersion(actualVersion, requiredVersion string) error {
	if actualVersion == "" {
		return NewCheckError("no actual version to validate", ErrorTypeParsing)
	}

	if requiredVersion == "" {
		return NewCheckError("no required version specified", ErrorTypeConfiguration)
	}

	// Parse the actual version
	actual, err := semver.ParseVersion(actualVersion)
	if err != nil {
		return NewCheckError("invalid actual version format: "+err.Error(), ErrorTypeParsing)
	}

	// Parse the required version constraint
	constraint, err := semver.ParseConstraint(requiredVersion)
	if err != nil {
		return NewCheckError("invalid required version constraint: "+err.Error(), ErrorTypeConfiguration)
	}

	// Check if actual version satisfies constraint
	if !constraint.IsSatisfiedBy(actual) {
		return NewCheckError("version does not satisfy constraint", ErrorTypeVersionMismatch)
	}

	return nil
}

// SetTimeout sets the default command timeout
func (c *Checker) SetTimeout(timeout time.Duration) {
	c.commandTimeout = timeout
}

// CheckMultipleTools runs checks for multiple tools concurrently
func (c *Checker) CheckMultipleTools(tools []manifest.ToolDefinition, platformInfo platform.PlatformInfo) []CheckResult {
	results := make([]CheckResult, len(tools))

	// For now, run sequentially - concurrency will be added in Phase 3.4
	for i, tool := range tools {
		results[i] = c.CheckTool(tool, platformInfo)
	}

	return results
}

