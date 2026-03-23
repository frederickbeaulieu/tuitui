// Package cli handles command-line argument parsing.
package cli

import (
	"fmt"
	"os"
	"slices"
)

type Args struct {
	RepoPath string
}

func Parse(version string) Args {
	args := os.Args[1:]

	handleVersion(args, version)

	var parsed Args
	for _, arg := range args {
		if arg[0] == '-' {
			fmt.Fprintf(os.Stderr, "Unknown flag: %s\n", arg)
			os.Exit(1)
		}
		if parsed.RepoPath == "" {
			parsed.RepoPath = arg
		}
	}

	return parsed
}

func handleVersion(args []string, version string) {
	if slices.Contains(args, "--version") || slices.Contains(args, "-v") {
		fmt.Println("tuitui " + version)
		os.Exit(0)
	}
}
