@echo off
REM Nanayam CLI Installer for Windows CMD
REM
REM Usage:
REM   curl -fsSL https://raw.githubusercontent.com/bytamilan/nanayam/main/install.cmd -o install.cmd && install.cmd

echo ===============================================
echo   Nanayam CLI Installer
echo ===============================================
echo.
echo This installer requires PowerShell.
echo Redirecting to PowerShell installer...
echo.

powershell -NoProfile -ExecutionPolicy Bypass -Command "$tmp = Join-Path $env:TEMP 'nanayam-install.ps1'; Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/bytamilan/nanayam/main/install.ps1' -OutFile $tmp; & $tmp %*"
