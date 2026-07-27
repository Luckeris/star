# Star ⭐

A simpler git-like version control system. Integrated with git, for uploading and pushing to your remote repos.

The `star` project is focused on making git more available and easy-to-use for new people in programming industry.

I have suffered too, so I had the vision to make it easier for myself and for other people.

---

## Features

- init (initialization of the .star structure)
- add (adds a specific file or by using the . argument, you can add whole folders)
- commit (commits with a message, saves and hashes the commit)
- help (prints out all the usable commands and their usage)
- log (prints all commits etc.)
- status (shows currently tracked files)
- branch (creates a new branch or shows your actual branch)
- checkout (switches to a different branch)

---

## Build & Installation

To build the project from source, you need to have [Go](https://go.dev/) installed on your system.

1. Navigate to the root directory of the project
2. Compile the binary using the Go compiler:

```bash
go build -o star
```

To use the star command globally from anywhere in your terminal, move the generated executable to a directory that is included in your system's PATH variable (e.g., /usr/local/bin on Linux/macOS, or add the project folder to Environment Variables on Windows).

---

## Usage

```bash
# 1. Initialize an empty star repository in your current directory

star init
```

```bash
# 2. Add some files or ignore unwanted ones via .starignore, then stage them all

star add .
```

```bash
# 3. Commit your staged changes with a descriptive message

star commit "Initial commit"
```

```bash
# 4. View your commit history

star log
```

```bash
# 5. Create and switch to a new branch for a new feature

star branch feature-auth
star checkout feature-auth
```
