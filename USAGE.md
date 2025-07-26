# GoopBot - Live Stream Notifications for Goop ### 👥 For Members:
- `!setbirthday <MM/DD>` - Set your birthday (e.g., 03/15) - requires "member" roleeators

## 🎯 What GoopBot Does

GoopBot automatically monitors Twitch streams and sends Discord notifications when users with the "Goop Creator" role go live. Perfect for gaming communities and content creator groups!

## 🚀 Features

✅ **Automatic Stream Monitoring** - Checks Twitch API every 5 minutes  
✅ **Role-Based System** - Only "Goop Creator" role holders can link their streams  
✅ **Rich Notifications** - Beautiful Discord embeds with stream details  
✅ **Birthday Celebrations** - Daily birthday notifications for members  
✅ **Redis Caching** - Prevents duplicate notifications  
✅ **Admin Controls** - Set notification channels and manual checks  
✅ **Real-time Updates** - Instant notifications when creators go live  

## 🔧 Setup Process

### 1. **Get API Credentials**
- **Twitch**: Get Client ID & Secret from [Twitch Developer Console](https://dev.twitch.tv/console)
- **Discord**: Create a bot at [Discord Developer Portal](https://discord.com/developers/applications)

### 2. **Configure Environment**
```bash
cp .env.example .env
# Edit .env with your actual credentials
```

### 3. **Install & Run Redis** (Optional but recommended)
```bash
# Windows: Download from GitHub releases
# macOS: brew install redis && brew services start redis  
# Linux: sudo apt-get install redis-server
```

### 4. **Build & Run**
```bash
go build
./GoopBot.exe  # Windows
```

## 📋 Commands

### 👑 For Goop Creators:
- `!linktwitch <username>` - Link your Twitch account
- `!unlinktwitch` - Unlink your Twitch account

### � For Members:
- `!setbirthday <MM/DD>` - Set your birthday (e.g., 03/15)

### �🛡️ For Admins & Server Owners:
- `!setnotifications #channel` - Set live notification channel
- `!setbirthdaychannel #channel` - Set birthday notification channel
- `!checkstreams` - Manually check stream status

### 🌍 For Everyone:
- `!help` - Show all commands
- `!gooplive` - Show currently live Goop Creators
- `!birthdays` - Show upcoming birthdays

## 🔄 How It Works

1. **Goop Creators link their Twitch accounts** using `!linktwitch`
2. **Members set their birthdays** using `!setbirthday MM/DD`
3. **Admins set notification channels** using `!setnotifications` and `!setbirthdaychannel`
4. **Bot monitors Twitch API** every 5 minutes automatically
5. **Bot checks birthdays** daily at midnight
6. **When someone goes live** → Rich notification sent to channel
7. **When it's someone's birthday** → Celebration message sent
8. **Redis caching** prevents spam notifications

## 📱 Example Workflow

**John (Goop Creator) links his account:**
```
John: !linktwitch johngamer123
Bot: ✅ Successfully linked your Twitch account: johngamer123
```

**Admin sets notification channels:**
```
Admin: !setnotifications #live-notifications  
Bot: ✅ Successfully set #live-notifications as the live notification channel

Admin: !setbirthdaychannel #birthdays
Bot: 🎂 Successfully set birthday notification channel!
```

**Member sets birthday:**
```
Member: !setbirthday 03/15
Bot: 🎂 Successfully set your birthday!
```

**When John goes live on Twitch:**
→ Bot automatically sends rich embed to #live-notifications with:
- John's Discord name and Twitch channel
- Stream title and game being played  
- Viewer count and direct Twitch link
- Live stream thumbnail

**When it's Member's birthday (March 15th):**
→ Bot automatically sends to #birthdays:
"🎉 **Happy Birthday** @Member! 🎂
Hope you have a wonderful day! 🎈"

## 🧪 Testing

Test your Twitch API setup:
```bash
# Set environment variables first
export TWITCH_CLIENT_ID="your_client_id"
export TWITCH_CLIENT_SECRET="your_client_secret"

# Run test
go run cmd/test_twitch/main.go
```

## 🗃️ Database Structure

- **GoopCreator**: Links Discord users (with Goop Creator role) to Twitch usernames
- **TwitchStream**: Tracks live status, viewer count, game, etc.
- **NotificationChannel**: Stores Discord channels for stream notifications
- **Birthday**: Stores user birthdays (month/day)
- **BirthdayChannel**: Stores Discord channels for birthday notifications

## ⚙️ Technical Details

- **Language**: Go 1.21+
- **Database**: SQLite with GORM
- **Cache**: Redis  
- **APIs**: Discord API, Twitch Helix API
- **Monitoring**: 5-minute stream intervals + daily birthday checks
- **Notifications**: Rich Discord embeds with live data + birthday celebrations

## 🔍 Troubleshooting

**Bot doesn't respond?**
- Check Discord bot permissions
- Verify token is correct
- Ensure bot is online

**No Twitch notifications?**
- Verify Twitch API credentials
- Check if streamers linked accounts correctly  
- Ensure notification channel is set

**Multiple notifications?**
- Redis caching should prevent this
- Check Redis connection

## 🚀 Ready for Production

The bot includes:
- Comprehensive error handling
- Detailed logging
- Rate limiting protection
- Efficient batch API calls
- Redis caching for performance
- Database migrations

Perfect for communities of any size! 🎮
