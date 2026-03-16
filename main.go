package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/frederickbeaulieu/tuitui/internal/app"
	"github.com/frederickbeaulieu/tuitui/internal/cli"
	"github.com/frederickbeaulieu/tuitui/internal/jj"
)

var version = "dev"

func main() {
	args := cli.Parse(version)

	runner := jj.NewRunner(args.RepoPath)

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
