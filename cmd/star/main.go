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

type IndexEntry struct {
	Path    string    `json:"path"`
	Hash    string    `json:"hash"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modtime"`
}

type Index struct {
	Entries []IndexEntry `json:"entries"`
}

type Commit struct {
	Message   string       `json:"message"`
	Timestamp time.Time    `json:"timestamp"`
	Files     []IndexEntry `json:"files"`
	Parent    string       `json:"parent"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "init":
		handleInit()
	case "help":
		printUsage()
	case "version":
		fmt.Println("star v0.1.0")
	case "hash-object":
		if len(os.Args) < 3 {
			fmt.Println("Usage: star hash-object <file>")
			return
		}
		handleHashObject(os.Args[2])
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Usage: star add <file>")
			return
		}
		handleAdd(os.Args[2])
	case "commit":
		if len(os.Args) < 3 {
			fmt.Println("Usage: star commit <message>")
			return
		}
		handleCommit(os.Args[2])
	case "log":
		handleLog()
	case "status":
		handleStatus()
	default:
		fmt.Println("Unknown command:", command)
	}
}

func printUsage() {
	fmt.Println("Usage: star <command>")
	fmt.Println("Available commands: help, version, init, hash-object, add, commit, log, status")
}

func handleInit() {
	// init base directory
	err := os.Mkdir(".star", 0755)
	if err != nil {
		if os.IsExist(err) {
			fmt.Println(".star is already initialized")
			return
		}
		fmt.Println("Error creating .star directory:", err)
		return
	}

	// init subdirectories
	slozky := []string{".star/objects", ".star/commits"}
	for _, slozka := range slozky {
		err := os.Mkdir(slozka, 0755)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}
	}

	// create empty HEAD
	error := os.WriteFile(".star/HEAD", []byte(""), 0644)
	if error != nil {
		fmt.Println("Error creating HEAD file:", error)
		return
	}

	// init empty index
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
}

func handleHashObject(path string) {
	// read target file
	file, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// calc hash
	hash := sha256.Sum256(file)
	hexString := hex.EncodeToString(hash[:])
	fmt.Println(hexString)

	// save object
	objectPath := filepath.Join(".star", "objects", hexString)
	err = os.WriteFile(objectPath, file, 0644)
	if err != nil {
		fmt.Println("Error creating object file:", err)
		return
	}
}

func handleAdd(path string) {
	// read existing index
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

	// read target file
	fileData, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// save blob object
	hash := sha256.Sum256(fileData)
	hexString := hex.EncodeToString(hash[:])
	objectPath := filepath.Join(".star", "objects", hexString)

	err = os.WriteFile(objectPath, fileData, 0644)
	if err != nil {
		fmt.Println("Error creating object file:", err)
		return
	}

	// get file metadata
	info, err := os.Stat(path)
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return
	}

	// check for duplicates
	nalezeno := false
	for i, entry := range mujIndex.Entries {
		if entry.Path == path {
			mujIndex.Entries[i].Hash = hexString
			mujIndex.Entries[i].Size = info.Size()
			mujIndex.Entries[i].ModTime = info.ModTime()
			nalezeno = true
			break
		}
	}

	// append new entry
	if !nalezeno {
		novyZaznam := IndexEntry{
			Path:    path,
			Hash:    hexString,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}
		mujIndex.Entries = append(mujIndex.Entries, novyZaznam)
	}

	// save updated index
	updatedIndexData, err := json.MarshalIndent(mujIndex, "", "  ")
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
}

func handleCommit(zprava string) {
	cas := time.Now()

	// read index
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

	if len(mujIndex.Entries) == 0 {
		fmt.Println("Nothing to commit (index is empty).")
		return
	}

	// get parent commit
	nactenaData, err := os.ReadFile(".star/HEAD")
	if err != nil {
		fmt.Println("Error reading HEAD file:", err)
		return
	}
	rodic := string(nactenaData)

	// create commit struct
	novyCommit := Commit{
		Message:   zprava,
		Timestamp: cas,
		Files:     mujIndex.Entries,
		Parent:    rodic,
	}

	// save commit object
	commitData, err := json.MarshalIndent(novyCommit, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling commit:", err)
		return
	}

	hash := sha256.Sum256(commitData)
	hexString := hex.EncodeToString(hash[:])
	commitPath := filepath.Join(".star", "commits", hexString+".json")

	err = os.WriteFile(commitPath, commitData, 0644)
	if err != nil {
		fmt.Println("Error creating commit file:", err)
		return
	}

	// update HEAD
	err = os.WriteFile(".star/HEAD", []byte(hexString), 0644)
	if err != nil {
		fmt.Println("Error updating HEAD file:", err)
		return
	}

	// clear index after commit
	prazdnyIndex := Index{Entries: []IndexEntry{}}
	vycistenyData, err := json.Marshal(prazdnyIndex)
	if err == nil {
		os.WriteFile(".star/index.json", vycistenyData, 0644)
	}

	fmt.Printf("Created commit %s\n", hexString)
}

func handleLog() {
	// get latest commit
	headData, err := os.ReadFile(".star/HEAD")
	if err != nil {
		fmt.Println("Error reading HEAD file:", err)
		return
	}

	commitHash := string(headData)

	if commitHash == "" {
		fmt.Println("No commits found.")
		return
	}

	// traverse history
	for {
		if commitHash == "" {
			break
		}
		commitPath := filepath.Join(".star", "commits", commitHash+".json")

		commitFile, err := os.ReadFile(commitPath)
		if err != nil {
			fmt.Printf("Error reading commit file for hash %s: %v\n", commitHash, err)
			return
		}

		commitData := Commit{}
		err = json.Unmarshal(commitFile, &commitData)
		if err != nil {
			fmt.Printf("Error unmarshaling commit data for hash %s: %v\n", commitHash, err)
			return
		}

		fmt.Printf("Commit: %s\n", commitHash)
		fmt.Printf("Timestamp: %s\n", commitData.Timestamp)
		fmt.Printf("Message: %s\n", commitData.Message)
		fmt.Println("----------------------------------------")

		commitHash = commitData.Parent
	}
}

func handleStatus() {
	// read index
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

	if len(mujIndex.Entries) == 0 {
		fmt.Println("No files are currently tracked (index is empty).")
		return
	}

	// print tracked files
	fmt.Printf("Tracked files:\n")
	for _, entry := range mujIndex.Entries {
		fmt.Printf("  added: %s\n", entry.Path)
	}
}
