// Package bot provides the Discord bot implementation.
package bot

import (
	"GoopBot/internal/twitch"
	"GoopBot/redis/redisutil"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/bwmarrin/discordgo"
)

// Bot represents our Discord bot instance
type Bot struct {
	discord      *discordgo.Session
	dbConn       *gorm.DB
	redis        *redis.Client
	twitchClient *twitch.Client
}

// GoopCreator represents a Discord user with the "Goop Creator" role (streamers)
type GoopCreator struct {
	gorm.Model
	DiscordID      string `gorm:"uniqueIndex" json:"discord_id"`
	Username       string `json:"username"`
	GuildID        string `json:"guild_id"`
	TwitchUsername string `json:"twitch_username"` // Their Twitch username
	IsActive       bool   `json:"is_active"`       // Whether notifications are enabled
}

// TwitchStream represents a Twitch stream status
type TwitchStream struct {
	gorm.Model
	TwitchUsername string     `gorm:"uniqueIndex" json:"twitch_username"`
	IsLive         bool       `json:"is_live"`
	LastChecked    *time.Time `json:"last_checked"`
	ViewerCount    int        `json:"viewer_count"`
	GameName       string     `json:"game_name"`
	StreamTitle    string     `json:"stream_title"`
	DiscordID      string     `json:"discord_id"` // Associated Discord user ID
}

// NotificationChannel represents channels where live notifications should be sent
type NotificationChannel struct {
	gorm.Model
	GuildID   string `json:"guild_id"`
	ChannelID string `gorm:"uniqueIndex" json:"channel_id"`
	IsActive  bool   `json:"is_active"`
}

// Birthday represents a user's birthday
type Birthday struct {
	gorm.Model
	DiscordID string    `gorm:"uniqueIndex" json:"discord_id"`
	Username  string    `json:"username"`
	GuildID   string    `json:"guild_id"`
	Month     int       `json:"month"`     // 1-12
	Day       int       `json:"day"`       // 1-31
	Year      *int      `json:"year"`      // Optional birth year
	LastSent  time.Time `json:"last_sent"` // When we last sent birthday message
}

// BirthdayChannel represents channels where birthday notifications should be sent
type BirthdayChannel struct {
	gorm.Model
	GuildID   string `json:"guild_id"`
	ChannelID string `gorm:"uniqueIndex" json:"channel_id"`
	IsActive  bool   `json:"is_active"`
}

// RoleMessage represents messages that grant roles when reacted to
type RoleMessage struct {
	gorm.Model
	GuildID   string `json:"guild_id"`
	ChannelID string `json:"channel_id"`
	MessageID string `gorm:"uniqueIndex" json:"message_id"`
	RoleName  string `json:"role_name"` // The role to grant (e.g., "member")
	IsActive  bool   `json:"is_active"`
}

// isUserAdmin checks if a user has admin permissions (either Administrator permission or server owner)
func (b *Bot) isUserAdmin(s *discordgo.Session, userID, channelID string) bool {
	// Get the guild from the channel
	channel, err := s.Channel(channelID)
	if err != nil {
		log.Printf("Error getting channel info: %v", err)
		return false
	}

	// Get guild information
	guild, err := s.Guild(channel.GuildID)
	if err != nil {
		log.Printf("Error getting guild info: %v", err)
		return false
	}

	// Check if user is the server owner
	if guild.OwnerID == userID {
		return true
	}

	// Check if user has Administrator permission
	permissions, err := s.UserChannelPermissions(userID, channelID)
	if err != nil {
		log.Printf("Error getting user permissions: %v", err)
		return false
	}

	return permissions&discordgo.PermissionAdministrator != 0
}

// handleReady handles the ready event when the bot connects
func (b *Bot) handleReady(s *discordgo.Session, r *discordgo.Ready) {
	log.Printf("Logged in as %s", s.State.User.Username)

	// Update status with proper activity type
	activity := &discordgo.Activity{
		Name: "Watching streams...",
		Type: discordgo.ActivityTypeWatching,
	}

	status := &discordgo.UpdateStatusData{
		Status:     "online",
		Activities: []*discordgo.Activity{activity},
	}

	if err := s.UpdateStatusComplex(*status); err != nil {
		log.Printf("Failed to set status: %v", err)
	}
}

