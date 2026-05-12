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
$ReleaseBaseUrl = if ($env:NANAYAM_RELEASE_BASE_URL) { $env:NANAYAM_RELEASE_BASE_URL.TrimEnd('/') } else { "https://github.com/$Repo/releases/download" }

$WithFabric = $false
$Setup = $false
$Version = "latest"
$Refresh = $false
$DevLocal = $false
$Source = ""

for ($i = 0; $i -lt $args.Length; $i++) {
    switch ($args[$i]) {
        "--with-fabric" { $WithFabric = $true }
        "--setup" { $Setup = $true }
        "--refresh" { $Refresh = $true }
        "--dev-local" { $DevLocal = $true }
        "--version" {
            if ($i + 1 -ge $args.Length) { throw "Missing value for --version" }
            $Version = $args[$i + 1]
            $i++
        }
        "--source" {
            if ($i + 1 -ge $args.Length) { throw "Missing value for --source" }
            $Source = $args[$i + 1]
            $i++
        }
    }
}

function Detect-Platform {
    $arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { throw "Only 64-bit Windows is supported" }
    return @{ Os = "windows"; Arch = $arch }
}

function Resolve-Version {
    $tag = $Version
    if ($tag -eq "latest") {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
        $tag = $release.tag_name
    } elseif (-not $tag.StartsWith("v")) {
        $tag = "v$tag"
    }
    return $tag
}

function Get-AssetName($tag, $platform) {
    return "nanayam_${tag}_$($platform.Os)_$($platform.Arch).zip"
}

function Get-ReleaseUrl($tag, $asset) {
    return "$ReleaseBaseUrl/$tag/$asset"
}

function Get-InstalledVersion($binaryPath) {
    if (Test-Path $binaryPath) {
        $versionLine = & $binaryPath version 2>$null | Select-Object -First 1
        if ($versionLine -match '^nanayam version\s+(\S+)') {
            return $Matches[1]
        }
    }
    return $null
}

function Download-Binary($platform, $dest, $tag) {
    $asset = Get-AssetName $tag $platform
    $url = Get-ReleaseUrl $tag $asset
    $tmpDir = Join-Path $env:TEMP ([System.Guid]::NewGuid().ToString())
    $archivePath = Join-Path $tmpDir $asset
    $extractDir = Join-Path $tmpDir "extract"
    New-Item -ItemType Directory -Force -Path $tmpDir, $extractDir | Out-Null

    Write-Host "[INFO] Downloading $BinaryName $tag for $($platform.Os)/$($platform.Arch)..."
    Write-Host "[INFO]   -> $url"
    Invoke-WebRequest -Uri $url -OutFile $archivePath -UseBasicParsing
    Expand-Archive -Path $archivePath -DestinationPath $extractDir -Force
    Move-Item -Force (Join-Path $extractDir $BinaryName) $dest
    Write-Host "[OK] Binary downloaded" -ForegroundColor Green
}

function Resolve-LocalSource {
    $candidate = if ($Source) { (Resolve-Path $Source).Path } else { (Get-Location).Path }
    if (Test-Path (Join-Path $candidate "cli\go.mod")) {
        return $candidate
    }
    if ((Split-Path $candidate -Leaf) -eq "cli" -and (Test-Path (Join-Path $candidate "go.mod"))) {
        return Split-Path $candidate -Parent
    }
    throw "Could not find a Nanayam repository. Use --source C:\path\to\nanayam with --dev-local."
}

function Build-LocalBinary($dest, $repoRoot) {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        throw "Go is required for --dev-local installs"
    }

    $buildVersion = (git -C $repoRoot describe --tags --always --dirty 2>$null)
    if (-not $buildVersion) { $buildVersion = "dev-local" }
    $buildCommit = (git -C $repoRoot rev-parse --short HEAD 2>$null)
    if (-not $buildCommit) { $buildCommit = "local" }
    $buildDate = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

    Write-Host "[INFO] Building local Nanayam CLI from $repoRoot..."
    Push-Location (Join-Path $repoRoot "cli")
    try {
        & go build -ldflags "-X github.com/bytamilan/nanayam/cli/cmd.version=$buildVersion -X github.com/bytamilan/nanayam/cli/cmd.commit=$buildCommit -X github.com/bytamilan/nanayam/cli/cmd.date=$buildDate" -o $dest .
    }
    finally {
        Pop-Location
    }
    Write-Host "[OK] Local build installed" -ForegroundColor Green
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
Write-Host "[INFO] Detected platform: $($platform.Os)/$($platform.Arch)"

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
$binaryPath = "$InstallDir\$BinaryName"

if ($DevLocal) {
    $repoRoot = Resolve-LocalSource
    Build-LocalBinary $binaryPath $repoRoot
}
else {
    $tag = Resolve-Version
    $installedVersion = Get-InstalledVersion $binaryPath
    if ($installedVersion -and $installedVersion -eq $tag -and -not $Refresh) {
        Write-Host "[OK] Nanayam $tag is already installed. Use --refresh to reinstall it." -ForegroundColor Green
    }
    else {
        Download-Binary $platform $binaryPath $tag
    }
}

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
Write-Host "Upgrade check: nanayam upgrade --check"
Write-Host "Help: nanayam --help"
