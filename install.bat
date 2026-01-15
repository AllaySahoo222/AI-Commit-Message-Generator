@echo off
REM Installation script for AI Commit Message Generator (Windows Batch)
REM Downloads latest release from GitHub and installs
REM Supports Windows (64-bit)

setlocal enabledelayedexpansion

echo AI Commit Message Generator - Installation Script
echo ==================================================
echo.

REM GitHub repository
set "GITHUB_REPO=AllaySahoo222/AI-Commit-Message-Generator"
set "GITHUB_API=https://api.github.com/repos/%GITHUB_REPO%"

REM Detect architecture (default to amd64)
set "ZIP_NAME=generate-commit-windows.zip"
echo Detected: Windows 64-bit
echo.

REM Get latest release tag
echo Fetching latest release...
for /f "tokens=2 delims=:," %%a in ('powershell -Command "(Invoke-RestMethod -Uri '%GITHUB_API%/releases/latest').tag_name"') do (
    set "LATEST_TAG=%%a"
    set "LATEST_TAG=!LATEST_TAG: =!"
    set "LATEST_TAG=!LATEST_TAG:"=!"
)

if "!LATEST_TAG!"=="" (
    echo Error: Failed to fetch latest release tag
    pause
    exit /b 1
)

echo Latest version: !LATEST_TAG!
echo.

REM Construct download URL
set "DOWNLOAD_URL=https://github.com/%GITHUB_REPO%/releases/download/!LATEST_TAG!/%ZIP_NAME%"

REM Create temporary directory
set "TEMP_DIR=%TEMP%\generate-commit-install"
if exist "%TEMP_DIR%" rmdir /s /q "%TEMP_DIR%"
mkdir "%TEMP_DIR%"

REM Download the release
echo Downloading %ZIP_NAME%...
powershell -Command "Invoke-WebRequest -Uri '%DOWNLOAD_URL%' -OutFile '%TEMP_DIR%\%ZIP_NAME%'"
if errorlevel 1 (
    echo Error: Failed to download %ZIP_NAME%
    pause
    exit /b 1
)

REM Extract the zip
echo Extracting...
powershell -Command "Expand-Archive -Path '%TEMP_DIR%\%ZIP_NAME%' -DestinationPath '%TEMP_DIR%' -Force"
if errorlevel 1 (
    echo Error: Failed to extract %ZIP_NAME%
    pause
    exit /b 1
)

REM Find the binary (should be generate-commit.exe)
set "BINARY_NAME=generate-commit.exe"
set "BINARY_PATH=%TEMP_DIR%\%BINARY_NAME%"
if not exist "%BINARY_PATH%" (
    echo Error: Binary not found in archive
    pause
    exit /b 1
)

REM Installation directory (user directory, no admin required)
set "INSTALL_DIR=%USERPROFILE%\bin"
set "INSTALL_PATH=%INSTALL_DIR%\%BINARY_NAME%"

REM Create directory if it doesn't exist
if not exist "%INSTALL_DIR%" (
    mkdir "%INSTALL_DIR%"
    echo Created directory: %INSTALL_DIR%
)

REM Check if already installed
if exist "%INSTALL_PATH%" (
    echo generate-commit is already installed at %INSTALL_PATH%
    set /p OVERWRITE="Overwrite? (y/N): "
    if /i not "!OVERWRITE!"=="y" (
        echo Installation cancelled.
        exit /b 0
    )
)

REM Copy binary
echo Installing to %INSTALL_PATH%...
copy "%BINARY_PATH%" "%INSTALL_PATH%" >nul
if errorlevel 1 (
    echo Error: Failed to copy binary
    pause
    exit /b 1
)

echo [OK] Installed successfully!
echo.

REM Check if already in PATH
echo %PATH% | findstr /C:"%INSTALL_DIR%" >nul
if errorlevel 1 (
    echo Adding to PATH...
    setx PATH "%PATH%;%INSTALL_DIR%" >nul
    echo [OK] Added to PATH (requires new terminal session)
    echo.
)

REM Cleanup
rmdir /s /q "%TEMP_DIR%"

echo Installation complete!
echo.
echo Next steps:
echo 1. Open a new Command Prompt window (to refresh PATH)
echo 2. Set your OLLAMA_API_KEY:
echo    set OLLAMA_API_KEY=your_api_key
echo    (Add this to your system environment variables for persistence)
echo.
echo 3. Navigate to a git repository and run:
echo    generate-commit init
echo.
echo 4. Start committing with AI-generated messages!
echo.
pause
