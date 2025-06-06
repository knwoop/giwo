# gwt - Git WorkTree Manager

A CLI tool for efficiently managing Git worktrees. Supports parallel work across multiple branches and manages the entire lifecycle of worktrees.

## Installation

```bash
git clone https://github.com/knwoop/gwt.git
cd gwt
make build
make install  # Optional: install to /usr/local/bin
```

## Commands

### `gwt create <branch-name>`

Create a new worktree based on the default branch.

```bash
gwt create feature-auth
gwt create bugfix-login --base develop
gwt create experiment-ui --force
```

**Options:**
- `--base <branch>` - Base branch to create worktree from (default: repository default branch)
- `--force` - Force creation even if directory exists

**Features:**
- Places worktree in `.worktree/<branch-name>`
- Automatically creates and switches to new branch
- Copies config files (.env, .gitignore, .editorconfig, etc.)
- Fetches default branch via GitHub API (requires GITHUB_TOKEN)

### `gwt remove <branch-name>`

Remove a worktree and optionally its local branch.

```bash
gwt remove feature-auth
gwt remove bugfix-login --keep-branch
gwt remove old-feature --force
```

**Aliases:** `rm`, `delete`

**Options:**
- `--force` - Force removal without confirmation
- `--keep-branch` - Keep the local branch after removing worktree

### `gwt list`

Display all worktrees with status information.

```bash
gwt list
gwt list --verbose
gwt list --format json
```

**Aliases:** `ls`

**Options:**
- `--verbose` - Show detailed information (commits, changes, etc.)
- `--format <table|json|simple>` - Output format

### `gwt status`

Show worktree statistics and recommendations.

```bash
gwt status
```

**Output:**
- Total worktrees count
- Active vs dirty worktrees
- Merged branches that can be cleaned
- Recommended actions

### `gwt clean`

Batch remove worktrees for merged branches.

```bash
gwt clean
gwt clean --dry-run
gwt clean --force
```

**Options:**
- `--dry-run` - Show what would be removed without actually removing
- `--force` - Force removal without confirmation

**Features:**
- Automatically detects merged branches
- Excludes main/master/develop branches
- Shows branch status before removal

### `gwt prune`

Remove administrative files for orphaned worktrees.

```bash
gwt prune
```

Wrapper around `git worktree prune -v`.

## GitHub Integration

Set `GITHUB_TOKEN` environment variable to enable:
- Automatic default branch detection
- Better API rate limits

```bash
export GITHUB_TOKEN=your_token_here
```

## Directory Structure

```
project-root/
├── .worktree/
│   ├── feature-auth/     # feature-auth branch
│   ├── bugfix-login/     # bugfix-login branch
│   └── experiment-ui/    # experiment-ui branch
└── main-worktree/        # Main worktree
```

## Examples

```bash
# Create a new feature branch worktree
gwt create feature-auth

# Work in the new worktree
cd .worktree/feature-auth

# List all worktrees
gwt list --verbose

# Check status and get recommendations
gwt status

# Clean up merged branches
gwt clean --dry-run
gwt clean

# Remove specific worktree
gwt remove feature-auth
```