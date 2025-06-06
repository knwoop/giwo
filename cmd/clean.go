package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/knwoop/gwt/pkg/worktree"
	"github.com/spf13/cobra"
)

var (
	cleanDryRun bool
	cleanForce  bool
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove worktrees for merged branches",
	Long: `Batch remove worktrees for branches that have been merged into the main branch.
This excludes main/master/develop branches by default.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := worktree.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize manager: %w", err)
		}

		mergedBranches, err := manager.GetMergedBranches()
		if err != nil {
			return fmt.Errorf("failed to get merged branches: %w", err)
		}

		if len(mergedBranches) == 0 {
			fmt.Println("🧹 No merged branches found to clean up")
			return nil
		}

		worktrees, err := manager.ListWorktrees()
		if err != nil {
			return fmt.Errorf("failed to list worktrees: %w", err)
		}

		worktreeMap := make(map[string]*worktree.Worktree)
		for _, wt := range worktrees {
			worktreeMap[wt.Branch] = wt
		}

		var toRemove []string
		for _, branch := range mergedBranches {
			if _, exists := worktreeMap[branch]; exists {
				toRemove = append(toRemove, branch)
			}
		}

		if len(toRemove) == 0 {
			fmt.Println("🧹 No worktrees found for merged branches")
			return nil
		}

		fmt.Printf("🧹 Found %d worktree(s) for merged branches:\n", len(toRemove))
		for _, branch := range toRemove {
			wt := worktreeMap[branch]
			status := "clean"
			if !wt.IsClean {
				status = "⚠️  dirty"
			}
			fmt.Printf("  - %s (%s)\n", branch, status)
		}

		if cleanDryRun {
			fmt.Printf("\n💡 Run without --dry-run to actually remove these worktrees\n")
			return nil
		}

		if !cleanForce {
			fmt.Printf("\nRemove %d worktree(s)? [y/N]: ", len(toRemove))
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			if strings.ToLower(strings.TrimSpace(response)) != "y" {
				fmt.Println("Operation cancelled")
				return nil
			}
		}

		removed := 0
		for _, branch := range toRemove {
			fmt.Printf("🗑️  Removing worktree '%s'...\n", branch)
			if err := manager.RemoveWorktree(branch, true, false); err != nil {
				fmt.Printf("⚠️  Failed to remove '%s': %v\n", branch, err)
				continue
			}
			removed++
		}

		fmt.Printf("✅ Successfully removed %d worktree(s)\n", removed)
		return nil
	},
}

func init() {
	cleanCmd.Flags().BoolVar(&cleanDryRun, "dry-run", false, "Show what would be removed without actually removing")
	cleanCmd.Flags().BoolVar(&cleanForce, "force", false, "Force removal without confirmation")
}