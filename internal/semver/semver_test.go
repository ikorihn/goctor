package semver

import (
	"testing"
)

func TestVersionParsing(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		expectError bool
		expected    Version
	}{
		{
			name:        "simple version",
			version:     "1.2.3",
			expectError: false,
			expected:    Version{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:        "version with prerelease",
			version:     "1.2.3-alpha.1",
			expectError: false,
			expected:    Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha.1"},
		},
		{
			name:        "version with build metadata",
			version:     "1.2.3+build.1",
			expectError: false,
			expected:    Version{Major: 1, Minor: 2, Patch: 3, Build: "build.1"},
		},
		{
			name:        "version with prerelease and build",
			version:     "1.2.3-beta.2+build.456",
			expectError: false,
			expected:    Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "beta.2", Build: "build.456"},
		},
		{
			name:        "two-part version",
			version:     "1.22",
			expectError: false,
			expected:    Version{Major: 1, Minor: 22, Patch: 0},
		},
		{
			name:        "single number version",
			version:     "24",
			expectError: false,
			expected:    Version{Major: 24, Minor: 0, Patch: 0},
		},
		{
			name:        "version with v prefix",
			version:     "v1.2.3",
			expectError: false,
			expected:    Version{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:        "invalid version - empty",
			version:     "",
			expectError: true,
		},
		{
			name:        "invalid version - non-numeric",
			version:     "a.b.c",
			expectError: true,
		},
		{
			name:        "invalid version - negative numbers",
			version:     "-1.2.3",
			expectError: true,
		},
		{
			name:        "invalid version - too many parts",
			version:     "1.2.3.4.5",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := ParseVersion(tt.version)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected parsing error for version '%s', got nil", tt.version)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no parsing error for version '%s', got: %v", tt.version, err)
				}

				if version.Major != tt.expected.Major ||
					version.Minor != tt.expected.Minor ||
					version.Patch != tt.expected.Patch ||
					version.Prerelease != tt.expected.Prerelease ||
					version.Build != tt.expected.Build {
					t.Errorf("Expected version %+v, got %+v", tt.expected, version)
				}
			}
		})
	}
}

func TestVersionComparison(t *testing.T) {
	tests := []struct {
		name     string
		version1 string
		version2 string
		expected int
	}{
		// Equal versions
		{"equal versions", "1.2.3", "1.2.3", 0},
		{"equal with prerelease", "1.2.3-alpha", "1.2.3-alpha", 0},

		// Major version differences
		{"major version greater", "2.0.0", "1.9.9", 1},
		{"major version lesser", "1.0.0", "2.0.0", -1},

		// Minor version differences
		{"minor version greater", "1.3.0", "1.2.9", 1},
		{"minor version lesser", "1.2.0", "1.3.0", -1},

		// Patch version differences
		{"patch version greater", "1.2.4", "1.2.3", 1},
		{"patch version lesser", "1.2.3", "1.2.4", -1},

		// Prerelease comparisons
		{"release vs prerelease", "1.2.3", "1.2.3-alpha", 1},
		{"prerelease vs release", "1.2.3-alpha", "1.2.3", -1},
		{"prerelease comparison", "1.2.3-alpha.1", "1.2.3-alpha.2", -1},
		{"prerelease comparison 2", "1.2.3-beta", "1.2.3-alpha", 1},

		// Different formats
		{"two vs three parts", "1.22", "1.22.0", 0},
		{"single vs three parts", "24", "24.0.0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1, err1 := ParseVersion(tt.version1)
			v2, err2 := ParseVersion(tt.version2)

			if err1 != nil || err2 != nil {
				t.Fatalf("Failed to parse versions: %v, %v", err1, err2)
			}

			result := v1.Compare(v2)
			if result != tt.expected {
				t.Errorf("Expected comparison result %d, got %d for %s vs %s",
					tt.expected, result, tt.version1, tt.version2)
			}
		})
	}
}

