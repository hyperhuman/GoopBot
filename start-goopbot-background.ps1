# GoopBot Background Runner
# This script runs GoopBot in the background without showing a console window

Set-Location "D:\MyGit\GoopBot\GoopBot"

# Load environment variables
if (Test-Path ".env") {
    Get-Content .env | ForEach-Object {
        if ($_ -match "^([^#=]+)=(.*)$") {
            $name = $matches[1].Trim()
            $value = $matches[2].Trim()
            $value = $value -replace '^"(.*)"$', '$1'
            $value = $value -replace "^'(.*)'$", '$1'
            [System.Environment]::SetEnvironmentVariable($name, $value, "Process")
        }
    }
}

# Start GoopBot hidden (no console window)
Start-Process -FilePath ".\GoopBot.exe" -WindowStyle Hidden -WorkingDirectory "D:\MyGit\GoopBot\GoopBot"
