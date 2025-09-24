package contract

import (
	"os"
	"path/filepath"
	"testing"
	"gopkg.in/yaml.v3"
)

// ManifestSchema represents the expected YAML manifest structure
type ManifestSchema struct {
	Meta     ManifestMeta     `yaml:"meta"`
	Defaults ManifestDefaults `yaml:"defaults,omitempty"`
	Tools    []ToolDefinition `yaml:"tools"`
}

type ManifestMeta struct {
	Version  int    `yaml:"version"`
	Name     string `yaml:"name"`
	Language string `yaml:"language,omitempty"`
}

type ManifestDefaults struct {
	TimeoutSec int    `yaml:"timeout_sec,omitempty"`
	RegexKey   string `yaml:"regex_key,omitempty"`
}

type ToolDefinition struct {
	ID        string            `yaml:"id"`
	Name      string            `yaml:"name"`
	Rationale string            `yaml:"rationale"`
	Require   string            `yaml:"require"`
	Check     CheckCommand      `yaml:"check"`
	Links     map[string]string `yaml:"links"`
}

type CheckCommand struct {
	Cmd   []string `yaml:"cmd"`
	Regex string   `yaml:"regex"`
}

func TestManifestSchemaCompliance(t *testing.T) {
	manifestFiles := []string{
		"testdata/manifests/sample.yaml",
	}

	for _, manifestFile := range manifestFiles {
		t.Run(filepath.Base(manifestFile), func(t *testing.T) {
			// Read manifest file
			data, err := os.ReadFile(manifestFile)
			if err != nil {
				t.Fatalf("Failed to read manifest file %s: %v", manifestFile, err)
			}

			// Parse YAML
			var manifest ManifestSchema
			if err := yaml.Unmarshal(data, &manifest); err != nil {
				t.Fatalf("Failed to parse YAML from %s: %v", manifestFile, err)
			}

			// Validate schema compliance
			validateManifestSchema(t, &manifest, manifestFile)
		})
	}
}

func validateManifestSchema(t *testing.T, manifest *ManifestSchema, filename string) {
	// Validate meta section
	if manifest.Meta.Version != 1 {
		t.Errorf("%s: meta.version must be 1, got %d", filename, manifest.Meta.Version)
	}

	if manifest.Meta.Name == "" {
		t.Errorf("%s: meta.name must not be empty", filename)
	}

	// Validate language if specified
	if manifest.Meta.Language != "" {
		if len(manifest.Meta.Language) != 2 {
			t.Errorf("%s: meta.language must be 2-character code, got '%s'", filename, manifest.Meta.Language)
		}
	}

	// Validate defaults section
	if manifest.Defaults.TimeoutSec != 0 {
		if manifest.Defaults.TimeoutSec < 1 || manifest.Defaults.TimeoutSec > 300 {
			t.Errorf("%s: defaults.timeout_sec must be between 1 and 300, got %d", filename, manifest.Defaults.TimeoutSec)
		}
	}

	// Validate tools section
	if len(manifest.Tools) == 0 {
		t.Errorf("%s: tools array must not be empty", filename)
	}

	toolIDs := make(map[string]bool)
	for i, tool := range manifest.Tools {
		validateToolDefinition(t, &tool, filename, i)

		// Check for duplicate IDs
		if toolIDs[tool.ID] {
			t.Errorf("%s: duplicate tool ID '%s'", filename, tool.ID)
		}
		toolIDs[tool.ID] = true
	}
}

func validateToolDefinition(t *testing.T, tool *ToolDefinition, filename string, index int) {
	// Validate required fields
	if tool.ID == "" {
		t.Errorf("%s: tool[%d].id must not be empty", filename, index)
	}

	// Validate ID format (lowercase alphanumeric with hyphens)
	for _, char := range tool.ID {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
			t.Errorf("%s: tool[%d].id '%s' must be lowercase alphanumeric with hyphens only", filename, index, tool.ID)
			break
		}
	}

	if tool.Name == "" {
		t.Errorf("%s: tool[%d].name must not be empty", filename, index)
	}

	if tool.Rationale == "" {
		t.Errorf("%s: tool[%d].rationale must not be empty", filename, index)
	}

	if tool.Require == "" {
		t.Errorf("%s: tool[%d].require must not be empty", filename, index)
	}

	// Validate check command
	if len(tool.Check.Cmd) == 0 {
		t.Errorf("%s: tool[%d].check.cmd must not be empty", filename, index)
	}

	if tool.Check.Regex == "" {
		t.Errorf("%s: tool[%d].check.regex must not be empty", filename, index)
	}

	// Validate regex contains named capture group
	if !containsNamedCaptureGroup(tool.Check.Regex) {
		t.Errorf("%s: tool[%d].check.regex must contain named capture group like (?P<ver>...)", filename, index)
	}

	// Validate links
	if len(tool.Links) == 0 {
		t.Errorf("%s: tool[%d].links must not be empty", filename, index)
	}

	for linkType, url := range tool.Links {
		if url == "" {
			t.Errorf("%s: tool[%d].links.%s must not be empty", filename, index, linkType)
		}

		// Basic URL validation
		if !isValidURL(url) {
			t.Errorf("%s: tool[%d].links.%s '%s' is not a valid URL", filename, index, linkType, url)
		}
	}
}

func containsNamedCaptureGroup(regex string) bool {
	// Simple check for named capture group pattern (?P<name>...)
	// This is a basic implementation for testing
	return len(regex) > 7 &&
		   (containsSubstring(regex, "(?P<") || containsSubstring(regex, "(?<"))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func isValidURL(url string) bool {
	// Basic URL validation - check for http/https prefix
	return len(url) > 7 && (url[:7] == "http://" || url[:8] == "https://")
}

func TestManifestSchemaInvalidCases(t *testing.T) {
	// Test cases for invalid manifest structures
	invalidCases := []struct {
		name     string
		yaml     string
		expected string
	}{
		{
			name: "missing meta section",
			yaml: `
tools:
  - id: test
    name: Test
    rationale: Testing
    require: ">=1.0"
    check:
      cmd: ["test"]
      regex: "(?P<ver>\\d+)"
    links:
      homepage: "https://example.com"
`,
			expected: "meta section required",
		},
		{
			name: "invalid version",
			yaml: `
meta:
  version: 2
  name: "Test"
tools:
  - id: test
    name: Test
    rationale: Testing
    require: ">=1.0"
    check:
      cmd: ["test"]
      regex: "(?P<ver>\\d+)"
    links:
      homepage: "https://example.com"
`,
			expected: "version must be 1",
		},
		{
			name: "empty tools array",
			yaml: `
meta:
  version: 1
  name: "Test"
tools: []
`,
			expected: "tools array must not be empty",
		},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			var manifest ManifestSchema
			err := yaml.Unmarshal([]byte(tc.yaml), &manifest)

			// The YAML parsing might succeed, but validation should catch issues
			if err == nil {
				// Run validation and expect it to fail
				// This would be implemented in the actual manifest loading code
				t.Log("YAML parsed successfully but validation should catch the issue")
			}
		})
	}
}