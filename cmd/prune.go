package cmd

import (
	"fmt"
	"os/exec"

	"github.com/knwoop/gwt/pkg/worktree"
	"github.com/spf13/cobra"
)

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove administrative files for orphaned worktrees",
	Long:  `Remove administrative files for orphaned worktrees. This is a wrapper around 'git worktree prune'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := worktree.New()
		if err != nil {
			return fmt.Errorf("failed to initialize manager: %w", err)
		}

		fmt.Println("🧹 Pruning orphaned worktree administrative files...")

		gitCmd := exec.Command("git", "worktree", "prune", "-v")
		gitCmd.Dir = manager.RepoRoot()

		output, err := gitCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to prune worktrees: %w", err)
		}

		if len(output) > 0 {
			fmt.Printf("%s", output)
		} else {
			fmt.Println("✅ No orphaned administrative files found")
		}

		return nil
	},
}
