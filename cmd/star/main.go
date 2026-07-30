// Command star is a simplified version control system binary entrypoint.
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Luckeris/star/internal/star"
)

const version = "star v0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	repo := star.NewRepository(".")
	command := os.Args[1]

	switch command {
	case "init":
		if err := repo.Init(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Initialized empty star repository in .star directory")

	case "help":
		printUsage()

	case "version":
		fmt.Println(version)

	case "hash-object":
		if len(os.Args) < 3 {
			fmt.Println("Usage: star hash-object <file>")
			os.Exit(1)
		}
		hash, err := repo.HashObject(os.Args[2])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(hash)

	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Usage: star add <file|directory>")
			os.Exit(1)
		}
		target := os.Args[2]
		if err := repo.Add(target); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Added %s to index\n", target)

	case "commit":
		if len(os.Args) < 3 {
			fmt.Println("Usage: star commit <message>")
			os.Exit(1)
		}
		hash, err := repo.Commit(os.Args[2])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created commit %s\n", hash)

	case "log":
		logs, err := repo.GetLog()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		for _, l := range logs {
			fmt.Printf("Commit: %s\n", l.Hash)
			fmt.Printf("Timestamp: %s\n", l.Commit.Timestamp)
			fmt.Printf("Message: %s\n", l.Commit.Message)
			fmt.Println("----------------------------------------")
		}

	case "status":
		tracked, err := repo.Status()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if len(tracked) == 0 {
			fmt.Println("No files are currently tracked (index is empty).")
			return
		}
		fmt.Println("Tracked files:")
		for _, path := range tracked {
			fmt.Printf("  added: %s\n", path)
		}

	case "checkout":
		if len(os.Args) < 3 {
			fmt.Println("Usage: star checkout <branch_or_commit-hash>")
			os.Exit(1)
		}
		target := os.Args[2]
		if err := repo.Checkout(target); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Switched to '%s'\n", target)

	case "branch":
		if len(os.Args) == 2 {
			branches, err := repo.ListBranches()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			for _, b := range branches {
				if b.IsCurrent {
					fmt.Printf("* %s\n", b.Name)
				} else {
					fmt.Printf("  %s\n", b.Name)
				}
			}
		} else {
			branchName := os.Args[2]
			if err := repo.CreateBranch(branchName); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Created branch '%s'\n", branchName)
		}

	case "remote":
		// Handle reading or setting remote URL ("star remote", "star remote <url>", or "star remote add <url>")
		if len(os.Args) == 2 {
			url, err := repo.GetRemote()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Remote URL: %s\n", url)
		} else {
			targetURL := os.Args[2]
			if targetURL == "add" && len(os.Args) >= 4 {
				targetURL = os.Args[3]
			}
			if err := repo.SetRemote(targetURL); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Set remote URL to '%s'\n", targetURL)
		}

	case "login", "config":
		handleLogin(repo)

	default:
		fmt.Println("Unknown command:", command)
		printUsage()
		os.Exit(1)
	}
}

func handleLogin(repo *star.Repository) {
	var name, email string

	if len(os.Args) >= 4 {
		name = os.Args[2]
		email = os.Args[3]
	} else {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter your author name (e.g., John Doe): ")
		inputName, _ := reader.ReadString('\n')
		name = strings.TrimSpace(inputName)

		fmt.Print("Enter your author email (e.g., john@example.com): ")
		inputEmail, _ := reader.ReadString('\n')
		email = strings.TrimSpace(inputEmail)
	}

	if name == "" || email == "" {
		fmt.Println("Error: Name and email cannot be empty.")
		os.Exit(1)
	}

	// 1. Save identity in Star config (.star/config.json)
	if err := repo.SetUserConfig(name, email); err != nil {
		fmt.Printf("Error saving Star configuration: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Saved author identity to Star: %s <%s>\n", name, email)

	// 2. Pass identity to underlying system Git
	gitCmdName := exec.Command("git", "config", "user.name", name)
	if err := gitCmdName.Run(); err != nil {
		fmt.Printf(" Warning: Failed to set Git user.name (%v). Is Git installed?\n", err)
	} else {
		fmt.Println("✓ Synced user.name to system Git")
	}

	gitCmdEmail := exec.Command("git", "config", "user.email", email)
	if err := gitCmdEmail.Run(); err != nil {
		fmt.Printf(" Warning: Failed to set Git user.email (%v). Is Git installed?\n", err)
	} else {
		fmt.Println("✓ Synced user.email to system Git")
	}

	fmt.Println("\nSuccess! Your identity is configured for commits and push.")
}

func printUsage() {
	fmt.Println("Usage: star <command> [arguments]")
	fmt.Println("Available commands: help, version, init, hash-object, add, commit, log, status, checkout, branch, remote, login")
}
