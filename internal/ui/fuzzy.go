// Package ui provides fuzzy search functionality.
package ui

import (
	"fmt"
	"strings"

	"github.com/knwoop/giwo/pkg/worktree"
	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
)

// FuzzyFinder provides fuzzy search functionality for worktrees using go-fuzzyfinder.
type FuzzyFinder struct {
	worktrees []*worktree.Worktree
}

// NewFuzzyFinder creates a new fuzzy finder.
func NewFuzzyFinder(worktrees []*worktree.Worktree) *FuzzyFinder {
	return &FuzzyFinder{
		worktrees: worktrees,
	}
}

// Search performs fuzzy search using go-fuzzyfinder.
func (f *FuzzyFinder) Search() (*worktree.Worktree, error) {
	if len(f.worktrees) == 0 {
		return nil, fmt.Errorf("no worktrees available")
	}

	// If only one worktree, return it directly
	if len(f.worktrees) == 1 {
		return f.worktrees[0], nil
	}

	// Use go-fuzzyfinder to search
	idx, err := fuzzyfinder.Find(
		f.worktrees,
		func(i int) string {
			return f.worktrees[i].Branch
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return f.formatWorktreePreview(f.worktrees[i])
		}),
		fuzzyfinder.WithHeader("Select Worktree"),
	)
	if err != nil {
		// go-fuzzyfinder returns specific error for user cancellation
		if err == fuzzyfinder.ErrAbort {
			return nil, nil
		}
		return nil, fmt.Errorf("fuzzy search failed: %w", err)
	}

	return f.worktrees[idx], nil
}

// formatWorktreePreview formats a worktree for the preview window.
func (f *FuzzyFinder) formatWorktreePreview(wt *worktree.Worktree) string {
	var lines []string

	// Branch and path info
	lines = append(lines, fmt.Sprintf("Branch: %s", wt.Branch))
	lines = append(lines, fmt.Sprintf("Path: %s", wt.Path))

	// Status info
	if wt.IsMain {
		lines = append(lines, "Type: Main worktree 🏠")
	} else {
		lines = append(lines, "Type: Feature worktree 🌱")
	}

	// Clean status
	if wt.IsClean {
		lines = append(lines, "Status: Clean ✅")
	} else {
		changes := wt.Added + wt.Modified + wt.Deleted
		lines = append(lines, fmt.Sprintf("Status: %d changes ⚠️", changes))
		if wt.Added > 0 {
			lines = append(lines, fmt.Sprintf("  Added: %d files", wt.Added))
		}
		if wt.Modified > 0 {
			lines = append(lines, fmt.Sprintf("  Modified: %d files", wt.Modified))
		}
		if wt.Deleted > 0 {
			lines = append(lines, fmt.Sprintf("  Deleted: %d files", wt.Deleted))
		}
	}

	// Remote sync status
	if wt.Ahead > 0 || wt.Behind > 0 {
		lines = append(lines, fmt.Sprintf("Sync: +%d/-%d commits 📡", wt.Ahead, wt.Behind))
	}

	// Last commit info
	if wt.LastCommit != "" {
		lines = append(lines, fmt.Sprintf("Last commit: %s", wt.LastCommit))
		if wt.CommitAge != "" {
			lines = append(lines, fmt.Sprintf("Commit age: %s", wt.CommitAge))
		}
	}

	return strings.Join(lines, "\n")
}
