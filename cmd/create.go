package cmd

import (
	"context"
	"fmt"

	"github.com/knwoop/giwo/internal/utils"
	"github.com/knwoop/giwo/pkg/github"
	"github.com/knwoop/giwo/pkg/worktree"
	"github.com/spf13/cobra"
)

var (
	createForce bool
	createBase  string
)

var createCmd = &cobra.Command{
	Use:   "create <branch-name>",
	Short: "Create a new worktree",
	Long: `Create a new worktree based on the current branch.
The worktree will be placed in .worktree/<branch-name> directory and 
automatically create and switch to the new branch.

By default, the new worktree will be created from the current branch.
Use --base to specify a different base branch.`,
	Args: cobra.ExactArgs(1),
	RunE: runCreateCommand,
}

func runCreateCommand(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	branchName := args[0]

	if err := utils.ValidateBranchName(branchName); err != nil {
		return fmt.Errorf("invalid branch name: %w", err)
	}

	manager, err := worktree.New()
	if err != nil {
		return fmt.Errorf("failed to initialize manager: %w", err)
	}

	baseBranch := createBase
	if baseBranch == "" {
		// Use current branch as default
		baseBranch, err = manager.GetCurrentBranch(ctx)
		if err != nil {
			// Fallback to repository default branch
			fmt.Printf("⚠️  Warning: failed to get current branch, trying repository default: %v\n", err)
			baseBranch, err = determineBaseBranch(ctx)
			if err != nil {
				fmt.Printf("⚠️  Warning: failed to determine base branch, using 'main': %v\n", err)
				baseBranch = "main"
			}
		}
	}

	fmt.Printf("🌱 Creating worktree '%s' based on '%s'...\n", branchName, baseBranch)

	if err := manager.Create(ctx, branchName, baseBranch, createForce); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	worktreePath := fmt.Sprintf("%s/%s", manager.WorktreeDir(), branchName)
	fmt.Printf("✅ Worktree created successfully at: %s\n", worktreePath)
	fmt.Printf("💡 Run 'cd %s' to switch to the new worktree\n", worktreePath)

	return nil
}

func determineBaseBranch(ctx context.Context) (string, error) {
	owner, repo, err := github.GetRepoInfo(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get repo info: %w", err)
	}

	client := github.New()
	baseBranch, err := client.GetDefaultBranch(ctx, owner, repo)
	if err != nil {
		return "", fmt.Errorf("failed to get default branch: %w", err)
	}

	return baseBranch, nil
}

func init() {
	createCmd.Flags().BoolVar(&createForce, "force", false, "Force creation even if directory exists")
	createCmd.Flags().StringVar(&createBase, "base", "", "Base branch to create worktree from (default: current branch)")
}
