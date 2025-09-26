package output

import (
	"fmt"
	"strings"

	"github.com/ikorihn/goctor/internal/checker"
	"github.com/ikorihn/goctor/internal/manifest"
)

// HumanFormatter provides human-readable output formatting
type HumanFormatter struct {
	colorEnabled bool
}

// NewHumanFormatter creates a new human-readable formatter
func NewHumanFormatter() *HumanFormatter {
	return &HumanFormatter{
		colorEnabled: true, // Can be disabled for non-terminal output
	}
}

// SetColorEnabled enables or disables color output
func (hf *HumanFormatter) SetColorEnabled(enabled bool) {
	hf.colorEnabled = enabled
}

// FormatEnvironmentReport formats a complete environment report
func (hf *HumanFormatter) FormatEnvironmentReport(report checker.EnvironmentReport) string {
	var output strings.Builder

	// Header
	output.WriteString(hf.formatHeader(report))
	output.WriteString("\n")

	// Summary
	output.WriteString(hf.formatSummary(report.Summary))
	output.WriteString("\n")

	// Individual tool results
	output.WriteString(hf.formatToolResults(report.Items))

	// Footer with recommendations
	if !report.IsSuccessful() {
		output.WriteString("\n")
		output.WriteString(hf.formatRecommendations(report.Items))
	}

	return output.String()
}

// FormatToolList formats a list of tools from a manifest
func (hf *HumanFormatter) FormatToolList(tools []manifest.ToolDefinition, manifestSource string) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("Tools defined in manifest (%s):\n\n", manifestSource))

	for i, tool := range tools {
		output.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, tool.Name, tool.ID))
		output.WriteString(fmt.Sprintf("   Required version: %s\n", tool.RequiredVersion))
		output.WriteString(fmt.Sprintf("   Rationale: %s\n", tool.Rationale))

		if len(tool.Links) > 0 {
			output.WriteString("   Links:\n")
			for linkType, url := range tool.Links {
				output.WriteString(fmt.Sprintf("     %s: %s\n", linkType, url))
			}
		}
		output.WriteString("\n")
	}

	return output.String()
}

// formatHeader creates the report header
func (hf *HumanFormatter) formatHeader(report checker.EnvironmentReport) string {
	var header strings.Builder

	header.WriteString("Development Environment Check\n")
	header.WriteString("============================\n")

	// Try to extract platform info if it's the right type
	if platformMap, ok := report.Platform.(map[string]interface{}); ok {
		if os, exists := platformMap["os"]; exists {
			if arch, exists := platformMap["arch"]; exists {
				header.WriteString(fmt.Sprintf("Platform: %s/%s\n", os, arch))
			}
		}
	}

	header.WriteString(fmt.Sprintf("Manifest: %s\n", report.ManifestSource))
	header.WriteString(fmt.Sprintf("Generated: %s\n", report.GeneratedAt.Format("2006-01-02 15:04:05")))

	return header.String()
}

// formatSummary creates the summary section
func (hf *HumanFormatter) formatSummary(summary checker.CheckSummary) string {
	var output strings.Builder

	output.WriteString("Summary:\n")
	output.WriteString("--------\n")

	output.WriteString(fmt.Sprintf("Total tools: %d\n", summary.Total))

	if summary.OK > 0 {
		output.WriteString(fmt.Sprintf("%s %d tools OK\n",
			hf.colorize("✓", "green"), summary.OK))
	}

	if summary.Missing > 0 {
		output.WriteString(fmt.Sprintf("%s %d tools missing\n",
			hf.colorize("✗", "red"), summary.Missing))
	}

	if summary.Outdated > 0 {
		output.WriteString(fmt.Sprintf("%s %d tools outdated\n",
			hf.colorize("⚠", "yellow"), summary.Outdated))
	}

	if summary.Errors > 0 {
		output.WriteString(fmt.Sprintf("%s %d tools with errors\n",
			hf.colorize("!", "red"), summary.Errors))
	}

	return output.String()
}

