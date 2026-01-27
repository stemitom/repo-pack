#Requires -Version 5.1
param(
    [string]$InstallDir,
    [switch]$NoModifyPath
)

$ErrorActionPreference = "Stop"
$Repo = "stemitom/repo-pack"
$Binary = "repo-pack"

function Main {
    $installDir = Get-InstallDir
    $platform = Get-Platform
    
    if (-not $platform) {
        Write-Error "unsupported platform: $env:PROCESSOR_ARCHITECTURE"
        exit 1
    }

    $version = Get-LatestVersion
    if (-not $version) {
        Write-Error "failed to get latest version"
        exit 1
    }

    Write-Host "installing repo-pack $version for $platform" -ForegroundColor Green

    $url = "https://github.com/$Repo/releases/download/$version/$Binary-$platform"
    $tmp = New-TemporaryFile
    $tmpPath = $tmp.FullName

    Write-Host "downloading $url" -ForegroundColor Green

    try {
        Invoke-WebRequest -Uri $url -OutFile $tmpPath -UseBasicParsing
    } catch {
        Write-Error "download failed: $_"
        exit 1
    }

    if (-not (Test-Path $installDir)) {
        New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    }

    $dest = Join-Path $installDir "$Binary.exe"
    Move-Item -Path $tmpPath -Destination $dest -Force

    Write-Host "installed to $dest" -ForegroundColor Green

    if (-not $NoModifyPath -and -not $env:REPO_PACK_NO_MODIFY_PATH) {
        Add-ToPath $installDir
    }
}

function Get-InstallDir {
    if ($InstallDir) {
        return $InstallDir
    }
    if ($env:REPO_PACK_INSTALL_DIR) {
        return $env:REPO_PACK_INSTALL_DIR
    }
    if ($env:XDG_BIN_HOME) {
        return $env:XDG_BIN_HOME
    }
    if ($env:XDG_DATA_HOME) {
        return Join-Path (Split-Path $env:XDG_DATA_HOME) "bin"
    }
    return Join-Path $env:USERPROFILE ".local\bin"
}

function Get-Platform {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "windows-x64.exe" }
        "x86" { return "windows-x64.exe" }
        "ARM64" { return "windows-x64.exe" }
        default { return $null }
    }
}

function Get-LatestVersion {
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing
        return $response.tag_name
    } catch {
        return $null
    }
}

function Add-ToPath($installDir) {
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    
    if ($currentPath -split ";" | Where-Object { $_ -eq $installDir }) {
        return
    }

    if ($env:GITHUB_PATH) {
        Add-Content -Path $env:GITHUB_PATH -Value $installDir
        return
    }

    $newPath = "$installDir;$currentPath"
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    $env:Path = "$installDir;$env:Path"

    Write-Host "added $installDir to PATH" -ForegroundColor Green
    Write-Host "restart your terminal for PATH changes to take effect" -ForegroundColor Yellow
}

Main
