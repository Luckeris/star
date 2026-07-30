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

func TestRemote(t *testing.T) {
	tempDir := t.TempDir()
	repo := star.NewRepository(tempDir)

	if err := repo.Init(); err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	// GetRemote before setting should return ErrNoRemote
	_, err := repo.GetRemote()
	if err != star.ErrNoRemote {
		t.Fatalf("expected ErrNoRemote, got: %v", err)
	}

	testURL := "https://github.com/example/star-repo.git"
	if err := repo.SetRemote(testURL); err != nil {
		t.Fatalf("failed to set remote: %v", err)
	}

	url, err := repo.GetRemote()
	if err != nil {
		t.Fatalf("failed to get remote: %v", err)
	}

	if url != testURL {
		t.Fatalf("expected remote %s, got %s", testURL, url)
	}
}

func TestUserConfig(t *testing.T) {
	tempDir := t.TempDir()
	repo := star.NewRepository(tempDir)

	if err := repo.Init(); err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	// GetUserConfig before setting should return error
	_, _, err := repo.GetUserConfig()
	if err == nil {
		t.Fatal("expected error when identity not configured, got nil")
	}

	name := "Jan Novak"
	email := "jan@example.com"
	if err := repo.SetUserConfig(name, email); err != nil {
		t.Fatalf("failed to set user config: %v", err)
	}

	gotName, gotEmail, err := repo.GetUserConfig()
	if err != nil {
		t.Fatalf("failed to get user config: %v", err)
	}

	if gotName != name || gotEmail != email {
		t.Fatalf("expected (%s, %s), got (%s, %s)", name, email, gotName, gotEmail)
	}
}
