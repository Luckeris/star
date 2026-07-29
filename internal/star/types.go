package star

import (
	"time"
)

// IndexEntry represents a tracked file's metadata and content hash.
type IndexEntry struct {
	Path    string    `json:"path"`
	Hash    string    `json:"hash"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modtime"`
}

// Index represents the staging area containing tracked files.
type Index struct {
	Entries []IndexEntry `json:"entries"`
}

// Commit represents a point-in-time snapshot of repository state.
type Commit struct {
	Message   string       `json:"message"`
	Timestamp time.Time    `json:"timestamp"`
	Files     []IndexEntry `json:"files"`
	Parent    string       `json:"parent"`
}

// CommitWithHash combines a commit hash with its parsed commit data.
type CommitWithHash struct {
	Hash   string
	Commit Commit
}

// BranchInfo holds information about a branch name and whether it is currently checked out.
type BranchInfo struct {
	Name      string
	IsCurrent bool
}

// Config holds repository configuration options such as remote URL.
type Config struct {
	RemoteURL string `json:"remote_url"`
}
