package star_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Luckeris/star/internal/star"
)

func TestInit(t *testing.T) {
	tempDir := t.TempDir()
	repo := star.NewRepository(tempDir)

	err := repo.Init()
	if err != nil {
		t.Fatalf("expected no error on init, got: %v", err)
	}

	// Verify .star directories exist
	expectedDirs := []string{
		filepath.Join(tempDir, ".star"),
		filepath.Join(tempDir, ".star", "objects"),
		filepath.Join(tempDir, ".star", "commits"),
		filepath.Join(tempDir, ".star", "refs", "heads"),
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("expected directory %s to exist", dir)
		}
	}

	// Re-init should return ErrAlreadyInitialized
	err = repo.Init()
	if err != star.ErrAlreadyInitialized {
		t.Errorf("expected ErrAlreadyInitialized, got: %v", err)
	}
}
