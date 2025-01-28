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

- Download files from public GitHub repositories.
- Preserve the directory structure starting from a specified base directory.
- Support for GitHub personal access tokens for private repositories (feature in progress).

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
./repo-pack --url <repository_url> [--limit <concurrent_download_limit>] [--style <progress_bar_style>] [--token <personal_access_token>]
```

- `--url`: The full URL to the GitHub repository directory you wish to download.
- `--token`: Your GitHub personal access token (optional, required for private repositories).
- `--limit`: The amount of concurrent download limits (optional, default is 10).

### Example

To download the `lua` directory from a repository:

```bash
./repo-pack --url https://github.com/JazzyGrim/dotfiles/tree/master/.config/nvim/lua
```

This will create a directory named `lua` in your current working directory and download all files under the `.config/nvim/lua` directory from the repository, preserving the structure under `lua`.

## Configuration

No additional configuration is required. However, you can set up a `.gitignore` file to ignore binaries or other directories as needed.

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

