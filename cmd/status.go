package cmd

import (
	"fmt"

	"github.com/knwoop/giwo/pkg/worktree"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show worktree statistics",
	Long:  `Display statistics about worktrees and provide recommended actions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := worktree.New()
		if err != nil {
			return fmt.Errorf("failed to initialize manager: %w", err)
		}

		ctx := cmd.Context()
		worktrees, err := manager.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list worktrees: %w", err)
		}

		stats := calculateStats(worktrees)

		fmt.Println("📊 Worktree Statistics")
		fmt.Printf("  Total worktrees: %d\n", stats.Total)
		fmt.Printf("  Active worktrees: %d\n", stats.Active)
		fmt.Printf("  Dirty worktrees: %d\n", stats.Dirty)
		fmt.Printf("  Main worktree: %s\n", formatBool(stats.MainExists))

		if stats.Dirty > 0 {
			fmt.Printf("\n⚠️  %d worktree(s) have uncommitted changes\n", stats.Dirty)
		}

		mergedBranches, err := manager.GetMergedBranches(ctx)
		if err == nil && len(mergedBranches) > 0 {
			fmt.Printf("\n🧹 %d merged branch(es) can be cleaned up:\n", len(mergedBranches))
			for _, branch := range mergedBranches {
				fmt.Printf("  - %s\n", branch)
			}
			fmt.Printf("\n💡 Run 'gwt clean' to remove merged worktrees\n")
		}

		if stats.Total == 1 && stats.MainExists {
			fmt.Printf("\n💡 Run 'gwt create <branch-name>' to create your first worktree\n")
		}

		return nil
	},
}

func calculateStats(worktrees []*worktree.Worktree) worktree.Stats {
	stats := worktree.Stats{
		Total: len(worktrees),
	}

	for _, wt := range worktrees {
		if wt.IsMain {
			stats.MainExists = true
		} else {
			stats.Active++
		}

		if !wt.IsClean {
			stats.Dirty++
		}
	}

	return stats
}

func formatBool(b bool) string {
	if b {
		return "✅ exists"
	}
	return "❌ missing"
}
