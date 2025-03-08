package articleservice

import (
	"github.com/rs/zerolog"

	"github.com/Vikot10/viarticles/internal/storage"
)

type ArticleService struct {
	logger *zerolog.Logger
	store  *storage.Storage
}

func New(logger *zerolog.Logger, store *storage.Storage) *ArticleService {
	as := &ArticleService{
		store: store,
	}
	l := logger.With().Str("service", "article service").Logger()
	as.logger = &l

	return as
}
