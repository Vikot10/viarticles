package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Address               string
	DatabasePath          string
	TelegramBotToken      string
	TelegramAllowedChatID int64
}

func Load() (Config, error) {
	cfg := Config{
		Address:      envOrDefault("VIARTICLES_ADDRESS", ":8080"),
		DatabasePath: envOrDefault("VIARTICLES_DATABASE_PATH", "./data/viarticles.db"),
		TelegramBotToken: strings.TrimSpace(
			os.Getenv("VIARTICLES_TELEGRAM_BOT_TOKEN"),
		),
	}

	chatID := strings.TrimSpace(os.Getenv("VIARTICLES_TELEGRAM_ALLOWED_CHAT_ID"))
	if chatID != "" {
		parsed, err := strconv.ParseInt(chatID, 10, 64)
		if err != nil {
			return Config{}, fmt.Errorf("VIARTICLES_TELEGRAM_ALLOWED_CHAT_ID must be an integer: %w", err)
		}
		cfg.TelegramAllowedChatID = parsed
	}

	return cfg, nil
}

func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
