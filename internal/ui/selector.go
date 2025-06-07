// Package ui provides user interface components for giwo.
package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/knwoop/giwo/pkg/worktree"
)

// Selector provides interactive selection functionality for worktrees.
type Selector struct {
	worktrees []*worktree.Worktree
}

// NewSelector creates a new selector with the given worktrees.
func NewSelector(worktrees []*worktree.Worktree) *Selector {
	return &Selector{
		worktrees: worktrees,
	}
}

// Select allows the user to interactively select a worktree.
// It returns the selected worktree or nil if cancelled.
func (s *Selector) Select() (*worktree.Worktree, error) {
	if len(s.worktrees) == 0 {
		return nil, fmt.Errorf("no worktrees available")
	}

	// If only one worktree (main), return it directly
	if len(s.worktrees) == 1 {
		return s.worktrees[0], nil
	}

	fmt.Println("📂 Available worktrees:")
	fmt.Println()

	// Display numbered list of worktrees
	for i, wt := range s.worktrees {
		status := s.formatWorktreeStatus(wt)
		fmt.Printf("  %d) %s %s\n", i+1, wt.Branch, status)
	}

	fmt.Println()
	fmt.Printf("Select worktree (1-%d, q to quit): ", len(s.worktrees))

	// Read user input
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	
	// Handle quit
	if strings.ToLower(input) == "q" || input == "" {
		return nil, nil
	}

	// Parse selection
	selection, err := strconv.Atoi(input)
	if err != nil {
		return nil, fmt.Errorf("invalid selection: %s", input)
	}

	if selection < 1 || selection > len(s.worktrees) {
		return nil, fmt.Errorf("selection out of range: %d", selection)
	}

	return s.worktrees[selection-1], nil
}

// SelectWithFilter allows the user to filter and select worktrees.
func (s *Selector) SelectWithFilter(filter string) (*worktree.Worktree, error) {
	if filter == "" {
		return s.Select()
	}

	// Filter worktrees based on branch name
	var filtered []*worktree.Worktree
	filter = strings.ToLower(filter)
	
	for _, wt := range s.worktrees {
		if strings.Contains(strings.ToLower(wt.Branch), filter) {
			filtered = append(filtered, wt)
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no worktrees match filter: %s", filter)
	}

	// If only one match, return it directly
	if len(filtered) == 1 {
		return filtered[0], nil
	}

	// Create new selector with filtered results
	filteredSelector := NewSelector(filtered)
	fmt.Printf("🔍 Filtered worktrees (matching '%s'):\n", filter)
	fmt.Println()
	
	return filteredSelector.Select()
}

// formatWorktreeStatus returns a formatted status string for a worktree.
func (s *Selector) formatWorktreeStatus(wt *worktree.Worktree) string {
	var parts []string

	if wt.IsMain {
		parts = append(parts, "🏠 main")
	} else {
		parts = append(parts, "🌱")
	}

	if !wt.IsClean {
		changes := wt.Added + wt.Modified + wt.Deleted
		parts = append(parts, fmt.Sprintf("⚠️  %d changes", changes))
	}

	if wt.Ahead > 0 || wt.Behind > 0 {
		parts = append(parts, fmt.Sprintf("📡 +%d/-%d", wt.Ahead, wt.Behind))
	}

	parts = append(parts, fmt.Sprintf("📁 %s", wt.Path))

	return strings.Join(parts, " ")
}