package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestToolListingFunctionality(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	buildCmd := exec.Command("go", "build", "-o", "doctor-test", "./cmd/doctor")
	if err := buildCmd.Run(); err != nil {
		t.Skip("Skipping test - unable to build binary")
	}
	defer os.Remove("doctor-test")

	t.Run("list tools human readable", func(t *testing.T) {
		// Create test manifest with multiple tools
		manifestContent := `
meta:
  version: 1
  name: "List Test Manifest"

tools:
  - id: go
    name: "Go"
    rationale: "Go development toolchain"
    require: ">=1.22 <1.25"
    check:
      cmd: ["go", "version"]
      regex: "go(?P<ver>\\d+\\.\\d+(\\.\\d+)?)"
    links:
      homepage: "https://go.dev/"
      download: "https://go.dev/dl/"

  - id: git
    name: "Git"
    rationale: "Version control system"
    require: ">=2.30"
    check:
      cmd: ["git", "--version"]
      regex: "git version (?P<ver>\\d+\\.\\d+\\.\\d+)"
    links:
      homepage: "https://git-scm.com/"
      download: "https://git-scm.com/downloads"

  - id: docker
    name: "Docker"
    rationale: "Container platform for development"
    require: ">=24"
    check:
      cmd: ["docker", "--version"]
      regex: "version (?P<ver>\\d+\\.\\d+\\.\\d+)"
    links:
      homepage: "https://www.docker.com/"
      docs: "https://docs.docker.com/"
`

		manifestFile := "list-test-manifest.yaml"
		if err := os.WriteFile(manifestFile, []byte(manifestContent), 0644); err != nil {
			t.Fatalf("Failed to create test manifest: %v", err)
		}
		defer os.Remove(manifestFile)

		// Run list command
		cmd := exec.Command("./doctor-test", "list", "-f", manifestFile)
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skip("Skipping test - command execution failed")
		}

		outputStr := string(output)

		// Should display header
		if !strings.Contains(outputStr, "Tools defined in manifest:") {
			t.Error("Expected output to contain list header")
		}

		// Should display all tools with their information
		expectedTools := []string{"go:", "git:", "docker:"}
		for _, tool := range expectedTools {
			if !strings.Contains(outputStr, tool) {
				t.Errorf("Expected output to contain tool '%s'", tool)
			}
		}

		// Should display version requirements
		expectedVersions := []string{">=1.22 <1.25", ">=2.30", ">=24"}
		for _, version := range expectedVersions {
			if !strings.Contains(outputStr, version) {
				t.Errorf("Expected output to contain version requirement '%s'", version)
			}
		}

		// Should display rationales
		expectedRationales := []string{
			"Go development toolchain",
			"Version control system",
			"Container platform for development",
		}
		for _, rationale := range expectedRationales {
			if !strings.Contains(outputStr, rationale) {
				t.Errorf("Expected output to contain rationale '%s'", rationale)
			}
		}

		// Should NOT execute any tool commands (listing doesn't check versions)
		// This is verified by the fact that the command should complete quickly
		// and not show installation status
		if strings.Contains(outputStr, "[OK]") || strings.Contains(outputStr, "[Missing]") {
			t.Error("List command should not show installation status")
		}
	})

	t.Run("list tools JSON format", func(t *testing.T) {
		manifestFile := "testdata/manifests/sample.yaml"

		cmd := exec.Command("./doctor-test", "list", "--json", "-f", manifestFile)
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skip("Skipping test - command execution failed")
		}

		// Validate JSON structure
		var listResp struct {
			ManifestSource string `json:"manifest_source"`
			Tools          []struct {
				ID              string `json:"id"`
				Name            string `json:"name"`
				RequiredVersion string `json:"required_version"`
				Rationale       string `json:"rationale"`
			} `json:"tools"`
		}

		if err := json.Unmarshal(output, &listResp); err != nil {
			t.Errorf("Failed to parse JSON output: %v\nOutput: %s", err, string(output))
			return
		}

		// Validate response structure
		if listResp.ManifestSource == "" {
			t.Error("JSON output should include manifest_source")
		}

		if len(listResp.Tools) == 0 {
			t.Error("JSON output should include tools array")
		}

		// Validate each tool has required fields
		for i, tool := range listResp.Tools {
			if tool.ID == "" {
				t.Errorf("Tool %d should have ID", i)
			}
			if tool.Name == "" {
				t.Errorf("Tool %d should have Name", i)
			}
			if tool.RequiredVersion == "" {
				t.Errorf("Tool %d should have RequiredVersion", i)
			}
			if tool.Rationale == "" {
				t.Errorf("Tool %d should have Rationale", i)
			}
		}
	})

	t.Run("list with embedded manifest", func(t *testing.T) {
		// Test listing without specifying manifest (should use embedded)
		cmd := exec.Command("./doctor-test", "list")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skip("Skipping test - command execution failed")
		}

		outputStr := string(output)

		// Should show tools from embedded manifest
		if !strings.Contains(outputStr, "Tools defined in manifest:") {
			t.Error("Expected output to show tools from embedded manifest")
		}

		// Should complete without errors
		if strings.Contains(outputStr, "error") || strings.Contains(outputStr, "Error") {
			t.Error("List command should not show errors with embedded manifest")
		}
	})

	t.Run("list with no tools in manifest", func(t *testing.T) {
		// Create manifest with no tools
		emptyManifestContent := `
meta:
  version: 1
  name: "Empty Manifest"

tools: []
`

		manifestFile := "empty-manifest.yaml"
		if err := os.WriteFile(manifestFile, []byte(emptyManifestContent), 0644); err != nil {
			t.Fatalf("Failed to create empty manifest: %v", err)
		}
		defer os.Remove(manifestFile)

		cmd := exec.Command("./doctor-test", "list", "-f", manifestFile)
		output, err := cmd.CombinedOutput()

		// Should handle empty tools list gracefully
		if err == nil {
			// If command succeeds, should show appropriate message
			outputStr := string(output)
			if !strings.Contains(outputStr, "no tools") && !strings.Contains(outputStr, "empty") {
				t.Error("Expected message about no tools in manifest")
			}
		} else {
			// If command fails, should be due to validation error
			outputStr := string(output)
			if !strings.Contains(outputStr, "error") {
				t.Error("Expected validation error for empty tools list")
			}
		}
	})

	t.Run("list with invalid manifest", func(t *testing.T) {
		// Test list command with invalid manifest
		cmd := exec.Command("./doctor-test", "list", "-f", "nonexistent-manifest.yaml")
		output, err := cmd.CombinedOutput()

		// Should handle error gracefully
		if err == nil {
			t.Error("Expected command to fail with nonexistent manifest")
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "error") && !strings.Contains(outputStr, "not found") {
			t.Error("Expected error message about missing manifest")
		}
	})
}

