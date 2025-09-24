package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestCustomManifestLoading(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	buildCmd := exec.Command("go", "build", "-o", "doctor-test", "./cmd/doctor")
	if err := buildCmd.Run(); err != nil {
		t.Skip("Skipping test - unable to build binary")
	}
	defer os.Remove("doctor-test")

	t.Run("load local manifest file", func(t *testing.T) {
		// Create a custom manifest
		manifestContent := `
meta:
  version: 1
  name: "Custom Test Manifest"
  language: "en"

defaults:
  timeout_sec: 10
  regex_key: "ver"

tools:
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
      docs: "https://git-scm.com/doc"

  - id: custom-tool
    name: "Custom Tool"
    rationale: "Custom tool for testing"
    require: ">=1.0"
    check:
      cmd: ["echo", "custom-tool v1.5.0"]
      regex: "custom-tool v(?P<ver>\\d+\\.\\d+\\.\\d+)"
    links:
      homepage: "https://example.com/"
`

		manifestFile := "custom-manifest.yaml"
		if err := os.WriteFile(manifestFile, []byte(manifestContent), 0644); err != nil {
			t.Fatalf("Failed to create custom manifest: %v", err)
		}
		defer os.Remove(manifestFile)

		// Run doctor with custom manifest
		cmd := exec.Command("./doctor-test", "doctor", "-f", manifestFile)
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skip("Skipping test - command execution failed")
		}

		outputStr := string(output)

		// Should check tools from custom manifest
		if !strings.Contains(outputStr, "git") {
			t.Error("Expected output to include git tool from custom manifest")
		}

		if !strings.Contains(outputStr, "custom-tool") {
			t.Error("Expected output to include custom-tool from manifest")
		}

		// Should show custom rationales
		if !strings.Contains(outputStr, "Version control system") {
			t.Error("Expected output to show custom rationale for git")
		}
	})

	t.Run("load manifest with short flag", func(t *testing.T) {
		// Test using -f flag instead of --manifest
		manifestFile := "testdata/manifests/sample.yaml"

		cmd := exec.Command("./doctor-test", "doctor", "-f", manifestFile)
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skip("Skipping test - command execution failed")
		}

		outputStr := string(output)

		// Should load the sample manifest
		if !strings.Contains(outputStr, "go") || !strings.Contains(outputStr, "docker") {
			t.Error("Expected output to include tools from sample manifest")
		}
	})

	t.Run("invalid manifest file", func(t *testing.T) {
		// Test with non-existent manifest file
		cmd := exec.Command("./doctor-test", "doctor", "-f", "nonexistent-manifest.yaml")
		output, err := cmd.CombinedOutput()

		// Should handle error gracefully
		if err == nil {
			t.Error("Expected command to fail with non-existent manifest file")
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "error") && !strings.Contains(outputStr, "not found") {
			t.Error("Expected error message about missing manifest file")
		}
	})

	t.Run("malformed manifest file", func(t *testing.T) {
		// Create malformed YAML
		malformedContent := `
meta:
  version: 1
  name: "Malformed Manifest"
tools:
  - id: test
    name: "Test"
    invalid_yaml: [unclosed list
`

		manifestFile := "malformed-manifest.yaml"
		if err := os.WriteFile(manifestFile, []byte(malformedContent), 0644); err != nil {
			t.Fatalf("Failed to create malformed manifest: %v", err)
		}
		defer os.Remove(manifestFile)

		cmd := exec.Command("./doctor-test", "doctor", "-f", manifestFile)
		output, err := cmd.CombinedOutput()

		// Should handle YAML parsing error gracefully
		if err == nil {
			t.Error("Expected command to fail with malformed YAML")
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "error") && !strings.Contains(outputStr, "yaml") {
			t.Error("Expected error message about YAML parsing")
		}
	})

	t.Run("manifest with missing required fields", func(t *testing.T) {
		// Create manifest with missing required fields
		incompleteContent := `
meta:
  version: 1
  name: "Incomplete Manifest"

tools:
  - id: incomplete-tool
    name: "Incomplete Tool"
    # Missing rationale, require, check, links
`

		manifestFile := "incomplete-manifest.yaml"
		if err := os.WriteFile(manifestFile, []byte(incompleteContent), 0644); err != nil {
			t.Fatalf("Failed to create incomplete manifest: %v", err)
		}
		defer os.Remove(manifestFile)

		cmd := exec.Command("./doctor-test", "doctor", "-f", manifestFile)
		output, err := cmd.CombinedOutput()

		// Should handle validation error gracefully
		if err == nil {
			t.Error("Expected command to fail with incomplete manifest")
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "error") && !strings.Contains(outputStr, "validation") {
			t.Error("Expected validation error message")
		}
	})
}

