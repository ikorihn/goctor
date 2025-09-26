package output

import (
	"encoding/json"
	"time"

	"github.com/ikorihn/goctor/internal/checker"
	"github.com/ikorihn/goctor/internal/manifest"
)

// JSONFormatter provides JSON output formatting
type JSONFormatter struct {
	prettyPrint bool
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{
		prettyPrint: true,
	}
}

// SetPrettyPrint enables or disables pretty-printed JSON output
func (jf *JSONFormatter) SetPrettyPrint(enabled bool) {
	jf.prettyPrint = enabled
}

// FormatEnvironmentReport formats a complete environment report as JSON
func (jf *JSONFormatter) FormatEnvironmentReport(report checker.EnvironmentReport) (string, error) {
	// Convert to JSON-friendly format
	jsonReport := JSONEnvironmentReport{
		SchemaVersion:  report.SchemaVersion,
		Platform:       report.Platform,
		Summary:        report.Summary,
		ManifestSource: report.ManifestSource,
		Items:          make([]JSONCheckResult, len(report.Items)),
		GeneratedAt:    report.GeneratedAt,
	}

	// Convert check results
	for i, item := range report.Items {
		jsonReport.Items[i] = jf.convertCheckResult(item)
	}

	return jf.marshalJSON(jsonReport)
}

// FormatToolList formats a list of tools as JSON
func (jf *JSONFormatter) FormatToolList(tools []manifest.ToolDefinition, manifestSource string) (string, error) {
	response := JSONToolListResponse{
		ManifestSource: manifestSource,
		Tools:          make([]JSONTool, len(tools)),
		GeneratedAt:    time.Now(),
	}

	for i, tool := range tools {
		response.Tools[i] = JSONTool{
			ID:              tool.ID,
			Name:            tool.Name,
			RequiredVersion: tool.RequiredVersion,
			Rationale:       tool.Rationale,
			CheckCommand:    tool.CheckCommand(),
			VersionRegex:    tool.VersionRegex(),
			Links:           tool.Links,
			TimeoutSeconds:  tool.TimeoutSeconds,
		}
	}

	return jf.marshalJSON(response)
}

// FormatCheckSummary formats just the summary as JSON
func (jf *JSONFormatter) FormatCheckSummary(summary checker.CheckSummary) (string, error) {
	return jf.marshalJSON(summary)
}

// convertCheckResult converts internal CheckResult to JSON-friendly format
func (jf *JSONFormatter) convertCheckResult(result checker.CheckResult) JSONCheckResult {
	return JSONCheckResult{
		ToolID:          result.ToolID,
		ToolName:        result.ToolName,
		Status:          result.Status.String(),
		RequiredVersion: result.RequiredVersion,
		ActualVersion:   result.ActualVersion,
		ErrorMessage:    result.ErrorMessage,
		Platform:        result.Platform,
		Links:           result.Links,
		CheckDuration:   result.CheckDuration,
	}
}

// marshalJSON marshals data to JSON with appropriate formatting
func (jf *JSONFormatter) marshalJSON(data interface{}) (string, error) {
	var jsonData []byte
	var err error

	if jf.prettyPrint {
		jsonData, err = json.MarshalIndent(data, "", "  ")
	} else {
		jsonData, err = json.Marshal(data)
	}

	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// JSONEnvironmentReport represents the JSON structure for environment reports
type JSONEnvironmentReport struct {
	SchemaVersion  int                `json:"schema_version"`
	Platform       interface{}        `json:"platform"`
	Summary        checker.CheckSummary `json:"summary"`
	ManifestSource string             `json:"manifest_source"`
	Items          []JSONCheckResult  `json:"items"`
	GeneratedAt    time.Time          `json:"generated_at"`
}

// JSONCheckResult represents the JSON structure for individual tool check results
type JSONCheckResult struct {
	ToolID          string            `json:"id"`
	ToolName        string            `json:"name"`
	Status          string            `json:"status"`
	RequiredVersion string            `json:"required_version"`
	ActualVersion   string            `json:"actual_version,omitempty"`
	ErrorMessage    string            `json:"error_message,omitempty"`
	Platform        string            `json:"platform"`
	Links           map[string]string `json:"links"`
	CheckDuration   time.Duration     `json:"check_duration_ms,omitempty"`
}

// JSONToolListResponse represents the JSON structure for tool list responses
type JSONToolListResponse struct {
	ManifestSource string     `json:"manifest_source"`
	Tools          []JSONTool `json:"tools"`
	GeneratedAt    time.Time  `json:"generated_at"`
}

// JSONTool represents the JSON structure for tool definitions
type JSONTool struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	RequiredVersion string            `json:"required_version"`
	Rationale       string            `json:"rationale"`
	CheckCommand    []string          `json:"check_command"`
	VersionRegex    string            `json:"version_regex"`
	Links           map[string]string `json:"links"`
	TimeoutSeconds  int               `json:"timeout_seconds,omitempty"`
}

// Validate validates the JSON environment report structure
func (jer *JSONEnvironmentReport) Validate() error {
	if jer.SchemaVersion != 1 {
		return &ValidationError{
			Field:   "schema_version",
			Message: "unsupported schema version",
		}
	}

	if jer.ManifestSource == "" {
		return &ValidationError{
			Field:   "manifest_source",
			Message: "manifest source cannot be empty",
		}
	}

	if jer.Items == nil {
		return &ValidationError{
			Field:   "items",
			Message: "items array cannot be nil",
		}
	}

	// Validate summary consistency
	if jer.Summary.Total != len(jer.Items) {
		return &ValidationError{
			Field:   "summary.total",
			Message: "total count does not match items length",
		}
	}

	return nil
}

// ValidationError represents JSON validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (ve *ValidationError) Error() string {
	return ve.Field + ": " + ve.Message
}

// FormatError formats an error response as JSON
func (jf *JSONFormatter) FormatError(err error, context string) (string, error) {
	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"message":     err.Error(),
			"context":     context,
			"timestamp":   time.Now(),
			"schema_version": 1,
		},
	}

	return jf.marshalJSON(errorResponse)
}

// FormatValidationErrors formats multiple validation errors as JSON
func (jf *JSONFormatter) FormatValidationErrors(errors []error) (string, error) {
	errorMessages := make([]string, len(errors))
	for i, err := range errors {
		errorMessages[i] = err.Error()
	}

	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"type":        "validation_error",
			"message":     "Multiple validation errors occurred",
			"details":     errorMessages,
			"timestamp":   time.Now(),
			"schema_version": 1,
		},
	}

	return jf.marshalJSON(errorResponse)
}