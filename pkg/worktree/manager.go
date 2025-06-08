// Package worktree provides functionality for managing Git worktrees.
package worktree

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/knwoop/giwo/internal/errors"
)

// Manager handles Git worktree operations.
type Manager struct {
	repoRoot    string
	worktreeDir string
}

// New creates a new Manager instance.
// It returns an error if the current directory is not in a Git repository.
func New() (*Manager, error) {
	repoRoot, err := getGitRoot()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrNotGitRepository, err)
	}

	worktreeDir := filepath.Join(repoRoot, ".worktree")

	return &Manager{
		repoRoot:    repoRoot,
		worktreeDir: worktreeDir,
	}, nil
}

// WorktreeDir returns the directory where worktrees are stored.
func (m *Manager) WorktreeDir() string {
	return m.worktreeDir
}

// RepoRoot returns the root directory of the Git repository.
func (m *Manager) RepoRoot() string {
	return m.repoRoot
}

// List returns all worktrees with their current status.
func (m *Manager) List(ctx context.Context) ([]*Worktree, error) {
	cmd := exec.CommandContext(ctx, "git", "worktree", "list", "--porcelain")
	cmd.Dir = m.repoRoot
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.NewGitError("worktree list", []string{"--porcelain"}, err)
	}

	worktrees, err := m.parseWorktreeList(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse worktree list: %w", err)
	}

	// Enrich each worktree with additional information
	for _, wt := range worktrees {
		if err := m.enrichWorktree(ctx, wt); err != nil {
			// Log warning but continue with other worktrees
			continue
		}
	}

	return worktrees, nil
}

// Create creates a new worktree and branch.
func (m *Manager) Create(ctx context.Context, branchName, baseBranch string, force bool) error {
	worktreePath := filepath.Join(m.worktreeDir, branchName)

	if !force {
		if _, err := os.Stat(worktreePath); err == nil {
			return fmt.Errorf("%w: %s", errors.ErrWorktreeExists, worktreePath)
		}
	}

	if err := os.MkdirAll(m.worktreeDir, 0o755); err != nil {
		return fmt.Errorf("failed to create worktree directory: %w", err)
	}

	if baseBranch == "" {
		baseBranch = "main"
	}

	// Fetch the latest changes
	if err := m.runGitCommand(ctx, "fetch", "--prune"); err != nil {
		return fmt.Errorf("failed to fetch: %w", err)
	}

	// Create the worktree
	args := []string{"worktree", "add", "-b", branchName, worktreePath, fmt.Sprintf("origin/%s", baseBranch)}
	if err := m.runGitCommand(ctx, args...); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	// Copy configuration files
	if err := m.copyConfigFiles(worktreePath); err != nil {
		// This is not a fatal error, just log a warning
		fmt.Printf("⚠️  Warning: failed to copy config files: %v\n", err)
	}

	return nil
}

// Remove removes a worktree and optionally its branch.
func (m *Manager) Remove(ctx context.Context, branchName string, force, keepBranch bool) error {
	worktreePath := filepath.Join(m.worktreeDir, branchName)

	if !force {
		if !m.confirmRemoval(branchName, worktreePath) {
			return errors.ErrOperationCancelled
		}
	}

	// Remove the worktree
	if err := m.runGitCommand(ctx, "worktree", "remove", worktreePath); err != nil {
		// Try with force flag
		if err := m.runGitCommand(ctx, "worktree", "remove", "--force", worktreePath); err != nil {
			return fmt.Errorf("failed to remove worktree: %w", err)
		}
	}

	// Remove the branch if requested
	if !keepBranch {
		if err := m.runGitCommand(ctx, "branch", "-D", branchName); err != nil {
			fmt.Printf("⚠️  Warning: failed to delete branch '%s': %v\n", branchName, err)
		}
	}

	return nil
}

