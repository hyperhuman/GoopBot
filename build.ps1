#!/usr/bin/env powershell
# Build script for GoopBot with CGO support

Write-Host "Setting up build environment..." -ForegroundColor Green

# Add MSYS2 GCC to PATH
$env:PATH = "C:\msys64\mingw64\bin;$env:PATH"

# Enable CGO for SQLite support
$env:CGO_ENABLED = 1

Write-Host "Building GoopBot..." -ForegroundColor Yellow
go build

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Build successful! Run './GoopBot.exe' to start the bot." -ForegroundColor Green
} else {
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}
