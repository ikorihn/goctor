package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/ikorihn/goctor/internal/checker"
	"github.com/ikorihn/goctor/internal/manifest"
	"github.com/ikorihn/goctor/internal/output"
	"github.com/ikorihn/goctor/internal/platform"
)

const (
	version = "1.0.0"
)

func main() {
	var (
		manifestFlag = flag.String("f", "", "manifest file path or URL")
		jsonFlag     = flag.Bool("json", false, "output JSON format")
		helpFlag     = flag.Bool("h", false, "show help")
		versionFlag  = flag.Bool("v", false, "show version")
	)

	flag.Parse()

	if *helpFlag {
		showHelp()
		return
	}

	if *versionFlag {
		fmt.Printf("doctor version %s\n", version)
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		args = []string{"doctor"} // Default command
	}

	command := args[0]

	switch command {
	case "doctor":
		exitCode := runDoctorCommand(*manifestFlag, *jsonFlag)
		os.Exit(exitCode)
	case "list":
		exitCode := runListCommand(*manifestFlag, *jsonFlag)
		os.Exit(exitCode)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		showHelp()
		os.Exit(1)
	}
}

func runDoctorCommand(manifestSource string, useJSON bool) int {
	// Load manifest
	loader := manifest.NewLoader()
	var m *manifest.Manifest
	var err error

	if manifestSource == "" {
		// Default to ./tools.yaml
		manifestSource = "./tools.yaml"
	}

	m, err = loader.LoadFromSource(manifestSource)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading manifest: %v\n", err)
		return 1
	}

	// Detect platform
	platformInfo := platform.DetectPlatform()
	if !platformInfo.IsSupported() {
		fmt.Fprintf(os.Stderr, "Unsupported platform: %s\n", platformInfo.String())
		return 1
	}

	// Create checker and run checks
	toolChecker := checker.NewChecker()
	results := make([]checker.CheckResult, len(m.Tools))

	for i, tool := range m.Tools {
		result := toolChecker.CheckTool(tool, platformInfo)
		results[i] = result
	}

	// Generate report
	report := checker.NewEnvironmentReport(platformInfo, manifestSource, results)

	// Output results
	if useJSON {
		jsonData, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating JSON output: %v\n", err)
			return 1
		}
		fmt.Println(string(jsonData))
	} else {
		formatter := output.NewHumanFormatter()
		output := formatter.FormatEnvironmentReport(*report)
		fmt.Print(output)
	}

	return report.GetExitCode()
}

func runListCommand(manifestSource string, useJSON bool) int {
	// Load manifest
	loader := manifest.NewLoader()
	var m *manifest.Manifest
	var err error

	if manifestSource == "" {
		// Default to ./tools.yaml
		manifestSource = "./tools.yaml"
	}

	m, err = loader.LoadFromSource(manifestSource)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading manifest: %v\n", err)
		return 1
	}

	// Output tool list
	if useJSON {
		listResponse := struct {
			ManifestSource string `json:"manifest_source"`
			Tools          []struct {
				ID              string `json:"id"`
				Name            string `json:"name"`
				RequiredVersion string `json:"required_version"`
				Rationale       string `json:"rationale"`
			} `json:"tools"`
		}{
			ManifestSource: manifestSource,
			Tools:          make([]struct {
				ID              string `json:"id"`
				Name            string `json:"name"`
				RequiredVersion string `json:"required_version"`
				Rationale       string `json:"rationale"`
			}, len(m.Tools)),
		}

		for i, tool := range m.Tools {
			listResponse.Tools[i] = struct {
				ID              string `json:"id"`
				Name            string `json:"name"`
				RequiredVersion string `json:"required_version"`
				Rationale       string `json:"rationale"`
			}{
				ID:              tool.ID,
				Name:            tool.Name,
				RequiredVersion: tool.RequiredVersion,
				Rationale:       tool.Rationale,
			}
		}

		jsonData, err := json.MarshalIndent(listResponse, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating JSON output: %v\n", err)
			return 1
		}
		fmt.Println(string(jsonData))
	} else {
		formatter := output.NewHumanFormatter()
		output := formatter.FormatToolList(m.Tools, manifestSource)
		fmt.Print(output)
	}

	return 0
}

func showHelp() {
	fmt.Print(`doctor - Development Environment Checker

USAGE:
    doctor [command] [flags]

COMMANDS:
    doctor    Check development environment (default)
    list      List tools defined in manifest

FLAGS:
    -f, --manifest PATH_OR_URL    Manifest file path or URL
    --json                        Output JSON format
    -h, --help                    Show help
    -v, --version                 Show version

EXAMPLES:
    doctor                                    # Check using ./tools.yaml
    doctor -f custom-manifest.yaml           # Check using custom manifest
    doctor --json                            # Output JSON format
    list                                     # List tools in ./tools.yaml
    list -f https://company.com/manifest.yaml # List tools from remote manifest
`)
}