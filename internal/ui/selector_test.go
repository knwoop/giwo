package ui

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/knwoop/giwo/pkg/worktree"
)

func TestNewSelector(t *testing.T) {
	worktrees := []*worktree.Worktree{
		{Branch: "main", Path: "/repo", IsMain: true},
		{Branch: "feature", Path: "/repo/.worktree/feature"},
	}

	selector := NewSelector(worktrees)
	
	if diff := cmp.Diff(len(worktrees), len(selector.worktrees)); diff != "" {
		t.Errorf("NewSelector worktree count mismatch (-want +got):\n%s", diff)
	}
}

func TestFormatWorktreeStatus(t *testing.T) {
	for name, tt := range map[string]struct {
		worktree *worktree.Worktree
		expected string
	}{
		"main worktree clean": {
			worktree: &worktree.Worktree{
				Branch:  "main",
				Path:    "/repo",
				IsMain:  true,
				IsClean: true,
			},
			expected: "🏠 main 📁 /repo",
		},
		"feature worktree with changes": {
			worktree: &worktree.Worktree{
				Branch:   "feature",
				Path:     "/repo/.worktree/feature",
				IsMain:   false,
				IsClean:  false,
				Added:    2,
				Modified: 1,
				Deleted:  0,
			},
			expected: "🌱 ⚠️  3 changes 📁 /repo/.worktree/feature",
		},
		"worktree with remote status": {
			worktree: &worktree.Worktree{
				Branch:  "feature",
				Path:    "/repo/.worktree/feature",
				IsMain:  false,
				IsClean: true,
				Ahead:   2,
				Behind:  1,
			},
			expected: "🌱 📡 +2/-1 📁 /repo/.worktree/feature",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			
			selector := NewSelector([]*worktree.Worktree{tt.worktree})
			result := selector.formatWorktreeStatus(tt.worktree)
			
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("formatWorktreeStatus mismatch (-want +got):\n%s", diff)
			}
		})
	}
}