func TestConstraintParsing(t *testing.T) {
	tests := []struct {
		name        string
		constraint  string
		expectError bool
		expected    Constraint
	}{
		{
			name:        "exact constraint",
			constraint:  "1.2.3",
			expectError: false,
			expected:    Constraint{Operator: OpEqual, Version: mustParseVersion("1.2.3")},
		},
		{
			name:        "greater than or equal",
			constraint:  ">=1.22",
			expectError: false,
			expected:    Constraint{Operator: OpGreaterEqual, Version: mustParseVersion("1.22.0")},
		},
		{
			name:        "greater than",
			constraint:  ">1.0.0",
			expectError: false,
			expected:    Constraint{Operator: OpGreater, Version: mustParseVersion("1.0.0")},
		},
		{
			name:        "less than",
			constraint:  "<2.0.0",
			expectError: false,
			expected:    Constraint{Operator: OpLess, Version: mustParseVersion("2.0.0")},
		},
		{
			name:        "less than or equal",
			constraint:  "<=1.25.0",
			expectError: false,
			expected:    Constraint{Operator: OpLessEqual, Version: mustParseVersion("1.25.0")},
		},
		{
			name:        "tilde constraint",
			constraint:  "~1.2.3",
			expectError: false,
			expected:    Constraint{Operator: OpTilde, Version: mustParseVersion("1.2.3")},
		},
		{
			name:        "caret constraint",
			constraint:  "^1.2.3",
			expectError: false,
			expected:    Constraint{Operator: OpCaret, Version: mustParseVersion("1.2.3")},
		},
		{
			name:        "invalid operator",
			constraint:  "=>1.2.3",
			expectError: true,
		},
		{
			name:        "invalid version",
			constraint:  ">=invalid.version",
			expectError: true,
		},
		{
			name:        "empty constraint",
			constraint:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := ParseConstraint(tt.constraint)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected parsing error for constraint '%s', got nil", tt.constraint)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no parsing error for constraint '%s', got: %v", tt.constraint, err)
				}

				if constraint.Operator != tt.expected.Operator {
					t.Errorf("Expected operator %v, got %v", tt.expected.Operator, constraint.Operator)
				}

				if constraint.Version.Compare(tt.expected.Version) != 0 {
					t.Errorf("Expected version %v, got %v", tt.expected.Version, constraint.Version)
				}
			}
		})
	}
}

func TestConstraintSatisfaction(t *testing.T) {
	tests := []struct {
		name       string
		constraint string
		version    string
		satisfied  bool
	}{
		// Exact constraints
		{"exact match", "1.2.3", "1.2.3", true},
		{"exact no match", "1.2.3", "1.2.4", false},

		// Greater than or equal
		{">=1.22 satisfied", ">=1.22", "1.22.0", true},
		{">=1.22 satisfied higher", ">=1.22", "1.23.0", true},
		{">=1.22 not satisfied", ">=1.22", "1.21.9", false},

		// Greater than
		{">1.0.0 satisfied", ">1.0.0", "1.0.1", true},
		{">1.0.0 not satisfied equal", ">1.0.0", "1.0.0", false},
		{">1.0.0 not satisfied lower", ">1.0.0", "0.9.9", false},

		// Less than
		{"<2.0.0 satisfied", "<2.0.0", "1.9.9", true},
		{"<2.0.0 not satisfied equal", "<2.0.0", "2.0.0", false},
		{"<2.0.0 not satisfied higher", "<2.0.0", "2.0.1", false},

		// Less than or equal
		{"<=1.25 satisfied", "<=1.25", "1.25.0", true},
		{"<=1.25 satisfied lower", "<=1.25", "1.24.0", true},
		{"<=1.25 not satisfied", "<=1.25", "1.26.0", false},

		// Tilde constraints (patch-level changes)
		{"~1.2.3 satisfied patch", "~1.2.3", "1.2.4", true},
		{"~1.2.3 not satisfied minor", "~1.2.3", "1.3.0", false},
		{"~1.2 satisfied minor", "~1.2", "1.2.9", true},

		// Caret constraints (compatible changes)
		{"^1.2.3 satisfied patch", "^1.2.3", "1.2.4", true},
		{"^1.2.3 satisfied minor", "^1.2.3", "1.3.0", true},
		{"^1.2.3 not satisfied major", "^1.2.3", "2.0.0", false},
		{"^0.2.3 only patch", "^0.2.3", "0.2.4", true},
		{"^0.2.3 not minor", "^0.2.3", "0.3.0", false},

		// Prerelease handling
		{"prerelease vs release", ">=1.0.0", "1.0.0-alpha", false},
		{"prerelease constraint", ">=1.0.0-alpha", "1.0.0-beta", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := ParseConstraint(tt.constraint)
			if err != nil {
				t.Fatalf("Failed to parse constraint '%s': %v", tt.constraint, err)
			}

			version, err := ParseVersion(tt.version)
			if err != nil {
				t.Fatalf("Failed to parse version '%s': %v", tt.version, err)
			}

			satisfied := constraint.IsSatisfiedBy(version)
			if satisfied != tt.satisfied {
				t.Errorf("Expected constraint '%s' satisfied by '%s' to be %t, got %t",
					tt.constraint, tt.version, tt.satisfied, satisfied)
			}
		})
	}
}

