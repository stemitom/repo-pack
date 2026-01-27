# repo-pack

Download files from GitHub repository directories, preserving directory structure.

## Features

- Download from public and private repositories
- Concurrent downloads with configurable limits
- Git LFS support
- Resume interrupted downloads
- Dry-run mode to preview files
- Progress bar with download statistics
- Graceful cancellation with Ctrl+C

## Installation

```console
$ curl -LsSf https://dub.sh/repo-pack | sh
```

Or build from source:

```console
$ cargo install --path .
```

## Usage

```console
$ repo-pack <URL> [OPTIONS]
```

### Examples

Download a directory:

```console
$ repo-pack https://github.com/astral-sh/uv/tree/main/crates/uv-fs/src
```

Download to a specific output directory:

```console
$ repo-pack https://github.com/owner/repo/tree/main/src -o ./my-project
```

Preview files without downloading:

```console
$ repo-pack https://github.com/owner/repo/tree/main/docs --dry-run
```

Resume an interrupted download:

```console
$ repo-pack https://github.com/owner/repo/tree/main/data --resume
```

Download from a private repository:

```console
$ repo-pack https://github.com/owner/repo/tree/main/config --token ghp_xxxx
```

### Options

| Option | Description |
|--------|-------------|
| `-o, --output <DIR>` | Output directory (default: `.`) |
| `-l, --limit <NUM>` | Concurrent download limit (default: `5`) |
| `-n, --dry-run` | Preview files without downloading |
| `-r, --resume` | Skip files that already exist |
| `-v, --verbose` | Show verbose output |
| `-q, --quiet` | Suppress non-error output |
| `--token <TOKEN>` | GitHub personal access token |
| `--no-progress` | Disable progress bar |

## Configuration

Configuration is stored at `~/.config/repo-pack/config.json`:

```json
{
  "concurrent_download_limit": 5,
  "progress_bar_style": "█",
  "github_token_path": "~/.github/token"
}
```

If `--token` is not provided, repo-pack reads from `github_token_path`.

## License

MIT
