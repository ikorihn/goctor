package manifest

import (
	"strings"
	"testing"
)

func TestManifestValidation(t *testing.T) {
	tests := []struct {
		name        string
		manifest    Manifest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid manifest",
			manifest: Manifest{
				Meta: ManifestMeta{
					Version:  1,
					Name:     "Test Manifest",
					Language: "en",
				},
				Defaults: ManifestDefaults{
					TimeoutSeconds: 5,
					RegexKey:       "ver",
				},
				Tools: []ToolDefinition{
					{
						ID:        "go",
						Name:      "Go",
						Rationale: "Go development",
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
				},
			},
			expectError: false,
		},
		{
			name: "invalid meta version",
			manifest: Manifest{
				Meta: ManifestMeta{
					Version:  2,
					Name:     "Test Manifest",
					Language: "en",
				},
				Tools: []ToolDefinition{
					{
						ID:        "go",
						Name:      "Go",
						Rationale: "Go development",
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
				},
			},
			expectError: true,
			errorMsg:    "unsupported manifest version",
		},
		{
			name: "empty tools list",
			manifest: Manifest{
				Meta: ManifestMeta{
					Version:  1,
					Name:     "Test Manifest",
					Language: "en",
				},
				Tools: []ToolDefinition{},
			},
			expectError: true,
			errorMsg:    "tools list cannot be empty",
		},
		{
			name: "duplicate tool IDs",
			manifest: Manifest{
				Meta: ManifestMeta{
					Version:  1,
					Name:     "Test Manifest",
					Language: "en",
				},
				Tools: []ToolDefinition{
					{
						ID:        "go",
						Name:      "Go",
						Rationale: "Go development",
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
					{
						ID:        "go",
						Name:      "Go Alternative",
						Rationale: "Another Go tool",
						RequiredVersion: ">=1.20",
						Check: CheckConfig{
							Command: []string{"go", "version"},
							Regex:   "go(?P<ver>\\d+\\.\\d+(\\.\\d+)?)",
						},
						Links: map[string]string{
							"homepage": "https://go.dev/",
						},
						TimeoutSeconds: 5,
					},
				},
			},
			expectError: true,
			errorMsg:    "duplicate tool ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.manifest.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error, got nil")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, got: %v", err)
				}
			}
		})
	}
}

func TestManifestMetaValidation(t *testing.T) {
	tests := []struct {
		name        string
		meta        ManifestMeta
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid meta",
			meta: ManifestMeta{
				Version:  1,
				Name:     "Test Manifest",
				Language: "en",
			},
			expectError: false,
		},
		{
			name: "valid meta without language",
			meta: ManifestMeta{
				Version: 1,
				Name:    "Test Manifest",
			},
			expectError: false,
		},
		{
			name: "invalid version",
			meta: ManifestMeta{
				Version:  0,
				Name:     "Test Manifest",
				Language: "en",
			},
			expectError: true,
			errorMsg:    "invalid version",
		},
		{
			name: "empty name",
			meta: ManifestMeta{
				Version:  1,
				Name:     "",
				Language: "en",
			},
			expectError: true,
			errorMsg:    "name cannot be empty",
		},
		{
			name: "invalid language code",
			meta: ManifestMeta{
				Version:  1,
				Name:     "Test Manifest",
				Language: "english",
			},
			expectError: true,
			errorMsg:    "invalid language code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.meta.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, got: %v", err)
				}
			}
		})
	}
}

func TestManifestDefaultsValidation(t *testing.T) {
	tests := []struct {
		name        string
		defaults    ManifestDefaults
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid defaults",
			defaults: ManifestDefaults{
				TimeoutSeconds: 5,
				RegexKey:       "ver",
			},
			expectError: false,
		},
		{
			name: "valid defaults with zero timeout (use system default)",
			defaults: ManifestDefaults{
				TimeoutSeconds: 0,
				RegexKey:       "version",
			},
			expectError: false,
		},
		{
			name: "invalid timeout - negative",
			defaults: ManifestDefaults{
				TimeoutSeconds: -1,
				RegexKey:       "ver",
			},
			expectError: true,
			errorMsg:    "timeout must be positive",
		},
		{
			name: "invalid timeout - too large",
			defaults: ManifestDefaults{
				TimeoutSeconds: 600,
				RegexKey:       "ver",
			},
			expectError: true,
			errorMsg:    "timeout too large",
		},
		{
			name: "empty regex key - allowed",
			defaults: ManifestDefaults{
				TimeoutSeconds: 5,
				RegexKey:       "",
			},
			expectError: false,
			errorMsg:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.defaults.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, got: %v", err)
				}
			}
		})
	}
}

