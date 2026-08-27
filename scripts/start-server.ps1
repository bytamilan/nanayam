# Nanayam - Start Server (Windows)
#
# Brings up the server-side stack (Fabric network + gateway, no client
# apps) with a single command, same as scripts/start-server.sh on
# Mac/Linux.
#
# Why this just delegates to the .sh script rather than reimplementing it:
# Hyperledger Fabric's peer/orderer/CA images are Linux-only, so Docker
# Desktop on Windows always runs them through its WSL2 backend regardless
# of what orchestrates the `docker` calls. A from-scratch PowerShell port of
# start-server.sh's docker-exec choreography would still depend on that same
# Linux layer, just via a second, harder-to-keep-in-sync implementation of
# the same commands - not less of a Windows dependency, only more code to
# maintain. So: find a bash (WSL, or Git for Windows' bash.exe, both a
# one-time install alongside Docker Desktop) and hand off to it.
#
# Usage:
#   .\scripts\start-server.ps1
#   .\scripts\start-server.ps1 --down
#   .\scripts\start-server.ps1 --clean

$ErrorActionPreference = "Stop"

function Test-DockerRunning {
    # $ErrorActionPreference only governs cmdlet errors, not the exit code
    # of a native executable like docker.exe, so check $LASTEXITCODE rather
    # than relying on try/catch here.
    docker info *> $null
    return ($LASTEXITCODE -eq 0)
}

function Find-Bash {
    # Prefer Git for Windows' bash.exe: it runs directly against the
    # Windows filesystem, so no path translation is needed.
    $gitBash = Get-Command bash.exe -ErrorAction SilentlyContinue
    if ($gitBash) { return @{ Kind = "bash"; Path = $gitBash.Source } }

    $gitBashDefault = "$env:ProgramFiles\Git\bin\bash.exe"
    if (Test-Path $gitBashDefault) { return @{ Kind = "bash"; Path = $gitBashDefault } }

    # Fall back to WSL if it's installed and has at least one distro.
    $wsl = Get-Command wsl.exe -ErrorAction SilentlyContinue
    if ($wsl) {
        wsl.exe -l -q *> $null
        if ($LASTEXITCODE -eq 0) { return @{ Kind = "wsl"; Path = $wsl.Source } }
    }

    return $null
}

Write-Host "===============================================" -ForegroundColor Blue
Write-Host "  Nanayam - Start Server" -ForegroundColor Blue
Write-Host "===============================================" -ForegroundColor Blue
Write-Host ""

if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Write-Error "Docker is required. Install Docker Desktop: https://docs.docker.com/desktop/install/windows-install/"
    exit 1
}

if (-not (Test-DockerRunning)) {
    Write-Error "Docker doesn't seem to be running. Start Docker Desktop and try again."
    exit 1
}

$bash = Find-Bash
if (-not $bash) {
    Write-Error @"
No bash interpreter found. Install one of:
  - Git for Windows (ships bash.exe): https://git-scm.com/download/win
  - WSL2 (recommended if you'll do more Fabric/Linux-container work): wsl --install

Docker Desktop on Windows requires one of these anyway to run Linux
containers, so this isn't an extra dependency for this script specifically.
"@
    exit 1
}

$repoRoot = Split-Path -Parent $PSScriptRoot

if ($bash.Kind -eq "wsl") {
    Push-Location $repoRoot
    try {
        & wsl.exe bash ./scripts/start-server.sh @args
    } finally {
        Pop-Location
    }
} else {
    # Forward slashes throughout, to sidestep bash treating a Windows
    # backslash path as escape sequences.
    $scriptPath = (Join-Path $repoRoot "scripts/start-server.sh") -replace '\\', '/'
    & $bash.Path $scriptPath @args
}

exit $LASTEXITCODE
