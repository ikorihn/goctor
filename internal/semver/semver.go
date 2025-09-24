package semver

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Version represents a semantic version
type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Build      string
}

// Operator represents comparison operators for version constraints
type Operator int

const (
	OpEqual Operator = iota
	OpGreater
	OpGreaterEqual
	OpLess
	OpLessEqual
	OpTilde
	OpCaret
	OpNotEqual
)

// Constraint represents a version constraint
type Constraint struct {
	Operator Operator
	Version  Version
}

var (
	// versionRegex matches semantic versions with optional v prefix
	versionRegex = regexp.MustCompile(`^v?(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:-([0-9A-Za-z\-\.]+))?(?:\+([0-9A-Za-z\-\.]+))?$`)

	// constraintRegex matches version constraints
	constraintRegex = regexp.MustCompile(`^(>=|<=|>|<|~|\^|!=)?(.+)$`)
)

// ParseVersion parses a version string into a Version struct
func ParseVersion(versionStr string) (Version, error) {
	if versionStr == "" {
		return Version{}, errors.New("version string cannot be empty")
	}

	matches := versionRegex.FindStringSubmatch(versionStr)
	if matches == nil {
		return Version{}, fmt.Errorf("invalid version format: %s", versionStr)
	}

	version := Version{}
	var err error

	// Parse major version (required)
	version.Major, err = strconv.Atoi(matches[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version: %s", matches[1])
	}

	// Parse minor version (optional, defaults to 0)
	if matches[2] != "" {
		version.Minor, err = strconv.Atoi(matches[2])
		if err != nil {
			return Version{}, fmt.Errorf("invalid minor version: %s", matches[2])
		}
	}

	// Parse patch version (optional, defaults to 0)
	if matches[3] != "" {
		version.Patch, err = strconv.Atoi(matches[3])
		if err != nil {
			return Version{}, fmt.Errorf("invalid patch version: %s", matches[3])
		}
	}

	// Parse prerelease (optional)
	if matches[4] != "" {
		version.Prerelease = matches[4]
	}

	// Parse build metadata (optional)
	if matches[5] != "" {
		version.Build = matches[5]
	}

	return version, nil
}

// ParseConstraint parses a constraint string into a Constraint struct
func ParseConstraint(constraintStr string) (Constraint, error) {
	if constraintStr == "" {
		return Constraint{}, errors.New("constraint string cannot be empty")
	}

	matches := constraintRegex.FindStringSubmatch(constraintStr)
	if matches == nil {
		return Constraint{}, fmt.Errorf("invalid constraint format: %s", constraintStr)
	}

	operatorStr := matches[1]
	versionStr := matches[2]

	// Determine operator
	var operator Operator
	switch operatorStr {
	case ">=":
		operator = OpGreaterEqual
	case ">":
		operator = OpGreater
	case "<=":
		operator = OpLessEqual
	case "<":
		operator = OpLess
	case "~":
		operator = OpTilde
	case "^":
		operator = OpCaret
	case "!=":
		operator = OpNotEqual
	case "", "=":
		operator = OpEqual
	default:
		return Constraint{}, fmt.Errorf("unknown operator: %s", operatorStr)
	}

	// Parse version
	version, err := ParseVersion(versionStr)
	if err != nil {
		return Constraint{}, fmt.Errorf("invalid version in constraint: %v", err)
	}

	return Constraint{
		Operator: operator,
		Version:  version,
	}, nil
}

// Compare compares this version with another version
// Returns -1 if this version is less, 0 if equal, 1 if greater
func (v Version) Compare(other Version) int {
	// Compare major version
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}

	// Compare minor version
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}

	// Compare patch version
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}

	// Compare prerelease
	return comparePrerelease(v.Prerelease, other.Prerelease)
}

// comparePrerelease compares prerelease versions
func comparePrerelease(pre1, pre2 string) int {
	// No prerelease is greater than any prerelease
	if pre1 == "" && pre2 == "" {
		return 0
	}
	if pre1 == "" {
		return 1
	}
	if pre2 == "" {
		return -1
	}

	// Compare prerelease strings lexicographically
	if pre1 < pre2 {
		return -1
	}
	if pre1 > pre2 {
		return 1
	}
	return 0
}

