package star_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Luckeris/star/internal/star"
)

func TestCommitAndLog(t *testing.T) {
	tempDir := t.TempDir()
	repo := star.NewRepository(tempDir)

	if err := repo.Init(); err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working dir: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	// Commit on empty index should fail
	_, err = repo.Commit("Empty commit")
	if err != star.ErrNothingToCommit {
		t.Errorf("expected ErrNothingToCommit, got: %v", err)
	}

	// Create and add file
	filePath := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(filePath, []byte("version 1"), 0644); err != nil {
		t.Fatalf("failed to write test.txt: %v", err)
	}

	if _, err := repo.Add("test.txt"); err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	// Check status
	tracked, err := repo.Status()
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}
	if len(tracked) != 1 || tracked[0] != "test.txt" {
		t.Errorf("unexpected tracked files: %v", tracked)
	}

	// Commit
	commitHash, err := repo.Commit("Initial commit")
	if err != nil {
		t.Fatalf("failed to create commit: %v", err)
	}

	if commitHash == "" {
		t.Fatal("expected non-empty commit hash")
	}

	// Verify index cleared after commit
	idx, err := repo.ReadIndex()
	if err != nil {
		t.Fatalf("failed to read index: %v", err)
	}
	if len(idx.Entries) != 0 {
		t.Errorf("expected empty index after commit, got %d entries", len(idx.Entries))
	}

	// Check log
	logs, err := repo.GetLog()
	if err != nil {
		t.Fatalf("failed to get logs: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(logs))
	}
	if logs[0].Commit.Message != "Initial commit" {
		t.Errorf("expected commit message 'Initial commit', got '%s'", logs[0].Commit.Message)
	}
}
