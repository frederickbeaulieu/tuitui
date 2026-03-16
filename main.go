package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/frederickbeaulieu/tuitui/internal/app"
	"github.com/frederickbeaulieu/tuitui/internal/jj"
)

func main() {
	repoPath := ""
	if len(os.Args) > 1 {
		repoPath = os.Args[1]
	}

	runner := jj.NewRunner(repoPath)

	_, err := runner.Run("root")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: not a jj repository (or jj is not installed)\n")
		os.Exit(1)
	}

	watcher := jj.NewRepoWatcher(runner)
	defer func() { _ = watcher.Close() }()

	model := app.New(runner, watcher)
	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
