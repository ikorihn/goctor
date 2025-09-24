package manifest

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Loader handles loading and parsing of manifest files
type Loader struct {
	httpClient *http.Client
}

// NewLoader creates a new manifest loader with default configuration
func NewLoader() *Loader {
	return &Loader{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// LoadFromFile loads a manifest from a local file
func (l *Loader) LoadFromFile(filePath string) (*Manifest, error) {
	if filePath == "" {
		return nil, errors.New("file path cannot be empty")
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("manifest file not found: %s", filePath)
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file %s: %v", filePath, err)
	}

	// Parse YAML
	manifest, err := l.parseYAML(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest file %s: %v", filePath, err)
	}

	return manifest, nil
}

// LoadFromURL loads a manifest from a remote URL
func (l *Loader) LoadFromURL(url string) (*Manifest, error) {
	if url == "" {
		return nil, errors.New("URL cannot be empty")
	}

	// Validate URL format
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("invalid URL format: %s", url)
	}

	// Make HTTP request
	resp, err := l.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest from %s: %v", url, err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch manifest from %s: HTTP %d", url, resp.StatusCode)
	}

	// Read response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response from %s: %v", url, err)
	}

	// Parse YAML
	manifest, err := l.parseYAML(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest from %s: %v", url, err)
	}

	return manifest, nil
}

// LoadFromSource loads a manifest from either a file path or URL
func (l *Loader) LoadFromSource(source string) (*Manifest, error) {
	if source == "" {
		return nil, errors.New("source cannot be empty")
	}

	// Determine if source is URL or file path
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return l.LoadFromURL(source)
	}

	return l.LoadFromFile(source)
}

// LoadEmbedded loads a manifest from embedded data (for default manifest)
func (l *Loader) LoadEmbedded(data []byte) (*Manifest, error) {
	if len(data) == 0 {
		return nil, errors.New("embedded data cannot be empty")
	}

	manifest, err := l.parseYAML(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded manifest: %v", err)
	}

	return manifest, nil
}

// parseYAML parses YAML data into a Manifest struct
func (l *Loader) parseYAML(data []byte) (*Manifest, error) {
	var manifest Manifest

	// Parse YAML
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("YAML parsing error: %v", err)
	}

	// Apply defaults to tools
	manifest.ApplyDefaults()

	// Validate the manifest
	if err := manifest.Validate(); err != nil {
		return nil, fmt.Errorf("manifest validation failed: %v", err)
	}

	return &manifest, nil
}

// MergeManifests merges multiple manifests with later ones taking precedence
func (l *Loader) MergeManifests(manifests ...*Manifest) (*Manifest, error) {
	if len(manifests) == 0 {
		return nil, errors.New("no manifests to merge")
	}

	// Start with the first manifest
	result := *manifests[0]

	// Merge each subsequent manifest
	for i := 1; i < len(manifests); i++ {
		if manifests[i] == nil {
			continue
		}
		result = result.Merge(*manifests[i])
	}

	// Apply defaults and validate the merged result
	result.ApplyDefaults()
	if err := result.Validate(); err != nil {
		return nil, fmt.Errorf("merged manifest validation failed: %v", err)
	}

	return &result, nil
}

// LoadMultipleSources loads and merges manifests from multiple sources
func (l *Loader) LoadMultipleSources(sources ...string) (*Manifest, error) {
	if len(sources) == 0 {
		return nil, errors.New("no sources provided")
	}

	manifests := make([]*Manifest, 0, len(sources))

	for i, source := range sources {
		if source == "" {
			continue
		}

		manifest, err := l.LoadFromSource(source)
		if err != nil {
			return nil, fmt.Errorf("failed to load manifest %d from %s: %v", i, source, err)
		}

		manifests = append(manifests, manifest)
	}

	if len(manifests) == 0 {
		return nil, errors.New("no valid manifests loaded")
	}

	return l.MergeManifests(manifests...)
}

// ValidateManifestStructure performs basic structural validation
func (l *Loader) ValidateManifestStructure(data []byte) error {
	var rawManifest map[string]interface{}

	if err := yaml.Unmarshal(data, &rawManifest); err != nil {
		return fmt.Errorf("invalid YAML structure: %v", err)
	}

	// Check for required top-level sections
	if _, exists := rawManifest["meta"]; !exists {
		return errors.New("missing required 'meta' section")
	}

	if _, exists := rawManifest["tools"]; !exists {
		return errors.New("missing required 'tools' section")
	}

	// Validate meta section structure
	meta, ok := rawManifest["meta"].(map[string]interface{})
	if !ok {
		return errors.New("'meta' section must be an object")
	}

	if _, exists := meta["version"]; !exists {
		return errors.New("missing required 'meta.version' field")
	}

	if _, exists := meta["name"]; !exists {
		return errors.New("missing required 'meta.name' field")
	}

	// Validate tools section structure
	tools, ok := rawManifest["tools"].([]interface{})
	if !ok {
		return errors.New("'tools' section must be an array")
	}

	if len(tools) == 0 {
		return errors.New("'tools' array cannot be empty")
	}

	// Validate each tool structure
	for i, tool := range tools {
		toolMap, ok := tool.(map[string]interface{})
		if !ok {
			return fmt.Errorf("tool %d must be an object", i)
		}

		requiredFields := []string{"id", "name", "rationale", "require", "check", "links"}
		for _, field := range requiredFields {
			if _, exists := toolMap[field]; !exists {
				return fmt.Errorf("tool %d missing required field: %s", i, field)
			}
		}
	}

	return nil
}

// GetEmbeddedManifest returns the default embedded manifest
func GetEmbeddedManifest() []byte {
	// This would typically be embedded using go:embed or similar
	// For now, we'll return a basic manifest
	return []byte(`
meta:
  version: 1
  name: "Default Development Tools"
  language: "en"

defaults:
  timeout_sec: 5
  regex_key: "ver"

tools:
  - id: go
    name: "Go"
    rationale: "Go development toolchain"
    require: ">=1.20"
    check:
      cmd: ["go", "version"]
      regex: "go(?P<ver>\\d+\\.\\d+(\\.\\d+)?)"
    links:
      homepage: "https://go.dev/"
      download: "https://go.dev/dl/"
      docs: "https://go.dev/doc/"

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
`)
}

// SetHTTPTimeout sets the timeout for HTTP requests
func (l *Loader) SetHTTPTimeout(timeout time.Duration) {
	l.httpClient.Timeout = timeout
}

// SetHTTPClient allows setting a custom HTTP client
func (l *Loader) SetHTTPClient(client *http.Client) {
	l.httpClient = client
}