package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"GoopBot/redis/redisutil"
	"GoopBot/storage/db"
	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
)

// Bot represents our Discord bot instance
type Bot struct {
	discord *discordgo.Session
	sqldb   *sql.DB
	redis   *redis.Client
}

type Config struct {
	DiscordToken string
	DBPath       string
	RedisAddr    string
}

// NewBot creates a new bot instance
func NewBot(discordToken string, dbPath string, redisAddr string) (*Bot, error) {
	// Initialize Discord session
	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	// Initialize database
	sqldb, err := db.NewSQLiteDB(db.Config{Path: dbPath})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Run migrations
	if err := db.Migrate(sqldb); err != nil {
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

	return &Bot{
		discord: dg,
		sqldb:   sqldb,
		redis:   redisClient,
	}, nil
}

func main() {
	config := Config{
		DiscordToken: os.Getenv("DISCORD_TOKEN"),
		DBPath:       "./twitchbot.db",
		RedisAddr:    os.Getenv("REDIS_ADDR"),
	}

	bot, err := NewBot(config.DiscordToken, config.DBPath, config.RedisAddr)
	if err != nil {
		log.Fatal(err)
	}

	if err := bot.discord.Open(); err != nil {
		log.Fatal(err)
	}

	log.Println("Bot is running!")
	select {}

}
