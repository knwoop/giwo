// Package ui provides fuzzy search functionality.
package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/knwoop/giwo/pkg/worktree"
)

// FuzzyFinder provides fuzzy search functionality for worktrees.
type FuzzyFinder struct {
	worktrees []*worktree.Worktree
	maxShow   int
}

// NewFuzzyFinder creates a new fuzzy finder.
func NewFuzzyFinder(worktrees []*worktree.Worktree) *FuzzyFinder {
	return &FuzzyFinder{
		worktrees: worktrees,
		maxShow:   10, // Show max 10 results
	}
}

// Search performs fuzzy search and returns an interactive selector.
func (f *FuzzyFinder) Search() (*worktree.Worktree, error) {
	if len(f.worktrees) == 0 {
		return nil, fmt.Errorf("no worktrees available")
	}

	// If only one worktree, return it directly
	if len(f.worktrees) == 1 {
		return f.worktrees[0], nil
	}

	fmt.Println("🔍 Fuzzy search for worktrees (type to filter, Enter to select)")
	fmt.Println("   Use numbers to select directly, 'q' to quit")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	
	for {
		// Show current matches
		matches := f.worktrees
		
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(input)
		
		// Handle quit
		if strings.ToLower(input) == "q" || input == "quit" {
			return nil, nil
		}

		// Handle empty input - show all
		if input == "" {
			matches = f.worktrees
		} else {
			// Handle direct number selection
			if num, err := strconv.Atoi(input); err == nil {
				if num >= 1 && num <= len(f.worktrees) {
					return f.worktrees[num-1], nil
				}
				fmt.Printf("Invalid selection: %d (range: 1-%d)\n", num, len(f.worktrees))
				continue
			}
			
			// Perform fuzzy search
			matches = f.fuzzySearch(input)
		}

		// Clear screen and show results
		fmt.Print("\033[2J\033[H") // Clear screen and move cursor to top
		fmt.Println("🔍 Fuzzy search for worktrees (type to filter, Enter to select)")
		fmt.Println("   Use numbers to select directly, 'q' to quit")
		fmt.Println()

		if len(matches) == 0 {
			fmt.Printf("No matches for: %s\n", input)
			fmt.Println()
			continue
		}

		// Show matches with numbers
		showCount := len(matches)
		if showCount > f.maxShow {
			showCount = f.maxShow
		}

		for i := 0; i < showCount; i++ {
			wt := matches[i]
			status := f.formatWorktreeForSearch(wt, input)
			fmt.Printf("  %d) %s\n", i+1, status)
		}

		if len(matches) > f.maxShow {
			fmt.Printf("  ... and %d more (refine search to see more)\n", len(matches)-f.maxShow)
		}
		fmt.Println()

		// If exact match, allow quick selection
		if len(matches) == 1 {
			fmt.Printf("Press Enter to select '%s' or continue typing: ", matches[0].Branch)
			line, _ := reader.ReadString('\n')
			line = strings.TrimSpace(line)
			if line == "" {
				return matches[0], nil
			}
			// Continue with the new input
			input = line
			matches = f.fuzzySearch(input)
		}
	}
}

// fuzzySearch performs fuzzy matching on worktree branch names.
func (f *FuzzyFinder) fuzzySearch(query string) []*worktree.Worktree {
	if query == "" {
		return f.worktrees
	}

	query = strings.ToLower(query)
	var matches []*worktree.Worktree
	
	// First pass: exact substring matches
	for _, wt := range f.worktrees {
		branchLower := strings.ToLower(wt.Branch)
		if strings.Contains(branchLower, query) {
			matches = append(matches, wt)
		}
	}

	// Second pass: fuzzy matches (characters appear in order)
	if len(matches) < f.maxShow {
		for _, wt := range f.worktrees {
			if f.isAlreadyMatched(wt, matches) {
				continue
			}
			if f.fuzzyMatch(strings.ToLower(wt.Branch), query) {
				matches = append(matches, wt)
				if len(matches) >= f.maxShow*2 { // Get some extra for sorting
					break
				}
			}
		}
	}

	return matches
}

// fuzzyMatch checks if all characters in query appear in target in order.
func (f *FuzzyFinder) fuzzyMatch(target, query string) bool {
	targetRunes := []rune(target)
	queryRunes := []rune(query)
	
	if len(queryRunes) == 0 {
		return true
	}
	
	queryIndex := 0
	for _, targetRune := range targetRunes {
		if queryIndex < len(queryRunes) && targetRune == queryRunes[queryIndex] {
			queryIndex++
		}
	}
	
	return queryIndex == len(queryRunes)
}

// isAlreadyMatched checks if a worktree is already in the matches slice.
func (f *FuzzyFinder) isAlreadyMatched(wt *worktree.Worktree, matches []*worktree.Worktree) bool {
	for _, match := range matches {
		if match.Path == wt.Path {
			return true
		}
	}
	return false
}

// formatWorktreeForSearch formats a worktree for display in search results.
func (f *FuzzyFinder) formatWorktreeForSearch(wt *worktree.Worktree, query string) string {
	// Highlight matching characters
	branch := f.highlightMatch(wt.Branch, query)
	
	var parts []string
	parts = append(parts, branch)

	if wt.IsMain {
		parts = append(parts, "🏠")
	} else {
		parts = append(parts, "🌱")
	}

	if !wt.IsClean {
		changes := wt.Added + wt.Modified + wt.Deleted
		parts = append(parts, fmt.Sprintf("⚠️%d", changes))
	}

	if wt.Ahead > 0 || wt.Behind > 0 {
		parts = append(parts, fmt.Sprintf("+%d/-%d", wt.Ahead, wt.Behind))
	}

	return strings.Join(parts, " ")
}

// highlightMatch highlights matching characters in the branch name.
func (f *FuzzyFinder) highlightMatch(branch, query string) string {
	if query == "" {
		return branch
	}

	// For simple substring matches, highlight the exact match
	queryLower := strings.ToLower(query)
	branchLower := strings.ToLower(branch)
	
	if idx := strings.Index(branchLower, queryLower); idx >= 0 {
		before := branch[:idx]
		match := branch[idx : idx+len(query)]
		after := branch[idx+len(query):]
		return before + "\033[1;33m" + match + "\033[0m" + after // Yellow highlight
	}

	return branch
}