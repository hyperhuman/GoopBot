// Package bot provides the Discord bot implementation.
package bot

import (
	"GoopBot/redis/redisutil"
	"GoopBot/storage/db"
	"context"
	"database/sql"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Bot represents our Discord bot instance
type Bot struct {
	discord *discordgo.Session
	dbConn  *sql.DB
	redis   *redis.Client
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

	// Handle commands using strings.HasPrefix
	if strings.HasPrefix(m.Content, "!help") {
		helpMessage := `
Available commands:
!help - Show this help message
!twitch <username> - Subscribe to a Twitch streamer
!status - Show current subscriptions
        `
		if _, err := s.ChannelMessageSend(m.ChannelID, helpMessage); err != nil {
			log.Printf("Failed to send help message: %v", err)
		}
	} else if strings.HasPrefix(m.Content, "!twitch ") {
		// Extract username from command
		username := strings.TrimSpace(m.Content[7:]) // 7 is length of "!twitch "
		if username != "" {
			// Add subscription logic here
			if _, err := s.ChannelMessageSend(m.ChannelID,
				fmt.Sprintf("Subscribed to %s's stream notifications", username)); err != nil {
				log.Printf("Failed to send subscription message: %v", err)
			}
		}
	}
}

// NewBot creates a new bot instance
func NewBot(discordToken string, dbPath string, redisAddr string) (*Bot, error) {
	// Initialize Discord session
	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	// Initialize database
	dbConn, err := db.NewSQLiteDB(db.Config{Path: dbPath})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Run migrations
	if err := db.Migrate(dbConn); err != nil {
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

	// Create bot instance
	bot := &Bot{
		discord: dg,
		dbConn:  dbConn,
		redis:   redisClient,
	}

	// Register event handlers
	dg.AddHandler(bot.handleReady)
	dg.AddHandler(bot.handleCommands)

	return bot, nil
}