// handleCommands handles incoming commands
func (b *Bot) handleCommands(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Get guild member to check roles
	member, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf("Failed to get guild member: %v", err)
		return
	}

	// Check if user has "Goop Creator" role
	hasGoopCreatorRole := b.hasRole(s, m.GuildID, member.Roles, "Goop Creator")

	// Check if user has "Member" role (case-sensitive)
	hasMemberRole := b.hasRole(s, m.GuildID, member.Roles, "member")

	// Handle commands using strings.HasPrefix
	if strings.HasPrefix(m.Content, "!help") {
		helpMessage := `
**Available commands:**
!help - Show this help message
!linktwitch <username> - Link your Twitch username (Goop Creator role required)
!unlinktwitch - Unlink your Twitch username
!setnotifications <channel> - Set notification channel for live streams (Admin only)
!gooplive - Show currently live Goop Creators
!checkstreams - Manually check stream status (Admin only)
!setbirthday <MM/DD> - Set your birthday (member role required)
!setbirthdaychannel <channel> - Set birthday notification channel (Admin only)
!birthdays - Show upcoming birthdays

**Role Management Commands (Admin only):**
!setrolemessage <message_id> [role_name] - Set a message to grant roles when reacted to (default: member)
!removerolemessage <message_id> - Remove role-granting from a message
!listrolemessages - List all active role-granting messages

**Auto-Role Feature:**
• React to designated messages to automatically get roles!
• Admins can set up which messages grant roles using !setrolemessage
        `
		if _, err := s.ChannelMessageSend(m.ChannelID, helpMessage); err != nil {
			log.Printf("Failed to send help message: %v", err)
		}
	} else if strings.HasPrefix(m.Content, "!linktwitch ") {
		if !hasGoopCreatorRole {
			if _, err := s.ChannelMessageSend(m.ChannelID,
				"❌ You need the 'Goop Creator' role to link your Twitch account!"); err != nil {
				log.Printf("Failed to send role error message: %v", err)
			}
			return
		}

		// Extract username from command
		username := strings.TrimSpace(m.Content[12:]) // 12 is length of "!linktwitch "
		if username != "" {
			if err := b.LinkTwitchAccount(m.Author.ID, m.Author.Username, m.GuildID, username); err != nil {
				errorMsg := fmt.Sprintf("❌ Failed to link Twitch account: %v", err)
				if _, err := s.ChannelMessageSend(m.ChannelID, errorMsg); err != nil {
					log.Printf("Failed to send error message: %v", err)
				}
			} else {
				successMsg := fmt.Sprintf("✅ Successfully linked your Twitch account: %s", username)
				if _, err := s.ChannelMessageSend(m.ChannelID, successMsg); err != nil {
					log.Printf("Failed to send success message: %v", err)
				}
			}
		}
	} else if strings.HasPrefix(m.Content, "!unlinktwitch") {
		if !hasGoopCreatorRole {
			if _, err := s.ChannelMessageSend(m.ChannelID,
				"❌ You need the 'Goop Creator' role to manage your Twitch account!"); err != nil {
				log.Printf("Failed to send role error message: %v", err)
			}
			return
		}

		if err := b.UnlinkTwitchAccount(m.Author.ID); err != nil {
			errorMsg := fmt.Sprintf("❌ Failed to unlink Twitch account: %v", err)
			if _, err := s.ChannelMessageSend(m.ChannelID, errorMsg); err != nil {
				log.Printf("Failed to send error message: %v", err)
			}
		} else {
			if _, err := s.ChannelMessageSend(m.ChannelID, "✅ Successfully unlinked your Twitch account"); err != nil {
				log.Printf("Failed to send success message: %v", err)
			}
		}
	} else if strings.HasPrefix(m.Content, "!setnotifications ") {
		// Check if user has admin permissions (including server owner)
		if !b.isUserAdmin(s, m.Author.ID, m.ChannelID) {
			if _, err := s.ChannelMessageSend(m.ChannelID,
				"❌ You need Administrator permissions or server ownership to set notification channels!"); err != nil {
				log.Printf("Failed to send permission error message: %v", err)
			}
			return
		}

		// Extract channel mention from command
		channelMention := strings.TrimSpace(m.Content[17:]) // 17 is length of "!setnotifications "
		channelID := strings.Trim(channelMention, "<>#")

		if channelID != "" {
			if err := b.SetNotificationChannel(m.GuildID, channelID); err != nil {
				errorMsg := fmt.Sprintf("❌ Failed to set notification channel: %v", err)
				if _, err := s.ChannelMessageSend(m.ChannelID, errorMsg); err != nil {
					log.Printf("Failed to send error message: %v", err)
				}
			} else {
				successMsg := fmt.Sprintf("✅ Successfully set <#%s> as the live notification channel", channelID)
				if _, err := s.ChannelMessageSend(m.ChannelID, successMsg); err != nil {
					log.Printf("Failed to send success message: %v", err)
				}
			}
		}
	} else if strings.HasPrefix(m.Content, "!gooplive") {
		// Show currently live Goop Creators
		liveCreators, err := b.GetLiveGoopCreators(m.GuildID)
		if err != nil {
			log.Printf("Failed to get live creators: %v", err)
			return
		}

		if len(liveCreators) == 0 {
			if _, err := s.ChannelMessageSend(m.ChannelID, "No Goop Creators are currently live 😴"); err != nil {
				log.Printf("Failed to send no live streamers message: %v", err)
			}
			return
		}

		message := "**🔴 Currently Live Goop Creators:**\n"
		for _, stream := range liveCreators {
			message += fmt.Sprintf("• **%s** - %s\n", stream.TwitchUsername, stream.StreamTitle)
			message += fmt.Sprintf("  └ Playing: %s | Viewers: %d\n", stream.GameName, stream.ViewerCount)
			message += fmt.Sprintf("  └ https://twitch.tv/%s\n\n", stream.TwitchUsername)
		}

		if _, err := s.ChannelMessageSend(m.ChannelID, message); err != nil {
			log.Printf("Failed to send live streamers message: %v", err)
		}
	} else if strings.HasPrefix(m.Content, "!checkstreams") {
		// Check if user has admin permissions (including server owner)
		if !b.isUserAdmin(s, m.Author.ID, m.ChannelID) {
			if _, err := s.ChannelMessageSend(m.ChannelID,
				"❌ You need Administrator permissions or server ownership to manually check streams!"); err != nil {
				log.Printf("Failed to send permission error message: %v", err)
			}
			return
		}

		if _, err := s.ChannelMessageSend(m.ChannelID, "🔄 Checking stream status..."); err != nil {
			log.Printf("Failed to send checking message: %v", err)
		}

		// Run stream check in background
		go func() {
			b.CheckStreamStatus()
			if _, err := s.ChannelMessageSend(m.ChannelID, "✅ Stream status check completed!"); err != nil {
				log.Printf("Failed to send completion message: %v", err)
			}
		}()
	} else if strings.HasPrefix(m.Content, "!setbirthday ") {
		// Check if user has member role
		if !hasMemberRole {
			if _, err := s.ChannelMessageSend(m.ChannelID,
				"❌ You need the 'member' role to set your birthday!"); err != nil {
				log.Printf("Failed to send role error message: %v", err)
			}
			return
		}

		// Extract birthday from command (MM/DD format)
		birthdayStr := strings.TrimSpace(m.Content[12:]) // 12 is length of "!setbirthday "
		if err := b.SetUserBirthday(m.Author.ID, m.Author.Username, m.GuildID, birthdayStr); err != nil {
			errorMsg := fmt.Sprintf("❌ Failed to set birthday: %v", err)
			if _, err := s.ChannelMessageSend(m.ChannelID, errorMsg); err != nil {
				log.Printf("Failed to send error message: %v", err)
			}
		} else {
			if _, err := s.ChannelMessageSend(m.ChannelID, "🎂 Successfully set your birthday!"); err != nil {
				log.Printf("Failed to send success message: %v", err)
			}
		}
	} else if strings.HasPrefix(m.Content, "!setbirthdaychannel ") {
		// Check if user has admin permissions (including server owner)
		if !b.isUserAdmin(s, m.Author.ID, m.ChannelID) {
			if _, err := s.ChannelMessageSend(m.ChannelID,
				"❌ You need Administrator permissions or server ownership to set birthday channels!"); err != nil {
				log.Printf("Failed to send permission error message: %v", err)
			}
			return
		}

		// Extract channel mention from command
		channelMention := strings.TrimSpace(m.Content[20:]) // 20 is length of "!setbirthdaychannel "
		channelID := strings.Trim(channelMention, "<>#")

		if channelID != "" {
			if err := b.SetBirthdayChannel(m.GuildID, channelID); err != nil {
				errorMsg := fmt.Sprintf("❌ Failed to set birthday channel: %v", err)
				if _, err := s.ChannelMessageSend(m.ChannelID, errorMsg); err != nil {
					log.Printf("Failed to send error message: %v", err)
				}
			} else {
				if _, err := s.ChannelMessageSend(m.ChannelID, "🎂 Successfully set birthday notification channel!"); err != nil {
					log.Printf("Failed to send success message: %v", err)
				}
			}
		} else {
			if _, err := s.ChannelMessageSend(m.ChannelID, "❌ Please mention a valid channel (e.g., #birthdays)"); err != nil {
				log.Printf("Failed to send invalid channel message: %v", err)
			}
		}
	} else if strings.HasPrefix(m.Content, "!birthdays") {
		// Show upcoming birthdays (available to everyone)
		birthdays, err := b.GetUpcomingBirthdays(m.GuildID)
		if err != nil {
			errorMsg := fmt.Sprintf("❌ Failed to get birthdays: %v", err)
			if _, err := s.ChannelMessageSend(m.ChannelID, errorMsg); err != nil {
				log.Printf("Failed to send error message: %v", err)
			}
			return
		}

		if len(birthdays) == 0 {
			if _, err := s.ChannelMessageSend(m.ChannelID, "🎂 No upcoming birthdays found!"); err != nil {
				log.Printf("Failed to send no birthdays message: %v", err)
			}
		} else {
			message := "🎂 **Upcoming Birthdays:**\n"
			for _, birthday := range birthdays {
				message += fmt.Sprintf("• %s - %02d/%02d\n", birthday.Username, birthday.Month, birthday.Day)
			}
			if _, err := s.ChannelMessageSend(m.ChannelID, message); err != nil {
				log.Printf("Failed to send birthdays message: %v", err)
			}
		}
	} else if strings.HasPrefix(m.Content, "!setrolemessage ") {
		// Check if user has admin permissions (including server owner)
		if !b.isUserAdmin(s, m.Author.ID, m.ChannelID) {
			if _, err := s.ChannelMessageSend(m.ChannelID,
				"❌ You need Administrator permissions or server ownership to set role messages!"); err != nil {
				log.Printf("Failed to send permission error message: %v", err)
			}
			return
		}

		// Parse command: !setrolemessage <message_id> [role_name]
		parts := strings.Fields(m.Content)
		if len(parts) < 2 {
			if _, err := s.ChannelMessageSend(m.ChannelID,
				"❌ Usage: !setrolemessage <message_id> [role_name]\nExample: !setrolemessage 123456789 member"); err != nil {
				log.Printf("Failed to send usage message: %v", err)
			}
			return
		}

		messageID := parts[1]
		roleName := "member" // Default role
		if len(parts) >= 3 {
			roleName = parts[2]
		}

		// Verify the role exists
		if b.getRoleID(s, m.GuildID, roleName) == "" {
			if _, err := s.ChannelMessageSend(m.ChannelID,
				fmt.Sprintf("❌ Role '%s' not found in this server!", roleName)); err != nil {
				log.Printf("Failed to send role not found message: %v", err)
			}
			return
		}

		// Verify the message exists and get its channel
		msg, err := s.ChannelMessage(m.ChannelID, messageID)
		if err != nil {
			// Try to find the message in other channels (this is a simplified approach)
			if _, err := s.ChannelMessageSend(m.ChannelID,
				"❌ Message not found! Make sure the message ID is correct and the message is in this channel."); err != nil {
				log.Printf("Failed to send message not found: %v", err)
			}
			return
		}

		if err := b.SetRoleMessage(m.GuildID, msg.ChannelID, messageID, roleName); err != nil {
			errorMsg := fmt.Sprintf("❌ Failed to set role message: %v", err)
			if _, err := s.ChannelMessageSend(m.ChannelID, errorMsg); err != nil {
				log.Printf("Failed to send error message: %v", err)
			}
		} else {
			successMsg := fmt.Sprintf("✅ Message %s will now grant the '%s' role when reacted to!", messageID, roleName)
			if _, err := s.ChannelMessageSend(m.ChannelID, successMsg); err != nil {
				log.Printf("Failed to send success message: %v", err)
			}
		}
	} else if strings.HasPrefix(m.Content, "!removerolemessage ") {
		// Check if user has admin permissions (including server owner)
		if !b.isUserAdmin(s, m.Author.ID, m.ChannelID) {
			if _, err := s.ChannelMessageSend(m.ChannelID,
				"❌ You need Administrator permissions or server ownership to remove role messages!"); err != nil {
				log.Printf("Failed to send permission error message: %v", err)
			}
			return
		}

		// Parse command: !removerolemessage <message_id>
		parts := strings.Fields(m.Content)
		if len(parts) < 2 {
			if _, err := s.ChannelMessageSend(m.ChannelID,
				"❌ Usage: !removerolemessage <message_id>"); err != nil {
				log.Printf("Failed to send usage message: %v", err)
			}
			return
		}

		messageID := parts[1]

		if err := b.RemoveRoleMessage(messageID); err != nil {
			errorMsg := fmt.Sprintf("❌ Failed to remove role message: %v", err)
			if _, err := s.ChannelMessageSend(m.ChannelID, errorMsg); err != nil {
				log.Printf("Failed to send error message: %v", err)
			}
		} else {
			if _, err := s.ChannelMessageSend(m.ChannelID, "✅ Role message removed successfully!"); err != nil {
				log.Printf("Failed to send success message: %v", err)
			}
		}
	} else if strings.HasPrefix(m.Content, "!listrolemessages") {
		// Check if user has admin permissions (including server owner)
		if !b.isUserAdmin(s, m.Author.ID, m.ChannelID) {
			if _, err := s.ChannelMessageSend(m.ChannelID,
				"❌ You need Administrator permissions or server ownership to list role messages!"); err != nil {
				log.Printf("Failed to send permission error message: %v", err)
			}
			return
		}

		roleMessages, err := b.GetRoleMessages(m.GuildID)
		if err != nil {
			errorMsg := fmt.Sprintf("❌ Failed to get role messages: %v", err)
			if _, err := s.ChannelMessageSend(m.ChannelID, errorMsg); err != nil {
				log.Printf("Failed to send error message: %v", err)
			}
			return
		}

		if len(roleMessages) == 0 {
			if _, err := s.ChannelMessageSend(m.ChannelID, "📝 No role messages are currently set up."); err != nil {
				log.Printf("Failed to send no role messages: %v", err)
			}
		} else {
			message := "📝 **Active Role Messages:**\n"
			for _, rm := range roleMessages {
				message += fmt.Sprintf("• Message ID: `%s` → Role: `%s` (in <#%s>)\n", rm.MessageID, rm.RoleName, rm.ChannelID)
			}
			if _, err := s.ChannelMessageSend(m.ChannelID, message); err != nil {
				log.Printf("Failed to send role messages list: %v", err)
			}
		}
	}
}

