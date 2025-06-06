package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/knwoop/gwt/internal/ui"
	"github.com/knwoop/gwt/pkg/worktree"
	"github.com/spf13/cobra"
)

var (
	switchFilter string
	switchPrint  bool
	switchFuzzy  bool
)

var switchCmd = &cobra.Command{
	Use:     "switch [filter]",
	Aliases: []string{"sw"},
	Short:   "Switch to a worktree interactively",
	Long: `Switch to a worktree using an interactive selector.
If a filter is provided, only worktrees matching the filter will be shown.
Use --fuzzy for an interactive fuzzy search interface similar to fzf.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSwitchCommand,
}

func runSwitchCommand(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	manager, err := worktree.New()
	if err != nil {
		return fmt.Errorf("failed to initialize manager: %w", err)
	}

	worktrees, err := manager.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	if len(worktrees) == 0 {
		fmt.Println("No worktrees found. Use 'gwt create <branch-name>' to create one.")
		return nil
	}

	var selected *worktree.Worktree

	// Use fuzzy search if requested
	if switchFuzzy {
		fuzzyFinder := ui.NewFuzzyFinder(worktrees)
		selected, err = fuzzyFinder.Search()
	} else {
		// Get filter from args or flag
		filter := switchFilter
		if len(args) > 0 {
			filter = args[0]
		}

		selector := ui.NewSelector(worktrees)

		if filter != "" {
			selected, err = selector.SelectWithFilter(filter)
		} else {
			selected, err = selector.Select()
		}
	}

	if err != nil {
		return fmt.Errorf("selection failed: %w", err)
	}

	if selected == nil {
		fmt.Println("Operation cancelled.")
		return nil
	}

	// If --print flag is set, just print the path
	if switchPrint {
		fmt.Println(selected.Path)
		return nil
	}

	// Check if we're already in the selected worktree
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if currentDir == selected.Path {
		fmt.Printf("Already in worktree '%s'\n", selected.Branch)
		return nil
	}

	// Try to change directory using a subshell
	fmt.Printf("🔄 Switching to worktree '%s' at %s\n", selected.Branch, selected.Path)

	// Since we can't change the parent shell's directory from a child process,
	// we'll provide instructions to the user
	fmt.Printf("💡 Run: cd %s\n", selected.Path)

	// Optionally, try to open a new shell in the directory
	if err := openShellInDirectory(selected.Path); err != nil {
		// If opening a new shell fails, that's okay - we've already given instructions
		fmt.Printf("⚠️  Could not open new shell: %v\n", err)
		fmt.Printf("📝 You can also copy and run: cd %s\n", selected.Path)
	}

	return nil
}

// openShellInDirectory attempts to open a new shell in the specified directory.
func openShellInDirectory(path string) error {
	// Try to determine the user's shell
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	// Try to open a new shell session
	cmd := exec.Command(shell)
	cmd.Dir = path
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("🐚 Opening new shell in %s (exit to return)\n", path)
	return cmd.Run()
}

func init() {
	switchCmd.Flags().StringVarP(&switchFilter, "filter", "f", "", "Filter worktrees by branch name")
	switchCmd.Flags().BoolVarP(&switchPrint, "print", "p", false, "Print the selected worktree path instead of switching")
	switchCmd.Flags().BoolVar(&switchFuzzy, "fuzzy", false, "Use interactive fuzzy search (like fzf)")
}
