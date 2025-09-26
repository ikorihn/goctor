package contract

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestGoctorCommandInterface(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		expectedExit   int
		outputFormat   string
	}{
		{
			name:           "doctor command with no flags",
			args:           []string{"doctor"},
			expectedOutput: "[OK]",
			expectedExit:   0,
			outputFormat:   "human",
		},
		{
			name:           "doctor command with JSON flag",
			args:           []string{"doctor", "--json"},
			expectedOutput: `{"schema_version":1`,
			expectedExit:   0,
			outputFormat:   "json",
		},
		{
			name:           "doctor command with custom manifest",
			args:           []string{"doctor", "-f", "testdata/manifests/sample.yaml"},
			expectedOutput: "[OK]",
			expectedExit:   0,
			outputFormat:   "human",
		},
		{
			name:           "doctor command with help flag",
			args:           []string{"doctor", "-h"},
			expectedOutput: "Usage:",
			expectedExit:   0,
			outputFormat:   "help",
		},
		{
			name:           "doctor command with version flag",
			args:           []string{"doctor", "-v"},
			expectedOutput: "doctor version",
			expectedExit:   0,
			outputFormat:   "version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test will fail until we implement the CLI
			cmd := exec.Command("./doctor", tt.args[1:]...)
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

				// Validate required fields according to contract
				requiredFields := []string{"schema_version", "platform", "summary", "manifest_source", "items", "generated_at"}
				for _, field := range requiredFields {
					if _, exists := result[field]; !exists {
						t.Errorf("JSON output missing required field: %s", field)
					}
				}
			}
		})
	}
}

func TestDoctorCommandExitCodes(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedExit int
		description  string
	}{
		{
			name:         "all tools satisfied",
			args:         []string{"doctor"},
			expectedExit: 0,
			description:  "Should return 0 when all tools meet requirements",
		},
		{
			name:         "missing or outdated tools",
			args:         []string{"doctor", "-f", "testdata/manifests/missing-tools.yaml"},
			expectedExit: 1,
			description:  "Should return 1 when tools are missing or outdated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./bin/goctor", tt.args[1:]...)
			_, err := cmd.CombinedOutput()

			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				}
			}

			if exitCode != tt.expectedExit {
				t.Errorf("%s: Expected exit code %d, got %d", tt.description, tt.expectedExit, exitCode)
			}
		})
	}
}

func TestMain(m *testing.M) {
	// Build the binary before running tests
	cmd := exec.Command("go", "build", "-o", "./bin", "./cmd/goctor")
	if err := cmd.Run(); err != nil {
		// This will fail until we implement main.go, which is expected for TDD
		// The tests should run but fail gracefully
	}

	// Run tests
	code := m.Run()

	// Cleanup
	os.Remove("./bin/goctor")

	os.Exit(code)
}
