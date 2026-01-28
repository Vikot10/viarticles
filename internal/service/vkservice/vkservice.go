package vkservice

import (
	"github.com/rs/zerolog"

	"github.com/Vikot10/viarticles/internal/storage"
)

const (
	vkAPIVersion = "5.199"
	vkAPIBaseURL = "https://api.vk.com/method"
)

type VkService struct {
	logger      *zerolog.Logger
	accessToken string
	store       *storage.Storage
}

func New(logger *zerolog.Logger, accessToken string, store *storage.Storage) *VkService {
	vk := VkService{
		accessToken: accessToken,
		store:       store,
	}
	l := logger.With().Str("service", "vk service").Logger()
	vk.logger = &l

	return &vk
}
