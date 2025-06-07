package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/knwoop/giwo/pkg/worktree"
	"github.com/spf13/cobra"
)

var (
	listVerbose bool
	listFormat  string
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all worktrees",
	Long:    `Display a list of all worktrees with their status information.`,
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

		if len(worktrees) == 0 {
			fmt.Println("No worktrees found")
			return nil
		}

		format := worktree.OutputFormat(listFormat)
		switch format {
		case worktree.OutputFormatJSON:
			return printJSON(worktrees)
		case worktree.OutputFormatSimple:
			return printSimple(worktrees)
		default:
			return printTable(worktrees, listVerbose)
		}
	},
}

func printTable(worktrees []*worktree.Worktree, verbose bool) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	if verbose {
		fmt.Fprintf(w, "BRANCH\tPATH\tSTATUS\tAHEAD/BEHIND\tCHANGES\tLAST COMMIT\tAGE\n")
		for _, wt := range worktrees {
			status := "🌱"
			if wt.IsMain {
				status = "🏠"
			} else if !wt.IsClean {
				status = "⚠️"
			}

			changes := fmt.Sprintf("M:%d A:%d D:%d", wt.Modified, wt.Added, wt.Deleted)
			if wt.IsClean {
				changes = "clean"
			}

			aheadBehind := ""
			if wt.Ahead > 0 || wt.Behind > 0 {
				aheadBehind = fmt.Sprintf("+%d/-%d", wt.Ahead, wt.Behind)
			} else {
				aheadBehind = "up-to-date"
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				wt.Branch, wt.Path, status, aheadBehind, changes,
				truncateString(wt.LastCommit, 50), wt.CommitAge)
		}
	} else {
		fmt.Fprintf(w, "BRANCH\tPATH\tSTATUS\n")
		for _, wt := range worktrees {
			status := "🌱"
			if wt.IsMain {
				status = "🏠 main"
			} else if !wt.IsClean {
				status = "⚠️  dirty"
			} else {
				status = "✅ clean"
			}

			fmt.Fprintf(w, "%s\t%s\t%s\n", wt.Branch, wt.Path, status)
		}
	}

	return nil
}

func printJSON(worktrees []*worktree.Worktree) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(worktrees)
}

func printSimple(worktrees []*worktree.Worktree) error {
	for _, wt := range worktrees {
		fmt.Printf("%s\t%s\n", wt.Branch, wt.Path)
	}
	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func init() {
	listCmd.Flags().BoolVarP(&listVerbose, "verbose", "v", false, "Show detailed information")
	listCmd.Flags().StringVar(&listFormat, "format", "table", "Output format (table, json, simple)")
}
