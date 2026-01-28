package telegramservice

import (
	"github.com/rs/zerolog"

	"github.com/Vikot10/viarticles/internal/storage"
)

const (
	tgAPIBaseURL = "https://api.telegram.org/bot"
)

type TelegramService struct {
	logger   *zerolog.Logger
	botToken string
	store    *storage.Storage
}

func New(logger *zerolog.Logger, botToken string, store *storage.Storage) *TelegramService {
	tg := &TelegramService{
		botToken: botToken,
		store:    store,
	}
	l := logger.With().Str("service", "telegram service").Logger()
	tg.logger = &l

	return tg
}
