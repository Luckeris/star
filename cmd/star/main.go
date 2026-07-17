package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// IndexEntry represents a single file tracked in the repository
type IndexEntry struct {
	Path    string    `json:"path"`
	Hash    string    `json:"hash"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modtime"`
}

// Index represents the entire index.json structure
type Index struct {
	Entries []IndexEntry `json:"entries"`
}

func main() {
	// Ensure the user provides at least one command
	if len(os.Args) < 2 {
		fmt.Println("Usage: star <command>")
		fmt.Println("Available commands: help, version, init, hash-object, add")
		return
	}

	command := os.Args[1]

	switch command {
	case "init":
		// 1. Create the base .star directory
		err := os.Mkdir(".star", 0755)
		if err != nil {
			if os.IsExist(err) {
				fmt.Println(".star is already initialized")
				return
			}
			fmt.Println("Error creating .star directory:", err)
			return
		}

		// 2. Create objects and commits subdirectories
		slozky := []string{".star/objects", ".star/commits"}
		for _, slozka := range slozky {
			err := os.Mkdir(slozka, 0755)
			if err != nil {
				fmt.Println("Error creating directory:", err)
				return
			}
		}

		// 3. Create an empty HEAD file
		error := os.WriteFile(".star/HEAD", []byte(""), 0644)
		if error != nil {
			fmt.Println("Error creating HEAD file:", error)
			return
		}

		// 4. Initialize an empty index.json
		mujIndex := Index{Entries: []IndexEntry{}}
		data, err := json.Marshal(mujIndex)
		if err != nil {
			fmt.Println("Error creating index file:", err)
			return
		}

		err = os.WriteFile(".star/index.json", data, 0644)
		if err != nil {
			fmt.Println("Error creating index file:", err)
			return
		}

		fmt.Println("Initialized empty star repository in .star directory")

	case "help":
		fmt.Println("Available commands: help, version, init, hash-object, add")

	case "version":
		fmt.Println("star v0.1.0")

	case "hash-object":
		// Validate argument count
		if len(os.Args) < 3 {
			fmt.Println("Usage: star hash-object <file>")
			return
		}
		path := os.Args[2]

		// 1. Read the target file
		file, err := os.ReadFile(path)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}

		// 2. Calculate SHA-256 hash
		hash := sha256.Sum256(file)
		hexString := hex.EncodeToString(hash[:])
		fmt.Println(hexString)

		// 3. Write the file content to .star/objects/<hash>
		objectPath := filepath.Join(".star", "objects", hexString)
		err = os.WriteFile(objectPath, file, 0644)
		if err != nil {
			fmt.Println("Error creating object file:", err)
			return
		}

	case "add":
		// Validate argument count
		if len(os.Args) < 3 {
			fmt.Println("Usage: star add <file>")
			return
		}

		// 1. Read and parse the existing index.json
		indexData, err := os.ReadFile(".star/index.json")
		if err != nil {
			fmt.Println("Error reading index file:", err)
			return
		}

		mujIndex := Index{}
		err = json.Unmarshal(indexData, &mujIndex)
		if err != nil {
			fmt.Println("Error unmarshaling index file:", err)
			return
		}

		// 2. Read the file being added
		path := os.Args[2]
		fileData, err := os.ReadFile(path)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}

		// 3. Calculate hash and save the object (blob)
		hash := sha256.Sum256(fileData)
		hexString := hex.EncodeToString(hash[:])
		objectPath := filepath.Join(".star", "objects", hexString)

		err = os.WriteFile(objectPath, fileData, 0644)
		if err != nil {
			fmt.Println("Error creating object file:", err)
			return
		}

		// 4. Get file metadata (size, mod time)
		info, err := os.Stat(path)
		if err != nil {
			fmt.Println("Error getting file info:", err)
			return
		}

		// 5. Append the new entry to the index
		novyZaznam := IndexEntry{
			Path:    path,
			Hash:    hexString,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}
		mujIndex.Entries = append(mujIndex.Entries, novyZaznam)

		// 6. Save the updated index back to disk
		updatedIndexData, err := json.Marshal(mujIndex)
		if err != nil {
			fmt.Println("Error marshaling updated index:", err)
			return
		}

		err = os.WriteFile(".star/index.json", updatedIndexData, 0644)
		if err != nil {
			fmt.Println("Error writing updated index file:", err)
			return
		}

		fmt.Printf("Added %s to index\n", path)

	default:
		fmt.Println("Unknown command:", command)
	}
}
