package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type Repository struct {
	DefaultBranch string `json:"default_branch"`
}

type Client struct {
	token string
}

func NewClient() *Client {
	return &Client{
		token: os.Getenv("GITHUB_TOKEN"),
	}
}

func (c *Client) GetDefaultBranch(owner, repo string) (string, error) {
	if c.token == "" {
		return c.fallbackDefaultBranch()
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.fallbackDefaultBranch()
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.fallbackDefaultBranch()
	}

	var repository Repository
	if err := json.NewDecoder(resp.Body).Decode(&repository); err != nil {
		return c.fallbackDefaultBranch()
	}

	return repository.DefaultBranch, nil
}

func (c *Client) fallbackDefaultBranch() (string, error) {
	branches := []string{"main", "master", "develop"}
	
	for _, branch := range branches {
		cmd := exec.Command("git", "rev-parse", "--verify", fmt.Sprintf("origin/%s", branch))
		if err := cmd.Run(); err == nil {
			return branch, nil
		}
	}
	
	return "main", nil
}

func GetRepoInfo() (string, string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get remote URL: %w", err)
	}

	remoteURL := strings.TrimSpace(string(output))
	owner, repo := parseGitHubURL(remoteURL)
	
	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("failed to parse GitHub repository info from: %s", remoteURL)
	}
	
	return owner, repo, nil
}

func parseGitHubURL(url string) (owner, repo string) {
	url = strings.TrimSpace(url)
	
	if strings.HasPrefix(url, "git@github.com:") {
		parts := strings.TrimPrefix(url, "git@github.com:")
		parts = strings.TrimSuffix(parts, ".git")
		segments := strings.Split(parts, "/")
		if len(segments) >= 2 {
			return segments[0], segments[1]
		}
	}
	
	if strings.HasPrefix(url, "https://github.com/") {
		parts := strings.TrimPrefix(url, "https://github.com/")
		parts = strings.TrimSuffix(parts, ".git")
		segments := strings.Split(parts, "/")
		if len(segments) >= 2 {
			return segments[0], segments[1]
		}
	}
	
	return "", ""
}