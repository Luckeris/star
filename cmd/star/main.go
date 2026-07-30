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
			handleCLIError(err)
		}
		fmt.Println("Initialized empty Star repository in .star/")
		fmt.Println(yellow("hint: run 'star login \"Your Name\" \"your.email@example.com\"' to set up your identity"))

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
			handleCLIError(err)
		}
		fmt.Printf("staged '%s' for commit\n", target)

	case "commit":
		if len(os.Args) < 3 {
			fmt.Println("Usage: star commit <message>")
			os.Exit(1)
		}
		hash, err := repo.Commit(os.Args[2])
		if err != nil {
			handleCLIError(err)
		}

		branchName := "main"
		if branches, err := repo.ListBranches(); err == nil {
			for _, b := range branches {
				if b.IsCurrent {
					branchName = b.Name
					break
				}
			}
		}
		fmt.Printf("[%s %s] %s\n", cyan(branchName), yellow(hash[:8]), os.Args[2])

	case "log":
		logs, err := repo.GetLog()
		if err != nil {
			handleCLIError(err)
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
			handleCLIError(err)
		}
		fmt.Printf("On branch %s\n", cyan(details.Branch))

		if len(details.Staged) == 0 && len(details.Modified) == 0 && len(details.Untracked) == 0 {
			fmt.Println("nothing to commit, working tree clean")
			return
		}

		if len(details.Staged) > 0 {
			fmt.Println("\n" + bold("Changes to be committed (staged):"))
			fmt.Println(cyan("  (use \"star commit <msg>\" to save staged changes)"))
			for _, entry := range details.Staged {
				fmt.Printf("  %s %s\n", green("staged:  "), entry.Path)
			}
		}

		if len(details.Modified) > 0 {
			fmt.Println("\n" + bold("Changes not staged for commit (modified):"))
			fmt.Println(cyan("  (use \"star add <file>...\" to update staged changes)"))
			for _, path := range details.Modified {
				fmt.Printf("  %s %s\n", red("modified:"), path)
			}
		}

		if len(details.Untracked) > 0 {
			fmt.Println("\n" + bold("Untracked files:"))
			fmt.Println(cyan("  (use \"star add <file>...\" to include in commit)"))
			for _, path := range details.Untracked {
				fmt.Printf("  %s%s\n", yellow("untracked: "), path)
			}
		}

	case "checkout":
		if len(os.Args) < 3 {
			fmt.Println("Usage: star checkout <branch_or_commit-hash>")
			os.Exit(1)
		}
		target := os.Args[2]
		if err := repo.Checkout(target); err != nil {
			handleCLIError(err)
		}
		fmt.Printf("Switched to '%s'\n", target)

	case "branch":
		if len(os.Args) == 2 {
			branches, err := repo.ListBranches()
			if err != nil {
				handleCLIError(err)
			}
			for _, b := range branches {
				if b.IsCurrent {
					fmt.Printf("* %s\n", green(b.Name))
				} else {
					fmt.Printf("  %s\n", b.Name)
				}
			}
		} else {
			branchName := os.Args[2]
			if err := repo.CreateBranch(branchName); err != nil {
				handleCLIError(err)
			}
			fmt.Printf("Created branch '%s'\n", branchName)
		}

	case "remote":
		if len(os.Args) == 2 {
			url, err := repo.GetRemote()
			if err != nil {
				handleCLIError(err)
			}
			fmt.Printf("origin\t%s\n", url)
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
				handleCLIError(err)
			}
			fmt.Printf("Set remote URL to '%s'\n", targetURL)
		}

	case "login", "config":
		handleLogin(repo)

	case "diff":
		diffs, err := repo.Diff()
		if err != nil {
			fmt.Printf("%s: %v\n", red("Error"), err)
			os.Exit(1)
		}
		if len(diffs) == 0 {
			fmt.Println("No modifications detected.")
			return
		}
		for _, diffFile := range diffs {
			fmt.Printf("%s %s\n", bold("diff --star"), cyan(diffFile.Path))
			for _, hunk := range diffFile.Hunks {
				if strings.HasPrefix(hunk, "+") {
					fmt.Println(green(hunk))
				} else if strings.HasPrefix(hunk, "-") {
					fmt.Println(red(hunk))
				} else {
					fmt.Println(hunk)
				}
			}
		}

	case "push":
		isForce := false
		if len(os.Args) >= 3 && (os.Args[2] == "--force" || os.Args[2] == "-f") {
			isForce = true
		}

		remoteURL, err := repo.GetRemote()
		if err != nil {
			handleCLIError(err)
		}
		branchName := "main"
		if branches, err := repo.ListBranches(); err == nil {
			for _, b := range branches {
				if b.IsCurrent {
					branchName = b.Name
					break
				}
			}
		}
		if isForce {
			fmt.Printf("Force-pushing branch '%s' to %s (%s)...\n", cyan(branchName), bold("origin"), yellow(remoteURL))
		} else {
			fmt.Printf("Pushing branch '%s' to %s (%s)...\n", cyan(branchName), bold("origin"), yellow(remoteURL))
		}

		if err := repo.Push(isForce); err != nil {
			handleCLIError(err)
		}
		fmt.Println("pushed to remote repository successfully")

	case "pull":
		remoteURL, err := repo.GetRemote()
		if err != nil {
			handleCLIError(err)
		}
		branchName := "main"
		if branches, err := repo.ListBranches(); err == nil {
			for _, b := range branches {
				if b.IsCurrent {
					branchName = b.Name
					break
				}
			}
		}
		fmt.Printf("Pulling remote changes from %s (%s) for branch '%s'...\n", bold("origin"), yellow(remoteURL), cyan(branchName))
		if err := repo.Pull(); err != nil {
			handleCLIError(err)
		}
		fmt.Println("pulled and integrated remote changes successfully")

	case "update", "self-update":
		fmt.Println("Checking for updates on GitHub...")
		msg, err := star.SelfUpdate(version)
		if err != nil {
			fmt.Printf("%s: %v\n", red("Update failed"), err)
			os.Exit(1)
		}
		fmt.Println(green(msg))

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
	fmt.Println(bold("Star Version Control System v0.1.0"))
	fmt.Println("A simpler, beginner-friendly VCS integrated with Git & GitHub.")
	fmt.Println()
	fmt.Println(bold("Usage:"))
	fmt.Println("  star <command> [arguments]")
	fmt.Println()
	fmt.Println(bold("Getting Started:"))
	fmt.Println("  " + yellow("init") + "          Initialize a new Star repository in current folder")
	fmt.Println("  " + yellow("login") + "         Configure your author name and email identity")
	fmt.Println("  " + yellow("remote") + " [url]  Show or set remote repository URL (GitHub)")
	fmt.Println()
	fmt.Println(bold("Daily Workflow:"))
	fmt.Println("  " + yellow("add") + " <path>    Stage a file or directory for commit")
	fmt.Println("  " + yellow("commit") + " <msg>  Record staged changes into a new commit")
	fmt.Println("  " + yellow("status") + "        Show branch status (staged, modified, untracked)")
	fmt.Println("  " + yellow("diff") + "          Show line-by-line file differences")
	fmt.Println("  " + yellow("pull") + "          Pull and integrate remote changes from GitHub")
	fmt.Println("  " + yellow("push") + " [--force] Push commit history to GitHub remote repository")
	fmt.Println("  " + yellow("log") + "           Display detailed commit history")
	fmt.Println()
	fmt.Println(bold("Branching & History:"))
	fmt.Println("  " + yellow("branch") + " [name] List branches or create a new branch")
	fmt.Println("  " + yellow("checkout") + " <ref> Switch to a branch or commit hash")
	fmt.Println()
	fmt.Println(bold("Utility:"))
	fmt.Println("  " + yellow("update") + "        Check GitHub for latest release and auto-update Star")
	fmt.Println("  " + yellow("version") + "       Display Star version")
	fmt.Println("  " + yellow("help") + "          Show this help message")
}
