// Package cli handles command-line argument parsing.
package cli

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

const logoArt = `
⠴⠶⣖⡋⠉⠍⠑⠢⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⠔⠊⠩⠉⢙⣲⠶⠦
⠀⠀⠀⢣⠀⠀⠀⠀⠘⢦⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⡴⠃⠀⠀⠀⠀⡜⠀⠀⠀
⠀⠀⣠⠼⡆⠀⢀⡔⠊⠉⠉⠓⠦⣄⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣠⠴⠚⠉⠉⠑⢢⡀⠀⢰⠧⣄⠀⠀
⠀⠀⠈⠚⡇⠀⠀⠃⠀⠀⠀⠀⠀⠀⠈⠓⢄⡀⠀⠀⠀⠀⠀⠀⠀ tuitui⠀⠀⠀⠀⠀⠀⠀⠀⢀⡠⠚⠁⠀⠀⠀⠀⠀⠀⠘⠀⠀⢸⠓⠁⠀⠀
⠀⠀⠀⠀⠘⣄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⠢⣀⠀⠀⠀⠀⠀ %s⠀⠀⠀⠀⠀⠀⣀⠔⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⠃⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠈⠲⣄⠀⠀⠀⠀⠀⠀⠒⠤⣀⠀⠀⠈⠳⣄⠀⠀⠀⠀⠀⠀  ⠀⠀⠀⠀⠀⠀⣠⠞⠁⠀⠀⣀⠤⠒⠀⠀⠀⠀⠀⠀⣠⠖⠁⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠉⠒⠤⢄⣀⣀⠀⠀⠀⢉⣓⠲⠀⣬⣕⡦⠀⠀⠀⠀  ⠀⠀⠀⠀⢴⣪⣥⠀⠖⣚⡉⠀⠀⠀⣀⣀⡠⠤⠒⠉⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠸⠿⠛⡶⠋⠁⠈⠑⣄⠘⢆⠀⠀⠀⠀⠀  ⠀⠀⠀⠀⠀⡰⠃⣠⠊⠁⠈⠙⢶⠛⠿⠇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⠎⢀⠞⠁⠀⠀⠀⠀⠈⢢⡈⠳⡀⠀⠀⠀  ⠀⠀⠀⢀⠞⢁⡔⠁⠀⠀⠀⠀⠈⠳⡀⠱⣄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠒⠓⠒⠛⠂⠀⠀⠀⠀⠀⠀⠀⠱⡄⠙⣄⠀⠀  ⠀⠀⣠⠋⢠⠎⠀⠀⠀⠀⠀⠀⠀⠐⠛⠒⠚⠒⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠦⢬⡦⠀  ⠀⢴⡥⠴⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀`

// centerVersion pads the version string so it appears centered
// below "tuitui" in the logo. The slot is 6 characters wide,
// matching the width of "tuitui" on the line above.
func centerVersion(version string) string {
	const slotWidth = 6
	n := len(version)
	if n >= slotWidth {
		return version[:slotWidth]
	}
	pad := (slotWidth - n) / 2
	right := slotWidth - pad - n
	return strings.Repeat(" ", pad) + version + strings.Repeat(" ", right)
}

func logo(version string) string {
	return fmt.Sprintf(logoArt, centerVersion(version))
}

type Args struct {
	RepoPath string
}

func Parse(version string) Args {
	args := os.Args[1:]

	handleHelp(args, version)
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
		fmt.Println(logo(version))
		os.Exit(0)
	}
}

func handleHelp(args []string, version string) {
	if slices.Contains(args, "--help") || slices.Contains(args, "-h") {
		fmt.Println(logo(version))
		fmt.Println()
		fmt.Println("A terminal user interface for Jujutsu (jj)")
		fmt.Println()
		fmt.Println("Usage: tuitui [path]")
		fmt.Println()
		fmt.Println("Arguments:")
		fmt.Println("  [path]    Path to a jj repository (defaults to current directory)")
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println("  -h, --help       Show this help message")
		fmt.Println("  -v, --version    Show version")
		os.Exit(0)
	}
}
