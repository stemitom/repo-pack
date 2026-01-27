# Repo-Pack

<!--toc:start-->

- [Repo-Pack](#repo-pack)
  - [Features](#features)
  - [Requirements](#requirements)
  - [Installation](#installation)
  - [Usage](#usage)
    - [Example](#example)
  - [Configuration](#configuration)
  - [Contributing](#contributing)
  - [License](#license)
  <!--toc:end-->

Repo-Pack is a Go-based tool designed to download files or directories from GitHub repositories, preserving the directory structure relative to a specified base directory. It's particularly useful for cloning parts of a repository or extracting specific files/directories without the need to clone the entire project.

## Features

- Download single files or entire directories from GitHub repositories
- Support for multiple URL formats:
  - Directory URLs (`/tree/`): `https://github.com/owner/repo/tree/main/path/to/dir`
  - File URLs (`/blob/`): `https://github.com/owner/repo/blob/main/path/to/file.txt`
  - Raw URLs: `https://raw.githubusercontent.com/owner/repo/main/path/to/file.txt`
- Enhanced progress display with:
  - Real-time download progress (files and bytes)
  - ETA calculation
  - Download speed (files/s and bytes/s)
  - Currently downloading files display
  - Colored output (with `--no-color` option to disable)
- Preserve the directory structure starting from a specified base directory
- Support for GitHub personal access tokens for private repositories
- Concurrent downloads with configurable limits
- Resume capability to skip already downloaded files
- Dry-run mode to preview files before downloading
- Custom output directory and filename support
- Verbose and quiet logging modes
- Git LFS (Large File Storage) support
- Graceful cancellation with Ctrl+C
- Comprehensive download summary

## Requirements

- Go 1.21.4 or higher

## Installation

- For mac/linux:
    `curl -LsSf https://dub.sh/repo-pack | sh`
- For windows:
    `curl -LsSf https://dub.sh/repo-pack-win | sh`

## Building

Clone the repository and build the binary:

```bash
git clone https://github.com/stemitom/repo-pack.git
cd repo-pack
go build -o repo-pack
```

## Usage

Run the tool with the required flags:

```bash
./repo-pack --url <repository_url> [OPTIONS]
```

### Required Flags

- `--url`: The full URL to the GitHub repository file or directory you wish to download
  - Directory: `https://github.com/owner/repo/tree/main/path/to/directory`
  - Single file: `https://github.com/owner/repo/blob/main/path/to/file.txt`
  - Raw file: `https://raw.githubusercontent.com/owner/repo/main/path/to/file.txt`

### Optional Flags

- `--token <token>`: Your GitHub personal access token (required for private repositories)
  - Can also be stored in `~/.github/token`
- `--output <directory>`: Output directory for downloaded files (default: current directory)
- `--output-file <filename>`: Custom filename for single file downloads (only works with `/blob/` or raw URLs)
- `--limit <number>`: Maximum concurrent downloads (default: 5, max recommended: 100)
- `--style <character>`: Progress bar style character (default: █)
- `--dry-run`: Preview files without downloading them
- `--resume`: Skip files that already exist locally
- `--verbose`: Enable verbose output (shows each file being downloaded)
- `--quiet`: Suppress non-error output
- `--no-color`: Disable colored output

### Examples

**Download a directory:**
```bash
./repo-pack --url https://github.com/JazzyGrim/dotfiles/tree/master/.config/nvim/lua
```

**Download a single file:**
```bash
./repo-pack --url https://github.com/owner/repo/blob/main/README.md
```

**Download a single file with custom name:**
```bash
./repo-pack --url https://github.com/owner/repo/blob/main/config.yaml --output-file my-config.yaml
```

**Download from raw URL:**
```bash
./repo-pack --url https://raw.githubusercontent.com/owner/repo/main/package.json
```

**Download to specific directory:**
```bash
./repo-pack --url https://github.com/owner/repo/tree/main/src --output ./my-project
```

**Preview files before downloading (dry-run):**
```bash
./repo-pack --url https://github.com/owner/repo/tree/main/docs --dry-run
```

**Resume interrupted download:**
```bash
./repo-pack --url https://github.com/owner/repo/tree/main/data --resume
```

**Download from private repository:**
```bash
./repo-pack --url https://github.com/owner/private-repo/tree/main/config --token YOUR_TOKEN
```

**Verbose output with higher concurrency:**
```bash
./repo-pack --url https://github.com/owner/repo/tree/main/assets --limit 20 --verbose
```

**Quiet mode (errors only):**
```bash
./repo-pack --url https://github.com/owner/repo/tree/main/files --quiet
```

**Disable colored output:**
```bash
./repo-pack --url https://github.com/owner/repo/tree/main/src --no-color
```

## Configuration

Repo-Pack automatically creates a configuration file at `~/.config/repo-pack/config.json` with default settings:

```json
{
  "concurrent_download_limit": 5,
  "progress_bar_style": "█",
  "github_token_path": "~/.github/token"
}
```

You can manually edit this file to change default settings. Command-line flags will override these defaults.

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