// IsSatisfiedBy checks if a version satisfies this constraint
func (c Constraint) IsSatisfiedBy(version Version) bool {
	comparison := version.Compare(c.Version)

	switch c.Operator {
	case OpEqual:
		return comparison == 0
	case OpGreater:
		return comparison > 0
	case OpGreaterEqual:
		return comparison >= 0
	case OpLess:
		return comparison < 0
	case OpLessEqual:
		return comparison <= 0
	case OpNotEqual:
		return comparison != 0
	case OpTilde:
		return c.tildeConstraint(version)
	case OpCaret:
		return c.caretConstraint(version)
	default:
		return false
	}
}

// tildeConstraint implements tilde (~) constraint logic
// ~1.2.3 := >=1.2.3 <1.3.0 (reasonably close to 1.2.3)
// ~1.2 := >=1.2.0 <1.3.0
// ~1 := >=1.0.0 <2.0.0
func (c Constraint) tildeConstraint(version Version) bool {
	if version.Major != c.Version.Major {
		return false
	}

	if version.Minor != c.Version.Minor {
		return false
	}

	// Allow any patch version >= constraint patch
	return version.Patch >= c.Version.Patch
}

// caretConstraint implements caret (^) constraint logic
// ^1.2.3 := >=1.2.3 <2.0.0 (compatible within same major version)
// ^0.2.3 := >=0.2.3 <0.3.0 (for 0.x, compatible within same minor)
// ^0.0.3 := >=0.0.3 <0.0.4 (for 0.0.x, compatible within same patch)
func (c Constraint) caretConstraint(version Version) bool {
	if version.Major != c.Version.Major {
		return false
	}

	// For major version 0, be more restrictive
	if c.Version.Major == 0 {
		if version.Minor != c.Version.Minor {
			return false
		}

		// For 0.0.x, only allow same patch
		if c.Version.Minor == 0 {
			return version.Patch == c.Version.Patch
		}

		// For 0.x.y, allow patch changes
		return version.Patch >= c.Version.Patch
	}

	// For major version > 0, allow minor and patch changes
	if version.Minor > c.Version.Minor {
		return true
	}
	if version.Minor == c.Version.Minor {
		return version.Patch >= c.Version.Patch
	}

	return false
}

// String returns the string representation of the version
func (v Version) String() string {
	result := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)

	if v.Prerelease != "" {
		result += "-" + v.Prerelease
	}

	if v.Build != "" {
		result += "+" + v.Build
	}

	return result
}

// String returns the string representation of the operator
func (op Operator) String() string {
	switch op {
	case OpEqual:
		return "="
	case OpGreater:
		return ">"
	case OpGreaterEqual:
		return ">="
	case OpLess:
		return "<"
	case OpLessEqual:
		return "<="
	case OpTilde:
		return "~"
	case OpCaret:
		return "^"
	case OpNotEqual:
		return "!="
	default:
		return "unknown"
	}
}

// String returns the string representation of the constraint
func (c Constraint) String() string {
	if c.Operator == OpEqual {
		return c.Version.String()
	}
	return c.Operator.String() + c.Version.String()
}

// SatisfiesAll checks if a version satisfies all constraints in a list
func SatisfiesAll(version Version, constraints []Constraint) bool {
	for _, constraint := range constraints {
		if !constraint.IsSatisfiedBy(version) {
			return false
		}
	}
	return true
}

// ParseConstraints parses multiple constraints from a space-separated string
func ParseConstraints(constraintStr string) ([]Constraint, error) {
	if constraintStr == "" {
		return nil, errors.New("constraint string cannot be empty")
	}

	parts := strings.Fields(constraintStr)
	constraints := make([]Constraint, len(parts))

	for i, part := range parts {
		constraint, err := ParseConstraint(part)
		if err != nil {
			return nil, fmt.Errorf("failed to parse constraint '%s': %v", part, err)
		}
		constraints[i] = constraint
	}

	return constraints, nil
}