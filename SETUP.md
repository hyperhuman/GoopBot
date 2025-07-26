# GoopBot Setup Guide

## Prerequisites

1. **Go 1.21+** installed
2. **Redis** server running (optional, but recommended for caching)
3. **Discord Bot** token
4. **Twitch API** credentials

## Step 1: Get Twitch API Credentials

1. Go to [Twitch Developer Console](https://dev.twitch.tv/console)
2. Click "Register Your Application"
3. Fill in:
   - **Name**: GoopBot (or your preferred name)
   - **OAuth Redirect URLs**: `http://localhost` (not used, but required)
   - **Category**: Application Integration
4. Click "Create"
5. Copy your **Client ID** and **Client Secret**

## Step 2: Create Discord Bot

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Click "New Application"
3. Give it a name (e.g., "GoopBot")
4. Go to the "Bot" section
5. Click "Add Bot"
6. Copy the **Bot Token**
7. Enable these intents:
   - Server Members Intent
   - Message Content Intent

## Step 3: Configure Environment

1. Copy `.env.example` to `.env`:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` with your credentials:
   ```env
   DISCORD_TOKEN=your_actual_discord_bot_token
   REDIS_ADDR=localhost:6379
   TWITCH_CLIENT_ID=your_actual_twitch_client_id
   TWITCH_CLIENT_SECRET=your_actual_twitch_client_secret
   ```

## Step 4: Install Redis (Optional but Recommended)

### Windows:
1. Download Redis from [GitHub releases](https://github.com/tporadowski/redis/releases)
2. Install and run `redis-server.exe`

### macOS:
```bash
brew install redis
brew services start redis
```

### Linux:
```bash
sudo apt-get install redis-server
sudo systemctl start redis
```

## Step 5: Install C Compiler (Required for SQLite)

GoopBot uses SQLite which requires CGO (C bindings). You need a C compiler installed:

### Windows:

**Option 1: Use the build script (Recommended)**
```bash
# If you have MSYS2 installed
.\build.ps1
```

**Option 2: Install MSYS2 manually**
```bash
# Install MSYS2
winget install --id=MSYS2.MSYS2

# Install GCC
C:\msys64\usr\bin\bash.exe -lc "pacman -S mingw-w64-x86_64-gcc --noconfirm"

# Build with CGO
$env:PATH = "C:\msys64\mingw64\bin;$env:PATH"
$env:CGO_ENABLED = 1
go build
```

**Option 3: Install TDM-GCC**
- Download from: https://jmeubank.github.io/tdm-gcc/
- Install and add to PATH

### macOS:
```bash
# Xcode Command Line Tools (if not already installed)
xcode-select --install
```

### Linux:
```bash
# Ubuntu/Debian
sudo apt-get install build-essential

# CentOS/RHEL
sudo yum groupinstall "Development Tools"
```

## Step 6: Build and Run

1. Build the bot:
   ```bash
   # Windows (with MSYS2)
   .\build.ps1
   
   # Or manually
   go build  # (after setting up C compiler)
   ```

2. Run the bot:
   ```bash
   ./GoopBot.exe  # Windows
   # or
   ./GoopBot      # Linux/macOS
   ```

3. (Optional) Test Twitch API integration:
   ```bash
   # Set environment variables first
   export TWITCH_CLIENT_ID=your_client_id
   export TWITCH_CLIENT_SECRET=your_client_secret
   
   # Build and run test utility
   go build -o test_twitch.exe cmd/test_twitch/main.go
   ./test_twitch.exe
   ```

## Step 6: Discord Server Setup

1. **Invite the bot** to your Discord server with these permissions:
   - Send Messages
   - Use Slash Commands
   - Read Message History
   - View Channels
   - Embed Links

2. **Create the required roles** in your Discord server:
   - **"Goop Creator"** role (for streamers)
   - **"member"** role (for birthday feature - exact name, all lowercase)

3. **Set up notification channels**:
   ```
   !setnotifications #live-notifications
   !setbirthdaychannel #birthdays
   ```

## Step 7: Link Streamers

Users with the "Goop Creator" role can link their Twitch accounts:
```
!linktwitch their_twitch_username
```

## Commands Reference

### For Goop Creators:
- `!linktwitch <username>` - Link your Twitch account
- `!unlinktwitch` - Unlink your Twitch account

### For Members:
- `!setbirthday <MM/DD>` - Set your birthday (e.g., !setbirthday 03/15) - requires "member" role

### For Admins & Server Owners:
- `!setnotifications #channel` - Set live notification channel
- `!setbirthdaychannel #channel` - Set birthday notification channel
- `!checkstreams` - Manually check stream status

### For Everyone:
- `!help` - Show all commands
- `!gooplive` - Show currently live Goop Creators
- `!birthdays` - Show upcoming birthdays

## How It Works

1. **Every 5 minutes**, the bot checks Twitch API for all linked streamers
2. **When someone goes live**, it sends a rich embed notification to the designated channel
3. **Daily at midnight**, the bot checks for birthdays and sends celebration messages
4. **Redis caching** prevents duplicate notifications
5. **Database storage** keeps track of all streamers, birthdays, and their status

## Troubleshooting

### Bot doesn't respond:
- Check if the bot has proper permissions
- Verify the Discord token is correct
- Make sure the bot is online in your server

### Twitch integration not working:
- Verify Twitch Client ID and Secret are correct
- Check bot logs for API errors
- Make sure streamers have linked their accounts correctly

### No notifications appearing:
- Ensure notification channel is set with `!setnotifications`
- Check if users have the "Goop Creator" role
- Verify their Twitch usernames are linked correctly

## Logs

The bot provides detailed logging:
- Stream status checks
- Going live notifications
- API errors
- Database operations

Monitor the console output to debug any issues.

## Production Deployment

For production use:
1. Use a proper Redis instance (not localhost)
2. Set up proper logging to files
3. Use environment variables for all configuration
4. Consider using Docker for deployment
5. Set up monitoring and health checks
