package worktree

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestOutputFormat(t *testing.T) {
	for name, tt := range map[string]struct {
		format   OutputFormat
		expected string
	}{
		"table format":  {OutputFormatTable, "table"},
		"json format":   {OutputFormatJSON, "json"},
		"simple format": {OutputFormatSimple, "simple"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result := string(tt.format)
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("String conversion mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestConfigFiles(t *testing.T) {
	// Test that config files list is not empty and contains expected files
	expectedFiles := []string{".env", ".gitignore"}

	for _, expected := range expectedFiles {
		found := false
		for _, file := range ConfigFiles {
			if file == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected config file %q not found in ConfigFiles", expected)
		}
	}

	if len(ConfigFiles) == 0 {
		t.Error("ConfigFiles should not be empty")
	}
}
