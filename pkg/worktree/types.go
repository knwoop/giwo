package worktree

import (
	"time"
)

type Worktree struct {
	Path       string    `json:"path"`
	Branch     string    `json:"branch"`
	IsMain     bool      `json:"is_main"`
	IsClean    bool      `json:"is_clean"`
	Ahead      int       `json:"ahead"`
	Behind     int       `json:"behind"`
	Modified   int       `json:"modified"`
	Added      int       `json:"added"`
	Deleted    int       `json:"deleted"`
	LastCommit string    `json:"last_commit"`
	CommitAge  string    `json:"commit_age"`
	CommitTime time.Time `json:"commit_time"`
}

type WorktreeStats struct {
	Total      int `json:"total"`
	Active     int `json:"active"`
	Merged     int `json:"merged"`
	Dirty      int `json:"dirty"`
	MainExists bool `json:"main_exists"`
}

type OutputFormat string

const (
	FormatTable  OutputFormat = "table"
	FormatJSON   OutputFormat = "json"
	FormatSimple OutputFormat = "simple"
)

var ConfigFiles = []string{
	".env",
	".env.local",
	".rgignore",
	".gitignore",
	".editorconfig",
	".prettierrc",
}