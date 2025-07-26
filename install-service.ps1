# GoopBot Windows Service Installer
# This script installs GoopBot as a Windows Service using NSSM (Non-Sucking Service Manager)

param(
    [Parameter(Mandatory=$false)]
    [string]$Action = "install"
)

$ServiceName = "GoopBot"
$ServiceDisplayName = "GoopBot Discord Stream Notifications"
$ServiceDescription = "Discord bot that monitors Twitch streams and sends birthday notifications"
$BotPath = Join-Path $PSScriptRoot "GoopBot.exe"
$LogPath = Join-Path $PSScriptRoot "logs"
$NSSMPath = Join-Path $PSScriptRoot "nssm.exe"

# Create logs directory if it doesn't exist
if (!(Test-Path $LogPath)) {
    New-Item -ItemType Directory -Path $LogPath -Force
}

function Install-NSSM {
    if (!(Test-Path $NSSMPath)) {
        Write-Host "Downloading NSSM (Non-Sucking Service Manager)..." -ForegroundColor Yellow
        
        # Download NSSM
        $nssmUrl = "https://nssm.cc/release/nssm-2.24.zip"
        $nssmZip = Join-Path $env:TEMP "nssm.zip"
        $nssmExtract = Join-Path $env:TEMP "nssm"
        
        try {
            Invoke-WebRequest -Uri $nssmUrl -OutFile $nssmZip
            Expand-Archive -Path $nssmZip -DestinationPath $nssmExtract -Force
            
            # Copy the appropriate NSSM executable
            if ([Environment]::Is64BitOperatingSystem) {
                Copy-Item (Join-Path $nssmExtract "nssm-2.24\win64\nssm.exe") $NSSMPath
            } else {
                Copy-Item (Join-Path $nssmExtract "nssm-2.24\win32\nssm.exe") $NSSMPath
            }
            
            Remove-Item $nssmZip -Force
            Remove-Item $nssmExtract -Recurse -Force
            
            Write-Host "✅ NSSM downloaded successfully" -ForegroundColor Green
        } catch {
            Write-Host "❌ Failed to download NSSM: $($_.Exception.Message)" -ForegroundColor Red
            exit 1
        }
    }
}

function Install-Service {
    Write-Host "Installing GoopBot as Windows Service..." -ForegroundColor Yellow
    
    # Check if service already exists
    $existingService = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
    if ($existingService) {
        Write-Host "Service already exists. Removing old service..." -ForegroundColor Yellow
        & $NSSMPath remove $ServiceName confirm
        Start-Sleep 2
    }
    
    # Install the service
    & $NSSMPath install $ServiceName $BotPath
    & $NSSMPath set $ServiceName DisplayName $ServiceDisplayName
    & $NSSMPath set $ServiceName Description $ServiceDescription
    & $NSSMPath set $ServiceName Start SERVICE_AUTO_START
    
    # Set working directory
    & $NSSMPath set $ServiceName AppDirectory $PSScriptRoot
    
    # Set up logging
    & $NSSMPath set $ServiceName AppStdout (Join-Path $LogPath "goopbot.log")
    & $NSSMPath set $ServiceName AppStderr (Join-Path $LogPath "goopbot-error.log")
    & $NSSMPath set $ServiceName AppRotateFiles 1
    & $NSSMPath set $ServiceName AppRotateOnline 1
    & $NSSMPath set $ServiceName AppRotateSeconds 86400  # Rotate daily
    & $NSSMPath set $ServiceName AppRotateBytes 10485760  # 10MB max
    
    # Set environment variables from .env file
    if (Test-Path ".env") {
        Write-Host "Loading environment variables from .env..." -ForegroundColor Cyan
        Get-Content .env | ForEach-Object {
            if ($_ -match "^([^#=]+)=(.*)$") {
                $name = $matches[1].Trim()
                $value = $matches[2].Trim()
                $value = $value -replace '^"(.*)"$', '$1'
                $value = $value -replace "^'(.*)'$", '$1'
                & $NSSMPath set $ServiceName AppEnvironmentExtra "$name=$value"
                Write-Host "  Set $name" -ForegroundColor Green
            }
        }
    }
    
    Write-Host "✅ Service installed successfully!" -ForegroundColor Green
    Write-Host "Service Name: $ServiceName" -ForegroundColor Cyan
    Write-Host "Display Name: $ServiceDisplayName" -ForegroundColor Cyan
    Write-Host "Logs: $LogPath" -ForegroundColor Cyan
}

function Start-ServiceNow {
    Write-Host "Starting GoopBot service..." -ForegroundColor Yellow
    Start-Service -Name $ServiceName
    Start-Sleep 3
    
    $service = Get-Service -Name $ServiceName
    if ($service.Status -eq "Running") {
        Write-Host "✅ GoopBot service is running!" -ForegroundColor Green
    } else {
        Write-Host "❌ Failed to start service. Status: $($service.Status)" -ForegroundColor Red
        Write-Host "Check logs in: $LogPath" -ForegroundColor Yellow
    }
}

function Remove-Service {
    Write-Host "Removing GoopBot service..." -ForegroundColor Yellow
    
    # Stop service if running
    $service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
    if ($service -and $service.Status -eq "Running") {
        Stop-Service -Name $ServiceName -Force
        Start-Sleep 3
    }
    
    # Remove service
    & $NSSMPath remove $ServiceName confirm
    Write-Host "✅ Service removed successfully!" -ForegroundColor Green
}

function Show-Status {
    $service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
    if ($service) {
        Write-Host "GoopBot Service Status: $($service.Status)" -ForegroundColor Cyan
        Write-Host "Display Name: $($service.DisplayName)" -ForegroundColor Cyan
        Write-Host "Start Type: $($service.StartType)" -ForegroundColor Cyan
    } else {
        Write-Host "GoopBot service is not installed" -ForegroundColor Yellow
    }
}

# Check if running as administrator
if (-NOT ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Host "❌ This script must be run as Administrator!" -ForegroundColor Red
    Write-Host "Right-click PowerShell and select 'Run as Administrator'" -ForegroundColor Yellow
    pause
    exit 1
}

# Main execution
switch ($Action.ToLower()) {
    "install" {
        Install-NSSM
        Install-Service
        Start-ServiceNow
    }
    "start" {
        Start-ServiceNow
    }
    "stop" {
        Stop-Service -Name $ServiceName -Force
        Write-Host "✅ Service stopped" -ForegroundColor Green
    }
    "restart" {
        Stop-Service -Name $ServiceName -Force
        Start-Sleep 2
        Start-ServiceNow
    }
    "remove" {
        Remove-Service
    }
    "status" {
        Show-Status
    }
    "logs" {
        if (Test-Path (Join-Path $LogPath "goopbot.log")) {
            Get-Content (Join-Path $LogPath "goopbot.log") -Tail 50
        } else {
            Write-Host "No logs found" -ForegroundColor Yellow
        }
    }
    default {
        Write-Host "Usage: .\install-service.ps1 [install|start|stop|restart|remove|status|logs]" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "Commands:" -ForegroundColor Yellow
        Write-Host "  install  - Install and start GoopBot as Windows Service"
        Write-Host "  start    - Start the service"
        Write-Host "  stop     - Stop the service"
        Write-Host "  restart  - Restart the service"
        Write-Host "  remove   - Remove the service"
        Write-Host "  status   - Show service status"
        Write-Host "  logs     - Show recent logs"
    }
}
