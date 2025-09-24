package manifest

import (
	"testing"
)

func TestToolDefinitionValidation(t *testing.T) {
	tests := []struct {
		name        string
		tool        ToolDefinition
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid tool definition",
			tool: ToolDefinition{
				ID:        "go",
				Name:      "Go",
				Rationale: "Go development toolchain",
				RequiredVersion: ">=1.22",
				Check: CheckConfig{
					Command: []string{"go", "version"},
					Regex:   "go(?P<ver>\\d+\\.\\d+(\\.\\d+)?)",
				},
				Links: map[string]string{
					"homepage": "https://go.dev/",
					"download": "https://go.dev/dl/",
				},
				TimeoutSeconds: 5,
			},
			expectError: false,
		},
		{
			name: "invalid ID - uppercase",
			tool: ToolDefinition{
				ID:        "Go-Tool",
				Name:      "Go",
				Rationale: "Go development toolchain",
				RequiredVersion: ">=1.22",
				Check: CheckConfig{
					Command: []string{"go", "version"},
					Regex:   "go(?P<ver>\\d+\\.\\d+(\\.\\d+)?)",
				},
				Links: map[string]string{
					"homepage": "https://go.dev/",
				},
				TimeoutSeconds: 5,
			},
			expectError: true,
			errorMsg:    "ID must be lowercase alphanumeric with hyphens",
		},
		{
			name: "invalid ID - special characters",
			tool: ToolDefinition{
				ID:        "go_tool!",
				Name:      "Go",
				Rationale: "Go development toolchain",
				RequiredVersion: ">=1.22",
				Check: CheckConfig{
					Command: []string{"go", "version"},
					Regex:   "go(?P<ver>\\d+\\.\\d+(\\.\\d+)?)",
				},
				Links: map[string]string{
					"homepage": "https://go.dev/",
				},
				TimeoutSeconds: 5,
			},
			expectError: true,
			errorMsg:    "ID must be lowercase alphanumeric with hyphens",
		},
		{
			name: "empty required fields",
			tool: ToolDefinition{
				ID:        "",
				Name:      "",
				Rationale: "",
				RequiredVersion: "",
				Check: CheckConfig{
					Command: []string{},
					Regex:   "",
				},
				Links:           map[string]string{},
				TimeoutSeconds:  0,
			},
			expectError: true,
			errorMsg:    "required fields cannot be empty",
		},
		{
			name: "invalid version regex - no named capture group",
			tool: ToolDefinition{
				ID:        "go",
				Name:      "Go",
				Rationale: "Go development toolchain",
				RequiredVersion: ">=1.22",
				Check: CheckConfig{
					Command: []string{"go", "version"},
					Regex:   "go\\d+\\.\\d+(\\.\\d+)?",
				},
				Links: map[string]string{
					"homepage": "https://go.dev/",
				},
				TimeoutSeconds: 5,
			},
			expectError: true,
			errorMsg:    "VersionRegex must contain named capture group",
		},
		{
			name: "invalid timeout - negative",
			tool: ToolDefinition{
				ID:        "go",
				Name:      "Go",
				Rationale: "Go development toolchain",
				RequiredVersion: ">=1.22",
				Check: CheckConfig{
					Command: []string{"go", "version"},
					Regex:   "go(?P<ver>\\d+\\.\\d+(\\.\\d+)?)",
				},
				Links: map[string]string{
					"homepage": "https://go.dev/",
				},
				TimeoutSeconds: -1,
			},
			expectError: true,
			errorMsg:    "TimeoutSeconds must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tool.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error, got nil")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, got: %v", err)
				}
			}
		})
	}
}