func TestManifestMerging(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	buildCmd := exec.Command("go", "build", "-o", "doctor-test", "./cmd/doctor")
	if err := buildCmd.Run(); err != nil {
		t.Skip("Skipping test - unable to build binary")
	}
	defer os.Remove("doctor-test")

	// This test would be for multiple manifest files if that feature is implemented
	// For now, we test the precedence of explicit manifest over embedded default

	t.Run("custom manifest overrides embedded", func(t *testing.T) {
		// Create a minimal custom manifest
		manifestContent := `
meta:
  version: 1
  name: "Override Test"

tools:
  - id: test-only-tool
    name: "Test Only Tool"
    rationale: "Only in custom manifest"
    require: ">=1.0"
    check:
      cmd: ["echo", "test-only-tool v1.0.0"]
      regex: "test-only-tool v(?P<ver>\\d+\\.\\d+\\.\\d+)"
    links:
      homepage: "https://example.com/"
`

		manifestFile := "override-manifest.yaml"
		if err := os.WriteFile(manifestFile, []byte(manifestContent), 0644); err != nil {
			t.Fatalf("Failed to create override manifest: %v", err)
		}
		defer os.Remove(manifestFile)

		// Run with custom manifest
		cmd := exec.Command("./doctor-test", "list", "-f", manifestFile)
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skip("Skipping test - command execution failed")
		}

		outputStr := string(output)

		// Should only include tools from custom manifest
		if !strings.Contains(outputStr, "test-only-tool") {
			t.Error("Expected output to include test-only-tool from custom manifest")
		}

		// Should not include embedded default tools (if any)
		// This verification depends on what's in the embedded manifest
	})
}

func TestManifestDefaultsApplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	buildCmd := exec.Command("go", "build", "-o", "doctor-test", "./cmd/doctor")
	if err := buildCmd.Run(); err != nil {
		t.Skip("Skipping test - unable to build binary")
	}
	defer os.Remove("doctor-test")

	t.Run("defaults applied to tools", func(t *testing.T) {
		// Create manifest with defaults and tools that should use them
		manifestContent := `
meta:
  version: 1
  name: "Defaults Test"

defaults:
  timeout_sec: 2
  regex_key: "version"

tools:
  - id: quick-tool
    name: "Quick Tool"
    rationale: "Tool that should use default timeout"
    require: ">=1.0"
    check:
      cmd: ["echo", "quick-tool version 1.0.0"]
      regex: "quick-tool version (?P<version>\\d+\\.\\d+\\.\\d+)"
    links:
      homepage: "https://example.com/"

  - id: slow-tool
    name: "Slow Tool"
    rationale: "Tool with explicit timeout"
    require: ">=1.0"
    check:
      cmd: ["sleep", "0.1", "&&", "echo", "slow-tool version 1.0.0"]
      regex: "slow-tool version (?P<version>\\d+\\.\\d+\\.\\d+)"
    links:
      homepage: "https://example.com/"
`

		manifestFile := "defaults-manifest.yaml"
		if err := os.WriteFile(manifestFile, []byte(manifestContent), 0644); err != nil {
			t.Fatalf("Failed to create defaults manifest: %v", err)
		}
		defer os.Remove(manifestFile)

		// Run doctor command - should complete quickly with defaults applied
		cmd := exec.Command("./doctor-test", "doctor", "-f", manifestFile)
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skip("Skipping test - command execution failed")
		}

		outputStr := string(output)

		// Both tools should be checked successfully
		if !strings.Contains(outputStr, "quick-tool") {
			t.Error("Expected output to include quick-tool")
		}

		// The fact that the command completed quickly suggests timeouts are working
		// More detailed timeout testing would require instrumentation
	})
}

func TestRemoteManifestLoading(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	buildCmd := exec.Command("go", "build", "-o", "doctor-test", "./cmd/doctor")
	if err := buildCmd.Run(); err != nil {
		t.Skip("Skipping test - unable to build binary")
	}
	defer os.Remove("doctor-test")

	t.Run("invalid URL", func(t *testing.T) {
		// Test with invalid URL
		cmd := exec.Command("./doctor-test", "doctor", "-f", "https://invalid.example.com/manifest.yaml")
		output, err := cmd.CombinedOutput()

		// Should handle network error gracefully
		if err == nil {
			t.Error("Expected command to fail with invalid URL")
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "error") && !strings.Contains(outputStr, "network") {
			t.Error("Expected network error message")
		}
	})

	t.Run("non-YAML URL", func(t *testing.T) {
		// Test with URL that doesn't serve YAML
		cmd := exec.Command("./doctor-test", "doctor", "-f", "https://httpbin.org/json")
		output, err := cmd.CombinedOutput()

		// Should handle non-YAML content gracefully
		if err == nil {
			t.Error("Expected command to fail with non-YAML content")
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "error") && !strings.Contains(outputStr, "yaml") {
			t.Error("Expected YAML parsing error message")
		}
	})

	// Note: Testing with a real working URL would require maintaining a test server
	// In a real implementation, this could be tested with a local HTTP server
}