package database

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rs/zerolog"
)

//go:embed migrations/*.sql
var fsMain embed.FS

func MakeMigration(pgConnection string, logger *zerolog.Logger) error {
	var d source.Driver
	var errIofs error

	d, errIofs = iofs.New(fsMain, "migrations")
	if errIofs != nil {
		return fmt.Errorf("error new iofs: %w", errIofs)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, pgConnection)
	if err != nil {
		return fmt.Errorf("error new migrate: %w", err)
	}
	defer m.Close()

	ver, dirty, errGetVersion := m.Version()
	if errGetVersion != nil && !errors.Is(errGetVersion, migrate.ErrNilVersion) {
		return fmt.Errorf("error get version: %w", errGetVersion)
	}
	if dirty {
		m.Force(int(ver - 1))
	}

	errUp := m.Up()
	if errUp != nil {
		if errors.Is(errUp, migrate.ErrNoChange) {
			logger.Info().Msg("no change for migrate")
			return nil
		}

		return fmt.Errorf("error up migrate: %w", errUp)
	}

	logger.Info().Msg("migrate done")

	return nil
}