// formatToolResults creates the detailed tool results section
func (hf *HumanFormatter) formatToolResults(items []checker.CheckResult) string {
	var output strings.Builder

	output.WriteString("\nDetailed Results:\n")
	output.WriteString("-----------------\n")

	for _, item := range items {
		output.WriteString(hf.formatSingleResult(item))
		output.WriteString("\n")
	}

	return output.String()
}

// formatSingleResult formats a single tool check result
func (hf *HumanFormatter) formatSingleResult(result checker.CheckResult) string {
	var output strings.Builder

	// Status icon and tool name
	icon := hf.getStatusIcon(result.Status)
	output.WriteString(fmt.Sprintf("%s %s (%s)\n",
		icon, result.ToolName, result.ToolID))

	// Version information
	if result.ActualVersion != "" {
		output.WriteString(fmt.Sprintf("  Installed: %s\n", result.ActualVersion))
	}
	output.WriteString(fmt.Sprintf("  Required:  %s\n", result.RequiredVersion))

	// Path information
	if result.CommandPath != "" {
		output.WriteString(fmt.Sprintf("  Path:      %s\n", result.CommandPath))
	}

	// Error message if present
	if result.ErrorMessage != "" {
		output.WriteString(fmt.Sprintf("  %s %s\n",
			hf.colorize("Error:", "red"), result.ErrorMessage))
	}

	// Status-specific messages
	switch result.Status {
	case checker.StatusNotFound:
		output.WriteString("  Tool not found in PATH\n")
	case checker.StatusOutdated:
		output.WriteString("  Installed version does not meet requirements\n")
	}

	return output.String()
}

// formatRecommendations creates recommendations for failed checks
func (hf *HumanFormatter) formatRecommendations(items []checker.CheckResult) string {
	var output strings.Builder

	output.WriteString("Recommendations:\n")
	output.WriteString("----------------\n")

	for _, item := range items {
		if item.Status == checker.StatusOK {
			continue
		}

		output.WriteString(fmt.Sprintf("\n%s (%s):\n", item.ToolName, item.ToolID))

		switch item.Status {
		case checker.StatusNotFound:
			output.WriteString("  Install this tool to continue development\n")
		case checker.StatusOutdated:
			output.WriteString(fmt.Sprintf("  Update to version %s or later\n", item.RequiredVersion))
		case checker.StatusError:
			output.WriteString("  Check tool installation and PATH configuration\n")
		}

		// Add helpful links
		if len(item.Links) > 0 {
			output.WriteString("  Links:\n")
			for linkType, url := range item.Links {
				output.WriteString(fmt.Sprintf("    %s: %s\n", strings.Title(linkType), url))
			}
		}
	}

	return output.String()
}

// getStatusIcon returns an appropriate icon for the status
func (hf *HumanFormatter) getStatusIcon(status checker.CheckStatus) string {
	switch status {
	case checker.StatusOK:
		return hf.colorize("✓", "green")
	case checker.StatusNotFound, checker.StatusMissing:
		return hf.colorize("✗", "red")
	case checker.StatusOutdated:
		return hf.colorize("⚠", "yellow")
	case checker.StatusError:
		return hf.colorize("!", "red")
	default:
		return hf.colorize("?", "gray")
	}
}

// colorize applies color codes to text if colors are enabled
func (hf *HumanFormatter) colorize(text, color string) string {
	if !hf.colorEnabled {
		return text
	}

	colorCodes := map[string]string{
		"red":    "\033[31m",
		"green":  "\033[32m",
		"yellow": "\033[33m",
		"blue":   "\033[34m",
		"gray":   "\033[90m",
		"reset":  "\033[0m",
	}

	if code, exists := colorCodes[color]; exists {
		return code + text + colorCodes["reset"]
	}

	return text
}

// FormatQuickSummary provides a brief one-line summary
func (hf *HumanFormatter) FormatQuickSummary(summary checker.CheckSummary) string {
	if summary.Missing == 0 && summary.Outdated == 0 && summary.Errors == 0 {
		return hf.colorize(fmt.Sprintf("✓ All %d tools are ready", summary.Total), "green")
	}

	issues := summary.Missing + summary.Outdated + summary.Errors
	return hf.colorize(fmt.Sprintf("✗ %d of %d tools need attention", issues, summary.Total), "red")
}