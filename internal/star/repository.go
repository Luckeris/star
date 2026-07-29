// Package star provides core functionality for the Star version control system.
package star

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	DirStar    = ".star"
	DirObjects = "objects"
	DirCommits = "commits"
	DirRefs    = "refs"
	DirHeads   = "heads"
	FileIndex  = "index.json"
	FileHead   = "HEAD"
	RefPrefix  = "ref: "
	RefHeads   = "refs/heads/"
)

var (
	// ErrAlreadyInitialized is returned when attempting to initialize an existing repository.
	ErrAlreadyInitialized = errors.New(".star is already initialized")
	// ErrNotInitialized is returned when executing commands in a non-repository folder.
	ErrNotInitialized = errors.New("repository not initialized (run 'star init' first)")
	// ErrEmptyCommitMessage is returned when a commit message is blank.
	ErrEmptyCommitMessage = errors.New("commit message cannot be empty")
	// ErrNothingToCommit is returned when committing an empty index.
	ErrNothingToCommit = errors.New("nothing to commit (index is empty)")
	// ErrNoCommits is returned when accessing history in a repository without commits.
	ErrNoCommits = errors.New("no commits found")
)

// Repository encapsulates the file system operations for a Star VCS repository.
type Repository struct {
	RootPath string
}

// NewRepository creates a new Repository instance targeting rootPath.
func NewRepository(rootPath string) *Repository {
	if rootPath == "" {
		rootPath = "."
	}
	return &Repository{RootPath: rootPath}
}

// Path returns a path joined relative to the .star directory within the repository root.
func (r *Repository) Path(elem ...string) string {
	slice := append([]string{r.RootPath, DirStar}, elem...)
	return filepath.Join(slice...)
}

// Init initializes the .star directory structure and default branch reference.
func (r *Repository) Init() error {
	starDir := r.Path()
	if _, err := os.Stat(starDir); err == nil {
		return ErrAlreadyInitialized
	}

	dirs := []string{
		r.Path(DirObjects),
		r.Path(DirCommits),
		r.Path(DirRefs),
		r.Path(DirRefs, DirHeads),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	headPath := r.Path(FileHead)
	if err := os.WriteFile(headPath, []byte(RefPrefix+RefHeads+"main"), 0644); err != nil {
		return fmt.Errorf("failed to create HEAD file: %w", err)
	}

	emptyIndex := Index{Entries: []IndexEntry{}}
	indexData, err := json.Marshal(emptyIndex)
	if err != nil {
		return fmt.Errorf("failed to marshal empty index: %w", err)
	}

	if err := os.WriteFile(r.Path(FileIndex), indexData, 0644); err != nil {
		return fmt.Errorf("failed to write index file: %w", err)
	}

	return nil
}

// ResolveHead resolves the current commit hash and reference path (if on a branch).
func (r *Repository) ResolveHead() (commitHash string, refPath string, err error) {
	headPath := r.Path(FileHead)
	headData, err := os.ReadFile(headPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", ErrNotInitialized
		}
		return "", "", fmt.Errorf("failed to read HEAD: %w", err)
	}

	headContent := strings.TrimSpace(string(headData))

	if strings.HasPrefix(headContent, RefPrefix) {
		relRefPath := strings.TrimPrefix(headContent, RefPrefix)
		refPath = r.Path(relRefPath)
		refData, err := os.ReadFile(refPath)
		if err != nil {
			if os.IsNotExist(err) {
				return "", refPath, nil // Branch exists, but no commits yet
			}
			return "", "", fmt.Errorf("failed to read ref file: %w", err)
		}
		return strings.TrimSpace(string(refData)), refPath, nil
	}

	// Detached HEAD pointing directly to a hash
	return headContent, "", nil
}
