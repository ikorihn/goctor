package manifest

import (
	"errors"
	"fmt"
	"regexp"
)

// Manifest represents the complete configuration for tool requirements
type Manifest struct {
	Meta     ManifestMeta     `yaml:"meta" json:"meta"`
	Defaults ManifestDefaults `yaml:"defaults,omitempty" json:"defaults,omitempty"`
	Tools    []ToolDefinition `yaml:"tools" json:"tools"`
}

// ManifestMeta contains metadata about the manifest
type ManifestMeta struct {
	Version  int    `yaml:"version" json:"version"`
	Name     string `yaml:"name" json:"name"`
	Language string `yaml:"language,omitempty" json:"language,omitempty"`
}

// ManifestDefaults contains default values applied to tool definitions
type ManifestDefaults struct {
	TimeoutSeconds int    `yaml:"timeout_sec,omitempty" json:"timeout_sec,omitempty"`
	RegexKey       string `yaml:"regex_key,omitempty" json:"regex_key,omitempty"`
}

// Validate performs comprehensive validation of the manifest
func (m *Manifest) Validate() error {
	if err := m.Meta.Validate(); err != nil {
		return fmt.Errorf("meta validation failed: %v", err)
	}

	if err := m.Defaults.Validate(); err != nil {
		return fmt.Errorf("defaults validation failed: %v", err)
	}

	if len(m.Tools) == 0 {
		return errors.New("tools list cannot be empty")
	}

	// Check for duplicate tool IDs
	toolIDs := make(map[string]bool)
	for i, tool := range m.Tools {
		if toolIDs[tool.ID] {
			return fmt.Errorf("duplicate tool ID: %s", tool.ID)
		}
		toolIDs[tool.ID] = true

		// Validate each tool
		if err := tool.Validate(); err != nil {
			return fmt.Errorf("tool %d (%s) validation failed: %v", i, tool.ID, err)
		}
	}

	return nil
}

// ApplyDefaults applies default values to tools that don't have explicit values
func (m *Manifest) ApplyDefaults() {
	for i := range m.Tools {
		m.Tools[i].ApplyDefaults(m.Defaults)
	}
}

// Merge combines this manifest with another, with the other taking precedence
func (m *Manifest) Merge(other Manifest) Manifest {
	result := Manifest{
		Meta:     other.Meta, // Use the other's metadata
		Defaults: m.mergeDefaults(other.Defaults),
		Tools:    make([]ToolDefinition, 0),
	}

	// Create a map of tools from the other manifest
	otherTools := make(map[string]ToolDefinition)
	for _, tool := range other.Tools {
		otherTools[tool.ID] = tool
	}

	// Add all tools from other manifest first
	for _, tool := range other.Tools {
		result.Tools = append(result.Tools, tool)
	}

	// Add tools from this manifest that aren't in the other
	for _, tool := range m.Tools {
		if _, exists := otherTools[tool.ID]; !exists {
			result.Tools = append(result.Tools, tool)
		}
	}

	return result
}

// mergeDefaults combines defaults with other taking precedence
func (m *Manifest) mergeDefaults(other ManifestDefaults) ManifestDefaults {
	result := m.Defaults

	if other.TimeoutSeconds > 0 {
		result.TimeoutSeconds = other.TimeoutSeconds
	}

	if other.RegexKey != "" {
		result.RegexKey = other.RegexKey
	}

	return result
}

// GetTool returns a tool by ID, or nil if not found
func (m *Manifest) GetTool(id string) *ToolDefinition {
	for i := range m.Tools {
		if m.Tools[i].ID == id {
			return &m.Tools[i]
		}
	}
	return nil
}

// Validate performs validation of the manifest metadata
func (mm *ManifestMeta) Validate() error {
	if mm.Version != 1 {
		return fmt.Errorf("unsupported manifest version: %d", mm.Version)
	}

	if mm.Name == "" {
		return errors.New("name cannot be empty")
	}

	// Validate language code if specified
	if mm.Language != "" {
		if len(mm.Language) != 2 {
			return fmt.Errorf("invalid language code: %s", mm.Language)
		}

		// Basic validation for ISO 639-1 language codes (2 letters)
		validLangRegex := regexp.MustCompile(`^[a-z]{2}$`)
		if !validLangRegex.MatchString(mm.Language) {
			return fmt.Errorf("invalid language code format: %s", mm.Language)
		}
	}

	return nil
}

// Validate performs validation of the manifest defaults
func (md *ManifestDefaults) Validate() error {
	if md.TimeoutSeconds < 0 {
		return errors.New("timeout must be positive")
	}

	if md.TimeoutSeconds > 300 {
		return errors.New("timeout too large")
	}

	// No validation for empty regex key since it's optional

	return nil
}

// GetDefaultTimeout returns the default timeout or a system default
func (md *ManifestDefaults) GetDefaultTimeout() int {
	if md.TimeoutSeconds > 0 {
		return md.TimeoutSeconds
	}
	return 5 // Default system timeout in seconds
}

// GetDefaultRegexKey returns the default regex key or a system default
func (md *ManifestDefaults) GetDefaultRegexKey() string {
	if md.RegexKey != "" {
		return md.RegexKey
	}
	return "ver" // Default regex capture group name
}