// handleMessageReactionAdd handles when a user adds a reaction to a message
func (b *Bot) handleMessageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	// Ignore reactions from bots
	if r.UserID == s.State.User.ID {
		return
	}

	// Check if this message is designated as a role-granting message
	var roleMessage RoleMessage
	if err := b.dbConn.Where("message_id = ? AND is_active = ?", r.MessageID, true).First(&roleMessage).Error; err != nil {
		// Message is not set up for role granting, ignore
		return
	}

	// Get guild member
	member, err := s.GuildMember(r.GuildID, r.UserID)
	if err != nil {
		log.Printf("Failed to get guild member for reaction: %v", err)
		return
	}

	// Check if user already has the target role
	if b.hasRole(s, r.GuildID, member.Roles, roleMessage.RoleName) {
		return // User already has the role
	}

	// Get the target role ID
	targetRoleID := b.getRoleID(s, r.GuildID, roleMessage.RoleName)
	if targetRoleID == "" {
		log.Printf("Could not find '%s' role in guild %s", roleMessage.RoleName, r.GuildID)
		return
	}

	// Assign the target role to the user
	if err := s.GuildMemberRoleAdd(r.GuildID, r.UserID, targetRoleID); err != nil {
		log.Printf("Failed to add %s role to user %s: %v", roleMessage.RoleName, r.UserID, err)
		return
	}

	log.Printf("Added '%s' role to user %s (%s) for reacting to message %s in guild %s", roleMessage.RoleName, member.User.Username, r.UserID, r.MessageID, r.GuildID)

	// Optionally send a DM to the user (uncomment if desired)
	/*
		channel, err := s.UserChannelCreate(r.UserID)
		if err == nil {
			s.ChannelMessageSend(channel.ID, fmt.Sprintf("🎉 Welcome! You've been given the '%s' role for participating in the server!", roleMessage.RoleName))
		}
	*/
}

