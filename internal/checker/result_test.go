package checker

import (
	"testing"
	"time"
)

func TestCheckResultStatusTransitions(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  CheckStatus
		installedVer   string
		requiredVer    string
		hasErrors      bool
		expectedStatus CheckStatus
		description    string
	}{
		{
			name:           "transition to OK",
			initialStatus:  StatusUnknown,
			installedVer:   "1.22.1",
			requiredVer:    ">=1.22",
			hasErrors:      false,
			expectedStatus: StatusOK,
			description:    "Tool found and meets requirements",
		},
		{
			name:           "transition to Missing",
			initialStatus:  StatusUnknown,
			installedVer:   "",
			requiredVer:    ">=1.22",
			hasErrors:      false,
			expectedStatus: StatusNotFound,
			description:    "Tool not found",
		},
		{
			name:           "transition to OK with version",
			initialStatus:  StatusUnknown,
			installedVer:   "1.21.0",
			requiredVer:    ">=1.22",
			hasErrors:      false,
			expectedStatus: StatusOK,
			description:    "Tool found - DetermineStatus doesn't do version validation",
		},
		{
			name:           "transition to Error",
			initialStatus:  StatusUnknown,
			installedVer:   "",
			requiredVer:    ">=1.22",
			hasErrors:      true,
			expectedStatus: StatusError,
			description:    "Error occurred during check",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &CheckResult{
				ToolID:          "test-tool",
				ToolName:        "Test Tool",
				Status:          tt.initialStatus,
				RequiredVersion: tt.requiredVer,
				ActualVersion:   "",
				ErrorMessage:    "",
				Platform:        "test-platform",
				Links: map[string]string{
					"homepage": "https://example.com",
				},
			}

			if tt.installedVer != "" {
				result.ActualVersion = tt.installedVer
			}

			if tt.hasErrors {
				result.ErrorMessage = "Command execution failed"
			}

			// This would be called by the checker logic
			result.DetermineStatus()

			if result.Status != tt.expectedStatus {
				t.Errorf("Expected status %v, got %v", tt.expectedStatus, result.Status)
			}
		})
	}
}

