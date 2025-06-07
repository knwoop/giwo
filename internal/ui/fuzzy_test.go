package ui

import (
	"strings"
	"testing"

	"github.com/knwoop/giwo/pkg/worktree"
)

func TestNewFuzzyFinder(t *testing.T) {
	worktrees := []*worktree.Worktree{
		{Branch: "main", Path: "/repo"},
		{Branch: "feature-auth", Path: "/repo/.worktree/feature-auth"},
	}

	finder := NewFuzzyFinder(worktrees)

	if finder == nil {
		t.Error("NewFuzzyFinder returned nil")
	}

	if len(finder.worktrees) != 2 {
		t.Errorf("Expected 2 worktrees, got %d", len(finder.worktrees))
	}
}

func TestFormatWorktreePreview(t *testing.T) {
	for name, tt := range map[string]struct {
		worktree *worktree.Worktree
		expected []string // Lines that should be present in the preview
	}{
		"main worktree clean": {
			worktree: &worktree.Worktree{
				Branch:  "main",
				Path:    "/repo",
				IsMain:  true,
				IsClean: true,
			},
			expected: []string{
				"Branch: main",
				"Path: /repo",
				"Type: Main worktree 🏠",
				"Status: Clean ✅",
			},
		},
		"feature worktree with changes": {
			worktree: &worktree.Worktree{
				Branch:   "feature-auth",
				Path:     "/repo/.worktree/feature-auth",
				IsMain:   false,
				IsClean:  false,
				Added:    2,
				Modified: 1,
				Deleted:  0,
			},
			expected: []string{
				"Branch: feature-auth",
				"Path: /repo/.worktree/feature-auth",
				"Type: Feature worktree 🌱",
				"Status: 3 changes ⚠️",
				"Added: 2 files",
				"Modified: 1 files",
			},
		},
		"worktree with remote status": {
			worktree: &worktree.Worktree{
				Branch:     "feature-sync",
				Path:       "/repo/.worktree/feature-sync",
				IsMain:     false,
				IsClean:    true,
				Ahead:      2,
				Behind:     1,
				LastCommit: "Add new feature",
				CommitAge:  "2h ago",
			},
			expected: []string{
				"Branch: feature-sync",
				"Path: /repo/.worktree/feature-sync",
				"Type: Feature worktree 🌱",
				"Status: Clean ✅",
				"Sync: +2/-1 commits 📡",
				"Last commit: Add new feature",
				"Commit age: 2h ago",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			finder := NewFuzzyFinder([]*worktree.Worktree{})
			result := finder.formatWorktreePreview(tt.worktree)

			for _, expectedLine := range tt.expected {
				if !strings.Contains(result, expectedLine) {
					t.Errorf("Expected preview to contain %q, got:\n%s", expectedLine, result)
				}
			}
		})
	}
}
