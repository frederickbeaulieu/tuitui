// Package editor resolves the user's preferred text editor and builds
// *exec.Cmd values that open files in it. The resolution precedence is:
//
//  1. $VISUAL
//  2. $EDITOR
//  3. `jj config get ui.editor`
//  4. "vi"
//
// The resolved editor string may itself be a shell command (for example
// `code --wait`). When it contains whitespace, Command shells out through
// `sh -c` so flags are honored; otherwise the editor is invoked directly.
package editor

import (
	"os"
	"os/exec"
	"strings"
	"sync"
)

// Resolver resolves and caches the user's editor.
type Resolver struct {
	jjConfig func() string // optional: returns editor from jj config, or ""

	once sync.Once
	cmd  string
}

// NewResolver creates a resolver. jjConfig is called at most once and
// may be nil. It should return the value of `jj config get ui.editor`
// or an empty string if unavailable.
func NewResolver(jjConfig func() string) *Resolver {
	return &Resolver{jjConfig: jjConfig}
}

// Resolve returns the editor command string (may contain args, e.g.
// "code --wait"). Never returns an empty string.
func (r *Resolver) Resolve() string {
	r.once.Do(func() {
		if v := strings.TrimSpace(os.Getenv("VISUAL")); v != "" {
			r.cmd = v
			return
		}
		if v := strings.TrimSpace(os.Getenv("EDITOR")); v != "" {
			r.cmd = v
			return
		}
		if r.jjConfig != nil {
			if v := strings.TrimSpace(r.jjConfig()); v != "" {
				r.cmd = v
				return
			}
		}
		r.cmd = "vi"
	})
	return r.cmd
}

// Command builds an *exec.Cmd that opens path in the resolved editor.
// If the editor string contains whitespace it is executed via `sh -c`
// so user-supplied flags (e.g. `code --wait`) are honored.
func (r *Resolver) Command(path string) *exec.Cmd {
	ed := r.Resolve()
	if strings.ContainsAny(ed, " \t") {
		// Quote path for shell.
		return exec.Command("sh", "-c", ed+` "$0"`, path)
	}
	return exec.Command(ed, path)
}