func TestManifestApplyDefaults(t *testing.T) {
	manifest := Manifest{
		Meta: ManifestMeta{
			Version: 1,
			Name:    "Test Manifest",
		},
		Defaults: ManifestDefaults{
			TimeoutSeconds: 10,
			RegexKey:       "version",
		},
		Tools: []ToolDefinition{
			{
				ID:        "tool1",
				Name:      "Tool 1",
				Rationale: "Testing",
				RequiredVersion: ">=1.0",
				Check: CheckConfig{
					Command: []string{"tool1", "--version"},
					Regex:   "(?P<ver>\\d+\\.\\d+\\.\\d+)",
				},
				Links: map[string]string{
					"homepage": "https://example.com",
				},
				TimeoutSeconds: 0, // Should use default
			},
			{
				ID:        "tool2",
				Name:      "Tool 2",
				Rationale: "Testing",
				RequiredVersion: ">=2.0",
				Check: CheckConfig{
					Command: []string{"tool2", "--version"},
					Regex:   "(?P<version>\\d+\\.\\d+\\.\\d+)",
				},
				Links: map[string]string{
					"homepage": "https://example.com",
				},
				TimeoutSeconds: 15, // Should keep explicit value
			},
		},
	}

	manifest.ApplyDefaults()

	// Check that defaults were applied correctly
	if manifest.Tools[0].TimeoutSeconds != 10 {
		t.Errorf("Expected tool1 timeout to be 10 (from defaults), got %d", manifest.Tools[0].TimeoutSeconds)
	}

	if manifest.Tools[1].TimeoutSeconds != 15 {
		t.Errorf("Expected tool2 timeout to remain 15 (explicit value), got %d", manifest.Tools[1].TimeoutSeconds)
	}
}

func TestManifestMerge(t *testing.T) {
	base := Manifest{
		Meta: ManifestMeta{
			Version: 1,
			Name:    "Base Manifest",
		},
		Tools: []ToolDefinition{
			{
				ID:        "go",
				Name:      "Go",
				Rationale: "Go development",
				RequiredVersion: ">=1.20",
				Check: CheckConfig{
					Command: []string{"go", "version"},
					Regex:   "go(?P<ver>\\d+\\.\\d+(\\.\\d+)?)",
				},
				Links: map[string]string{
					"homepage": "https://go.dev/",
				},
				TimeoutSeconds: 5,
			},
			{
				ID:        "git",
				Name:      "Git",
				Rationale: "Version control",
				RequiredVersion: ">=2.30",
				Check: CheckConfig{
					Command: []string{"git", "--version"},
					Regex:   "git version (?P<ver>\\d+\\.\\d+\\.\\d+)",
				},
				Links: map[string]string{
					"homepage": "https://git-scm.com/",
				},
				TimeoutSeconds: 5,
			},
		},
	}

	override := Manifest{
		Meta: ManifestMeta{
			Version: 1,
			Name:    "Override Manifest",
		},
		Tools: []ToolDefinition{
			{
				ID:        "go",
				Name:      "Go (Updated)",
				Rationale: "Updated Go development",
				RequiredVersion: ">=1.22",
				Check: CheckConfig{
					Command: []string{"go", "version"},
					Regex:   "go(?P<ver>\\d+\\.\\d+(\\.\\d+)?)",
				},
				Links: map[string]string{
					"homepage": "https://go.dev/",
					"download": "https://go.dev/dl/",
				},
				TimeoutSeconds: 10,
			},
			{
				ID:        "docker",
				Name:      "Docker",
				Rationale: "Containerization",
				RequiredVersion: ">=24",
				Check: CheckConfig{
					Command: []string{"docker", "--version"},
					Regex:   "version (?P<ver>\\d+\\.\\d+\\.\\d+)",
				},
				Links: map[string]string{
					"homepage": "https://docker.com/",
				},
				TimeoutSeconds: 15,
			},
		},
	}

	merged := base.Merge(override)

	// Check that Go tool was updated
	goTool := findToolByID(merged.Tools, "go")
	if goTool == nil {
		t.Fatal("Go tool not found in merged manifest")
	}

	if goTool.Name != "Go (Updated)" {
		t.Errorf("Expected merged Go tool name to be 'Go (Updated)', got '%s'", goTool.Name)
	}

	if goTool.RequiredVersion != ">=1.22" {
		t.Errorf("Expected merged Go tool version to be '>=1.22', got '%s'", goTool.RequiredVersion)
	}

	// Check that Git tool was preserved
	gitTool := findToolByID(merged.Tools, "git")
	if gitTool == nil {
		t.Fatal("Git tool not found in merged manifest")
	}

	if gitTool.RequiredVersion != ">=2.30" {
		t.Errorf("Expected Git tool version to be preserved as '>=2.30', got '%s'", gitTool.RequiredVersion)
	}

	// Check that Docker tool was added
	dockerTool := findToolByID(merged.Tools, "docker")
	if dockerTool == nil {
		t.Fatal("Docker tool not found in merged manifest")
	}

	if dockerTool.RequiredVersion != ">=24" {
		t.Errorf("Expected Docker tool version to be '>=24', got '%s'", dockerTool.RequiredVersion)
	}

	// Check total number of tools
	expectedCount := 3 // go (updated), git (preserved), docker (added)
	if len(merged.Tools) != expectedCount {
		t.Errorf("Expected %d tools in merged manifest, got %d", expectedCount, len(merged.Tools))
	}
}

func findToolByID(tools []ToolDefinition, id string) *ToolDefinition {
	for i := range tools {
		if tools[i].ID == id {
			return &tools[i]
		}
	}
	return nil
}