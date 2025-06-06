// Package utils provides validation and utility functions for gwt.
package utils

import (
	"regexp"
	"strings"

	"github.com/knwoop/gwt/internal/errors"
)

// Git branch name restrictions based on git-check-ref-format.
var (
	// invalidBranchChars contains characters that are not allowed in branch names.
	invalidBranchChars = []string{"..", "~", "^", ":", "?", "*", "[", "\\"}

	// reservedNames contains branch names that are reserved and cannot be used.
	reservedNames = []string{"HEAD", "head"}

	// refsPattern matches branch names starting with "refs/".
	refsPattern = regexp.MustCompile(`^refs/`)
)

// ValidateBranchName validates a Git branch name according to Git naming rules.
// It returns a ValidationError if the name is invalid.
func ValidateBranchName(name string) error {
	if name == "" {
		return errors.NewValidationError("branch_name", name, errors.ErrInvalidBranchName)
	}

	// Check for spaces
	if strings.Contains(name, " ") {
		return errors.NewValidationError("branch_name", name,
			errors.ErrInvalidBranchName)
	}

	// Check for invalid characters
	for _, char := range invalidBranchChars {
		if strings.Contains(name, char) {
			return errors.NewValidationError("branch_name", name,
				errors.ErrInvalidBranchName)
		}
	}

	// Check for leading/trailing dashes or periods
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") ||
		strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") {
		return errors.NewValidationError("branch_name", name,
			errors.ErrInvalidBranchName)
	}

	// Check for reserved names
	for _, reserved := range reservedNames {
		if strings.EqualFold(name, reserved) {
			return errors.NewValidationError("branch_name", name,
				errors.ErrInvalidBranchName)
		}
	}

	// Check for refs/ prefix
	if refsPattern.MatchString(name) {
		return errors.NewValidationError("branch_name", name,
			errors.ErrInvalidBranchName)
	}

	return nil
}

// SanitizeBranchName converts an arbitrary string into a valid Git branch name.
// It replaces invalid characters and formats according to Git naming conventions.
func SanitizeBranchName(name string) string {
	// Trim whitespace
	name = strings.TrimSpace(name)

	// Replace spaces and underscores with dashes
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")

	// Replace invalid characters with dashes
	for _, char := range invalidBranchChars {
		name = strings.ReplaceAll(name, char, "-")
	}

	// Remove leading/trailing dashes and periods
	name = strings.Trim(name, "-.")

	// Provide default name if result is empty
	if name == "" {
		return "unnamed-branch"
	}

	return name
}
