// Package star provides core functionality for the Star version control system.
package star

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	DirStar        = ".star"
	DirObjects     = "objects"
	DirCommits     = "commits"
	DirRefs        = "refs"
	DirHeads       = "heads"
	FileIndex      = "index.json"
	FileHead       = "HEAD"
	FileConfig     = "config.json"
	FileStarIgnore = ".starignore"
	RefPrefix      = "ref: "
	RefHeads       = "refs/heads/"
)

var (
	// ErrAlreadyInitialized is returned when attempting to initialize an existing repository.
	ErrAlreadyInitialized = errors.New(".star is already initialized")
	// ErrNotInitialized is returned when executing commands in a non-repository folder.
	ErrNotInitialized = errors.New("repository not initialized (run 'star init' first)")
	// ErrEmptyCommitMessage is returned when a commit message is blank.
	ErrEmptyCommitMessage = errors.New("commit message cannot be empty")
	// ErrNothingToCommit is returned when committing an empty index.
	ErrNothingToCommit = errors.New("nothing to commit (index is empty)")
	// ErrNoCommits is returned when accessing history in a repository without commits.
	ErrNoCommits = errors.New("no commits found")
	// ErrNoRemote is returned when remote URL is not configured.
	ErrNoRemote = errors.New("no remote URL configured (run 'star remote add <url>')")
	// ErrEverythingUpToDate is returned when pushing to a remote that is already up to date.
	ErrEverythingUpToDate = errors.New("everything up-to-date")
)

// Repository encapsulates the file system operations for a Star VCS repository.
type Repository struct {
	RootPath string
}

// IsStarOrGitPath returns true if the path is inside .star or .git directories.
func IsStarOrGitPath(path string) bool {
	clean := filepath.ToSlash(filepath.Clean(path))
	return clean == ".star" || strings.HasPrefix(clean, ".star/") || clean == ".git" || strings.HasPrefix(clean, ".git/")
}

// FindRepositoryRoot walks up parent directories starting from startPath until it finds a .star directory.
func FindRepositoryRoot(startPath string) (string, error) {
	if startPath == "" || startPath == "." {
		cwd, err := os.Getwd()
		if err == nil {
			startPath = cwd
		}
	}
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}

	dir := absPath
	for {
		starDir := filepath.Join(dir, DirStar)
		if info, err := os.Stat(starDir); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", ErrNotInitialized
}

// NewRepository creates a new Repository instance targeting rootPath (automatically discovering root if "." is passed).
func NewRepository(rootPath string) *Repository {
	if rootPath == "" || rootPath == "." {
		if foundRoot, err := FindRepositoryRoot("."); err == nil {
			rootPath = foundRoot
		} else {
			rootPath = "."
		}
	}
	return &Repository{RootPath: rootPath}
}

// Path returns a path joined relative to the .star directory within the repository root.
func (r *Repository) Path(elem ...string) string {
	parts := make([]string, 0, 2+len(elem))
	parts = append(parts, r.RootPath, DirStar)
	parts = append(parts, elem...)
	return filepath.Join(parts...)
}

// Init initializes the .star directory structure and default branch reference.
func (r *Repository) Init() error {
	starDir := r.Path()
	if _, err := os.Stat(starDir); err == nil {
		return ErrAlreadyInitialized
	}

	dirs := []string{
		r.Path(DirObjects),
		r.Path(DirCommits),
		r.Path(DirRefs),
		r.Path(DirRefs, DirHeads),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	headPath := r.Path(FileHead)
	if err := os.WriteFile(headPath, []byte(RefPrefix+RefHeads+"main"), 0644); err != nil {
		return fmt.Errorf("failed to create HEAD file: %w", err)
	}

	emptyIndex := Index{Entries: []IndexEntry{}}
	indexData, err := json.Marshal(emptyIndex)
	if err != nil {
		return fmt.Errorf("failed to marshal empty index: %w", err)
	}

	if err := os.WriteFile(r.Path(FileIndex), indexData, 0644); err != nil {
		return fmt.Errorf("failed to write index file: %w", err)
	}

	// Auto-create .gitignore with .star/ if not present
	gitignorePath := filepath.Join(r.RootPath, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		_ = os.WriteFile(gitignorePath, []byte(".star/\n"), 0644)
	}

	return nil
}

// ResolveHead resolves the current commit hash and reference path (if on a branch).
func (r *Repository) ResolveHead() (commitHash string, refPath string, err error) {
	headPath := r.Path(FileHead)
	headData, err := os.ReadFile(headPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", ErrNotInitialized
		}
		return "", "", fmt.Errorf("failed to read HEAD: %w", err)
	}

	headContent := strings.TrimSpace(string(headData))

	if strings.HasPrefix(headContent, RefPrefix) {
		relRefPath := strings.TrimPrefix(headContent, RefPrefix)
		refPath = r.Path(relRefPath)
		refData, err := os.ReadFile(refPath)
		if err != nil {
			if os.IsNotExist(err) {
				return "", refPath, nil // Branch exists, but no commits yet
			}
			return "", "", fmt.Errorf("failed to read ref file: %w", err)
		}
		return strings.TrimSpace(string(refData)), refPath, nil
	}

	// Detached HEAD pointing directly to a hash
	return headContent, "", nil
}

// ReadConfig reads and parses .star/config.json. If it does not exist, an empty Config struct is returned.
func (r *Repository) ReadConfig() (*Config, error) {
	configPath := r.Path(FileConfig)
	cfgData, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(cfgData, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// WriteConfig saves the Config struct into .star/config.json.
func (r *Repository) WriteConfig(cfg *Config) error {
	cfgData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := r.Path(FileConfig)
	if err := os.WriteFile(configPath, cfgData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SetRemote saves the remote repository URL to .star/config.json.
func (r *Repository) SetRemote(remoteURL string) error {
	remoteURL = strings.TrimSpace(remoteURL)
	if remoteURL == "" {
		return errors.New("remote URL cannot be empty")
	}

	cfg, err := r.ReadConfig()
	if err != nil {
		return err
	}

	cfg.RemoteURL = remoteURL
	return r.WriteConfig(cfg)
}

// GetRemote reads the configured remote repository URL from .star/config.json.
func (r *Repository) GetRemote() (string, error) {
	cfg, err := r.ReadConfig()
	if err != nil {
		return "", err
	}

	if cfg.RemoteURL == "" {
		return "", ErrNoRemote
	}

	return cfg.RemoteURL, nil
}

// SetUserConfig sets the user name and email in .star/config.json.
func (r *Repository) SetUserConfig(name, email string) error {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	if name == "" || email == "" {
		return errors.New("user name and email cannot be empty")
	}

	cfg, err := r.ReadConfig()
	if err != nil {
		return err
	}

	cfg.UserName = name
	cfg.UserEmail = email
	return r.WriteConfig(cfg)
}

// GetUserConfig reads the configured user name and email from .star/config.json.
func (r *Repository) GetUserConfig() (name string, email string, err error) {
	cfg, err := r.ReadConfig()
	if err != nil {
		return "", "", err
	}

	if cfg.UserName == "" || cfg.UserEmail == "" {
		return "", "", errors.New("user identity not configured (run 'star login' or 'star config')")
	}

	return cfg.UserName, cfg.UserEmail, nil
}
