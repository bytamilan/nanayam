@echo off
REM Nanayam - Start Server (Windows CMD)
REM
REM Thin wrapper: this needs PowerShell, so it just re-launches
REM start-server.ps1 from this same scripts\ directory.
REM
REM Usage:
REM   scripts\start-server.cmd
REM   scripts\start-server.cmd --down
REM   scripts\start-server.cmd --clean

powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0start-server.ps1" %*
