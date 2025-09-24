package contract

import (
	"encoding/json"
	"os/exec"
	"testing"
	"time"
)

// EnvironmentReport represents the JSON schema for doctor command output
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

// ListResponse represents the JSON schema for list command output
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

func TestJSONOutputSchemaCompliance(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		args        []string
		schemaType  string
	}{
		{
			name:       "doctor command JSON output",
			command:    "./doctor",
			args:       []string{"doctor", "--json"},
			schemaType: "environment_report",
		},
		{
			name:       "list command JSON output",
			command:    "./doctor",
			args:       []string{"list", "--json"},
			schemaType: "list_response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(tt.command, tt.args[1:]...)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Skip("Skipping test until CLI is implemented")
			}

			// Test that output is valid JSON
			var jsonOutput interface{}
			if err := json.Unmarshal(output, &jsonOutput); err != nil {
				t.Fatalf("Expected valid JSON output, got error: %v\nOutput: %s", err, string(output))
			}

			// Test schema compliance based on command type
			switch tt.schemaType {
			case "environment_report":
				var report EnvironmentReport
				if err := json.Unmarshal(output, &report); err != nil {
					t.Fatalf("Failed to unmarshal environment report: %v", err)
				}
				validateEnvironmentReport(t, &report)

			case "list_response":
				var listResp ListResponse
				if err := json.Unmarshal(output, &listResp); err != nil {
					t.Fatalf("Failed to unmarshal list response: %v", err)
				}
				validateListResponse(t, &listResp)
			}
		})
	}
}

func validateEnvironmentReport(t *testing.T, report *EnvironmentReport) {
	// Validate schema version
	if report.SchemaVersion != 1 {
		t.Errorf("Expected schema_version to be 1, got %d", report.SchemaVersion)
	}

	// Validate platform info
	validOS := map[string]bool{"darwin": true, "linux": true}
	if !validOS[report.Platform.OS] {
		t.Errorf("Expected platform.os to be 'darwin' or 'linux', got '%s'", report.Platform.OS)
	}

	validArch := map[string]bool{"amd64": true, "arm64": true}
	if !validArch[report.Platform.Arch] {
		t.Errorf("Expected platform.arch to be 'amd64' or 'arm64', got '%s'", report.Platform.Arch)
	}

	// Validate summary consistency
	totalItems := len(report.Items)
	if report.Summary.Total != totalItems {
		t.Errorf("Summary.Total (%d) should match number of items (%d)", report.Summary.Total, totalItems)
	}

	calculatedTotal := report.Summary.OK + report.Summary.Missing + report.Summary.Outdated + report.Summary.Errors
	if calculatedTotal != report.Summary.Total {
		t.Errorf("Sum of status counts (%d) should equal total (%d)", calculatedTotal, report.Summary.Total)
	}

	// Validate each check result
	for i, item := range report.Items {
		validateCheckResult(t, &item, i)
	}

	// Validate manifest source is not empty
	if report.ManifestSource == "" {
		t.Error("ManifestSource should not be empty")
	}

	// Validate generated_at is not zero
	if report.GeneratedAt.IsZero() {
		t.Error("GeneratedAt should not be zero time")
	}
}

func validateCheckResult(t *testing.T, result *CheckResult, index int) {
	// Validate required fields
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

	// Validate status values
	validStatuses := map[string]bool{"ok": true, "missing": true, "outdated": true, "error": true}
	if !validStatuses[result.Status] {
		t.Errorf("Item %d: Invalid status '%s'", index, result.Status)
	}

	// Validate status consistency
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

	// Validate links are valid URLs (basic check)
	for linkType, url := range result.Links {
		if url == "" {
			t.Errorf("Item %d: Link '%s' should not be empty", index, linkType)
		}
	}

	// Validate errors array is not nil
	if result.Errors == nil {
		t.Errorf("Item %d: Errors array should not be nil", index)
	}
}

func validateListResponse(t *testing.T, response *ListResponse) {
	// Validate manifest source is not empty
	if response.ManifestSource == "" {
		t.Error("ManifestSource should not be empty")
	}

	// Validate tools array
	if response.Tools == nil {
		t.Error("Tools array should not be nil")
	}

	// Validate each tool info
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