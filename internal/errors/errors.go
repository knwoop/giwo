// Package errors defines custom error types for gwt.
package errors

import (
	"errors"
	"fmt"
)

// Sentinel errors following the style guide.
var (
	ErrNotGitRepository     = errors.New("not in a git repository")
	ErrWorktreeExists       = errors.New("worktree already exists")
	ErrWorktreeNotFound     = errors.New("worktree not found")
	ErrBranchNotFound       = errors.New("branch not found")
	ErrInvalidBranchName    = errors.New("invalid branch name")
	ErrGitHubAPIUnavailable = errors.New("github API unavailable")
	ErrOperationCancelled   = errors.New("operation cancelled by user")
)

// ValidationError represents a validation error with details.
type ValidationError struct {
	Field string
	Value string
	Err   error
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s=%q: %v", e.Field, e.Value, e.Err)
}

// Unwrap returns the underlying error.
func (e *ValidationError) Unwrap() error {
	return e.Err
}

// GitError represents a git operation error.
type GitError struct {
	Operation string
	Args      []string
	Err       error
}

// Error implements the error interface.
func (e *GitError) Error() string {
	return fmt.Sprintf("git %s failed: %v", e.Operation, e.Err)
}

// Unwrap returns the underlying error.
func (e *GitError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new validation error.
func NewValidationError(field, value string, err error) *ValidationError {
	return &ValidationError{
		Field: field,
		Value: value,
		Err:   err,
	}
}

// NewGitError creates a new git error.
func NewGitError(operation string, args []string, err error) *GitError {
	return &GitError{
		Operation: operation,
		Args:      args,
		Err:       err,
	}
}