// GetMergedBranches returns a list of branches that have been merged.
func (m *Manager) GetMergedBranches(ctx context.Context) ([]string, error) {
	// Try main first, then master
	for _, mainBranch := range []string{"main", "master"} {
		cmd := exec.CommandContext(ctx, "git", "branch", "--merged", fmt.Sprintf("origin/%s", mainBranch))
		cmd.Dir = m.repoRoot
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		return m.parseBranchList(string(output)), nil
	}

	return nil, fmt.Errorf("failed to determine merged branches: no main/master branch found")
}

// GetRepoInfo extracts GitHub repository information from Git remote.
func (m *Manager) GetRepoInfo() (owner, repo string, err error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = m.repoRoot
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get remote URL: %w", err)
	}

	remoteURL := strings.TrimSpace(string(output))
	owner, repo = parseGitHubURL(remoteURL)

	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("failed to parse GitHub repository info from: %s", remoteURL)
	}

	return owner, repo, nil
}

// parseWorktreeList parses the output of 'git worktree list --porcelain'.
func (m *Manager) parseWorktreeList(output string) ([]*Worktree, error) {
	var worktrees []*Worktree
	lines := strings.Split(output, "\n")

	var current *Worktree
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if current != nil {
				worktrees = append(worktrees, current)
				current = nil
			}
			continue
		}

		if strings.HasPrefix(line, "worktree ") {
			path := strings.TrimPrefix(line, "worktree ")
			current = &Worktree{Path: path}
		} else if strings.HasPrefix(line, "branch ") && current != nil {
			branch := strings.TrimPrefix(line, "branch ")
			if strings.HasPrefix(branch, "refs/heads/") {
				branch = strings.TrimPrefix(branch, "refs/heads/")
			}
			current.Branch = branch
		} else if strings.HasPrefix(line, "HEAD ") && current != nil {
			current.Branch = "HEAD"
		}
	}

	if current != nil {
		worktrees = append(worktrees, current)
	}

	return worktrees, nil
}

// enrichWorktree adds status information to a worktree.
func (m *Manager) enrichWorktree(ctx context.Context, wt *Worktree) error {
	wt.IsMain = wt.Path == m.repoRoot

	if err := m.getGitStatus(ctx, wt); err != nil {
		return err
	}

	if err := m.getCommitInfo(ctx, wt); err != nil {
		return err
	}

	if err := m.getRemoteStatus(ctx, wt); err != nil {
		return err
	}

	return nil
}

// getGitStatus populates the status fields of a worktree.
func (m *Manager) getGitStatus(ctx context.Context, wt *Worktree) error {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = wt.Path
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	statusOutput := strings.TrimSpace(string(output))
	wt.IsClean = len(statusOutput) == 0

	if !wt.IsClean {
		lines := strings.Split(statusOutput, "\n")
		for _, line := range lines {
			if len(line) < 2 {
				continue
			}

			status := line[:2]
			switch {
			case strings.Contains(status, "M"):
				wt.Modified++
			case strings.Contains(status, "A"):
				wt.Added++
			case strings.Contains(status, "D"):
				wt.Deleted++
			}
		}
	}

	return nil
}

// getCommitInfo populates commit-related fields of a worktree.
func (m *Manager) getCommitInfo(ctx context.Context, wt *Worktree) error {
	cmd := exec.CommandContext(ctx, "git", "log", "-1", "--format=%s|%ct")
	cmd.Dir = wt.Path
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	parts := strings.Split(strings.TrimSpace(string(output)), "|")
	if len(parts) >= 2 {
		wt.LastCommit = parts[0]
		if timestamp, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
			wt.CommitTime = time.Unix(timestamp, 0)
			wt.CommitAge = formatTimeAgo(wt.CommitTime)
		}
	}

	return nil
}

