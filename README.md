# tuitui

A terminal user interface for [Jujutsu (jj)](https://github.com/jj-vcs/jj) version control, built in Go with [Bubble Tea v2](https://charm.land/bubbletea).

## Features

- **Graph log** — displays the jj commit graph with full DAG structure, colors, and working copy indicator
- **File browser** — lists changed files per revision with status indicators (A/M/D/R)
- **Diff panel** — shows syntax-highlighted diffs piped through [delta](https://github.com/dandavison/delta), with inline and side-by-side layouts
- **Live updates** — polls the repository for changes and auto-refreshes
- **Vim-style navigation** — `j`/`k`, `ctrl+u`/`ctrl+d`, `g`/`G`
- **Tokyo Night** color palette

## Requirements

- Go 1.26+
- [jj](https://github.com/jj-vcs/jj) (tested with 0.39.0)
- [delta](https://github.com/dandavison/delta) for syntax-highlighted diffs

## Install

```sh
go install github.com/frederickbeaulieu/tuitui@latest
```

Or build from source:

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

## Keybindings

| Key | Action |
|---|---|
| `q` / `ctrl+c` | Quit |
| `h` | Back |
| `l` | Open |
| `j` / `k` | Navigate / scroll |
| `ctrl+d` / `ctrl+u` | Half-page down / up |
| `g` / `G` | Jump to top / bottom |

### Diff panel extras

| Key | Action |
|---|---|
| `s` | Toggle side-by-side / inline |
| `z` | Toggle full file / changes only |

## Contributing

Contributions are welcome. Please open an issue to discuss your idea before submitting a pull request.

## License

[MIT](LICENSE)
