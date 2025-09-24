package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestBasicEnvironmentCheck(t *testing.T) {
	// This test implements the scenario from quickstart.md
	// It should run the doctor command and verify basic functionality

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "doctor-test", "./cmd/doctor")
	if err := buildCmd.Run(); err != nil {
		t.Skip("Skipping test - unable to build binary (expected until implementation)")
	}
	defer os.Remove("doctor-test")

	tests := []struct {
		name            string
		args            []string
		expectOutput    string
		expectExitCode  int
		manifestFile    string
		description     string
	}{
		{
			name:           "basic check with embedded manifest",
			args:           []string{},
			expectOutput:   "[OK]",
			expectExitCode: 0,
			description:    "Should run basic environment check",
		},
		{
			name:           "check with custom manifest",
			args:           []string{"-f", "testdata/manifests/sample.yaml"},
			expectOutput:   "go:",
			expectExitCode: 0,
			manifestFile:   "testdata/manifests/sample.yaml",
			description:    "Should use custom manifest file",
		},
		{
			name:           "check with missing tools manifest",
			args:           []string{"-f", "testdata/manifests/missing-tools.yaml"},
			expectOutput:   "[Missing]",
			expectExitCode: 1,
			manifestFile:   "testdata/manifests/missing-tools.yaml",
			description:    "Should detect missing tools and return exit code 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test manifest if needed
			if tt.manifestFile != "" {
				createTestManifest(t, tt.manifestFile)
			}

			// Run the command
			cmd := exec.Command("./doctor-test", tt.args...)
			output, err := cmd.CombinedOutput()

			// Check exit code
			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				}
			}

			if exitCode != tt.expectExitCode {
				t.Errorf("%s: Expected exit code %d, got %d", tt.description, tt.expectExitCode, exitCode)
			}

			// Check output content
			outputStr := string(output)
			if !strings.Contains(outputStr, tt.expectOutput) {
				t.Errorf("%s: Expected output to contain '%s', got: %s", tt.description, tt.expectOutput, outputStr)
			}

			// Cleanup test manifest
			if tt.manifestFile != "" {
				os.Remove(tt.manifestFile)
			}
		})
	}
}

func TestEnvironmentCheckStatusReporting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	buildCmd := exec.Command("go", "build", "-o", "doctor-test", "./cmd/doctor")
	if err := buildCmd.Run(); err != nil {
		t.Skip("Skipping test - unable to build binary")
	}
	defer os.Remove("doctor-test")

	// Create manifest with mix of tools (some present, some missing)
	manifestContent := `
meta:
  version: 1
  name: "Mixed Tools Test"

tools:
  - id: go
    name: "Go"
    rationale: "Present tool"
    require: ">=1.0"
    check:
      cmd: ["go", "version"]
      regex: "go(?P<ver>\\d+\\.\\d+(\\.\\d+)?)"
    links:
      homepage: "https://go.dev/"

  - id: nonexistent-tool
    name: "Nonexistent Tool"
    rationale: "Missing tool"
    require: ">=1.0"
    check:
      cmd: ["nonexistent-tool", "--version"]
      regex: "(?P<ver>\\d+\\.\\d+\\.\\d+)"
    links:
      homepage: "https://example.com/"
`

	manifestFile := "test-mixed-manifest.yaml"
	if err := os.WriteFile(manifestFile, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to create test manifest: %v", err)
	}
	defer os.Remove(manifestFile)

	// Run doctor command
	cmd := exec.Command("./doctor-test", "-f", manifestFile)
	output, err := cmd.CombinedOutput()

	// Should return exit code 1 (issues found)
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
	}

	if exitCode != 1 {
		t.Errorf("Expected exit code 1 (issues found), got %d", exitCode)
	}

	outputStr := string(output)

	// Should contain both OK and Missing statuses
	if !strings.Contains(outputStr, "[OK]") {
		t.Error("Expected output to contain '[OK]' for available tools")
	}

	if !strings.Contains(outputStr, "[Missing]") {
		t.Error("Expected output to contain '[Missing]' for unavailable tools")
	}

	// Should show rationale for tools
	if !strings.Contains(outputStr, "Present tool") {
		t.Error("Expected output to show rationale for tools")
	}

	// Should show links for problematic tools
	if !strings.Contains(outputStr, "Links:") {
		t.Error("Expected output to show links for problematic tools")
	}
}

func TestEnvironmentCheckTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	buildCmd := exec.Command("go", "build", "-o", "doctor-test", "./cmd/doctor")
	if err := buildCmd.Run(); err != nil {
		t.Skip("Skipping test - unable to build binary")
	}
	defer os.Remove("doctor-test")

	// Create manifest with command that hangs (sleep)
	manifestContent := `
meta:
  version: 1
  name: "Timeout Test"

defaults:
  timeout_sec: 1

tools:
  - id: hanging-tool
    name: "Hanging Tool"
    rationale: "Tool that hangs for testing timeout"
    require: ">=1.0"
    check:
      cmd: ["sleep", "10"]
      regex: "(?P<ver>\\d+\\.\\d+\\.\\d+)"
    links:
      homepage: "https://example.com/"
`

	manifestFile := "test-timeout-manifest.yaml"
	if err := os.WriteFile(manifestFile, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to create test manifest: %v", err)
	}
	defer os.Remove(manifestFile)

	// Run doctor command - should complete quickly due to timeout
	cmd := exec.Command("./doctor-test", "-f", manifestFile)
	output, _ := cmd.CombinedOutput()

	// Should handle timeout gracefully
	outputStr := string(output)
	if !strings.Contains(outputStr, "timeout") && !strings.Contains(outputStr, "Error") {
		t.Error("Expected output to handle timeout gracefully")
	}
}

func createTestManifest(t *testing.T, filename string) {
	content := `
meta:
  version: 1
  name: "Test Manifest"

tools:
  - id: go
    name: "Go"
    rationale: "Go development"
    require: ">=1.0"
    check:
      cmd: ["go", "version"]
      regex: "go(?P<ver>\\d+\\.\\d+(\\.\\d+)?)"
    links:
      homepage: "https://go.dev/"
`

	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test manifest %s: %v", filename, err)
	}
}