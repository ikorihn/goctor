package manifest

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// CheckConfig represents the check configuration for a tool
type CheckConfig struct {
	Command []string `yaml:"cmd" json:"cmd"`
	Regex   string   `yaml:"regex" json:"regex"`
}

// ToolDefinition represents a development tool with its requirements and detection logic
type ToolDefinition struct {
	ID              string            `yaml:"id" json:"id"`
	Name            string            `yaml:"name" json:"name"`
	Rationale       string            `yaml:"rationale" json:"rationale"`
	RequiredVersion string            `yaml:"require" json:"require"`
	Check           CheckConfig       `yaml:"check" json:"check"`
	Links           map[string]string `yaml:"links" json:"links"`
	TimeoutSeconds  int               `yaml:"timeout_sec,omitempty" json:"timeout_seconds,omitempty"`
}

// CheckCommand returns the command to execute for version checking
func (td *ToolDefinition) CheckCommand() []string {
	return td.Check.Command
}

// VersionRegex returns the regex pattern for version extraction
func (td *ToolDefinition) VersionRegex() string {
	return td.Check.Regex
}

// Validate performs comprehensive validation of the tool definition
func (td *ToolDefinition) Validate() error {
	if err := td.validateRequiredFields(); err != nil {
		return err
	}

	if err := td.validateID(); err != nil {
		return err
	}

	if err := td.ValidateVersionConstraint(); err != nil {
		return err
	}

	if err := td.ValidateRegex(); err != nil {
		return err
	}

	if err := td.ValidateLinks(); err != nil {
		return err
	}

	if err := td.validateTimeout(); err != nil {
		return err
	}

	return nil
}

// validateRequiredFields checks that all required fields are not empty
func (td *ToolDefinition) validateRequiredFields() error {
	if td.ID == "" || td.Name == "" || td.Rationale == "" || td.RequiredVersion == "" ||
	   len(td.Check.Command) == 0 || td.Check.Regex == "" || len(td.Links) == 0 {
		return errors.New("required fields cannot be empty")
	}
	return nil
}

// validateID checks that the ID follows the required format
func (td *ToolDefinition) validateID() error {
	if td.ID == "" {
		return errors.New("ID cannot be empty")
	}

	// ID must be lowercase alphanumeric with hyphens only
	validIDRegex := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !validIDRegex.MatchString(td.ID) {
		return errors.New("ID must be lowercase alphanumeric with hyphens")
	}

	return nil
}

// ValidateVersionConstraint validates the semantic version constraint
func (td *ToolDefinition) ValidateVersionConstraint() error {
	if td.RequiredVersion == "" {
		return errors.New("version constraint cannot be empty")
	}

	// Basic validation for common semver constraint patterns
	// This is a simplified validation - full semver parsing happens in the semver package
	validPatterns := []string{
		`^\d+(\.\d+)*$`,                           // 1.2.3
		`^>=\d+(\.\d+)*$`,                        // >=1.2.3
		`^>\d+(\.\d+)*$`,                         // >1.2.3
		`^<=\d+(\.\d+)*$`,                        // <=1.2.3
		`^<\d+(\.\d+)*$`,                         // <1.2.3
		`^~\d+(\.\d+)*$`,                         // ~1.2.3
		`^\^\d+(\.\d+)*$`,                        // ^1.2.3
		`^>=\d+(\.\d+)* <\d+(\.\d+)*$`,          // >=1.2 <1.3
	}

	for _, pattern := range validPatterns {
		matched, _ := regexp.MatchString(pattern, td.RequiredVersion)
		if matched {
			return nil
		}
	}

	return fmt.Errorf("invalid version constraint format: %s", td.RequiredVersion)
}

// ValidateRegex validates the version extraction regular expression
func (td *ToolDefinition) ValidateRegex() error {
	if td.Check.Regex == "" {
		return errors.New("empty regex")
	}

	// Check if regex is valid
	_, err := regexp.Compile(td.Check.Regex)
	if err != nil {
		return fmt.Errorf("malformed regex: %v", err)
	}

	// Check if regex contains named capture group
	if !strings.Contains(td.Check.Regex, "(?P<") && !strings.Contains(td.Check.Regex, "(?<") {
		return errors.New("VersionRegex must contain named capture group")
	}

	return nil
}

// ValidateLinks validates the URLs in the links map
func (td *ToolDefinition) ValidateLinks() error {
	if len(td.Links) == 0 {
		return errors.New("links cannot be empty")
	}

	for linkType, linkURL := range td.Links {
		if linkURL == "" {
			return fmt.Errorf("link %s cannot be empty", linkType)
		}

		// Validate URL format
		if !isValidURL(linkURL) {
			return fmt.Errorf("invalid URL for %s: %s", linkType, linkURL)
		}
	}

	return nil
}

// validateTimeout checks that timeout is positive if specified
func (td *ToolDefinition) validateTimeout() error {
	if td.TimeoutSeconds < 0 {
		return errors.New("TimeoutSeconds must be positive")
	}
	return nil
}

// isValidURL performs basic URL validation
func isValidURL(urlStr string) bool {
	if urlStr == "" {
		return false
	}

	// Basic check for http/https
	if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") {
		// Use Go's url.Parse for more thorough validation
		_, err := url.Parse(urlStr)
		return err == nil
	}

	return false
}

// ApplyDefaults applies default values from manifest defaults
func (td *ToolDefinition) ApplyDefaults(defaults ManifestDefaults) {
	if td.TimeoutSeconds == 0 && defaults.TimeoutSeconds > 0 {
		td.TimeoutSeconds = defaults.TimeoutSeconds
	}

	// If the regex uses the default capture group name, no change needed
	// This is handled during parsing where the regex key can be used
}