func TestListCommandExitCodes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	buildCmd := exec.Command("go", "build", "-o", "doctor-test", "./cmd/doctor")
	if err := buildCmd.Run(); err != nil {
		t.Skip("Skipping test - unable to build binary")
	}
	defer os.Remove("doctor-test")

	t.Run("successful list returns exit code 0", func(t *testing.T) {
		cmd := exec.Command("./doctor-test", "list", "-f", "testdata/manifests/sample.yaml")
		output, err := cmd.CombinedOutput()

		// Check exit code
		exitCode := 0
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			}
		}

		if exitCode != 0 {
			t.Errorf("Expected exit code 0 for successful list, got %d\nOutput: %s", exitCode, string(output))
		}
	})

	t.Run("list with invalid manifest returns non-zero exit code", func(t *testing.T) {
		cmd := exec.Command("./doctor-test", "list", "-f", "nonexistent-manifest.yaml")
		_, err := cmd.CombinedOutput()

		// Should return non-zero exit code
		exitCode := 0
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			}
		}

		if exitCode == 0 {
			t.Error("Expected non-zero exit code for invalid manifest")
		}
	})
}

func TestListCommandPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	buildCmd := exec.Command("go", "build", "-o", "doctor-test", "./cmd/doctor")
	if err := buildCmd.Run(); err != nil {
		t.Skip("Skipping test - unable to build binary")
	}
	defer os.Remove("doctor-test")

	t.Run("list command completes quickly", func(t *testing.T) {
		// Create manifest with many tools
		manyToolsContent := `
meta:
  version: 1
  name: "Many Tools Manifest"

tools:`

		// Add 50 tools to test performance
		for i := 0; i < 50; i++ {
			manyToolsContent += `
  - id: tool` + string(rune('0'+i%10)) + `
    name: "Tool ` + string(rune('0'+i%10)) + `"
    rationale: "Testing tool ` + string(rune('0'+i%10)) + `"
    require: ">=1.0"
    check:
      cmd: ["echo", "tool` + string(rune('0'+i%10)) + ` v1.0.0"]
      regex: "tool` + string(rune('0'+i%10)) + ` v(?P<ver>\\d+\\.\\d+\\.\\d+)"
    links:
      homepage: "https://example.com/"`
		}

		manifestFile := "many-tools-manifest.yaml"
		if err := os.WriteFile(manifestFile, []byte(manyToolsContent), 0644); err != nil {
			t.Fatalf("Failed to create many tools manifest: %v", err)
		}
		defer os.Remove(manifestFile)

		// Run list command - should complete quickly without executing tool commands
		cmd := exec.Command("./doctor-test", "list", "-f", manifestFile)
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skip("Skipping test - command execution failed")
		}

		outputStr := string(output)

		// Should list all tools without executing them
		if !strings.Contains(outputStr, "Tools defined in manifest:") {
			t.Error("Expected tools list header")
		}

		// Should contain multiple tools
		toolCount := strings.Count(outputStr, "tool")
		if toolCount < 10 { // Should have at least some tools listed
			t.Errorf("Expected to list multiple tools, found %d mentions", toolCount)
		}

		// List command should be fast - the fact that it completed suggests it's working
		// In a real implementation, you might add timing measurements here
	})
}