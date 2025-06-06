package cmd

import (
	"fmt"

	"github.com/knwoop/gwt/internal/utils"
	"github.com/knwoop/gwt/pkg/github"
	"github.com/knwoop/gwt/pkg/worktree"
	"github.com/spf13/cobra"
)

var (
	createForce bool
	createBase  string
)

var createCmd = &cobra.Command{
	Use:   "create <branch-name>",
	Short: "Create a new worktree",
	Long: `Create a new worktree based on the default branch.
The worktree will be placed in .worktree/<branch-name> directory and 
automatically create and switch to the new branch.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		branchName := args[0]
		
		if err := utils.ValidateBranchName(branchName); err != nil {
			return fmt.Errorf("invalid branch name: %w", err)
		}
		
		manager, err := worktree.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize manager: %w", err)
		}

		baseBranch := createBase
		if baseBranch == "" {
			owner, repo, err := github.GetRepoInfo()
			if err != nil {
				fmt.Printf("⚠️  Warning: failed to get repo info, using 'main' as base: %v\n", err)
				baseBranch = "main"
			} else {
				client := github.NewClient()
				baseBranch, err = client.GetDefaultBranch(owner, repo)
				if err != nil {
					fmt.Printf("⚠️  Warning: failed to get default branch, using 'main': %v\n", err)
					baseBranch = "main"
				}
			}
		}

		fmt.Printf("🌱 Creating worktree '%s' based on '%s'...\n", branchName, baseBranch)
		
		if err := manager.CreateWorktree(branchName, baseBranch, createForce); err != nil {
			return fmt.Errorf("failed to create worktree: %w", err)
		}

		worktreePath := fmt.Sprintf("%s/%s", manager.GetWorktreeDir(), branchName)
		fmt.Printf("✅ Worktree created successfully at: %s\n", worktreePath)
		fmt.Printf("💡 Run 'cd %s' to switch to the new worktree\n", worktreePath)
		
		return nil
	},
}

func init() {
	createCmd.Flags().BoolVar(&createForce, "force", false, "Force creation even if directory exists")
	createCmd.Flags().StringVar(&createBase, "base", "", "Base branch to create worktree from (default: repository default branch)")
}