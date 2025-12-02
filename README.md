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

Repo-Pack is a Go-based tool designed to download files from a specified GitHub repository directory, preserving the directory structure relative to a specified base directory. It's particularly useful for cloning parts of a repository or extracting specific directories without the need to clone the entire project.

## Features

- Download files from public and private GitHub repositories
- Preserve the directory structure starting from a specified base directory
- Support for GitHub personal access tokens for private repositories
- Concurrent downloads with configurable limits
- Resume capability to skip already downloaded files
- Dry-run mode to preview files before downloading
- Custom output directory support
- Verbose and quiet logging modes
- Progress bar with real-time download statistics
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

- `--url`: The full URL to the GitHub repository directory you wish to download
  - Example: `https://github.com/owner/repo/tree/main/path/to/directory`

### Optional Flags

- `--token <token>`: Your GitHub personal access token (required for private repositories)
  - Can also be stored in `~/.github/token`
- `--output <directory>`: Output directory for downloaded files (default: current directory)
- `--limit <number>`: Maximum concurrent downloads (default: 5, max recommended: 100)
- `--style <character>`: Progress bar style character (default: █)
- `--dry-run`: Preview files without downloading them
- `--resume`: Skip files that already exist locally
- `--verbose`: Enable verbose output (shows each file being downloaded)
- `--quiet`: Suppress non-error output

### Examples

**Basic usage:**
```bash
./repo-pack --url https://github.com/JazzyGrim/dotfiles/tree/master/.config/nvim/lua
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

