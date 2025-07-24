# GoopBot Startup Script
# This script loads environment variables from .env and starts the bot

Write-Host "Loading environment variables from .env file..." -ForegroundColor Cyan

# Load .env file
if (Test-Path ".env") {
    Get-Content .env | ForEach-Object {
        if ($_ -match "^([^#=]+)=(.*)$") {
            $name = $matches[1].Trim()
            $value = $matches[2].Trim()
            # Remove quotes if present
            $value = $value -replace '^"(.*)"$', '$1'
            $value = $value -replace "^'(.*)'$", '$1'
            [System.Environment]::SetEnvironmentVariable($name, $value, "Process")
            Write-Host "  $name = $value" -ForegroundColor Green
        }
    }
} else {
    Write-Host "Error: .env file not found!" -ForegroundColor Red
    exit 1
}

Write-Host "`nStarting GoopBot..." -ForegroundColor Cyan

# Start the bot
.\GoopBot.exe