func TestMultipleConstraints(t *testing.T) {
	tests := []struct {
		name        string
		constraints []string
		version     string
		satisfied   bool
	}{
		{
			name:        "range constraint satisfied",
			constraints: []string{">=1.22", "<1.25"},
			version:     "1.23.0",
			satisfied:   true,
		},
		{
			name:        "range constraint not satisfied - too low",
			constraints: []string{">=1.22", "<1.25"},
			version:     "1.21.0",
			satisfied:   false,
		},
		{
			name:        "range constraint not satisfied - too high",
			constraints: []string{">=1.22", "<1.25"},
			version:     "1.25.0",
			satisfied:   false,
		},
		{
			name:        "complex constraints",
			constraints: []string{">=1.20", "<2.0", "!=1.21.5"},
			version:     "1.22.0",
			satisfied:   true,
		},
		{
			name:        "complex constraints excluded",
			constraints: []string{">=1.20", "<2.0", "!=1.21.5"},
			version:     "1.21.5",
			satisfied:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraints := make([]Constraint, len(tt.constraints))
			for i, constraintStr := range tt.constraints {
				constraint, err := ParseConstraint(constraintStr)
				if err != nil {
					t.Fatalf("Failed to parse constraint '%s': %v", constraintStr, err)
				}
				constraints[i] = constraint
			}

			version, err := ParseVersion(tt.version)
			if err != nil {
				t.Fatalf("Failed to parse version '%s': %v", tt.version, err)
			}

			satisfied := SatisfiesAll(version, constraints)
			if satisfied != tt.satisfied {
				t.Errorf("Expected constraints %v satisfied by '%s' to be %t, got %t",
					tt.constraints, tt.version, tt.satisfied, satisfied)
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		version  Version
		expected string
	}{
		{Version{1, 2, 3, "", ""}, "1.2.3"},
		{Version{1, 2, 3, "alpha", ""}, "1.2.3-alpha"},
		{Version{1, 2, 3, "", "build.1"}, "1.2.3+build.1"},
		{Version{1, 2, 3, "beta.2", "build.456"}, "1.2.3-beta.2+build.456"},
		{Version{24, 0, 0, "", ""}, "24.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.version.String()
			if result != tt.expected {
				t.Errorf("Expected version string '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// Benchmark tests for performance-critical operations
func BenchmarkParseVersion(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseVersion("1.22.3")
	}
}

func BenchmarkParseConstraint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseConstraint(">=1.22 <1.25")
	}
}

func BenchmarkVersionCompare(b *testing.B) {
	v1, _ := ParseVersion("1.22.3")
	v2, _ := ParseVersion("1.22.4")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v1.Compare(v2)
	}
}

func BenchmarkConstraintSatisfaction(b *testing.B) {
	constraint, _ := ParseConstraint(">=1.22")
	version, _ := ParseVersion("1.22.3")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		constraint.IsSatisfiedBy(version)
	}
}

// Helper function for tests
func mustParseVersion(s string) Version {
	v, err := ParseVersion(s)
	if err != nil {
		panic(err)
	}
	return v
}