// hasRole checks if a user has a specific role
func (b *Bot) hasRole(s *discordgo.Session, guildID string, userRoles []string, roleName string) bool {
	// Get all guild roles
	roles, err := s.GuildRoles(guildID)
	if err != nil {
		log.Printf("Failed to get guild roles: %v", err)
		return false
	}

	// Find the role ID for the role name
	var targetRoleID string
	for _, role := range roles {
		if role.Name == roleName {
			targetRoleID = role.ID
			break
		}
	}

	if targetRoleID == "" {
		return false
	}

	// Check if user has this role
	for _, roleID := range userRoles {
		if roleID == targetRoleID {
			return true
		}
	}

	return false
}

// getRoleID gets the role ID for a given role name
func (b *Bot) getRoleID(s *discordgo.Session, guildID string, roleName string) string {
	// Get all guild roles
	roles, err := s.GuildRoles(guildID)
	if err != nil {
		log.Printf("Failed to get guild roles: %v", err)
		return ""
	}

	// Find the role ID for the role name
	for _, role := range roles {
		if role.Name == roleName {
			return role.ID
		}
	}

	return ""
}

// NewBot creates a new bot instance
func NewBot(discordToken string, dbPath string, redisAddr string, twitchClientID string, twitchClientSecret string) (*Bot, error) {
	// Initialize Discord session
	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	// Set up Discord intents (required for message handling and reactions)
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages | discordgo.IntentsMessageContent | discordgo.IntentsGuildMessageReactions

	// Initialize database using GORM
	dbConn, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Run migrations
	if err := dbConn.AutoMigrate(&GoopCreator{}, &TwitchStream{}, &NotificationChannel{}, &Birthday{}, &BirthdayChannel{}, &RoleMessage{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Initialize Redis
	ctx := context.Background()
	redisClient, err := redisutil.NewRedisClient(ctx, redisutil.Config{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// Initialize Twitch API client
	twitchClient, err := twitch.NewClient(twitch.Config{
		ClientID:     twitchClientID,
		ClientSecret: twitchClientSecret,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Twitch client: %w", err)
	}

	// Open Discord session
	if err := dg.Open(); err != nil {
		log.Fatal(err)
	}

	// Create bot instance
	bot := &Bot{
		discord:      dg,
		dbConn:       dbConn,
		redis:        redisClient,
		twitchClient: twitchClient,
	}

	// Register event handlers
	dg.AddHandler(bot.handleReady)
	dg.AddHandler(bot.handleCommands)
	dg.AddHandler(bot.handleMessageReactionAdd)

	return bot, nil
}

// Close cleans up the bot resources
func (b *Bot) Close() {
	b.discord.Close()
	b.redis.Close()
}

// Run starts the bot and blocks until it is stopped
func (b *Bot) Run() {
	log.Println("Bot is running!")

	// Start stream monitoring every 5 minutes
	b.StartStreamMonitoring(5 * time.Minute)

	// Start birthday monitoring (daily checks)
	b.StartBirthdayMonitoring()

	// Run an initial check after 10 seconds
	time.AfterFunc(10*time.Second, func() {
		log.Println("Running initial stream status check...")
		b.CheckStreamStatus()
	})

	select {} // Block forever
}

// LinkTwitchAccount links a Discord user's Twitch account (for Goop Creators)
func (b *Bot) LinkTwitchAccount(discordID, username, guildID, twitchUsername string) error {
	creator := GoopCreator{
		DiscordID:      discordID,
		Username:       username,
		GuildID:        guildID,
		TwitchUsername: twitchUsername,
		IsActive:       true,
	}

	// Use Save to update if exists or create if doesn't
	return b.dbConn.Save(&creator).Error
}

// UnlinkTwitchAccount removes a Discord user's Twitch link
func (b *Bot) UnlinkTwitchAccount(discordID string) error {
	return b.dbConn.Where("discord_id = ?", discordID).Delete(&GoopCreator{}).Error
}

// SetNotificationChannel sets a channel for live notifications
func (b *Bot) SetNotificationChannel(guildID, channelID string) error {
	// First, deactivate all existing notification channels for this guild
	if err := b.dbConn.Model(&NotificationChannel{}).
		Where("guild_id = ?", guildID).
		Update("is_active", false).Error; err != nil {
		return err
	}

	// Create or update the new notification channel
	channel := NotificationChannel{
		GuildID:   guildID,
		ChannelID: channelID,
		IsActive:  true,
	}

	return b.dbConn.Save(&channel).Error
}

// GetLiveGoopCreators returns all live Goop Creators for a guild
func (b *Bot) GetLiveGoopCreators(guildID string) ([]TwitchStream, error) {
	var streams []TwitchStream
	err := b.dbConn.Table("twitch_streams").
		Select("twitch_streams.*").
		Joins("JOIN goop_creators ON twitch_streams.discord_id = goop_creators.discord_id").
		Where("goop_creators.guild_id = ? AND goop_creators.is_active = ? AND twitch_streams.is_live = ?",
			guildID, true, true).
		Find(&streams).Error

	return streams, err
}

// UpdateStreamStatus updates the live status of a streamer and sends notifications if needed
func (b *Bot) UpdateStreamStatus(twitchUsername string, isLive bool, viewerCount int, gameName, streamTitle string) error {
	now := time.Now()

	// Check previous status from Redis cache first
	ctx := context.Background()
	previousStatus, _ := redisutil.GetStreamStatus(ctx, b.redis, twitchUsername)
	wasLiveBefore := previousStatus != nil && previousStatus.IsLive

	// Get previous status from database as backup
	var previousStream TwitchStream
	prevExists := b.dbConn.Where("twitch_username = ?", twitchUsername).First(&previousStream).Error == nil
	if !wasLiveBefore && prevExists {
		wasLiveBefore = previousStream.IsLive
	}

	// Find the Discord ID for this Twitch username
	var creator GoopCreator
	if err := b.dbConn.Where("twitch_username = ?", twitchUsername).First(&creator).Error; err != nil {
		// If no creator found, still update the stream but don't send notifications
		log.Printf("No Goop Creator found for Twitch username: %s", twitchUsername)
	}

	// Update or create the stream status
	stream := TwitchStream{
		TwitchUsername: twitchUsername,
		IsLive:         isLive,
		LastChecked:    &now,
		ViewerCount:    viewerCount,
		GameName:       gameName,
		StreamTitle:    streamTitle,
		DiscordID:      creator.DiscordID,
	}

	if err := b.dbConn.Save(&stream).Error; err != nil {
		return err
	}

	// If streamer just went live (wasn't live before and is live now), send notifications
	if isLive && !wasLiveBefore && creator.DiscordID != "" {
		log.Printf("🔴 %s just went LIVE! Sending notifications...", creator.Username)
		go b.sendGoingLiveNotifications(creator.GuildID, twitchUsername, creator.Username, gameName, streamTitle, viewerCount)
	} else if isLive && wasLiveBefore {
		log.Printf("📺 %s is still live (updating data only)", creator.Username)
	} else if !isLive && wasLiveBefore {
		log.Printf("⚫ %s went offline", creator.Username)
	}

	return nil
}

// sendGoingLiveNotifications sends notifications when a Goop Creator goes live
func (b *Bot) sendGoingLiveNotifications(guildID, twitchUsername, discordUsername, gameName, streamTitle string, viewerCount int) {
	// Get notification channels for this guild
	var channels []NotificationChannel
	if err := b.dbConn.Where("guild_id = ? AND is_active = ?", guildID, true).Find(&channels).Error; err != nil {
		log.Printf("Failed to get notification channels for guild %s: %v", guildID, err)
		return
	}

	if len(channels) == 0 {
		log.Printf("No active notification channels found for guild %s", guildID)
		return
	}

	// Create notification embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🔴 %s is now LIVE!", discordUsername),
		Description: streamTitle,
		Color:       0x9146FF, // Twitch purple
		URL:         fmt.Sprintf("https://twitch.tv/%s", twitchUsername),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Twitch Channel",
				Value:  fmt.Sprintf("[%s](https://twitch.tv/%s)", twitchUsername, twitchUsername),
				Inline: true,
			},
			{
				Name:   "Game/Category",
				Value:  gameName,
				Inline: true,
			},
			{
				Name:   "Viewers",
				Value:  fmt.Sprintf("%d", viewerCount),
				Inline: true,
			},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: fmt.Sprintf("https://static-cdn.jtvnw.net/previews-ttv/live_user_%s-320x180.jpg", strings.ToLower(twitchUsername)),
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "GoopBot Live Notifications",
		},
	}

	// Send notifications to all active notification channels
	for _, channel := range channels {
		if _, err := b.discord.ChannelMessageSendEmbed(channel.ChannelID, embed); err != nil {
			log.Printf("Failed to send notification to channel %s: %v", channel.ChannelID, err)
		}
	}

	log.Printf("Sent going live notifications for %s (%s) to %d channels", discordUsername, twitchUsername, len(channels))
}

// CheckStreamStatus method you can call periodically to check Twitch API
func (b *Bot) CheckStreamStatus() {
	// Get all active Goop Creators
	var creators []GoopCreator
	if err := b.dbConn.Where("is_active = ?", true).Find(&creators).Error; err != nil {
		log.Printf("Failed to get active creators: %v", err)
		return
	}

	if len(creators) == 0 {
		log.Println("No active Goop Creators to check")
		return
	}

	// Extract usernames for batch API call
	usernames := make([]string, len(creators))
	for i, creator := range creators {
		usernames[i] = creator.TwitchUsername
	}

	log.Printf("Checking stream status for %d creators: %v", len(usernames), usernames)

	// Get stream data from Twitch API
	streams, err := b.twitchClient.GetMultipleStreams(usernames)
	if err != nil {
		log.Printf("Failed to get stream data from Twitch API: %v", err)
		return
	}

	// Create a map of live streams for quick lookup
	liveStreams := make(map[string]*twitch.StreamData)
	for i := range streams {
		liveStreams[strings.ToLower(streams[i].UserLogin)] = &streams[i]
	}

	// Process each creator
	for _, creator := range creators {
		username := strings.ToLower(creator.TwitchUsername)

		if streamData, isLive := liveStreams[username]; isLive {
			// Creator is live
			log.Printf("%s is LIVE: %s", creator.TwitchUsername, streamData.Title)

			// Update stream status in database
			if err := b.UpdateStreamStatus(
				creator.TwitchUsername,
				true,
				streamData.ViewerCount,
				streamData.GameName,
				streamData.Title,
			); err != nil {
				log.Printf("Failed to update stream status for %s: %v", creator.TwitchUsername, err)
			}

			// Cache in Redis to avoid duplicate notifications
			ctx := context.Background()
			if err := redisutil.SetStreamStatus(ctx, b.redis, creator.TwitchUsername, true); err != nil {
				log.Printf("Failed to cache stream status for %s: %v", creator.TwitchUsername, err)
			}
		} else {
			// Creator is offline
			log.Printf("%s is offline", creator.TwitchUsername)

			// Update stream status in database
			if err := b.UpdateStreamStatus(
				creator.TwitchUsername,
				false,
				0,
				"",
				"",
			); err != nil {
				log.Printf("Failed to update stream status for %s: %v", creator.TwitchUsername, err)
			}

			// Update Redis cache
			ctx := context.Background()
			if err := redisutil.SetStreamStatus(ctx, b.redis, creator.TwitchUsername, false); err != nil {
				log.Printf("Failed to cache stream status for %s: %v", creator.TwitchUsername, err)
			}
		}
	}

	log.Printf("Stream status check completed. Found %d live streams out of %d creators", len(streams), len(creators))
}

// StartStreamMonitoring starts periodic stream status monitoring
func (b *Bot) StartStreamMonitoring(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			b.CheckStreamStatus()
		}
	}()
	log.Printf("Started stream monitoring with %v interval", interval)
}

