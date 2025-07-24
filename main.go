package main

import (
	"log"
	"os"

	"GoopBot/internal/bot"
)

type Config struct {
	DiscordToken       string
	DBPath             string
	RedisAddr          string
	TwitchClientID     string
	TwitchClientSecret string
}

func main() {
	config := Config{
		DiscordToken:       os.Getenv("DISCORD_TOKEN"),
		DBPath:             "./GoopBot.db",
		RedisAddr:          os.Getenv("REDIS_ADDR"),
		TwitchClientID:     os.Getenv("TWITCH_CLIENT_ID"),
		TwitchClientSecret: os.Getenv("TWITCH_CLIENT_SECRET"),
	}

	bot, err := bot.NewBot(config.DiscordToken, config.DBPath, config.RedisAddr, config.TwitchClientID, config.TwitchClientSecret)
	if err != nil {
		log.Fatal(err)
	}

	// Main event loop
	bot.Run()

	log.Println("Bot is running!")
	select {}

}
