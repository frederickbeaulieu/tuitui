package jj

import "strings"

// IsInteractive reports whether a jj invocation with the given args is
// likely to need a real terminal (because jj will spawn $EDITOR or its
// built-in diff editor).
//
// The rules are derived from the flags the user actually typed:
//
//   - Any -i / --interactive flag  → interactive.
//   - commit / describe / split without a message-providing flag
//     (-m, --message, --stdin, --no-edit) → interactive (opens $EDITOR).
//   - diffedit is always interactive.
//   - resolve with no revision args is interactive (merge tool).
//   - Everything else is non-interactive.
//
// args is the argument list as it would be passed to `jj` (i.e. the
// first element is the subcommand, e.g. "commit").
func IsInteractive(args []string) bool {
	sub, rest := subcommand(args)
	if sub == "" {
		return false
	}

	if hasAnyFlag(rest, "-i", "--interactive") {
		return true
	}

	switch sub {
	case "diffedit":
		return true
	case "commit", "describe", "split":
		if hasMessageFlag(rest) {
			return false
		}
		return true
	case "resolve":
		// `jj resolve` with no explicit path/revset launches the merge tool.
		// If the user passed any positional arg, assume they know what
		// they're doing and still treat it as interactive — resolve only
		// ever runs the merge tool.
		return true
	}
	return false
}

// subcommand returns the first non-global-flag token and the remaining args.
// It skips a handful of global flags that take values so they don't get
// mistaken for the subcommand. We don't need to be exhaustive here — users
// of the cmdbar don't type global flags, and the runner injects them itself.
func subcommand(args []string) (string, []string) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if !strings.HasPrefix(a, "-") {
			return a, args[i+1:]
		}
		// Skip values for known global flags that take an argument.
		switch a {
		case "-R", "--repository", "--at-op", "--at-operation", "--config", "--config-file":
			i++ // skip the value
		}
	}
	return "", nil
}

func hasAnyFlag(args []string, flags ...string) bool {
	for _, a := range args {
		// Stop scanning after `--`.
		if a == "--" {
			return false
		}
		for _, f := range flags {
			if a == f {
				return true
			}
			// Support --flag=value form for long flags.
			if strings.HasPrefix(f, "--") && strings.HasPrefix(a, f+"=") {
				return true
			}
		}
	}
	return false
}

// hasMessageFlag returns true if args contain any flag that would
// suppress the interactive editor for commit/describe/split.
func hasMessageFlag(args []string) bool {
	return hasAnyFlag(args,
		"-m", "--message",
		"--stdin",
		"--no-edit",
	)
}
