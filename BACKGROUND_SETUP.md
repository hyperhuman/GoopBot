# GoopBot Background & Auto-Start Setup Guide

## 🚀 Method 1: Windows Service (Recommended)

**Best for:** Production use, server environments, always-on operation

### Installation:
1. **Run PowerShell as Administrator** (Right-click → "Run as Administrator")
2. Navigate to GoopBot folder: `cd "d:\MyGit\GoopBot\GoopBot"`
3. Run: `.\install-service.ps1 install`

### What it does:
- ✅ Downloads NSSM (service manager) automatically
- ✅ Installs GoopBot as proper Windows Service  
- ✅ Auto-starts on boot (even before user login)
- ✅ Automatic restart if crashes
- ✅ Proper logging to `logs/` folder
- ✅ Loads environment variables from `.env`

### Management Commands:
```powershell
.\install-service.ps1 start     # Start service
.\install-service.ps1 stop      # Stop service  
.\install-service.ps1 restart   # Restart service
.\install-service.ps1 status    # Check status
.\install-service.ps1 logs      # View recent logs
.\install-service.ps1 remove    # Uninstall service
```

### Service Management:
- **Services App**: Search "Services" → Find "GoopBot Discord Stream Notifications"
- **Task Manager**: Services tab → Look for "GoopBot"
- **Command Line**: `sc query GoopBot` or `Get-Service GoopBot`

---

## 🏠 Method 2: Startup Shortcut (Simple)

**Best for:** Personal use, single-user systems

### Installation:
1. Open PowerShell in GoopBot folder
2. Run: `.\startup-helper.ps1 install`

### What it does:
- ✅ Creates startup shortcut in your Startup folder
- ✅ Starts when you log in to Windows
- ✅ Runs in background (hidden console)
- ✅ Simple to install/remove

### Management Commands:
```powershell
.\startup-helper.ps1 start      # Start now
.\startup-helper.ps1 stop       # Stop GoopBot
.\startup-helper.ps1 restart    # Restart
.\startup-helper.ps1 status     # Check status
.\startup-helper.ps1 remove     # Remove startup shortcut
```

---

## 🔧 Manual Background Start

For immediate testing:
```powershell
# Start hidden (no console window)
Start-Process -FilePath ".\start-bot.ps1" -WindowStyle Hidden

# Or start minimized
Start-Process -FilePath ".\start-bot.ps1" -WindowStyle Minimized
```

---

## 📊 Monitoring Your Bot

### Check if running:
```powershell
Get-Process -Name "GoopBot" -ErrorAction SilentlyContinue
```

### View logs (Service method):
- Location: `d:\MyGit\GoopBot\GoopBot\logs\`
- Files: `goopbot.log` (normal), `goopbot-error.log` (errors)

### Task Manager:
- Look for "GoopBot.exe" in Processes tab
- Check CPU/Memory usage

---

## 🎯 Recommendations

### For Personal Use:
- Use **Startup Shortcut method** - simple and effective
- Starts when you log in
- Easy to manage

### For Server/Production:
- Use **Windows Service method** - most reliable  
- Starts before user login
- Automatic restart on failure
- Proper logging and monitoring

### Testing First:
1. Test bot manually: `.\start-bot.ps1`
2. Verify Discord commands work
3. Then set up auto-start

---

## 🚨 Troubleshooting

### Bot not starting:
1. Check `.env` file has correct credentials
2. Ensure Redis is running: `.\Redis\redis-cli.exe ping`
3. Check logs for errors
4. Test manual start first

### Service issues:
- Run PowerShell as Administrator
- Check Windows Event Viewer for service errors
- Verify NSSM installed correctly

### Startup shortcut issues:
- Check if shortcut exists in Startup folder
- Verify PowerShell execution policy: `Get-ExecutionPolicy`
- May need: `Set-ExecutionPolicy RemoteSigned -Scope CurrentUser`

---

## 💡 Pro Tips

1. **Always test manually first** before setting up auto-start
2. **Monitor logs** especially first few days after setup
3. **Redis must be running** - consider making it a service too
4. **Keep `.env` file secure** - contains sensitive tokens
5. **Regular backups** of `GoopBot.db` database file

Your GoopBot will now run 24/7 and automatically restart with your PC! 🎉
