# ClickUp CLI Installation Script for Windows
#
# This script installs the ClickUp CLI (cu) on Windows systems
# Usage: 
#   irm https://raw.githubusercontent.com/timimsms/cu/main/scripts/install.ps1 | iex
#
# You can also specify a version:
#   $env:CU_VERSION="v1.0.0"; irm https://raw.githubusercontent.com/timimsms/cu/main/scripts/install.ps1 | iex
#
# Or install to a custom directory:
#   $env:CU_INSTALL_DIR="C:\Program Files\cu"; irm https://raw.githubusercontent.com/timimsms/cu/main/scripts/install.ps1 | iex

param(
    [string]$Version = $env:CU_VERSION,
    [string]$InstallDir = $env:CU_INSTALL_DIR
)

# Configuration
$RepoOwner = "timimsms"
$RepoName = "cu"
$BinaryName = "cu"

# Set defaults
if (-not $Version) {
    $Version = "latest"
}

if (-not $InstallDir) {
    $InstallDir = "$env:LOCALAPPDATA\Programs\cu"
}

# Error handling
$ErrorActionPreference = "Stop"

# Helper functions
function Write-Info {
    param([string]$Message)
    Write-Host $Message -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host $Message -ForegroundColor Green
}

function Write-Error {
    param([string]$Message)
    Write-Host "Error: $Message" -ForegroundColor Red
}

function Write-Warning {
    param([string]$Message)
    Write-Host $Message -ForegroundColor Yellow
}

# Detect architecture
function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "x86_64" }
        "x86" { return "i386" }
        "ARM64" { return "arm64" }
        default {
            Write-Error "Unsupported architecture: $arch"
            exit 1
        }
    }
}

# Get the latest version from GitHub
function Get-LatestVersion {
    try {
        $latestUrl = "https://api.github.com/repos/$RepoOwner/$RepoName/releases/latest"
        $response = Invoke-RestMethod -Uri $latestUrl -UseBasicParsing
        return $response.tag_name
    }
    catch {
        Write-Error "Failed to get latest version: $_"
        exit 1
    }
}

# Download file with progress
function Download-File {
    param(
        [string]$Url,
        [string]$Output
    )
    
    try {
        Write-Info "Downloading from $Url..."
        
        # Use Invoke-WebRequest with progress
        $ProgressPreference = 'Continue'
        Invoke-WebRequest -Uri $Url -OutFile $Output -UseBasicParsing
        
        if (-not (Test-Path $Output)) {
            throw "Download failed - file not found"
        }
    }
    catch {
        Write-Error "Failed to download file: $_"
        exit 1
    }
}

# Calculate SHA256 hash
function Get-FileHash256 {
    param([string]$FilePath)
    
    $hash = Get-FileHash -Path $FilePath -Algorithm SHA256
    return $hash.Hash.ToLower()
}

# Verify checksum
function Verify-Checksum {
    param(
        [string]$FilePath,
        [string]$ChecksumsUrl
    )
    
    Write-Info "Verifying checksum..."
    
    # Download checksums file
    $checksumsFile = Join-Path $env:TEMP "checksums.txt"
    Download-File -Url $ChecksumsUrl -Output $checksumsFile
    
    # Read checksums
    $checksums = Get-Content $checksumsFile
    
    # Find expected checksum
    $fileName = Split-Path $FilePath -Leaf
    $expectedLine = $checksums | Where-Object { $_ -match [regex]::Escape($fileName) }
    
    if (-not $expectedLine) {
        Write-Warning "Could not find checksum for $fileName, skipping verification"
        Remove-Item $checksumsFile -Force
        return
    }
    
    $expectedChecksum = ($expectedLine -split '\s+')[0]
    
    # Calculate actual checksum
    $actualChecksum = Get-FileHash256 -FilePath $FilePath
    
    # Compare
    if ($expectedChecksum -ne $actualChecksum) {
        Write-Error "Checksum verification failed!"
        Write-Error "Expected: $expectedChecksum"
        Write-Error "Actual:   $actualChecksum"
        Remove-Item $checksumsFile -Force
        exit 1
    }
    
    Write-Success "Checksum verified ✓"
    Remove-Item $checksumsFile -Force
}

