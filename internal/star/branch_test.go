package star_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Luckeris/star/internal/star"
)

func TestBranchAndCheckout(t *testing.T) {
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

	// Branching before commit should fail
	err = repo.CreateBranch("feature")
	if err == nil {
		t.Fatal("expected error creating branch with no commits, got nil")
	}

	// Create initial commit
	filePath := filepath.Join(tempDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("v1"), 0644); err != nil {
		t.Fatalf("failed to write file.txt: %v", err)
	}

	if err := repo.Add("file.txt"); err != nil {
		t.Fatalf("failed to add file.txt: %v", err)
	}

	_, err = repo.Commit("commit 1")
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Create branch
	if err := repo.CreateBranch("feature"); err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	branches, err := repo.ListBranches()
	if err != nil {
		t.Fatalf("failed to list branches: %v", err)
	}

	if len(branches) != 2 { // main and feature
		t.Fatalf("expected 2 branches, got %d", len(branches))
	}

	// Switch to feature branch
	if err := repo.Checkout("feature"); err != nil {
		t.Fatalf("failed to checkout feature: %v", err)
	}

	branches, _ = repo.ListBranches()
	for _, b := range branches {
		if b.Name == "feature" && !b.IsCurrent {
			t.Errorf("expected branch feature to be marked current")
		}
	}
}
