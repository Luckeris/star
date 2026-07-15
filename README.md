# Star ⭐

A simpler, beginner-friendly alternative to Git commands — built in Go.

`star` is focused on making common version-control actions easier to understand and use, especially for people who are still learning Git workflows.

---

## Why Star?

Git is powerful, but the learning curve can feel steep at first.  
Star is my attempt to build a cleaner command-line experience for common tasks while learning Go and CLI development.

---

## Goals

- Make basic version-control workflows easier
- Provide clearer command usage for beginners
- Keep the tool lightweight and practical
- Learn and improve Go through real project development

---

## Current Status

🚧 **Work in progress**  
This project is actively being developed and improved over time.

---

## Planned Features

- Initialize repository helpers
- Simplified add/commit workflow
- Basic branch and status commands
- Easier remote setup helpers
- Better command help and error messages

> Feature list may change as the project evolves.

---

## Tech Stack

- **Language:** Go
- **Type:** CLI tool
- **Focus:** Developer productivity / learning tool

---

## Getting Started

### 1) Clone the repository

```bash
git clone https://github.com/Luckeris/star.git
cd star
```

### 2) Build the project

```bash
go build -o star .
```

### 3) Run it

```bash
./star
```

On Windows:

```powershell
star.exe
```

---

## Development

If you are contributing or testing locally:

```bash
go run .
```

Format code:

```bash
go fmt ./...
```

---

## Roadmap

- [ ] Add stable command structure
- [ ] Improve CLI help output
- [ ] Add more Git-like workflow commands
- [ ] Improve error handling and validation
- [ ] Add tests for core command behavior
- [ ] Write usage examples for each command

---

## Contributing

Feedback, issues, and suggestions are welcome.

If you try Star and something feels confusing or broken, please open an issue:
- what command you ran
- what you expected
- what happened instead

This helps me improve the tool faster.

---

## Disclaimer

Star is a learning project and is **not a full replacement for Git**.  
Use it for experimentation, learning, and lightweight workflows.

---

## License

This project is open-source and available under the MIT License.
