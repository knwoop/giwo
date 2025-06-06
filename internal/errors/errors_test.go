package errors

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestValidationError(t *testing.T) {
	for name, tt := range map[string]struct {
		field    string
		value    string
		err      error
		expected string
	}{
		"simple validation error": {
			field:    "branch_name",
			value:    "invalid..name",
			err:      ErrInvalidBranchName,
			expected: `validation failed for branch_name="invalid..name": invalid branch name`,
		},
		"empty field": {
			field:    "",
			value:    "test",
			err:      ErrInvalidBranchName,
			expected: `validation failed for ="test": invalid branch name`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := NewValidationError(tt.field, tt.value, tt.err)

			if diff := cmp.Diff(tt.expected, err.Error()); diff != "" {
				t.Errorf("Error() mismatch (-want +got):\n%s", diff)
			}

			if !errors.Is(err, tt.err) {
				t.Errorf("Expected error to wrap %v", tt.err)
			}
		})
	}
}

func TestGitError(t *testing.T) {
	for name, tt := range map[string]struct {
		operation string
		args      []string
		err       error
		expected  string
	}{
		"worktree command error": {
			operation: "worktree",
			args:      []string{"add", "path"},
			err:       errors.New("failed to create"),
			expected:  "git worktree failed: failed to create",
		},
		"fetch command error": {
			operation: "fetch",
			args:      []string{"--prune"},
			err:       errors.New("network error"),
			expected:  "git fetch failed: network error",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := NewGitError(tt.operation, tt.args, tt.err)

			if diff := cmp.Diff(tt.expected, err.Error()); diff != "" {
				t.Errorf("Error() mismatch (-want +got):\n%s", diff)
			}

			if !errors.Is(err, tt.err) {
				t.Errorf("Expected error to wrap %v", tt.err)
			}
		})
	}
}
