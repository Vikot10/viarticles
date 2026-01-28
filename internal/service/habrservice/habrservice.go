package habrservice

import (
	"github.com/rs/zerolog"

	"github.com/Vikot10/viarticles/internal/storage"
)

const (
	habrBaseURL = "https://habr.com"
	habrAPIURL  = "https://habr.com/kek/v2"
)

type HabrService struct {
	logger    *zerolog.Logger
	authToken string
	store     *storage.Storage
}

func New(logger *zerolog.Logger, authToken string, store *storage.Storage) *HabrService {
	h := &HabrService{
		authToken: authToken,
		store:     store,
	}
	l := logger.With().Str("service", "habr service").Logger()
	h.logger = &l

	return h
}
