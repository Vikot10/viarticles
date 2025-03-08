package vkservice

import (
	"github.com/rs/zerolog"
)

type VkService struct {
	logger      *zerolog.Logger
	accessToken string
}

func New(logger *zerolog.Logger, accessToken string) *VkService {
	vk := VkService{
		accessToken: "",
	}
	l := logger.With().Str("service", "vk service").Logger()
	vk.logger = &l

	return &vk
}
