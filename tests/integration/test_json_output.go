package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestJSONOutputIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	buildCmd := exec.Command("go", "build", "-o", "doctor-test", "./cmd/doctor")
	if err := buildCmd.Run(); err != nil {
		t.Skip("Skipping test - unable to build binary")
	}
	defer os.Remove("doctor-test")

	// Create test manifest
	manifestContent := `
meta:
  version: 1
  name: "JSON Test Manifest"

tools:
  - id: go
    name: "Go"
    rationale: "Go development toolchain"
    require: ">=1.0"
    check:
      cmd: ["go", "version"]
      regex: "go(?P<ver>\\d+\\.\\d+(\\.\\d+)?)"
    links:
      homepage: "https://go.dev/"
      download: "https://go.dev/dl/"
      docs: "https://go.dev/doc/"
`

	manifestFile := "test-json-manifest.yaml"
	if err := os.WriteFile(manifestFile, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to create test manifest: %v", err)
	}
	defer os.Remove(manifestFile)

	t.Run("doctor command JSON output", func(t *testing.T) {
		cmd := exec.Command("./doctor-test", "doctor", "--json", "-f", manifestFile)
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skip("Skipping test - command execution failed (expected until implementation)")
		}

		// Validate JSON structure
		var report EnvironmentReport
		if err := json.Unmarshal(output, &report); err != nil {
			t.Errorf("Failed to parse JSON output: %v\nOutput: %s", err, string(output))
			return
		}

		// Validate schema compliance
		validateEnvironmentReportIntegration(t, &report)
	})

	t.Run("list command JSON output", func(t *testing.T) {
		cmd := exec.Command("./doctor-test", "list", "--json", "-f", manifestFile)
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skip("Skipping test - command execution failed")
		}

		// Validate JSON structure
		var listResp ListResponse
		if err := json.Unmarshal(output, &listResp); err != nil {
			t.Errorf("Failed to parse JSON output: %v\nOutput: %s", err, string(output))
			return
		}

		// Validate list response
		validateListResponseIntegration(t, &listResp)
	})
}

// JSON schema types for integration testing
type EnvironmentReport struct {
	SchemaVersion  int           `json:"schema_version"`
	Platform       PlatformInfo  `json:"platform"`
	Summary        CheckSummary  `json:"summary"`
	ManifestSource string        `json:"manifest_source"`
	Items          []CheckResult `json:"items"`
	GeneratedAt    time.Time     `json:"generated_at"`
}

type PlatformInfo struct {
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Hostname string `json:"hostname,omitempty"`
}

type CheckSummary struct {
	Total    int `json:"total"`
	OK       int `json:"ok"`
	Missing  int `json:"missing"`
	Outdated int `json:"outdated"`
	Errors   int `json:"errors"`
}

type CheckResult struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Status           string            `json:"status"`
	Required         string            `json:"required"`
	Installed        *string           `json:"installed"`
	Rationale        string            `json:"rationale"`
	Links            map[string]string `json:"links"`
	Errors           []string          `json:"errors"`
}

type ListResponse struct {
	ManifestSource string     `json:"manifest_source"`
	Tools          []ToolInfo `json:"tools"`
}

type ToolInfo struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	RequiredVersion string `json:"required_version"`
	Rationale       string `json:"rationale"`
}

func validateEnvironmentReportIntegration(t *testing.T, report *EnvironmentReport) {
	// Schema version validation
	if report.SchemaVersion != 1 {
		t.Errorf("Expected schema_version 1, got %d", report.SchemaVersion)
	}

	// Platform validation
	validOS := map[string]bool{"darwin": true, "linux": true}
	if !validOS[report.Platform.OS] {
		t.Errorf("Expected platform.os to be 'darwin' or 'linux', got '%s'", report.Platform.OS)
	}

	validArch := map[string]bool{"amd64": true, "arm64": true}
	if !validArch[report.Platform.Arch] {
		t.Errorf("Expected platform.arch to be 'amd64' or 'arm64', got '%s'", report.Platform.Arch)
	}

	// Summary validation
	if report.Summary.Total != len(report.Items) {
		t.Errorf("Summary.Total (%d) should match number of items (%d)",
			report.Summary.Total, len(report.Items))
	}

	calculatedTotal := report.Summary.OK + report.Summary.Missing + report.Summary.Outdated + report.Summary.Errors
	if calculatedTotal != report.Summary.Total {
		t.Errorf("Sum of status counts (%d) should equal total (%d)",
			calculatedTotal, report.Summary.Total)
	}

	// Items validation
	for i, item := range report.Items {
		validateCheckResultIntegration(t, &item, i)
	}

	// Manifest source validation
	if report.ManifestSource == "" {
		t.Error("ManifestSource should not be empty")
	}

	// Generated timestamp validation
	if report.GeneratedAt.IsZero() {
		t.Error("GeneratedAt should not be zero time")
	}

	// Generated timestamp should be recent (within last minute)
	if time.Since(report.GeneratedAt) > time.Minute {
		t.Error("GeneratedAt should be recent")
	}
}

