package cmd

import (
	"fmt"

	"github.com/knwoop/gwt/pkg/worktree"
	"github.com/spf13/cobra"
)

var (
	removeForce      bool
	removeKeepBranch bool
)

var removeCmd = &cobra.Command{
	Use:     "remove <branch-name>",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a worktree",
	Long: `Remove the specified worktree and optionally delete the associated local branch.
By default, the local branch will be deleted unless --keep-branch is specified.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		branchName := args[0]
		
		manager, err := worktree.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize manager: %w", err)
		}

		fmt.Printf("🗑️  Removing worktree '%s'...\n", branchName)
		
		if err := manager.RemoveWorktree(branchName, removeForce, removeKeepBranch); err != nil {
			return fmt.Errorf("failed to remove worktree: %w", err)
		}

		if removeKeepBranch {
			fmt.Printf("✅ Worktree removed successfully (branch kept)\n")
		} else {
			fmt.Printf("✅ Worktree and branch removed successfully\n")
		}
		
		return nil
	},
}

func init() {
	removeCmd.Flags().BoolVar(&removeForce, "force", false, "Force removal without confirmation")
	removeCmd.Flags().BoolVar(&removeKeepBranch, "keep-branch", false, "Keep the local branch after removing worktree")
}