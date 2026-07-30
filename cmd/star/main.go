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

	case "help", "--help", "-h":
		printUsage()

	case "version", "--version", "-v":
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
		fmt.Printf("Created commit %s\n", hash[:8])

	case "log":
		logs, err := repo.GetLog()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		for _, l := range logs {
			fmt.Printf("Commit: %s\n", l.Hash)
			if l.Commit.AuthorName != "" {
				fmt.Printf("Author: %s <%s>\n", l.Commit.AuthorName, l.Commit.AuthorEmail)
			}
			fmt.Printf("Timestamp: %s\n", l.Commit.Timestamp.Format("2006-01-02 15:04:05"))
			fmt.Printf("Message: %s\n", l.Commit.Message)
			fmt.Println("----------------------------------------")
		}

	case "status":
		details, err := repo.GetStatusDetails()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("On branch %s\n", details.Branch)

		if len(details.Staged) == 0 && len(details.Modified) == 0 && len(details.Untracked) == 0 {
			fmt.Println("Working tree clean (nothing to commit).")
			return
		}

		if len(details.Staged) > 0 {
			fmt.Println("\nChanges to be committed (staged):")
			for _, entry := range details.Staged {
				fmt.Printf("  added:    %s\n", entry.Path)
			}
		}

		if len(details.Modified) > 0 {
			fmt.Println("\nChanges not staged for commit (modified):")
			for _, path := range details.Modified {
				fmt.Printf("  modified: %s\n", path)
			}
		}

		if len(details.Untracked) > 0 {
			fmt.Println("\nUntracked files:")
			for _, path := range details.Untracked {
				fmt.Printf("  untracked: %s\n", path)
			}
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
			if targetURL == "add" {
				if len(os.Args) < 4 {
					fmt.Println("Usage: star remote add <url>")
					os.Exit(1)
				}
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

	case "push":
		fmt.Println("Pushing repository history to remote...")
		if err := repo.Push(); err != nil {
			fmt.Printf("Error during push: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ Successfully pushed to remote repository!")

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
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  init          Initialize a new star repository")
	fmt.Println("  add <path>    Stage a file or directory for commit")
	fmt.Println("  commit <msg>  Record staged changes into a new commit")
	fmt.Println("  log           Display commit history")
	fmt.Println("  status        Show tracked files and current branch")
	fmt.Println("  branch [name] List branches or create a new one")
	fmt.Println("  checkout <ref> Switch to a branch or commit")
	fmt.Println("  remote [url]  Show or set remote repository URL")
	fmt.Println("  push          Push commit history to remote repository")
	fmt.Println("  login         Configure author name and email")
	fmt.Println("  hash-object   Compute SHA-256 hash of a file")
	fmt.Println("  version       Display current Star version")
	fmt.Println("  help          Show this help message")
}
