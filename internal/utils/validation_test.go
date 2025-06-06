package utils

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestValidateBranchName(t *testing.T) {
	for name, test := range map[string]struct {
		branchName string
		wantError  bool
	}{
		"valid simple name":       {"feature-auth", false},
		"valid with numbers":      {"feature-123", false},
		"valid with underscores":  {"feature_auth", false},
		"empty name":              {"", true},
		"name with spaces":        {"feature auth", true},
		"name with double dots":   {"feature..auth", true},
		"name starting with dash": {"-feature", true},
		"name ending with dash":   {"feature-", true},
		"name starting with dot":  {".feature", true},
		"name ending with dot":    {"feature.", true},
		"reserved name HEAD":      {"HEAD", true},
		"reserved name head":      {"head", true},
		"name with refs prefix":   {"refs/heads/feature", true},
		"name with tilde":         {"feature~1", true},
		"name with caret":         {"feature^1", true},
		"name with colon":         {"feature:auth", true},
		"name with question mark": {"feature?", true},
		"name with asterisk":      {"feature*", true},
		"name with bracket":       {"feature[1]", true},
		"name with backslash":     {"feature\\auth", true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := ValidateBranchName(test.branchName)
			if test.wantError && err == nil {
				t.Errorf("ValidateBranchName(%q) expected error but got none", test.branchName)
			}
			if !test.wantError && err != nil {
				t.Errorf("ValidateBranchName(%q) unexpected error: %v", test.branchName, err)
			}
		})
	}
}

func TestSanitizeBranchName(t *testing.T) {
	for name, tt := range map[string]struct {
		input    string
		expected string
	}{
		"simple name":                       {"feature", "feature"},
		"name with spaces":                  {"feature auth", "feature-auth"},
		"name with underscores":             {"feature_auth", "feature-auth"},
		"name with double dots":             {"feature..auth", "feature-auth"},
		"name with leading/trailing dashes": {"-feature-", "feature"},
		"name with leading/trailing dots":   {".feature.", "feature"},
		"empty string":                      {"", "unnamed-branch"},
		"only invalid chars":                {"..~~", "unnamed-branch"},
		"mixed invalid chars":               {"feature~auth:test", "feature-auth-test"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			result := SanitizeBranchName(tt.input)
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("SanitizeBranchName(%q) mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}
