# tuitui

A terminal user interface for [Jujutsu (jj)](https://github.com/jj-vcs/jj) version control, built in Go with [Bubble Tea v2](https://charm.land/bubbletea).

## Features

- **Graph log** — displays the jj commit graph with full DAG structure, colors, and working copy indicator
- **File browser** — lists changed files per revision with status indicators (A/M/D/R)
- **Diff panel** — shows syntax-highlighted diffs piped through [delta](https://github.com/dandavison/delta), with inline and side-by-side layouts
- **Command bar** — run any jj command with `:`; includes dynamic completion powered by jj's built-in shell completion engine, ghost text, and a navigable suggestions dropdown
- **Live updates** — polls the repository for changes and auto-refreshes
- **Vim-style navigation** — `j`/`k`, `ctrl+u`/`ctrl+d`, `g`/`G`
- **Tokyo Night** color palette

## Requirements

- Go 1.26+
- [jj](https://github.com/jj-vcs/jj) (tested with 0.39.0)
- [delta](https://github.com/dandavison/delta) for syntax-highlighted diffs

## Install

### Homebrew (macOS)

```sh
brew install frederickbeaulieu/tap/tuitui
```

### Go

```sh
go install github.com/frederickbeaulieu/tuitui@latest
```

### Download a binary

Pre-built binaries for Linux, macOS, and Windows are available on the
[Releases](https://github.com/frederickbeaulieu/tuitui/releases) page.

Download the archive for your platform, extract it, and place the `tuitui`
binary somewhere on your `PATH`.

### Build from source

```sh
git clone https://github.com/frederickbeaulieu/tuitui.git
cd tuitui
go build -o tuitui .
```

## Usage

Run inside any jj repository:

```sh
tuitui
```

Or specify a repo path:

```sh
tuitui /path/to/repo
```

Print the version:

```sh
tuitui --version
```

## Keybindings

| Key | Action |
|---|---|
| `q` / `ctrl+c` | Quit |
| `h` / `left` | Back |
| `l` / `right` | Open |
| `j` / `k` / `up` / `down` | Navigate / scroll |
| `ctrl+d` / `ctrl+u` | Half-page down / up |
| `g` / `G` | Jump to top / bottom |

### Log panel

| Key | Action |
|---|---|
| `z` | Toggle all revisions / current tree |

### Diff panel

| Key | Action |
|---|---|
| `s` | Toggle side-by-side / inline |
| `z` | Toggle full file / changes only |

### Command bar

Open with `:` to run any jj command. Completions appear automatically as you type.

| Key | Action |
|---|---|
| `:` | Open command bar |
| `tab` | Accept completion |
| `ctrl+n` / `ctrl+p` | Navigate suggestions |
| `ctrl+space` | Toggle suggestions dropdown |
| `enter` | Run command |
| `esc` | Close command bar |

## Contributing

Contributions are welcome. Please open an issue to discuss your idea before submitting a pull request.

## License

[MIT](LICENSE)
