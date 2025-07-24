# GoopBot

A Discord bot that automatically notifies your server when users with the "Goop Creator" role go live on Twitch.

## Features

🔴 **Automatic Live Notifications** - Monitors Twitch streams every 5 minutes  
👑 **Role-Based System** - Only "Goop Creator" role holders can link streams  
📺 **Rich Discord Embeds** - Beautiful notifications with stream details  
⚡ **Redis Caching** - Prevents duplicate notifications  
🛡️ **Admin Controls** - Configure channels and manual checks  

## Quick Start

1. **Get Credentials**:
   - [Twitch Developer Console](https://dev.twitch.tv/console) → Client ID & Secret
   - [Discord Developer Portal](https://discord.com/developers/applications) → Bot Token

2. **Configure**:
   ```bash
   cp .env.example .env
   # Edit .env with your credentials
   ```

3. **Build & Run**:
   ```bash
   # Windows (requires C compiler)
   .\build.ps1
   ./GoopBot.exe
   
   # Linux/macOS
   go build
   ./GoopBot
   ```

4. **Setup Discord**:
   ```
   !setnotifications #live-notifications
   !linktwitch your_twitch_username
   ```

## Commands

- `!linktwitch <username>` - Link Twitch account (Goop Creator role)
- `!setnotifications #channel` - Set notification channel (Admin)
- `!gooplive` - Show currently live creators
- `!help` - Show all commands

## Documentation

- **[SETUP.md](SETUP.md)** - Detailed setup instructions
- **[USAGE.md](USAGE.md)** - Complete feature guide
- **[.env.example](.env.example)** - Configuration template

## Testing

Test Twitch API integration:
```bash
go build -o test_twitch.exe cmd/test_twitch/main.go
./test_twitch.exe
```

Perfect for gaming communities and content creator groups! 🎮
