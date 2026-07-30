package star

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// HashObject calculates the SHA-256 hash of a file and stores it in .star/objects.
func (r *Repository) HashObject(filePath string) (string, error) {
	fileData, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	defer fileData.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, fileData); err != nil {
		return "", err
	}
	hashHex := hex.EncodeToString(hasher.Sum(nil))

	objectPath := r.Path(DirObjects, hashHex)

	// Skip writing if object with this hash already exists (content-addressable)
	if _, err := os.Stat(objectPath); err == nil {
		return hashHex, nil
	}

	if _, err := fileData.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to seek file: %w", err)
	}

	outFile, err := os.Create(objectPath)
	if err != nil {
		return "", fmt.Errorf("failed to create object file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, fileData); err != nil {
		return "", fmt.Errorf("failed to save object file: %w", err)
	}
	return hashHex, nil
}

// ReadIndex reads and parses the repository index.json file.
func (r *Repository) ReadIndex() (*Index, error) {
	indexPath := r.Path(FileIndex)
	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotInitialized
		}
		return nil, fmt.Errorf("failed to read index file: %w", err)
	}

	idx := &Index{}
	if err := json.Unmarshal(indexData, idx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal index file: %w", err)
	}

	return idx, nil
}

// WriteIndex saves the given Index struct into index.json.
func (r *Repository) WriteIndex(idx *Index) error {
	indexData, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	indexPath := r.Path(FileIndex)
	if err := os.WriteFile(indexPath, indexData, 0644); err != nil {
		return fmt.Errorf("failed to write index file: %w", err)
	}

	return nil
}

// Add stages a single file or recursively stages all files within a directory.
func (r *Repository) Add(targetPath string) error {
	idx, err := r.ReadIndex()
	if err != nil {
		return err
	}

	info, err := os.Stat(targetPath)
	if err != nil {
		return fmt.Errorf("failed to stat target path %s: %w", targetPath, err)
	}

	var pathsToAdd []string
	if info.IsDir() {
		err := filepath.WalkDir(targetPath, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			// Skip .git and .star directories
			if d.IsDir() {
				name := d.Name()
				if name == ".git" || name == DirStar {
					return filepath.SkipDir
				}
				return nil
			}
			pathsToAdd = append(pathsToAdd, p)
			return nil
		})
		if err != nil {
			return fmt.Errorf("error walking directory %s: %w", targetPath, err)
		}
	} else {
		pathsToAdd = append(pathsToAdd, targetPath)
	}
	// Build a lookup map for O(1) entry search instead of O(n) linear scan
	entryIndex := make(map[string]int, len(idx.Entries))
	for i, entry := range idx.Entries {
		entryIndex[entry.Path] = i
	}

	for _, p := range pathsToAdd {
		relPath := filepath.Clean(p)
		// Ignore path if it is within .star
		if strings.HasPrefix(relPath, DirStar) {
			continue
		}

		hashHex, err := r.HashObject(relPath)
		if err != nil {
			return err
		}

		fileInfo, err := os.Stat(relPath)
		if err != nil {
			return fmt.Errorf("failed to stat file %s: %w", relPath, err)
		}

		if i, exists := entryIndex[relPath]; exists {
			idx.Entries[i].Hash = hashHex
			idx.Entries[i].Size = fileInfo.Size()
			idx.Entries[i].ModTime = fileInfo.ModTime()
		} else {
			entryIndex[relPath] = len(idx.Entries)
			idx.Entries = append(idx.Entries, IndexEntry{
				Path:    relPath,
				Hash:    hashHex,
				Size:    fileInfo.Size(),
				ModTime: fileInfo.ModTime(),
			})
		}
	}

	return r.WriteIndex(idx)
}
