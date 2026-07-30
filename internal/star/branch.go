package star

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	// ErrBranchAlreadyExists is returned when attempting to create a branch that already exists.
	ErrBranchAlreadyExists = errors.New("branch already exists")
	// ErrCommitNotFound is returned when a target commit or branch hash cannot be found.
	ErrCommitNotFound = errors.New("commit or branch not found")
)

// validateBranchName checks that a branch name does not contain path traversal characters.
func validateBranchName(name string) error {
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, " ") {
		return fmt.Errorf("invalid branch name '%s': must not contain '..', '/', '\\' or spaces", name)
	}
	return nil
}

// isValidHash checks if a string consists only of valid hex characters (SHA-256 hash or hash prefix).
func isValidHash(s string) bool {
	if len(s) < 4 || len(s) > 64 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// CreateBranch creates a new branch pointing to the current HEAD commit.
func (r *Repository) CreateBranch(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("branch name cannot be empty")
	}

	if err := validateBranchName(name); err != nil {
		return err
	}

	commitHash, _, err := r.ResolveHead()
	if err != nil {
		return err
	}

	if commitHash == "" {
		return errors.New("cannot create branch: no commits yet")
	}

	branchPath := r.Path(DirRefs, DirHeads, name)
	if _, err := os.Stat(branchPath); err == nil {
		return fmt.Errorf("%w: %s", ErrBranchAlreadyExists, name)
	}

	if err := os.WriteFile(branchPath, []byte(commitHash), 0644); err != nil {
		return fmt.Errorf("failed to create branch ref: %w", err)
	}

	return nil
}

// DeleteBranch removes a branch reference file from .star/refs/heads/.
func (r *Repository) DeleteBranch(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("branch name cannot be empty")
	}

	if err := validateBranchName(name); err != nil {
		return err
	}

	branches, err := r.ListBranches()
	if err == nil {
		for _, b := range branches {
			if b.Name == name && b.IsCurrent {
				return fmt.Errorf("cannot delete branch '%s' because it is currently checked out", name)
			}
		}
	}

	branchPath := r.Path(DirRefs, DirHeads, name)
	if _, err := os.Stat(branchPath); os.IsNotExist(err) {
		return fmt.Errorf("branch '%s' not found", name)
	}

	if err := os.Remove(branchPath); err != nil {
		return fmt.Errorf("failed to delete branch ref: %w", err)
	}

	return nil
}

// ListBranches returns all available branch names and indicates which branch is currently active.
func (r *Repository) ListBranches() ([]BranchInfo, error) {
	headPath := r.Path(FileHead)
	headData, err := os.ReadFile(headPath)
	currentBranch := ""
	if err == nil {
		headContent := strings.TrimSpace(string(headData))
		if strings.HasPrefix(headContent, RefPrefix+RefHeads) {
			currentBranch = strings.TrimPrefix(headContent, RefPrefix+RefHeads)
		}
	}

	headsDir := r.Path(DirRefs, DirHeads)
	entries, err := os.ReadDir(headsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotInitialized
		}
		return nil, fmt.Errorf("failed to read branches: %w", err)
	}

	var branches []BranchInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		branches = append(branches, BranchInfo{
			Name:      name,
			IsCurrent: name == currentBranch,
		})
	}

	return branches, nil
}

// Checkout switches to specified target (branch name or commit hash) and restores working files.
func (r *Repository) Checkout(target string) error {
	target = strings.TrimSpace(target)
	if target == "" {
		return errors.New("checkout target cannot be empty")
	}

	var commitHash string
	isBranch := false

	// Check if target is a valid branch name and if the branch ref file exists
	if err := validateBranchName(target); err == nil {
		branchPath := r.Path(DirRefs, DirHeads, target)
		if _, statErr := os.Stat(branchPath); statErr == nil {
			isBranch = true
			hashData, err := os.ReadFile(branchPath)
			if err != nil {
				return fmt.Errorf("failed to read branch file: %w", err)
			}
			commitHash = strings.TrimSpace(string(hashData))
		}
	}

	if !isBranch {
		if !isValidHash(target) {
			return fmt.Errorf("%w: %s", ErrCommitNotFound, target)
		}
		commitHash = target
	}

	commitPath := r.Path(DirCommits, commitHash+".json")
	commitFile, err := os.ReadFile(commitPath)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrCommitNotFound, target)
	}

	var commitData Commit
	if err := json.Unmarshal(commitFile, &commitData); err != nil {
		return fmt.Errorf("failed to parse commit data: %w", err)
	}

	// Remove files tracked in the current commit that do not exist in the target commit
	currentHeadHash, _, _ := r.ResolveHead()
	if currentHeadHash != "" && currentHeadHash != commitHash {
		currentCommitPath := r.Path(DirCommits, currentHeadHash+".json")
		if currentCommitFile, err := os.ReadFile(currentCommitPath); err == nil {
			var currentCommit Commit
			if err := json.Unmarshal(currentCommitFile, &currentCommit); err == nil {
				targetFiles := make(map[string]bool, len(commitData.Files))
				for _, f := range commitData.Files {
					targetFiles[f.Path] = true
				}
				for _, f := range currentCommit.Files {
					if !targetFiles[f.Path] {
						if IsStarOrGitPath(f.Path) {
							continue
						}
						staleFilePath := filepath.Clean(filepath.Join(r.RootPath, f.Path))
						rel, err := filepath.Rel(r.RootPath, staleFilePath)
						if err == nil && !strings.HasPrefix(rel, "..") && rel != ".." {
							_ = os.Remove(staleFilePath)
						}
					}
				}
			}
		}
	}

	// Restore files from commit object store into working directory
	for _, file := range commitData.Files {
		if IsStarOrGitPath(file.Path) {
			continue
		}
		targetFilePath := filepath.Clean(filepath.Join(r.RootPath, file.Path))
		rel, err := filepath.Rel(r.RootPath, targetFilePath)
		if err != nil || strings.HasPrefix(rel, "..") || rel == ".." {
			return fmt.Errorf("invalid path traversal attempt: %s", file.Path)
		}

		if err := os.MkdirAll(filepath.Dir(targetFilePath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for file %s: %w", file.Path, err)
		}

		objectPath := r.Path(DirObjects, file.Hash)
		objFile, err := os.Open(objectPath)
		if err != nil {
			return fmt.Errorf("failed to read object for file %s: %w", file.Path, err)
		}

		dstFile, err := os.Create(targetFilePath)
		if err != nil {
			objFile.Close()
			return fmt.Errorf("failed to restore file %s: %w", file.Path, err)
		}

		_, copyErr := io.Copy(dstFile, objFile)
		objFile.Close()
		dstFile.Close()

		if copyErr != nil {
			return fmt.Errorf("failed to restore file %s: %w", file.Path, copyErr)
		}
	}

	// Update index to match checked-out commit
	newIndex := Index{Entries: commitData.Files}
	if err := r.WriteIndex(&newIndex); err != nil {
		return fmt.Errorf("failed to update index during checkout: %w", err)
	}

	// Update HEAD
	headPath := r.Path(FileHead)
	if isBranch {
		if err := os.WriteFile(headPath, []byte(RefPrefix+RefHeads+target), 0644); err != nil {
			return fmt.Errorf("failed to update HEAD to branch: %w", err)
		}
	} else {
		if err := os.WriteFile(headPath, []byte(commitHash), 0644); err != nil {
			return fmt.Errorf("failed to update HEAD to hash: %w", err)
		}
	}

	return nil
}
