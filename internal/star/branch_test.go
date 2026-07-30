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

	if _, err := repo.Add("file.txt"); err != nil {
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

	// Cannot delete current active branch
	err = repo.DeleteBranch("feature")
	if err == nil {
		t.Fatal("expected error deleting current branch, got nil")
	}

	// Switch back to main branch
	if err := repo.Checkout("main"); err != nil {
		t.Fatalf("failed to checkout main: %v", err)
	}

	// Delete feature branch
	if err := repo.DeleteBranch("feature"); err != nil {
		t.Fatalf("failed to delete feature branch: %v", err)
	}

	branches, _ = repo.ListBranches()
	if len(branches) != 1 {
		t.Fatalf("expected 1 branch after deletion, got %d", len(branches))
	}
}

func TestCheckoutStaleFileCleanup(t *testing.T) {
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

	// Commit 1 on main branch: contains file1.txt
	file1Path := filepath.Join(tempDir, "file1.txt")
	if err := os.WriteFile(file1Path, []byte("file1 content"), 0644); err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}
	if _, err := repo.Add("file1.txt"); err != nil {
		t.Fatalf("failed to add file1: %v", err)
	}
	commit1Hash, err := repo.Commit("commit 1 on main")
	if err != nil {
		t.Fatalf("failed to create commit 1: %v", err)
	}

	// Create feature branch pointing to commit 1
	if err := repo.CreateBranch("feature"); err != nil {
		t.Fatalf("failed to create feature branch: %v", err)
	}

	// Commit 2 on main branch: adds file2.txt
	file2Path := filepath.Join(tempDir, "file2.txt")
	if err := os.WriteFile(file2Path, []byte("file2 content"), 0644); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}
	if _, err := repo.Add("file2.txt"); err != nil {
		t.Fatalf("failed to add file2: %v", err)
	}
	_, err = repo.Commit("commit 2 on main")
	if err != nil {
		t.Fatalf("failed to create commit 2: %v", err)
	}

	// Both file1.txt and file2.txt exist on main branch
	if _, err := os.Stat(file2Path); os.IsNotExist(err) {
		t.Fatalf("file2.txt should exist on main branch before checkout")
	}

	// Switch back to feature branch (which only has commit 1 / file1.txt)
	if err := repo.Checkout("feature"); err != nil {
		t.Fatalf("failed to checkout feature branch: %v", err)
	}

	// Verify file2.txt was cleaned up automatically
	if _, err := os.Stat(file2Path); !os.IsNotExist(err) {
		t.Errorf("file2.txt should have been cleaned up after checkout to feature branch")
	}

	_ = commit1Hash
}
