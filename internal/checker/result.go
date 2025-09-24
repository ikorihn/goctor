package checker

import (
	"errors"
	"fmt"
	"time"
)

// CheckStatus represents the possible states of a tool check
type CheckStatus int

const (
	StatusUnknown CheckStatus = iota
	StatusOK
	StatusMissing
	StatusOutdated
	StatusError
	StatusNotFound // Alias for StatusMissing for backwards compatibility
)

// ErrorType represents different categories of check errors
type ErrorType int

const (
	ErrorTypeConfiguration ErrorType = iota
	ErrorTypeExecution
	ErrorTypeParsing
	ErrorTypeTimeout
	ErrorTypeVersionMismatch
)

// String returns the string representation of the check status
func (cs CheckStatus) String() string {
	switch cs {
	case StatusOK:
		return "ok"
	case StatusMissing:
		return "missing"
	case StatusNotFound:
		return "not_found"
	case StatusOutdated:
		return "outdated"
	case StatusError:
		return "error"
	case StatusUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// CheckError represents an error that occurred during tool checking
type CheckError struct {
	Message string
	Type    ErrorType
}

func (ce CheckError) Error() string {
	return ce.Message
}

// NewCheckError creates a new CheckError with the specified message and type
func NewCheckError(message string, errorType ErrorType) CheckError {
	return CheckError{
		Message: message,
		Type:    errorType,
	}
}

// CheckResult represents the outcome of verifying a single tool installation
type CheckResult struct {
	ToolID          string            `json:"id"`
	ToolName        string            `json:"name"`
	Status          CheckStatus       `json:"status"`
	RequiredVersion string            `json:"required"`
	ActualVersion   string            `json:"actual_version"`
	ErrorMessage    string            `json:"error_message,omitempty"`
	Platform        string            `json:"platform"`
	Links           map[string]string `json:"links"`
	CheckDuration   time.Duration     `json:"check_duration,omitempty"`
}

// EnvironmentReport represents a comprehensive summary of all tool checks
type EnvironmentReport struct {
	SchemaVersion  int           `json:"schema_version"`
	Platform       interface{}   `json:"platform"` // Use interface{} to avoid circular import
	Summary        CheckSummary  `json:"summary"`
	ManifestSource string        `json:"manifest_source"`
	Items          []CheckResult `json:"items"`
	GeneratedAt    time.Time     `json:"generated_at"`
}

// CheckSummary provides statistical summary of tool verification results
type CheckSummary struct {
	Total    int `json:"total"`
	OK       int `json:"ok"`
	Missing  int `json:"missing"`
	Outdated int `json:"outdated"`
	Errors   int `json:"errors"`
}

// Validate performs validation of the check result
func (cr *CheckResult) Validate() error {
	// Check required fields
	if cr.ToolID == "" || cr.ToolName == "" || cr.RequiredVersion == "" {
		return errors.New("required fields cannot be empty")
	}

	if len(cr.Links) == 0 {
		return errors.New("links cannot be empty")
	}

	// Validate status consistency
	switch cr.Status {
	case StatusOK:
		if cr.ActualVersion == "" {
			return errors.New("OK status must have actual version")
		}
	case StatusNotFound, StatusMissing:
		if cr.ActualVersion != "" {
			return errors.New("Missing/NotFound status should not have actual version")
		}
	case StatusOutdated:
		if cr.ActualVersion == "" {
			return errors.New("Outdated status must have actual version")
		}
	case StatusError:
		if cr.ErrorMessage == "" {
			return errors.New("Error status must have error message")
		}
	}

	return nil
}

// DetermineStatus sets the status based on current state
func (cr *CheckResult) DetermineStatus() {
	// If there are errors, status is Error
	if cr.ErrorMessage != "" {
		cr.Status = StatusError
		return
	}

	// If no version found, status is Missing/NotFound
	if cr.ActualVersion == "" {
		cr.Status = StatusNotFound
		return
	}

	// If we have a version, assume it's OK for this simple implementation
	// Real version validation happens in the checker
	cr.Status = StatusOK
}

// AddError adds an error message and sets status to Error
func (cr *CheckResult) AddError(errMsg string) {
	cr.ErrorMessage = errMsg
	cr.Status = StatusError
}

// HasErrors returns true if the check result has any errors
func (cr *CheckResult) HasErrors() bool {
	return cr.ErrorMessage != ""
}

// Validate performs validation of the environment report
func (er *EnvironmentReport) Validate() error {
	if er.SchemaVersion != 1 {
		return fmt.Errorf("unsupported schema version: %d", er.SchemaVersion)
	}

	if er.ManifestSource == "" {
		return errors.New("manifest source cannot be empty")
	}

	if er.Items == nil {
		return errors.New("items cannot be nil")
	}

	// Validate summary consistency
	if er.Summary.Total != len(er.Items) {
		return errors.New("summary total mismatch")
	}

	calculatedTotal := er.Summary.OK + er.Summary.Missing + er.Summary.Outdated + er.Summary.Errors
	if calculatedTotal != er.Summary.Total {
		return errors.New("summary counts don't add up to total")
	}

	// Validate each item
	for i, item := range er.Items {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("item %d validation failed: %v", i, err)
		}
	}

	// Validate platform info - skip validation since it's interface{}
	// Platform validation is handled by the platform package itself

	if er.GeneratedAt.IsZero() {
		return errors.New("generated_at cannot be zero time")
	}

	return nil
}


// CalculateCheckSummary calculates summary statistics from check results
func CalculateCheckSummary(items []CheckResult) CheckSummary {
	summary := CheckSummary{
		Total: len(items),
	}

	for _, item := range items {
		switch item.Status {
		case StatusOK:
			summary.OK++
		case StatusMissing:
			summary.Missing++
		case StatusOutdated:
			summary.Outdated++
		case StatusError:
			summary.Errors++
		}
	}

	return summary
}

// NewEnvironmentReport creates a new environment report with current timestamp
func NewEnvironmentReport(platform interface{}, manifestSource string, items []CheckResult) *EnvironmentReport {
	summary := CalculateCheckSummary(items)

	return &EnvironmentReport{
		SchemaVersion:  1,
		Platform:       platform,
		Summary:        summary,
		ManifestSource: manifestSource,
		Items:          items,
		GeneratedAt:    time.Now(),
	}
}

// IsSuccessful returns true if all tools meet requirements (no missing, outdated, or errors)
func (er *EnvironmentReport) IsSuccessful() bool {
	return er.Summary.Missing == 0 && er.Summary.Outdated == 0 && er.Summary.Errors == 0
}

// GetExitCode returns the appropriate exit code for the report
func (er *EnvironmentReport) GetExitCode() int {
	if er.IsSuccessful() {
		return 0
	}
	return 1
}