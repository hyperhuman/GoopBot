// Test script for Twitch API integration
// Run with: go run cmd/test_twitch/main.go
package main

import (
	"GoopBot/internal/twitch"
	"fmt"
	"log"
	"os"
)

func main() {
	clientID := os.Getenv("TWITCH_CLIENT_ID")
	clientSecret := os.Getenv("TWITCH_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("Please set TWITCH_CLIENT_ID and TWITCH_CLIENT_SECRET environment variables")
	}

	// Create Twitch client
	client, err := twitch.NewClient(twitch.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})
	if err != nil {
		log.Fatalf("Failed to create Twitch client: %v", err)
	}

	// Test with some popular streamers
	testStreamers := []string{"shroud", "ninja", "pokimane", "xqc"}

	fmt.Println("Testing Twitch API integration...")
	fmt.Println("=====================================")

	for _, streamer := range testStreamers {
		isLive, streamData, err := client.IsUserLive(streamer)
		if err != nil {
			fmt.Printf("❌ Error checking %s: %v\n", streamer, err)
			continue
		}

		if isLive {
			fmt.Printf("🔴 %s is LIVE!\n", streamer)
			fmt.Printf("   Title: %s\n", streamData.Title)
			fmt.Printf("   Game: %s\n", streamData.GameName)
			fmt.Printf("   Viewers: %d\n", streamData.ViewerCount)
		} else {
			fmt.Printf("⚫ %s is offline\n", streamer)
		}
		fmt.Println()
	}

	fmt.Println("✅ Twitch API test completed successfully!")
}
