package star

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// Commit creates a new commit object with the staged index entries and updates HEAD/ref.
func (r *Repository) Commit(message string) (string, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return "", ErrEmptyCommitMessage
	}

	idx, err := r.ReadIndex()
	if err != nil {
		return "", err
	}

	if len(idx.Entries) == 0 {
		return "", ErrNothingToCommit
	}

	parentHash, refPath, err := r.ResolveHead()
	if err != nil {
		return "", err
	}

	cfg, _ := r.ReadConfig()

	newCommit := Commit{
		Message:     message,
		AuthorName:  cfg.UserName,
		AuthorEmail: cfg.UserEmail,
		Timestamp:   time.Now().UTC(),
		Files:       idx.Entries,
		Parent:      parentHash,
	}

	commitData, err := json.MarshalIndent(newCommit, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal commit data: %w", err)
	}

	hash := sha256.Sum256(commitData)
	hashHex := hex.EncodeToString(hash[:])
	commitPath := r.Path(DirCommits, hashHex+".json")

	if err := os.WriteFile(commitPath, commitData, 0644); err != nil {
		return "", fmt.Errorf("failed to write commit file: %w", err)
	}

	// Update reference or detached HEAD
	if refPath != "" {
		if err := os.WriteFile(refPath, []byte(hashHex), 0644); err != nil {
			return "", fmt.Errorf("failed to update branch reference: %w", err)
		}
	} else {
		if err := os.WriteFile(r.Path(FileHead), []byte(hashHex), 0644); err != nil {
			return "", fmt.Errorf("failed to update HEAD file: %w", err)
		}
	}

	// Clear index after successful commit
	emptyIndex := Index{Entries: []IndexEntry{}}
	if err := r.WriteIndex(&emptyIndex); err != nil {
		return "", fmt.Errorf("failed to clear index after commit: %w", err)
	}

	return hashHex, nil
}

// GetLog traverses the commit history starting from HEAD and returns all commits.
func (r *Repository) GetLog() ([]CommitWithHash, error) {
	currentHash, _, err := r.ResolveHead()
	if err != nil {
		return nil, err
	}

	if currentHash == "" {
		return nil, ErrNoCommits
	}

	visited := make(map[string]bool)
	var logs []CommitWithHash
	for currentHash != "" {
		// Cycle detection: stop if we have already seen this commit hash
		if visited[currentHash] {
			break
		}
		visited[currentHash] = true

		commitPath := r.Path(DirCommits, currentHash+".json")
		commitFile, err := os.ReadFile(commitPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read commit object for hash %s: %w", currentHash, err)
		}

		var commitData Commit
		if err := json.Unmarshal(commitFile, &commitData); err != nil {
			return nil, fmt.Errorf("failed to parse commit data for hash %s: %w", currentHash, err)
		}

		logs = append(logs, CommitWithHash{
			Hash:   currentHash,
			Commit: commitData,
		})

		currentHash = commitData.Parent
	}

	return logs, nil
}

// Status returns the list of currently tracked files in the index.
func (r *Repository) Status() ([]string, error) {
	idx, err := r.ReadIndex()
	if err != nil {
		return nil, err
	}

	var tracked []string
	for _, entry := range idx.Entries {
		tracked = append(tracked, entry.Path)
	}

	return tracked, nil
}
