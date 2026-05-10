@echo off
REM Nanayam CLI Installer for Windows CMD
REM
REM Usage:
REM   curl -fsSL https://nanayam.io/install.cmd -o install.cmd && install.cmd

echo ===============================================
echo   Nanayam CLI Installer
echo ===============================================
echo.
echo This installer requires PowerShell.
echo Redirecting to PowerShell installer...
echo.

powershell -Command "& {irm https://nanayam.io/install.ps1 | iex}"
