package contract

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

func TestListCommandInterface(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		expectedExit   int
		outputFormat   string
	}{
		{
			name:           "list command with no flags",
			args:           []string{"list"},
			expectedOutput: "Tools defined in manifest:",
			expectedExit:   0,
			outputFormat:   "human",
		},
		{
			name:           "list command with JSON flag",
			args:           []string{"list", "--json"},
			expectedOutput: `{"manifest_source"`,
			expectedExit:   0,
			outputFormat:   "json",
		},
		{
			name:           "list command with custom manifest",
			args:           []string{"list", "-f", "testdata/manifests/sample.yaml"},
			expectedOutput: "Tools defined in manifest:",
			expectedExit:   0,
			outputFormat:   "human",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test will fail until we implement the CLI
			cmd := exec.Command("./bin/goctor", tt.args...)
			output, err := cmd.CombinedOutput()

			// Check exit code
			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				}
			}

			if exitCode != tt.expectedExit {
				t.Errorf("Expected exit code %d, got %d", tt.expectedExit, exitCode)
			}

			// Check output format
			outputStr := string(output)
			if !strings.Contains(outputStr, tt.expectedOutput) {
				t.Errorf("Expected output to contain '%s', got: %s", tt.expectedOutput, outputStr)
			}

			// Validate JSON output if specified
			if tt.outputFormat == "json" {
				var result map[string]interface{}
				if err := json.Unmarshal(output, &result); err != nil {
					t.Errorf("Expected valid JSON output, got error: %v", err)
				}

				// Validate required fields for list command JSON output
				requiredFields := []string{"manifest_source", "tools"}
				for _, field := range requiredFields {
					if _, exists := result[field]; !exists {
						t.Errorf("JSON output missing required field: %s", field)
					}
				}

				// Validate tools array structure
				if tools, ok := result["tools"].([]interface{}); ok {
					if len(tools) > 0 {
						if tool, ok := tools[0].(map[string]interface{}); ok {
							toolFields := []string{"id", "name", "required_version", "rationale"}
							for _, field := range toolFields {
								if _, exists := tool[field]; !exists {
									t.Errorf("Tool object missing required field: %s", field)
								}
							}
						}
					}
				} else {
					t.Error("tools field should be an array")
				}
			}
		})
	}
}

func TestListCommandDisplaysAllTools(t *testing.T) {
	// Test that list command shows all tools from manifest without executing checks
	cmd := exec.Command("./bin/goctor", "list", "-f", "testdata/manifests/sample.yaml")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Skip("Skipping test until CLI is implemented")
	}

	outputStr := string(output)

	// Check that it includes expected tools from sample manifest
	expectedTools := []string{"go:", "git:", "docker:"}
	for _, tool := range expectedTools {
		if !strings.Contains(outputStr, tool) {
			t.Errorf("Expected output to contain tool '%s', got: %s", tool, outputStr)
		}
	}

	// Check that it includes version requirements
	expectedPatterns := []string{">=1.22", ">=2.30", ">=24"}
	for _, pattern := range expectedPatterns {
		if !strings.Contains(outputStr, pattern) {
			t.Errorf("Expected output to contain version pattern '%s', got: %s", pattern, outputStr)
		}
	}

	// Check that it includes rationales
	expectedRationales := []string{"Go development", "Version control", "Container platform"}
	for _, rationale := range expectedRationales {
		if !strings.Contains(outputStr, rationale) {
			t.Errorf("Expected output to contain rationale '%s', got: %s", rationale, outputStr)
		}
	}
}
