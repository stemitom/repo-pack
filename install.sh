#!/bin/sh
set -eu

REPO="stemitom/repo-pack"
BINARY="repo-pack"

main() {
    install_dir=$(get_install_dir)
    platform=$(detect_platform)
    
    if [ -z "$platform" ]; then
        err "unsupported platform: $(uname -s)-$(uname -m)"
    fi

    version=$(get_latest_version)
    if [ -z "$version" ]; then
        err "failed to get latest version"
    fi

    say "installing repo-pack $version for $platform"

    tmp=$(mktemp -d)
    trap 'rm -rf "$tmp"' EXIT

    url="https://github.com/$REPO/releases/download/$version/$BINARY-$platform"
    say "downloading $url"
    
    download "$url" "$tmp/$BINARY"

    mkdir -p "$install_dir"
    chmod +x "$tmp/$BINARY"
    mv "$tmp/$BINARY" "$install_dir/$BINARY"

    say "installed to $install_dir/$BINARY"

    if [ "${REPO_PACK_NO_MODIFY_PATH:-}" != "1" ]; then
        add_to_path "$install_dir"
    fi
}

get_install_dir() {
    if [ -n "${REPO_PACK_INSTALL_DIR:-}" ]; then
        echo "$REPO_PACK_INSTALL_DIR"
    elif [ -n "${XDG_BIN_HOME:-}" ]; then
        echo "$XDG_BIN_HOME"
    elif [ -n "${XDG_DATA_HOME:-}" ]; then
        echo "$XDG_DATA_HOME/../bin"
    else
        echo "$HOME/.local/bin"
    fi
}

detect_platform() {
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    arch=$(uname -m)

    case "$os" in
        linux)
            case "$arch" in
                x86_64|amd64) echo "linux-x64" ;;
                aarch64|arm64) echo "linux-arm64" ;;
                *) return 1 ;;
            esac
            ;;
        darwin)
            case "$arch" in
                x86_64|amd64) echo "macos-x64" ;;
                aarch64|arm64) echo "macos-arm64" ;;
                *) return 1 ;;
            esac
            ;;
        *) return 1 ;;
    esac
}

get_latest_version() {
    download "https://api.github.com/repos/$REPO/releases/latest" - 2>/dev/null | grep '"tag_name"' | cut -d'"' -f4
}

download() {
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$1" ${2:+-o "$2"}
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "${2:--}" "$1"
    else
        err "curl or wget required"
    fi
}

add_to_path() {
    _install_dir="$1"
    
    case ":$PATH:" in
        *":$_install_dir:"*) return ;;
    esac

    if [ -n "${GITHUB_PATH:-}" ]; then
        echo "$_install_dir" >> "$GITHUB_PATH"
        return
    fi

    _env_dir="${XDG_CONFIG_HOME:-$HOME/.config}/repo-pack"
    _env_file="$_env_dir/env"
    
    mkdir -p "$_env_dir"
    cat > "$_env_file" << EOF
# repo-pack shell setup
case ":\$PATH:" in
    *":$_install_dir:"*) ;;
    *) export PATH="$_install_dir:\$PATH" ;;
esac
EOF

    _source_line=". \"$_env_file\""
    
    for _profile in "$HOME/.profile" "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.zshrc" "$HOME/.zshenv"; do
        if [ -f "$_profile" ]; then
            if ! grep -qF "$_env_file" "$_profile" 2>/dev/null; then
                echo "$_source_line" >> "$_profile"
            fi
        fi
    done

    if [ -d "$HOME/.config/fish" ]; then
        mkdir -p "$HOME/.config/fish/conf.d"
        cat > "$HOME/.config/fish/conf.d/repo-pack.fish" << EOF
if not contains "$_install_dir" \$PATH
    set -gx PATH "$_install_dir" \$PATH
end
EOF
    fi

    say "added $_install_dir to PATH via $_env_file"
    say "restart your shell or run: . \"$_env_file\""
}

say() {
    printf '\033[0;32m%s\033[0m\n' "$*"
}

err() {
    printf '\033[0;31merror: %s\033[0m\n' "$*" >&2
    exit 1
}

main "$@"
