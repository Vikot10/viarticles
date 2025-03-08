package main

import (
	"os"

	"github.com/rs/zerolog"
)

func createLogger(isDebug bool) *zerolog.Logger {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	return &logger
}
