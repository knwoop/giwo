package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "giwo",
	Short: "Git WorkTree Manager - Efficiently manage Git worktrees",
	Long: `giwo is a CLI tool for efficiently managing Git worktrees.
It supports parallel work across multiple branches and manages 
the entire lifecycle of worktrees.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(pruneCmd)
	rootCmd.AddCommand(switchCmd)
}
