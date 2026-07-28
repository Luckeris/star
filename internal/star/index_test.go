package star_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Luckeris/star/internal/star"
)

func TestAddSingleFileAndDirectory(t *testing.T) {
	tempDir := t.TempDir()
	repo := star.NewRepository(tempDir)

	if err := repo.Init(); err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	// Create test files
	file1Path := filepath.Join(tempDir, "file1.txt")
	if err := os.WriteFile(file1Path, []byte("hello star"), 0644); err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}

	subDir := filepath.Join(tempDir, "docs")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subDir: %v", err)
	}

	file2Path := filepath.Join(subDir, "doc.txt")
	if err := os.WriteFile(file2Path, []byte("documentation"), 0644); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working dir: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to chdir to tempDir: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	if err := repo.Add("file1.txt"); err != nil {
		t.Fatalf("failed to add file1.txt: %v", err)
	}

	idx, err := repo.ReadIndex()
	if err != nil {
		t.Fatalf("failed to read index: %v", err)
	}

	if len(idx.Entries) != 1 {
		t.Fatalf("expected 1 index entry, got %d", len(idx.Entries))
	}
	if idx.Entries[0].Path != "file1.txt" {
		t.Errorf("expected path file1.txt, got %s", idx.Entries[0].Path)
	}

	// Add directory recursively (add .)
	if err := repo.Add("."); err != nil {
		t.Fatalf("failed to add directory: %v", err)
	}

	idx, err = repo.ReadIndex()
	if err != nil {
		t.Fatalf("failed to read index: %v", err)
	}

	if len(idx.Entries) != 2 {
		t.Fatalf("expected 2 index entries after adding '.', got %d", len(idx.Entries))
	}
}
