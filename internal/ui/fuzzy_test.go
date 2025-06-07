package ui

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/knwoop/giwo/pkg/worktree"
)

func TestFuzzyMatch(t *testing.T) {
	for name, tt := range map[string]struct {
		target   string
		query    string
		expected bool
	}{
		"exact match":           {"feature", "feature", true},
		"substring match":       {"feature-auth", "auth", true},
		"fuzzy match":          {"feature-auth", "feath", true},
		"case insensitive":     {"feature-auth", "feath", true},
		"no match":             {"feature", "bugfix", false},
		"empty query":          {"feature", "", true},
		"partial fuzzy":        {"feature-authentication", "featauth", true},
		"reverse order":        {"feature-auth", "authfeat", false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			
			finder := NewFuzzyFinder([]*worktree.Worktree{})
			// Convert to lowercase for the test since fuzzyMatch expects lowercase input
			target := strings.ToLower(tt.target)
			query := strings.ToLower(tt.query)
			result := finder.fuzzyMatch(target, query)
			
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("fuzzyMatch(%q, %q) mismatch (-want +got):\n%s", tt.target, tt.query, diff)
			}
		})
	}
}

func TestFuzzySearch(t *testing.T) {
	worktrees := []*worktree.Worktree{
		{Branch: "main", Path: "/repo"},
		{Branch: "feature-auth", Path: "/repo/.worktree/feature-auth"},
		{Branch: "feature-ui", Path: "/repo/.worktree/feature-ui"},
		{Branch: "bugfix-login", Path: "/repo/.worktree/bugfix-login"},
		{Branch: "hotfix-critical", Path: "/repo/.worktree/hotfix-critical"},
	}

	for name, tt := range map[string]struct {
		query          string
		expectedCount  int
		expectedFirst  string
	}{
		"empty query returns all": {
			query:         "",
			expectedCount: 5,
			expectedFirst: "main",
		},
		"exact substring match": {
			query:         "feature",
			expectedCount: 2,
			expectedFirst: "feature-auth",
		},
		"fuzzy match": {
			query:         "feath",
			expectedCount: 1,
			expectedFirst: "feature-auth",
		},
		"no matches": {
			query:         "nonexistent",
			expectedCount: 0,
		},
		"case insensitive": {
			query:         "FEATURE",
			expectedCount: 2,
			expectedFirst: "feature-auth",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			
			finder := NewFuzzyFinder(worktrees)
			results := finder.fuzzySearch(tt.query)
			
			if diff := cmp.Diff(tt.expectedCount, len(results)); diff != "" {
				t.Errorf("fuzzySearch count mismatch (-want +got):\n%s", diff)
			}
			
			if tt.expectedCount > 0 && len(results) > 0 {
				if diff := cmp.Diff(tt.expectedFirst, results[0].Branch); diff != "" {
					t.Errorf("fuzzySearch first result mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestHighlightMatch(t *testing.T) {
	for name, tt := range map[string]struct {
		branch   string
		query    string
		expected string
	}{
		"exact match": {
			branch:   "feature",
			query:    "feature",
			expected: "\033[1;33mfeature\033[0m",
		},
		"substring match": {
			branch:   "feature-auth",
			query:    "auth",
			expected: "feature-\033[1;33mauth\033[0m",
		},
		"no match": {
			branch:   "feature",
			query:    "bugfix",
			expected: "feature",
		},
		"empty query": {
			branch:   "feature",
			query:    "",
			expected: "feature",
		},
		"case insensitive": {
			branch:   "Feature-Auth",
			query:    "auth",
			expected: "Feature-\033[1;33mAuth\033[0m",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			
			finder := NewFuzzyFinder([]*worktree.Worktree{})
			result := finder.highlightMatch(tt.branch, tt.query)
			
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("highlightMatch mismatch (-want +got):\n%s", diff)
			}
		})
	}
}