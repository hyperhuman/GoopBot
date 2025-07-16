package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	LogLevel string `json:"log_level"`
	// Add other config fields as needed
}

func LoadConfig() (Config, error) {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	// Read config file
	configFile := "config.json"
	data, err := os.ReadFile(configFile)
	if err != nil {
		return Config{}, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("unmarshaling config: %w", err)
	}

	return cfg, nil
}
