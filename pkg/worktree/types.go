// Package worktree provides types and utilities for managing Git worktrees.
package worktree

import (
	"time"
)

// OutputFormat represents the output format for worktree listings.
type OutputFormat string

// Output format constants.
const (
	OutputFormatTable  OutputFormat = "table"
	OutputFormatJSON   OutputFormat = "json"
	OutputFormatSimple OutputFormat = "simple"
)

// Config file names that should be copied to new worktrees.
var ConfigFiles = []string{
	".editorconfig",
	".env",
	".env.local",
	".gitignore",
	".prettierrc",
	".rgignore",
}

// Worktree represents a Git worktree with its current status.
// Fields are ordered by importance: identifying fields first, then status fields.
type Worktree struct {
	// Identifying fields
	Path   string `json:"path"`
	Branch string `json:"branch"`

	// Status flags
	IsMain  bool `json:"is_main"`
	IsClean bool `json:"is_clean"`

	// Sync status with remote
	Ahead  int `json:"ahead"`
	Behind int `json:"behind"`

	// Local changes count
	Added    int `json:"added"`
	Modified int `json:"modified"`
	Deleted  int `json:"deleted"`

	// Commit information
	LastCommit string    `json:"last_commit"`
	CommitAge  string    `json:"commit_age"`
	CommitTime time.Time `json:"commit_time"`
}

// Stats represents statistics about all worktrees.
type Stats struct {
	Total      int  `json:"total"`
	Active     int  `json:"active"`
	Dirty      int  `json:"dirty"`
	Merged     int  `json:"merged"`
	MainExists bool `json:"main_exists"`
}