// Birthday-related methods

// SetUserBirthday sets a user's birthday
func (b *Bot) SetUserBirthday(discordID, username, guildID, birthdayStr string) error {
	// Parse birthday string (MM/DD format)
	parts := strings.Split(birthdayStr, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid birthday format, use MM/DD (e.g., 03/15)")
	}

	month, err := strconv.Atoi(parts[0])
	if err != nil || month < 1 || month > 12 {
		return fmt.Errorf("invalid month, must be 01-12")
	}

	day, err := strconv.Atoi(parts[1])
	if err != nil || day < 1 || day > 31 {
		return fmt.Errorf("invalid day, must be 01-31")
	}

	// Create or update birthday record
	birthday := Birthday{
		DiscordID: discordID,
		Username:  username,
		GuildID:   guildID,
		Month:     month,
		Day:       day,
	}

	// Use GORM's Upsert functionality
	result := b.dbConn.Where("discord_id = ?", discordID).Assign(birthday).FirstOrCreate(&birthday)
	return result.Error
}

// SetBirthdayChannel sets the birthday notification channel for a guild
func (b *Bot) SetBirthdayChannel(guildID, channelID string) error {
	// First, deactivate any existing birthday channels for this guild
	if err := b.dbConn.Model(&BirthdayChannel{}).Where("guild_id = ?", guildID).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate existing birthday channels: %w", err)
	}

	// Create or update the new birthday channel
	channel := BirthdayChannel{
		GuildID:   guildID,
		ChannelID: channelID,
		IsActive:  true,
	}

	result := b.dbConn.Where("channel_id = ?", channelID).Assign(channel).FirstOrCreate(&channel)
	return result.Error
}

