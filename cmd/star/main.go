package main

import (
	"fmt"
	"os"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: star <command>")
		fmt.Println("Available commands: help, version, init")
		return
	}
	command := os.Args[1]
	switch command {
	case "init":
		err := os.Mkdir(".star", 0755)
		if err != nil {
			fmt.Println("Error creating .star directory:", err)
			return
		}
		fmt.Println("Initialized empty star repository in .star directory")
	case "help":
		fmt.Println("Available commands: help, version, init")
	case "version":
		fmt.Println("star v0.1.0")
	default:
		fmt.Println("Unknown command:", command)
	}

}