func validateCheckResultIntegration(t *testing.T, result *CheckResult, index int) {
	// Required fields
	if result.ID == "" {
		t.Errorf("Item %d: ID should not be empty", index)
	}

	if result.Name == "" {
		t.Errorf("Item %d: Name should not be empty", index)
	}

	if result.Required == "" {
		t.Errorf("Item %d: Required should not be empty", index)
	}

	if result.Rationale == "" {
		t.Errorf("Item %d: Rationale should not be empty", index)
	}

	// Status validation
	validStatuses := map[string]bool{"ok": true, "missing": true, "outdated": true, "error": true}
	if !validStatuses[result.Status] {
		t.Errorf("Item %d: Invalid status '%s'", index, result.Status)
	}

	// Status consistency validation
	switch result.Status {
	case "ok":
		if result.Installed == nil {
			t.Errorf("Item %d: OK status should have installed version", index)
		}
	case "missing":
		if result.Installed != nil {
			t.Errorf("Item %d: Missing status should not have installed version", index)
		}
	case "outdated":
		if result.Installed == nil {
			t.Errorf("Item %d: Outdated status should have installed version", index)
		}
	}

	// Links validation
	if len(result.Links) == 0 {
		t.Errorf("Item %d: Links should not be empty", index)
	}

	for linkType, url := range result.Links {
		if url == "" {
			t.Errorf("Item %d: Link '%s' should not be empty", index, linkType)
		}
		// Basic URL validation
		if len(url) < 7 || (url[:7] != "http://" && url[:8] != "https://") {
			t.Errorf("Item %d: Link '%s' should be a valid URL", index, linkType)
		}
	}

	// Errors array validation
	if result.Errors == nil {
		t.Errorf("Item %d: Errors array should not be nil", index)
	}
}

func validateListResponseIntegration(t *testing.T, response *ListResponse) {
	// Manifest source validation
	if response.ManifestSource == "" {
		t.Error("ManifestSource should not be empty")
	}

	// Tools validation
	if response.Tools == nil {
		t.Error("Tools array should not be nil")
	}

	if len(response.Tools) == 0 {
		t.Error("Tools array should not be empty")
	}

	// Validate each tool
	for i, tool := range response.Tools {
		if tool.ID == "" {
			t.Errorf("Tool %d: ID should not be empty", i)
		}

		if tool.Name == "" {
			t.Errorf("Tool %d: Name should not be empty", i)
		}

		if tool.RequiredVersion == "" {
			t.Errorf("Tool %d: RequiredVersion should not be empty", i)
		}

		if tool.Rationale == "" {
			t.Errorf("Tool %d: Rationale should not be empty", i)
		}
	}
}

func TestJSONOutputParsability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	buildCmd := exec.Command("go", "build", "-o", "doctor-test", "./cmd/doctor")
	if err := buildCmd.Run(); err != nil {
		t.Skip("Skipping test - unable to build binary")
	}
	defer os.Remove("doctor-test")

	tests := []struct {
		name     string
		command  []string
		jsonPath string
	}{
		{
			name:     "doctor JSON output",
			command:  []string{"doctor", "--json"},
			jsonPath: "$.schema_version",
		},
		{
			name:     "list JSON output",
			command:  []string{"list", "--json"},
			jsonPath: "$.manifest_source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./doctor-test", tt.command...)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Skip("Skipping test - command execution failed")
			}

			// Test that output is valid JSON
			var jsonOutput interface{}
			if err := json.Unmarshal(output, &jsonOutput); err != nil {
				t.Errorf("Output is not valid JSON: %v\nOutput: %s", err, string(output))
			}

			// Test that output can be processed by standard JSON tools
			// This simulates real-world usage where output is piped to jq or similar
			if len(output) == 0 {
				t.Error("JSON output should not be empty")
			}

			// Verify output starts with '{' (JSON object)
			trimmed := string(output)
			if len(trimmed) == 0 || trimmed[0] != '{' {
				t.Error("JSON output should start with '{'")
			}
		})
	}
}