// GetUpcomingBirthdays gets upcoming birthdays for a guild (next 30 days)
func (b *Bot) GetUpcomingBirthdays(guildID string) ([]Birthday, error) {
	var birthdays []Birthday

	// Get current date
	now := time.Now()
	currentMonth := int(now.Month())
	currentDay := now.Day()

	// Query for birthdays in current month from today onwards, or next month
	query := b.dbConn.Where("guild_id = ?", guildID)

	// This is a simplified version - for production you'd want more sophisticated date logic
	query = query.Where("(month = ? AND day >= ?) OR (month = ?)",
		currentMonth, currentDay, (currentMonth%12)+1)

	if err := query.Order("month ASC, day ASC").Find(&birthdays).Error; err != nil {
		return nil, fmt.Errorf("failed to get birthdays: %w", err)
	}

	return birthdays, nil
}

// CheckBirthdays checks for today's birthdays and sends notifications
func (b *Bot) CheckBirthdays() {
	now := time.Now()
	currentMonth := int(now.Month())
	currentDay := now.Day()

	// Get all birthdays for today
	var birthdays []Birthday
	if err := b.dbConn.Where("month = ? AND day = ?", currentMonth, currentDay).Find(&birthdays).Error; err != nil {
		log.Printf("Failed to get today's birthdays: %v", err)
		return
	}

	for _, birthday := range birthdays {
		// Check if we already sent a birthday message today
		if birthday.LastSent.Format("2006-01-02") == now.Format("2006-01-02") {
			continue
		}

		// Get birthday channel for this guild
		var channel BirthdayChannel
		if err := b.dbConn.Where("guild_id = ? AND is_active = ?", birthday.GuildID, true).First(&channel).Error; err != nil {
			log.Printf("No active birthday channel found for guild %s", birthday.GuildID)
			continue
		}

		// Send birthday message
		message := fmt.Sprintf("🎉 **Happy Birthday** <@%s>! 🎂\nHope you have a wonderful day! 🎈", birthday.DiscordID)

		if _, err := b.discord.ChannelMessageSend(channel.ChannelID, message); err != nil {
			log.Printf("Failed to send birthday message for %s: %v", birthday.Username, err)
			continue
		}

		// Update last sent timestamp
		birthday.LastSent = now
		if err := b.dbConn.Save(&birthday).Error; err != nil {
			log.Printf("Failed to update birthday last sent for %s: %v", birthday.Username, err)
		}

		log.Printf("Sent birthday message for %s in guild %s", birthday.Username, birthday.GuildID)
	}
}

