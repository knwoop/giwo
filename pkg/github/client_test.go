package github

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseGitHubURL(t *testing.T) {
	for name, tt := range map[string]struct {
		url           string
		expectedOwner string
		expectedRepo  string
	}{
		"SSH URL":                {"git@github.com:knwoop/gwt.git", "knwoop", "gwt"},
		"SSH URL without .git":   {"git@github.com:knwoop/gwt", "knwoop", "gwt"},
		"HTTPS URL":              {"https://github.com/knwoop/gwt.git", "knwoop", "gwt"},
		"HTTPS URL without .git": {"https://github.com/knwoop/gwt", "knwoop", "gwt"},
		"Invalid URL":            {"not-a-github-url", "", ""},
		"Empty URL":              {"", "", ""},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			owner, repo := parseGitHubURL(tt.url)
			if diff := cmp.Diff(tt.expectedOwner, owner); diff != "" {
				t.Errorf("parseGitHubURL(%q) owner mismatch (-want +got):\n%s", tt.url, diff)
			}
			if diff := cmp.Diff(tt.expectedRepo, repo); diff != "" {
				t.Errorf("parseGitHubURL(%q) repo mismatch (-want +got):\n%s", tt.url, diff)
			}
		})
	}
}
