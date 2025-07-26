# Simple Startup Script for GoopBot
# This creates a startup shortcut and background runner

param(
    [Parameter(Mandatory=$false)]
    [string]$Action = "install"
)

$StartupFolder = [System.Environment]::GetFolderPath('Startup')
$ShortcutPath = Join-Path $StartupFolder "GoopBot.lnk"
$BotDirectory = $PSScriptRoot
$StartScript = Join-Path $BotDirectory "start-goopbot-background.ps1"

function Create-BackgroundRunner {
    # Create a script that runs GoopBot in background
    $backgroundScript = @"
# GoopBot Background Runner
# This script runs GoopBot in the background without showing a console window

Set-Location "$BotDirectory"

# Load environment variables
if (Test-Path ".env") {
    Get-Content .env | ForEach-Object {
        if (`$_ -match "^([^#=]+)=(.*)$") {
            `$name = `$matches[1].Trim()
            `$value = `$matches[2].Trim()
            `$value = `$value -replace '^"(.*)"$', '`$1'
            `$value = `$value -replace "^'(.*)'$", '`$1'
            [System.Environment]::SetEnvironmentVariable(`$name, `$value, "Process")
        }
    }
}

# Start GoopBot hidden (no console window)
Start-Process -FilePath ".\GoopBot.exe" -WindowStyle Hidden -WorkingDirectory "$BotDirectory"
"@

    Set-Content -Path $StartScript -Value $backgroundScript -Encoding UTF8
    Write-Host "✅ Created background runner script: $StartScript" -ForegroundColor Green
}

function Install-StartupShortcut {
    Create-BackgroundRunner
    
    # Create startup shortcut
    $WScriptShell = New-Object -ComObject WScript.Shell
    $Shortcut = $WScriptShell.CreateShortcut($ShortcutPath)
    $Shortcut.TargetPath = "powershell.exe"
    $Shortcut.Arguments = "-WindowStyle Hidden -ExecutionPolicy Bypass -File `"$StartScript`""
    $Shortcut.WorkingDirectory = $BotDirectory
    $Shortcut.IconLocation = "shell32.dll,25"  # Robot icon
    $Shortcut.Description = "Start GoopBot in background"
    $Shortcut.Save()
    
    Write-Host "✅ Startup shortcut created!" -ForegroundColor Green
    Write-Host "Location: $ShortcutPath" -ForegroundColor Cyan
    Write-Host "GoopBot will now start automatically when you log in to Windows" -ForegroundColor Green
}

function Remove-StartupShortcut {
    if (Test-Path $ShortcutPath) {
        Remove-Item $ShortcutPath -Force
        Write-Host "✅ Startup shortcut removed" -ForegroundColor Green
    } else {
        Write-Host "No startup shortcut found" -ForegroundColor Yellow
    }
    
    if (Test-Path $StartScript) {
        Remove-Item $StartScript -Force
        Write-Host "✅ Background runner script removed" -ForegroundColor Green
    }
}

function Start-GoopBotBackground {
    & $StartScript
    Write-Host "✅ GoopBot started in background" -ForegroundColor Green
    Write-Host "Check Task Manager to see if GoopBot.exe is running" -ForegroundColor Cyan
}

function Stop-GoopBot {
    $processes = Get-Process -Name "GoopBot" -ErrorAction SilentlyContinue
    if ($processes) {
        $processes | Stop-Process -Force
        Write-Host "✅ GoopBot stopped ($($processes.Count) process(es) terminated)" -ForegroundColor Green
    } else {
        Write-Host "GoopBot is not currently running" -ForegroundColor Yellow
    }
}

function Show-Status {
    $processes = Get-Process -Name "GoopBot" -ErrorAction SilentlyContinue
    if ($processes) {
        Write-Host "✅ GoopBot is running" -ForegroundColor Green
        Write-Host "Process ID(s): $($processes.Id -join ', ')" -ForegroundColor Cyan
        Write-Host "Started: $($processes[0].StartTime)" -ForegroundColor Cyan
    } else {
        Write-Host "❌ GoopBot is not running" -ForegroundColor Red
    }
    
    if (Test-Path $ShortcutPath) {
        Write-Host "✅ Startup shortcut is installed" -ForegroundColor Green
    } else {
        Write-Host "❌ Startup shortcut is not installed" -ForegroundColor Yellow
    }
}

# Main execution
switch ($Action.ToLower()) {
    "install" {
        Install-StartupShortcut
        Write-Host ""
        Write-Host "Want to start GoopBot now? Run: .\startup-helper.ps1 start" -ForegroundColor Cyan
    }
    "start" {
        Start-GoopBotBackground
    }
    "stop" {
        Stop-GoopBot
    }
    "restart" {
        Stop-GoopBot
        Start-Sleep 2
        Start-GoopBotBackground
    }
    "remove" {
        Stop-GoopBot
        Remove-StartupShortcut
    }
    "status" {
        Show-Status
    }
    default {
        Write-Host "Usage: .\startup-helper.ps1 [install|start|stop|restart|remove|status]" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "Commands:" -ForegroundColor Yellow
        Write-Host "  install  - Create startup shortcut for auto-start"
        Write-Host "  start    - Start GoopBot in background now"
        Write-Host "  stop     - Stop GoopBot"
        Write-Host "  restart  - Restart GoopBot"
        Write-Host "  remove   - Remove startup shortcut and stop bot"
        Write-Host "  status   - Show current status"
    }
}