// StartBirthdayMonitoring starts daily birthday checking
func (b *Bot) StartBirthdayMonitoring() {
	// Check birthdays once at startup
	go b.CheckBirthdays()

	// Then check every 24 hours at midnight
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			b.CheckBirthdays()
		}
	}()
	log.Println("Started birthday monitoring")
}

// Role message management methods

// SetRoleMessage sets a message to grant a role when reacted to
func (b *Bot) SetRoleMessage(guildID, channelID, messageID, roleName string) error {
	roleMessage := RoleMessage{
		GuildID:   guildID,
		ChannelID: channelID,
		MessageID: messageID,
		RoleName:  roleName,
		IsActive:  true,
	}

	// Use GORM's Upsert functionality to update if exists or create if doesn't
	result := b.dbConn.Where("message_id = ?", messageID).Assign(roleMessage).FirstOrCreate(&roleMessage)
	return result.Error
}

// RemoveRoleMessage removes role-granting functionality from a message
func (b *Bot) RemoveRoleMessage(messageID string) error {
	return b.dbConn.Where("message_id = ?", messageID).Delete(&RoleMessage{}).Error
}

// GetRoleMessages gets all active role messages for a guild
func (b *Bot) GetRoleMessages(guildID string) ([]RoleMessage, error) {
	var roleMessages []RoleMessage
	err := b.dbConn.Where("guild_id = ? AND is_active = ?", guildID, true).Find(&roleMessages).Error
	return roleMessages, err
}