# Add to PATH
function Add-ToPath {
    param([string]$Directory)
    
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    
    if ($userPath -notlike "*$Directory*") {
        Write-Info "Adding $Directory to PATH..."
        
        $newPath = $userPath
        if ($newPath -and $newPath[-1] -ne ';') {
            $newPath += ';'
        }
        $newPath += $Directory
        
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        
        # Update current session
        $env:Path = [Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + $newPath
        
        Write-Success "Added to PATH ✓"
        Write-Warning "Note: You may need to restart your terminal for PATH changes to take effect"
    }
    else {
        Write-Info "$Directory is already in PATH"
    }
}

# Main installation function
function Install-ClickUpCLI {
    Write-Host ""
    Write-Info "ClickUp CLI Installer for Windows"
    Write-Info "================================="
    Write-Host ""
    
    # Detect architecture
    $arch = Get-Architecture
    Write-Info "Detected architecture: $arch"
    
    # Get version to install
    if ($Version -eq "latest") {
        Write-Info "Fetching latest version..."
        $Version = Get-LatestVersion
    }
    Write-Info "Installing version: $Version"
    
    # Construct download URL
    $platform = "windows_$arch"
    $archiveName = "${BinaryName}_${platform}.zip"
    $downloadUrl = "https://github.com/$RepoOwner/$RepoName/releases/download/$Version/$archiveName"
    $checksumsUrl = "https://github.com/$RepoOwner/$RepoName/releases/download/$Version/checksums.txt"
    
    # Create temp directory
    $tempDir = Join-Path $env:TEMP "cu-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
    
    try {
        Push-Location $tempDir
        
        # Download archive
        $archivePath = Join-Path $tempDir $archiveName
        Download-File -Url $downloadUrl -Output $archivePath
        
        # Verify checksum
        Verify-Checksum -FilePath $archivePath -ChecksumsUrl $checksumsUrl
        
        # Extract archive
        Write-Info "Extracting archive..."
        Expand-Archive -Path $archivePath -DestinationPath $tempDir -Force
        
        # Find the binary
        $binaryPath = Join-Path $tempDir "$BinaryName.exe"
        if (-not (Test-Path $binaryPath)) {
            # Try without .exe extension
            $binaryPath = Join-Path $tempDir $BinaryName
            if (-not (Test-Path $binaryPath)) {
                Write-Error "Binary $BinaryName not found in archive"
                exit 1
            }
        }
        
        # Create install directory
        if (-not (Test-Path $InstallDir)) {
            Write-Info "Creating directory: $InstallDir"
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        }
        
        # Install binary
        $targetPath = Join-Path $InstallDir "$BinaryName.exe"
        Write-Info "Installing $BinaryName to $InstallDir..."
        
        # Stop if binary is running
        $process = Get-Process -Name $BinaryName -ErrorAction SilentlyContinue
        if ($process) {
            Write-Warning "Stopping running $BinaryName process..."
            Stop-Process -Name $BinaryName -Force
            Start-Sleep -Seconds 1
        }
        
        # Copy binary
        Copy-Item -Path $binaryPath -Destination $targetPath -Force
        
        # Verify installation
        if (Test-Path $targetPath) {
            Write-Success "Installation successful! ✓"
            Write-Host ""
            
            # Add to PATH
            Add-ToPath -Directory $InstallDir
            
            # Show version
            Write-Info "Installed version:"
            & $targetPath --version
            
            Write-Host ""
            Write-Info "Get started with:"
            Write-Host "  $BinaryName --help"
            Write-Host "  $BinaryName auth login"
            
            # Install shell completions (optional)
            Write-Host ""
            Write-Info "To enable PowerShell completions, run:"
            Write-Host "  $BinaryName completion powershell | Out-String | Invoke-Expression"
            Write-Host ""
            Write-Host "To make completions persistent, add the above line to your PowerShell profile:"
            Write-Host "  notepad `$PROFILE"
        }
        else {
            Write-Error "Installation failed"
            exit 1
        }
    }
    finally {
        Pop-Location
        Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

# Check if running as administrator (not required, but show warning if needed)
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")

if (-not $isAdmin -and $InstallDir -like "$env:ProgramFiles*") {
    Write-Warning "Installing to $InstallDir requires administrator privileges"
    Write-Warning "Run this script as administrator or choose a different install directory"
    exit 1
}

# Run installation
try {
    Install-ClickUpCLI
}
catch {
    Write-Error "Installation failed: $_"
    exit 1
}