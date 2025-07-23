package main

import (
	"log"
	"os"

	"GoopBot/internal/bot"
	"github.com/bwmarrin/discordgo"

	"gorm.io/gorm"
)

type Config struct {
	DiscordToken string
	DBPath       string
	RedisAddr    string
}

type Post struct {
	gorm.Model
	Title  string
	Slug   string
	Author discordgo.User
}

func main() {
	config := Config{
		DiscordToken: os.Getenv("DISCORD_TOKEN"),
		DBPath:       "./GoopBot.db",
		RedisAddr:    os.Getenv("REDIS_ADDR"),
	}

	bot, err := bot.NewBot(config.DiscordToken, config.DBPath, config.RedisAddr)
	if err != nil {
		log.Fatal(err)
	}

	// Main event loop
	bot.Run()

	log.Println("Bot is running!")
	select {}

}
