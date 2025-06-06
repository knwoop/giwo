// Package github provides GitHub API integration for gwt.
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const (
	// GitHubAPIBaseURL is the base URL for GitHub API.
	GitHubAPIBaseURL = "https://api.github.com"

	// DefaultRequestTimeout is the default timeout for HTTP requests.
	DefaultRequestTimeout = 10 * time.Second
)

// Repository represents a GitHub repository response.
type Repository struct {
	DefaultBranch string `json:"default_branch"`
}

// Client handles GitHub API interactions.
type Client struct {
	token      string
	httpClient *http.Client
}

// New creates a new GitHub client.
// It uses the GITHUB_TOKEN environment variable for authentication.
func New() *Client {
	return &Client{
		token: os.Getenv("GITHUB_TOKEN"),
		httpClient: &http.Client{
			Timeout: DefaultRequestTimeout,
		},
	}
}

// GetDefaultBranch returns the default branch for a GitHub repository.
// It falls back to local Git inspection if the API is unavailable.
func (c *Client) GetDefaultBranch(ctx context.Context, owner, repo string) (string, error) {
	if c.token == "" {
		return c.fallbackDefaultBranch(ctx)
	}

	url := fmt.Sprintf("%s/repos/%s/%s", GitHubAPIBaseURL, owner, repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "gwt-cli")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Fall back to local detection on network errors
		return c.fallbackDefaultBranch(ctx)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.fallbackDefaultBranch(ctx)
	}

	var repository Repository
	if err := json.NewDecoder(resp.Body).Decode(&repository); err != nil {
		return c.fallbackDefaultBranch(ctx)
	}

	return repository.DefaultBranch, nil
}

// fallbackDefaultBranch determines the default branch by checking local Git references.
func (c *Client) fallbackDefaultBranch(ctx context.Context) (string, error) {
	candidates := []string{"main", "master", "develop"}

	for _, branch := range candidates {
		cmd := exec.CommandContext(ctx, "git", "rev-parse", "--verify", fmt.Sprintf("origin/%s", branch))
		if err := cmd.Run(); err == nil {
			return branch, nil
		}
	}

	// Default to "main" if no remote branches are found
	return "main", nil
}

// GetRepoInfo extracts GitHub repository information from Git remote configuration.
func GetRepoInfo(ctx context.Context) (owner, repo string, err error) {
	cmd := exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
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

// parseGitHubURL extracts owner and repository name from a GitHub URL.
// It supports both SSH and HTTPS formats.
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
