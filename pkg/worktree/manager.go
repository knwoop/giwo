package worktree

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Manager struct {
	repoRoot    string
	worktreeDir string
}

func NewManager() (*Manager, error) {
	repoRoot, err := getGitRoot()
	if err != nil {
		return nil, fmt.Errorf("not in a git repository: %w", err)
	}

	worktreeDir := filepath.Join(repoRoot, ".worktree")
	
	return &Manager{
		repoRoot:    repoRoot,
		worktreeDir: worktreeDir,
	}, nil
}

func (m *Manager) GetWorktreeDir() string {
	return m.worktreeDir
}

func (m *Manager) GetRepoRoot() string {
	return m.repoRoot
}

func (m *Manager) ListWorktrees() ([]*Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = m.repoRoot
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	var worktrees []*Worktree
	lines := strings.Split(string(output), "\n")
	
	var currentWorktree *Worktree
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if currentWorktree != nil {
				worktrees = append(worktrees, currentWorktree)
				currentWorktree = nil
			}
			continue
		}

		if strings.HasPrefix(line, "worktree ") {
			path := strings.TrimPrefix(line, "worktree ")
			currentWorktree = &Worktree{
				Path: path,
			}
		} else if strings.HasPrefix(line, "branch ") && currentWorktree != nil {
			branch := strings.TrimPrefix(line, "branch ")
			if strings.HasPrefix(branch, "refs/heads/") {
				branch = strings.TrimPrefix(branch, "refs/heads/")
			}
			currentWorktree.Branch = branch
		} else if strings.HasPrefix(line, "HEAD ") && currentWorktree != nil {
			currentWorktree.Branch = "HEAD"
		}
	}
	
	if currentWorktree != nil {
		worktrees = append(worktrees, currentWorktree)
	}

	for _, wt := range worktrees {
		if err := m.enrichWorktreeInfo(wt); err != nil {
			continue
		}
	}

	return worktrees, nil
}

func (m *Manager) enrichWorktreeInfo(wt *Worktree) error {
	wt.IsMain = wt.Path == m.repoRoot

	if err := m.getGitStatus(wt); err != nil {
		return err
	}

	if err := m.getCommitInfo(wt); err != nil {
		return err
	}

	if err := m.getRemoteStatus(wt); err != nil {
		return err
	}

	return nil
}

func (m *Manager) getGitStatus(wt *Worktree) error {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = wt.Path
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	wt.IsClean = len(strings.TrimSpace(string(output))) == 0
	
	lines := strings.Split(string(output), "\n")
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

	return nil
}

func (m *Manager) getCommitInfo(wt *Worktree) error {
	cmd := exec.Command("git", "log", "-1", "--format=%s|%ct")
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

func (m *Manager) getRemoteStatus(wt *Worktree) error {
	if wt.Branch == "HEAD" || wt.Branch == "" {
		return nil
	}

	cmd := exec.Command("git", "rev-list", "--count", "--left-right", fmt.Sprintf("origin/%s...HEAD", wt.Branch))
	cmd.Dir = wt.Path
	output, err := cmd.Output()
	if err != nil {
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

func (m *Manager) CreateWorktree(branchName, baseBranch string, force bool) error {
	worktreePath := filepath.Join(m.worktreeDir, branchName)
	
	if !force {
		if _, err := os.Stat(worktreePath); err == nil {
			return fmt.Errorf("worktree directory already exists: %s", worktreePath)
		}
	}

	if err := os.MkdirAll(m.worktreeDir, 0755); err != nil {
		return fmt.Errorf("failed to create worktree directory: %w", err)
	}

	if baseBranch == "" {
		baseBranch = "main"
	}

	cmd := exec.Command("git", "fetch", "--prune")
	cmd.Dir = m.repoRoot
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch: %w", err)
	}

	cmd = exec.Command("git", "worktree", "add", "-b", branchName, worktreePath, fmt.Sprintf("origin/%s", baseBranch))
	cmd.Dir = m.repoRoot
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	if err := m.copyConfigFiles(worktreePath); err != nil {
		fmt.Printf("⚠️  Warning: failed to copy config files: %v\n", err)
	}

	return nil
}

func (m *Manager) RemoveWorktree(branchName string, force, keepBranch bool) error {
	worktreePath := filepath.Join(m.worktreeDir, branchName)
	
	if !force {
		fmt.Printf("Remove worktree '%s' at %s? [y/N]: ", branchName, worktreePath)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			return fmt.Errorf("operation cancelled")
		}
	}

	cmd := exec.Command("git", "worktree", "remove", worktreePath)
	cmd.Dir = m.repoRoot
	if err := cmd.Run(); err != nil {
		cmd = exec.Command("git", "worktree", "remove", "--force", worktreePath)
		cmd.Dir = m.repoRoot
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to remove worktree: %w", err)
		}
	}

	if !keepBranch {
		cmd = exec.Command("git", "branch", "-D", branchName)
		cmd.Dir = m.repoRoot
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️  Warning: failed to delete branch '%s': %v\n", branchName, err)
		}
	}

	return nil
}

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

func getGitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func copyFile(src, dst string) error {
	cmd := exec.Command("cp", src, dst)
	return cmd.Run()
}

func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)
	
	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%dm ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		return fmt.Sprintf("%dh ago", hours)
	} else {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	}
}

func (m *Manager) GetMergedBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "--merged", "origin/main")
	cmd.Dir = m.repoRoot
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("git", "branch", "--merged", "origin/master")
		cmd.Dir = m.repoRoot
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get merged branches: %w", err)
		}
	}

	var branches []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		branch := strings.TrimSpace(line)
		branch = strings.TrimPrefix(branch, "* ")
		if branch != "" && branch != "main" && branch != "master" && branch != "develop" {
			branches = append(branches, branch)
		}
	}

	return branches, nil
}

func (m *Manager) GetRepoInfo() (string, string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = m.repoRoot
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get remote URL: %w", err)
	}

	remoteURL := strings.TrimSpace(string(output))
	owner, repo := parseGitHubURL(remoteURL)
	
	return owner, repo, nil
}

func parseGitHubURL(url string) (owner, repo string) {
	sshRegex := regexp.MustCompile(`git@github\.com:([^/]+)/(.+?)(?:\.git)?$`)
	httpsRegex := regexp.MustCompile(`https://github\.com/([^/]+)/(.+?)(?:\.git)?$`)
	
	if matches := sshRegex.FindStringSubmatch(url); len(matches) == 3 {
		return matches[1], matches[2]
	}
	
	if matches := httpsRegex.FindStringSubmatch(url); len(matches) == 3 {
		return matches[1], matches[2]
	}
	
	return "", ""
}