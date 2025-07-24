# Discord Bot Troubleshooting Guide

## Bot Setup Checklist

### 1. Discord Developer Portal Settings
Make sure your bot has these settings enabled in the Discord Developer Portal:
- **MESSAGE CONTENT INTENT** ✅ (Required for reading message content)
- **SERVER MEMBERS INTENT** ✅ (Recommended for role checking)
- **GUILD MESSAGES INTENT** ✅ (Required for receiving messages)

### 2. Bot Permissions in Discord Server
Your bot needs these permissions in your Discord server:
- **Read Messages** ✅
- **Send Messages** ✅
- **Read Message History** ✅
- **Use Slash Commands** ✅
- **Embed Links** ✅
- **Attach Files** ✅
- **Manage Roles** (if you want the bot to assign roles)

### 3. Bot Invite URL
Use this URL format to invite your bot with proper permissions:
```
https://discord.com/api/oauth2/authorize?client_id=YOUR_BOT_CLIENT_ID&permissions=412384383040&scope=bot
```
Replace `YOUR_BOT_CLIENT_ID` with your actual bot's client ID.

### 4. Testing Commands
Try these commands in your Discord server:
- `!help` - Shows available commands
- `!linktwitch your_twitch_username` - Links your Twitch account
- `!setnotifications` - Sets current channel for notifications (Admin only)

### 5. Common Issues

**Bot doesn't respond to commands:**
- Check if bot is online (green dot next to bot name)
- Verify MESSAGE CONTENT INTENT is enabled in Developer Portal
- Make sure bot has Read Messages permission in the channel
- Commands are case-sensitive and must start with `!`

**"Missing client ID" error:**
- Make sure .env file has correct TWITCH_CLIENT_ID
- Restart bot using `.\start-bot.ps1` to reload environment variables

**Permission errors:**
- Admin commands require Administrator permission in Discord
- Make sure you have the "Goop Creator" role created in your server

### 6. Bot Status Check
You can check if the bot is working by:
1. Look for the bot in your server member list
2. Check if it has a green "Online" status
3. Try the `!help` command first

### 7. Logs
Check the PowerShell terminal where the bot is running for error messages.
The bot will log when commands are received and processed.
