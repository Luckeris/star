# Star ⭐

A simpler git-like version control system. Integrated with git, for uploading and pushing to your remote repos.

The `star` project is focused on making git more available and easy-to-use for new people in programming industry.

I have suffered too, so I had the vision to make it easier for myself and for other people.

---

## Features

- **`init`**: Initialize the `.star` repository structure.
- **`login`**: Configure author name and email identity.
- **`remote`**: Show or configure remote repository URL (`star remote <url>`).
- **`add`**: Stage a specific file or an entire directory (using `.`). Respects `.starignore`.
- **`commit`**: Record staged changes into a new commit with a message.
- **`push`**: Push commit history to remote repository.
- **`log`**: Display commit history.
- **`status`**: Show branch status with staged, modified, and untracked files.
- **`branch`**: Create a new branch or list existing branches.
- **`checkout`**: Switch to a branch or commit hash.
- **`hash-object`**: Compute the SHA-256 hash of a file and save it to object store.
- **`version`**: Display current Star version.
- **`update`**: Check GitHub for the latest release and auto-update Star executable.
- **`help`**: Print available commands and usage instructions.

---

## Build & Installation

To build the project from source, ensure you have [Go](https://go.dev/) installed on your system.

1. Clone or navigate to the root directory of the repository.
2. Compile the binary using the Go compiler:

```bash
go build -o star ./cmd/star
```

To run unit tests across all packages:

```bash
go test ./...
```

To use `star` globally from anywhere in your terminal, move the generated executable to a directory included in your system's `PATH` variable (e.g., `/usr/local/bin` on Linux/macOS).

---

## Usage

```bash
# 1. Initialize an empty star repository in your current directory
star init

# 2. Configure author name and email identity
star login "Your Name" "your.email@example.com"

# 3. Configure remote repository URL
star remote https://github.com/username/repository.git

# 4. Stage a file or stage all files in current directory recursively
star add .

# 5. Commit your staged changes with a descriptive message
star commit "Initial commit"

# 6. Push commit history to GitHub
star push

# 7. View your commit history
star log

# 8. Check branch and working tree status
star status

# 9. Create and switch to a new branch
star branch feature-auth
star checkout feature-auth
```

---

## Project Structure

```text
star/
├── cmd/
│   └── star/             # Main executable CLI entrypoint
└── internal/
    └── star/             # Core VCS domain engine & unit tests
```