func TestToolDefinitionSemverConstraintValidation(t *testing.T) {
	tests := []struct {
		name        string
		constraint  string
		expectError bool
	}{
		{"valid constraint - exact", "1.22.0", false},
		{"valid constraint - greater than", ">=1.22", false},
		{"valid constraint - range", ">=1.22 <1.25", false},
		{"valid constraint - tilde", "~1.22.0", false},
		{"valid constraint - caret", "^1.22.0", false},
		{"invalid constraint - empty", "", true},
		{"invalid constraint - malformed", ">=1.22.x", true},
		{"invalid constraint - invalid operator", "=>1.22", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := ToolDefinition{
				ID:        "test",
				Name:      "Test",
				Rationale: "Testing",
				RequiredVersion: tt.constraint,
				Check: CheckConfig{
					Command: []string{"test", "--version"},
					Regex:   "(?P<ver>\\d+\\.\\d+\\.\\d+)",
				},
				Links: map[string]string{
					"homepage": "https://example.com",
				},
				TimeoutSeconds: 5,
			}

			err := tool.ValidateVersionConstraint()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error for constraint '%s', got nil", tt.constraint)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error for constraint '%s', got: %v", tt.constraint, err)
				}
			}
		})
	}
}

func TestToolDefinitionRegexValidation(t *testing.T) {
	tests := []struct {
		name        string
		regex       string
		expectError bool
		errorType   string
	}{
		{
			name:        "valid regex with named group",
			regex:       "go(?P<ver>\\d+\\.\\d+(\\.\\d+)?)",
			expectError: false,
		},
		{
			name:        "valid regex with different named group",
			regex:       "version (?P<version>\\d+\\.\\d+\\.\\d+)",
			expectError: false,
		},
		{
			name:        "invalid regex - no named group",
			regex:       "go\\d+\\.\\d+(\\.\\d+)?",
			expectError: true,
			errorType:   "missing_named_group",
		},
		{
			name:        "invalid regex - malformed",
			regex:       "go(?P<ver>\\d+\\.\\d+(\\.\\d+)?",
			expectError: true,
			errorType:   "malformed_regex",
		},
		{
			name:        "empty regex",
			regex:       "",
			expectError: true,
			errorType:   "empty_regex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := ToolDefinition{
				ID:        "test",
				Name:      "Test",
				Rationale: "Testing",
				RequiredVersion: ">=1.0",
				Check: CheckConfig{
					Command: []string{"test", "--version"},
					Regex:   tt.regex,
				},
				Links: map[string]string{
					"homepage": "https://example.com",
				},
				TimeoutSeconds: 5,
			}

			err := tool.ValidateRegex()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error for regex '%s', got nil", tt.regex)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error for regex '%s', got: %v", tt.regex, err)
				}
			}
		})
	}
}

func TestToolDefinitionLinkValidation(t *testing.T) {
	tests := []struct {
		name        string
		links       map[string]string
		expectError bool
	}{
		{
			name: "valid links",
			links: map[string]string{
				"homepage": "https://example.com",
				"download": "https://example.com/download",
				"docs":     "https://docs.example.com",
			},
			expectError: false,
		},
		{
			name: "valid http links",
			links: map[string]string{
				"homepage": "http://example.com",
				"download": "http://example.com/download",
			},
			expectError: false,
		},
		{
			name: "invalid links - not URLs",
			links: map[string]string{
				"homepage": "not-a-url",
				"download": "example.com",
			},
			expectError: true,
		},
		{
			name:        "empty links",
			links:       map[string]string{},
			expectError: true,
		},
		{
			name: "empty link value",
			links: map[string]string{
				"homepage": "",
				"download": "https://example.com",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := ToolDefinition{
				ID:        "test",
				Name:      "Test",
				Rationale: "Testing",
				RequiredVersion: ">=1.0",
				Check: CheckConfig{
					Command: []string{"test", "--version"},
					Regex:   "(?P<ver>\\d+\\.\\d+\\.\\d+)",
				},
				Links:           tt.links,
				TimeoutSeconds:  5,
			}

			err := tool.ValidateLinks()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error for links %v, got nil", tt.links)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error for links %v, got: %v", tt.links, err)
				}
			}
		})
	}
}