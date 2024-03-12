# Set the repository and binary name
$REPO = "stemitom/repo-pack"
$BINARY_NAME = "repo-pack"

# Colors
$RED = [System.ConsoleColor]::Red
$GREEN = [System.ConsoleColor]::Green
$YELLOW = [System.ConsoleColor]::Yellow
$BLUE = [System.ConsoleColor]::Blue
$NC = [System.ConsoleColor]::White

# Check if the binary exists and delete it
if (Test-Path "C:\Program Files\$BINARY_NAME.exe") {
    Write-Host "Removing existing $BINARY_NAME binary..." -ForegroundColor $YELLOW
    Remove-Item "C:\Program Files\$BINARY_NAME.exe"
}

Write-Host "Getting the latest release information..." -ForegroundColor $BLUE
$LATEST_VERSION = (Invoke-RestMethod -Uri "https://api.github.com/repos/$REPO/releases/latest").tag_name

Write-Host "Getting the machine architecture..." -ForegroundColor $BLUE
$ARCH = $env:PROCESSOR_ARCHITECTURE
$KERNEL = "Windows"

Write-Host "Downloading the binary..." -ForegroundColor $GREEN
$DOWNLOAD_URL = "https://github.com/$REPO/releases/download/$LATEST_VERSION/${BINARY_NAME}_${KERNEL}_${ARCH}.zip"
Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile "${BINARY_NAME}.zip"

Write-Host "Extracting the binary..." -ForegroundColor $GREEN
Expand-Archive -Path "${BINARY_NAME}.zip"

Write-Host "Making the binary executable..." -ForegroundColor $GREEN

Write-Host "Install the binary? (y/n) " -ForegroundColor $YELLOW -NoNewline
$INSTALL = Read-Host

if ($INSTALL -eq "y") {
    Write-Host "Installing the binary..." -ForegroundColor $GREEN
    Move-Item -Path "${BINARY_NAME}.exe" -Destination "C:\Program Files\"
} else {
    Write-Host "Skipping installation..." -ForegroundColor $RED
}

Write-Host "Cleaning up..." -ForegroundColor $GREEN
Remove-Item "${BINARY_NAME}.zip"

Write-Host "Done!" -ForegroundColor $GREEN
