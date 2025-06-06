package utils

import (
	"fmt"
	"regexp"
	"strings"
)

func ValidateBranchName(name string) error {
	if name == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	if strings.Contains(name, " ") {
		return fmt.Errorf("branch name cannot contain spaces")
	}

	invalidChars := []string{"..", "~", "^", ":", "?", "*", "[", "\\", ".."}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("branch name contains invalid character: %s", char)
		}
	}

	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return fmt.Errorf("branch name cannot start or end with a dash")
	}

	if strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") {
		return fmt.Errorf("branch name cannot start or end with a period")
	}

	reserved := []string{"HEAD", "head"}
	for _, reserved := range reserved {
		if strings.EqualFold(name, reserved) {
			return fmt.Errorf("branch name '%s' is reserved", name)
		}
	}

	if matched, _ := regexp.MatchString(`^refs/`, name); matched {
		return fmt.Errorf("branch name cannot start with 'refs/'")
	}

	return nil
}

func SanitizeBranchName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	
	invalidChars := []string{"..", "~", "^", ":", "?", "*", "[", "\\"}
	for _, char := range invalidChars {
		name = strings.ReplaceAll(name, char, "-")
	}
	
	name = strings.Trim(name, "-.")
	
	if name == "" {
		name = "unnamed-branch"
	}
	
	return name
}