func TestCheckResultValidation(t *testing.T) {
	tests := []struct {
		name        string
		result      CheckResult
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid check result",
			result: CheckResult{
				ToolID:          "go",
				ToolName:        "Go",
				Status:          StatusOK,
				RequiredVersion: ">=1.22",
				ActualVersion:   "1.22.1",
				ErrorMessage:    "",
				Platform:        "darwin/amd64",
				Links: map[string]string{
					"homepage": "https://go.dev/",
				},
				CheckDuration: time.Millisecond * 100,
			},
			expectError: false,
		},
		{
			name: "missing required fields",
			result: CheckResult{
				ToolID:          "",
				ToolName:        "",
				Status:          StatusOK,
				RequiredVersion: "",
				ActualVersion:   "",
				ErrorMessage:    "",
				Platform:        "",
				Links:           map[string]string{},
			},
			expectError: true,
			errorMsg:    "required fields cannot be empty",
		},
		{
			name: "invalid status consistency - OK without version",
			result: CheckResult{
				ToolID:          "go",
				ToolName:        "Go",
				Status:          StatusOK,
				RequiredVersion: ">=1.22",
				ActualVersion:   "",
				ErrorMessage:    "",
				Platform:        "darwin/amd64",
				Links: map[string]string{
					"homepage": "https://go.dev/",
				},
			},
			expectError: true,
			errorMsg:    "OK status must have actual version",
		},
		{
			name: "invalid status consistency - Missing with version",
			result: CheckResult{
				ToolID:          "go",
				ToolName:        "Go",
				Status:          StatusMissing,
				RequiredVersion: ">=1.22",
				ActualVersion:   "1.22.1",
				ErrorMessage:    "",
				Platform:        "darwin/amd64",
				Links: map[string]string{
					"homepage": "https://go.dev/",
				},
			},
			expectError: true,
			errorMsg:    "Missing/NotFound status should not have actual version",
		},
		{
			name: "invalid status consistency - Outdated without version",
			result: CheckResult{
				ToolID:          "go",
				ToolName:        "Go",
				Status:          StatusOutdated,
				RequiredVersion: ">=1.22",
				ActualVersion:   "",
				ErrorMessage:    "",
				Platform:        "darwin/amd64",
				Links: map[string]string{
					"homepage": "https://go.dev/",
				},
			},
			expectError: true,
			errorMsg:    "Outdated status must have actual version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.result.Validate()

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

func TestCheckStatusString(t *testing.T) {
	tests := []struct {
		status   CheckStatus
		expected string
	}{
		{StatusOK, "ok"},
		{StatusMissing, "missing"},
		{StatusOutdated, "outdated"},
		{StatusError, "error"},
		{StatusUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.status.String() != tt.expected {
				t.Errorf("Expected status string '%s', got '%s'", tt.expected, tt.status.String())
			}
		})
	}
}

func TestCheckResultAddError(t *testing.T) {
	result := CheckResult{
		ToolID:       "test",
		ToolName:     "Test Tool",
		Status:       StatusUnknown,
		ErrorMessage: "",
	}

	// Add error
	result.AddError("test error message")
	if result.ErrorMessage != "test error message" {
		t.Errorf("Expected error message 'test error message', got '%s'", result.ErrorMessage)
	}

	// Check that status is automatically set to Error when errors are added
	if result.Status != StatusError {
		t.Errorf("Expected status to be Error when errors are present, got %v", result.Status)
	}
}

func TestCheckResultHasErrors(t *testing.T) {
	result := CheckResult{
		ErrorMessage: "",
	}

	if result.HasErrors() {
		t.Error("Expected HasErrors() to return false for empty error message")
	}

	result.ErrorMessage = "some error"
	if !result.HasErrors() {
		t.Error("Expected HasErrors() to return true when error message is present")
	}
}

func TestEnvironmentReportValidation(t *testing.T) {
	tests := []struct {
		name        string
		report      EnvironmentReport
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid environment report",
			report: EnvironmentReport{
				SchemaVersion:  1,
				Platform:       map[string]interface{}{"os": "darwin", "arch": "arm64", "hostname": "test-host"},
				Summary:        CheckSummary{Total: 1, OK: 1, Missing: 0, Outdated: 0, Errors: 0},
				ManifestSource: "embedded",
				Items: []CheckResult{
					{
						ToolID:          "go",
						ToolName:        "Go",
						Status:          StatusOK,
						RequiredVersion: ">=1.22",
						ActualVersion:   "1.22.1",
						ErrorMessage:    "",
						Platform:        "darwin/amd64",
						Links:           map[string]string{"homepage": "https://go.dev/"},
					},
				},
				GeneratedAt: time.Now(),
			},
			expectError: false,
		},
		{
			name: "invalid schema version",
			report: EnvironmentReport{
				SchemaVersion:  2,
				Platform:       map[string]interface{}{"os": "darwin", "arch": "arm64"},
				Summary:        CheckSummary{Total: 0, OK: 0, Missing: 0, Outdated: 0, Errors: 0},
				ManifestSource: "embedded",
				Items:          []CheckResult{},
				GeneratedAt:    time.Now(),
			},
			expectError: true,
			errorMsg:    "unsupported schema version",
		},
		{
			name: "summary mismatch",
			report: EnvironmentReport{
				SchemaVersion:  1,
				Platform:       map[string]interface{}{"os": "darwin", "arch": "arm64"},
				Summary:        CheckSummary{Total: 2, OK: 1, Missing: 0, Outdated: 0, Errors: 0},
				ManifestSource: "embedded",
				Items: []CheckResult{
					{
						ToolID:          "go",
						ToolName:        "Go",
						Status:          StatusOK,
						RequiredVersion: ">=1.22",
						ActualVersion:   "1.22.1",
						ErrorMessage:    "",
						Platform:        "darwin/amd64",
						Links:           map[string]string{"homepage": "https://go.dev/"},
					},
				},
				GeneratedAt: time.Now(),
			},
			expectError: true,
			errorMsg:    "summary total mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.report.Validate()

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

func TestCheckSummaryCalculation(t *testing.T) {
	items := []CheckResult{
		{Status: StatusOK},
		{Status: StatusOK},
		{Status: StatusMissing},
		{Status: StatusOutdated},
		{Status: StatusError},
	}

	summary := CalculateCheckSummary(items)

	expected := CheckSummary{
		Total:    5,
		OK:       2,
		Missing:  1,
		Outdated: 1,
		Errors:   1,
	}

	if summary != expected {
		t.Errorf("Expected summary %+v, got %+v", expected, summary)
	}
}

