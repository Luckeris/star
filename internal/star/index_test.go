package star_test

import (
	"os"
	"path/filepath"
	"strings"
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

func TestStarIgnore(t *testing.T) {
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

	// Create .starignore
	ignoreContent := "# Ignore logs and temp directory\n*.log\ntemp/\n"
	if err := os.WriteFile(filepath.Join(tempDir, ".starignore"), []byte(ignoreContent), 0644); err != nil {
		t.Fatalf("failed to write .starignore: %v", err)
	}

	// Create normal file
	if err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}

	// Create ignored log file
	if err := os.WriteFile(filepath.Join(tempDir, "app.log"), []byte("log data"), 0644); err != nil {
		t.Fatalf("failed to write app.log: %v", err)
	}

	// Create ignored directory and file inside it
	tempFolder := filepath.Join(tempDir, "temp")
	if err := os.MkdirAll(tempFolder, 0755); err != nil {
		t.Fatalf("failed to mkdir temp: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempFolder, "cache.txt"), []byte("cache"), 0644); err != nil {
		t.Fatalf("failed to write cache.txt: %v", err)
	}

	// Add all
	if err := repo.Add("."); err != nil {
		t.Fatalf("failed to add .: %v", err)
	}

	idx, err := repo.ReadIndex()
	if err != nil {
		t.Fatalf("failed to read index: %v", err)
	}

	// Should contain main.go and .starignore, but NOT app.log or temp/cache.txt
	for _, entry := range idx.Entries {
		if entry.Path == "app.log" || strings.HasPrefix(entry.Path, "temp") {
			t.Errorf("expected path %s to be ignored by .starignore", entry.Path)
		}
	}
}
