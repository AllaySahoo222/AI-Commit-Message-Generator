# Installation script for AI Commit Message Generator (Windows PowerShell)
# Downloads latest release from GitHub and installs
# Supports Windows (64-bit/ARM64)

$ErrorActionPreference = "Stop"

# GitHub repository
$GITHUB_REPO = "AllaySahoo222/AI-Commit-Message-Generator"
$GITHUB_API = "https://api.github.com/repos/$GITHUB_REPO"

Write-Host "AI Commit Message Generator - Installation Script" -ForegroundColor Cyan
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host ""

# Detect architecture
$arch = $env:PROCESSOR_ARCHITECTURE
$isArm64 = $false

# Check for ARM64
if ($env:PROCESSOR_ARCHITEW6432 -eq "ARM64" -or $arch -eq "ARM64") {
    $isArm64 = $true
    $ZIP_NAME = "generate-commit-windows-arm64.zip"
    Write-Host "Detected: Windows ARM64" -ForegroundColor Green
} elseif ($arch -eq "AMD64" -or $arch -eq "x86_64") {
    $ZIP_NAME = "generate-commit-windows.zip"
    Write-Host "Detected: Windows 64-bit" -ForegroundColor Green
} else {
    Write-Host "Error: Unsupported architecture: $arch" -ForegroundColor Red
    exit 1
}

# Get latest release tag
Write-Host "Fetching latest release..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$GITHUB_API/releases/latest"
    $LATEST_TAG = $response.tag_name
    Write-Host "Latest version: $LATEST_TAG" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "Error: Failed to fetch latest release tag" -ForegroundColor Red
    Write-Host $_.Exception.Message
    exit 1
}

# Construct download URL
$DOWNLOAD_URL = "https://github.com/$GITHUB_REPO/releases/download/$LATEST_TAG/$ZIP_NAME"

# Create temporary directory
$TEMP_DIR = Join-Path $env:TEMP "generate-commit-install"
if (Test-Path $TEMP_DIR) {
    Remove-Item -Path $TEMP_DIR -Recurse -Force
}
New-Item -ItemType Directory -Path $TEMP_DIR | Out-Null

try {
    # Download the release
    Write-Host "Downloading $ZIP_NAME..." -ForegroundColor Cyan
    $zipPath = Join-Path $TEMP_DIR $ZIP_NAME
    Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile $zipPath

    # Extract the zip
    Write-Host "Extracting..." -ForegroundColor Cyan
    Expand-Archive -Path $zipPath -DestinationPath $TEMP_DIR -Force

    # Find the binary (should be generate-commit.exe)
    $BINARY_NAME = "generate-commit.exe"
    $binaryPath = Join-Path $TEMP_DIR $BINARY_NAME
    if (-not (Test-Path $binaryPath)) {
        Write-Host "Error: Binary not found in archive" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "Error: Failed to download or extract $ZIP_NAME" -ForegroundColor Red
    Write-Host $_.Exception.Message
    exit 1
}

# Installation directory options
$userBin = Join-Path $env:USERPROFILE "bin"
$localBin = "C:\Program Files\generate-commit"

Write-Host ""
Write-Host "Choose installation location:"
Write-Host "1. User directory: $userBin (Recommended - no admin required)"
Write-Host "2. System directory: $localBin (Requires admin)"
Write-Host ""
$choice = Read-Host "Your choice (1/2)"

if ($choice -eq "1") {
    $INSTALL_DIR = $userBin
    $INSTALL_PATH = Join-Path $INSTALL_DIR $BINARY_NAME
    
    # Create directory if it doesn't exist
    if (-not (Test-Path $INSTALL_DIR)) {
        New-Item -ItemType Directory -Path $INSTALL_DIR | Out-Null
        Write-Host "Created directory: $INSTALL_DIR" -ForegroundColor Green
    }
    
    # Check if already installed
    if (Test-Path $INSTALL_PATH) {
        Write-Host "generate-commit is already installed at $INSTALL_PATH" -ForegroundColor Yellow
        $response = Read-Host "Overwrite? (y/N)"
        if ($response -notmatch "^[Yy]$") {
            Write-Host "Installation cancelled."
            exit 0
        }
    }
    
    # Copy binary
    Write-Host "Installing to $INSTALL_PATH..." -ForegroundColor Cyan
    Copy-Item $binaryPath $INSTALL_PATH -Force
    Write-Host "✓ Installed successfully!" -ForegroundColor Green
    
    # Add to PATH if not already there
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($currentPath -notlike "*$INSTALL_DIR*") {
        Write-Host ""
        Write-Host "Adding to PATH..." -ForegroundColor Cyan
        [Environment]::SetEnvironmentVariable("Path", "$currentPath;$INSTALL_DIR", "User")
        Write-Host "✓ Added to PATH (requires new terminal session)" -ForegroundColor Green
    }
    
} elseif ($choice -eq "2") {
    $INSTALL_DIR = $localBin
    $INSTALL_PATH = Join-Path $INSTALL_DIR $BINARY_NAME
    
    # Check for admin rights
    $isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    if (-not $isAdmin) {
        Write-Host "Error: Administrator rights required for system installation" -ForegroundColor Red
        Write-Host "Please run PowerShell as Administrator, or choose option 1." -ForegroundColor Yellow
        exit 1
    }
    
    # Create directory if it doesn't exist
    if (-not (Test-Path $INSTALL_DIR)) {
        New-Item -ItemType Directory -Path $INSTALL_DIR | Out-Null
        Write-Host "Created directory: $INSTALL_DIR" -ForegroundColor Green
    }
    
    # Check if already installed
    if (Test-Path $INSTALL_PATH) {
        Write-Host "generate-commit is already installed at $INSTALL_PATH" -ForegroundColor Yellow
        $response = Read-Host "Overwrite? (y/N)"
        if ($response -notmatch "^[Yy]$") {
            Write-Host "Installation cancelled."
            exit 0
        }
    }
    
    # Copy binary
    Write-Host "Installing to $INSTALL_PATH..." -ForegroundColor Cyan
    Copy-Item $binaryPath $INSTALL_PATH -Force
    Write-Host "✓ Installed successfully!" -ForegroundColor Green
    
    # Add to system PATH if not already there
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
    if ($currentPath -notlike "*$INSTALL_DIR*") {
        Write-Host ""
        Write-Host "Adding to system PATH..." -ForegroundColor Cyan
        [Environment]::SetEnvironmentVariable("Path", "$currentPath;$INSTALL_DIR", "Machine")
        Write-Host "✓ Added to PATH (requires new terminal session)" -ForegroundColor Green
    }
} else {
    Write-Host "Invalid choice. Installation cancelled." -ForegroundColor Red
    exit 1
}

# Cleanup
Remove-Item -Path $TEMP_DIR -Recurse -Force -ErrorAction SilentlyContinue

# Verify installation
Write-Host ""
Write-Host "Installation complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:"
Write-Host "1. Open a new terminal/PowerShell window (to refresh PATH)"
Write-Host "2. Set your OLLAMA_API_KEY:"
Write-Host "   `$env:OLLAMA_API_KEY = 'your_api_key'"
Write-Host "   (Add this to your PowerShell profile for persistence)"
Write-Host ""
Write-Host "3. Navigate to a git repository and run:"
Write-Host "   generate-commit init"
Write-Host ""
Write-Host "4. Start committing with AI-generated messages!"