// getRemoteStatus populates remote tracking information of a worktree.
func (m *Manager) getRemoteStatus(ctx context.Context, wt *Worktree) error {
	if wt.Branch == "HEAD" || wt.Branch == "" {
		return nil
	}

	cmd := exec.CommandContext(ctx, "git", "rev-list", "--count", "--left-right",
		fmt.Sprintf("origin/%s...HEAD", wt.Branch))
	cmd.Dir = wt.Path
	output, err := cmd.Output()
	if err != nil {
		// Not an error if remote branch doesn't exist
		return nil
	}

	parts := strings.Fields(strings.TrimSpace(string(output)))
	if len(parts) >= 2 {
		if behind, err := strconv.Atoi(parts[0]); err == nil {
			wt.Behind = behind
		}
		if ahead, err := strconv.Atoi(parts[1]); err == nil {
			wt.Ahead = ahead
		}
	}

	return nil
}

// runGitCommand runs a git command in the repository root.
func (m *Manager) runGitCommand(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = m.repoRoot
	if err := cmd.Run(); err != nil {
		return errors.NewGitError(args[0], args[1:], err)
	}
	return nil
}

// copyConfigFiles copies configuration files to the new worktree.
func (m *Manager) copyConfigFiles(destPath string) error {
	for _, file := range ConfigFiles {
		srcPath := filepath.Join(m.repoRoot, file)
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue
		}

		destFilePath := filepath.Join(destPath, file)
		if err := copyFile(srcPath, destFilePath); err != nil {
			return err
		}
	}
	return nil
}

// confirmRemoval prompts the user for confirmation.
func (m *Manager) confirmRemoval(branchName, worktreePath string) bool {
	fmt.Printf("Remove worktree '%s' at %s? [y/N]: ", branchName, worktreePath)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	return strings.ToLower(strings.TrimSpace(response)) == "y"
}

// parseBranchList parses the output of 'git branch --merged'.
func (m *Manager) parseBranchList(output string) []string {
	var branches []string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		branch := strings.TrimSpace(line)
		branch = strings.TrimPrefix(branch, "* ")
		if branch != "" && !isProtectedBranch(branch) {
			branches = append(branches, branch)
		}
	}
	return branches
}

// Helper functions

// GetCurrentBranch returns the current branch name.
func (m *Manager) GetCurrentBranch(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = m.repoRoot
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	
	branch := strings.TrimSpace(string(output))
	if branch == "HEAD" {
		// We're in detached HEAD state, try to get symbolic name
		cmd = exec.CommandContext(ctx, "git", "describe", "--contains", "--all", "HEAD")
		cmd.Dir = m.repoRoot
		output, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("in detached HEAD state and cannot determine branch")
		}
		branch = strings.TrimSpace(string(output))
		// Remove refs/heads/ prefix if present
		if strings.HasPrefix(branch, "heads/") {
			branch = strings.TrimPrefix(branch, "heads/")
		}
	}
	
	return branch, nil
}

// getGitRoot returns the root directory of the Git repository.
func getGitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	cmd := exec.Command("cp", src, dst)
	return cmd.Run()
}

// formatTimeAgo formats a time duration as a human-readable string.
func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%dm ago", minutes)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		return fmt.Sprintf("%dh ago", hours)
	default:
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	}
}

// parseGitHubURL extracts owner and repo from a GitHub URL.
func parseGitHubURL(url string) (owner, repo string) {
	url = strings.TrimSpace(url)

	// SSH format: git@github.com:owner/repo.git
	sshRegex := regexp.MustCompile(`git@github\.com:([^/]+)/(.+?)(?:\.git)?$`)
	if matches := sshRegex.FindStringSubmatch(url); len(matches) == 3 {
		return matches[1], matches[2]
	}

	// HTTPS format: https://github.com/owner/repo.git
	httpsRegex := regexp.MustCompile(`https://github\.com/([^/]+)/(.+?)(?:\.git)?$`)
	if matches := httpsRegex.FindStringSubmatch(url); len(matches) == 3 {
		return matches[1], matches[2]
	}

	return "", ""
}

// isProtectedBranch returns true if the branch should not be automatically removed.
func isProtectedBranch(branch string) bool {
	protected := []string{"main", "master", "develop", "dev"}
	for _, p := range protected {
		if branch == p {
			return true
		}
	}
	return false
}
