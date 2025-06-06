#!/bin/bash
# gwt shell integration for directory switching
# Source this file in your shell configuration (.bashrc, .zshrc, etc.)

# Function to switch to a worktree directory
gwt-switch() {
    local selected_path
    
    # Use gwt switch with --print flag to get the path
    selected_path=$(gwt switch --print "$@")
    
    # Check if a path was returned (not cancelled)
    if [[ -n "$selected_path" && "$selected_path" != "Operation cancelled." ]]; then
        echo "🔄 Switching to: $selected_path"
        cd "$selected_path" || {
            echo "❌ Failed to change to directory: $selected_path"
            return 1
        }
        
        # Show current branch and status
        if command -v git >/dev/null 2>&1; then
            echo "📍 Current branch: $(git branch --show-current 2>/dev/null || echo 'unknown')"
        fi
    else
        echo "❌ No worktree selected or operation cancelled"
        return 1
    fi
}

# Alias for convenience
alias gws='gwt-switch'

# Fuzzy search version
gwt-fuzzy() {
    gwt-switch --fuzzy "$@"
}

# Alias for fuzzy search
alias gwf='gwt-fuzzy'

echo "🎉 gwt shell integration loaded!"
echo "   Use 'gwt-switch' or 'gws' to switch directories"
echo "   Use 'gwt-fuzzy' or 'gwf' for fuzzy search"