# Nanayam CLI Installer for Windows PowerShell
#
# Usage:
#   irm https://raw.githubusercontent.com/bytamilan/nanayam/main/install.ps1 | iex
#   irm https://raw.githubusercontent.com/bytamilan/nanayam/main/install.ps1 | iex -Command { & $args[0] --with-fabric }

$ErrorActionPreference = "Stop"

$Repo = "bytamilan/nanayam"
$BinaryName = "nanayam.exe"
$InstallDir = "$env:USERPROFILE\.nanayam\bin"
$FabricBinDir = "$env:USERPROFILE\.nanayam\fabric-bin"

$WithFabric = $false
$Setup = $false
$Version = "latest"

foreach ($arg in $args) {
    switch ($arg) {
        "--with-fabric" { $WithFabric = $true }
        "--setup" { $Setup = $true }
    }
}

function Detect-Platform {
    $arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
    return "windows-$arch"
}

function Download-Binary($platform, $dest) {
    $tag = $Version
    if ($tag -eq "latest") {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
        $tag = $release.tag_name
    }
    $url = "https://github.com/$Repo/releases/download/$tag/${BinaryName}-${platform}"
    Write-Host "[INFO] Downloading $BinaryName $tag for $platform..."
    Invoke-WebRequest -Uri $url -OutFile $dest -UseBasicParsing
    Write-Host "[OK] Binary downloaded" -ForegroundColor Green
}

function Add-ToPath {
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if (-not $currentPath.Contains(".nanayam\bin")) {
        [Environment]::SetEnvironmentVariable("Path", "$currentPath;$InstallDir", "User")
        Write-Host "[OK] Added $InstallDir to user PATH" -ForegroundColor Green
        Write-Host "[WARN] Restart your terminal to apply PATH changes" -ForegroundColor Yellow
    }
}

Write-Host "===============================================" -ForegroundColor Blue
Write-Host "  Nanayam CLI Installer" -ForegroundColor Blue
Write-Host "===============================================" -ForegroundColor Blue
Write-Host ""

$platform = Detect-Platform
Write-Host "[INFO] Detected platform: $platform"

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
$binaryPath = "$InstallDir\$BinaryName"

Download-Binary $platform $binaryPath
Add-ToPath

if ($WithFabric) {
    Write-Host ""
    Write-Host "[INFO] Fabric binary download for Windows is not yet automated."
    Write-Host "       Please download manually from https://github.com/hyperledger/fabric/releases"
}

if ($Setup) {
    Write-Host ""
    Write-Host "[INFO] Running prerequisites check..."
    & $binaryPath prerequisites --auto
}

Write-Host ""
Write-Host "===============================================" -ForegroundColor Green
Write-Host "  Installation Complete!" -ForegroundColor Green
Write-Host "===============================================" -ForegroundColor Green
Write-Host ""
Write-Host "Run: nanayam version"
Write-Host "Help: nanayam --help"
