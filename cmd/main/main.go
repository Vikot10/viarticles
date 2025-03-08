package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"github.com/Vikot10/viarticles/internal/application"
	"github.com/Vikot10/viarticles/internal/config"
	"github.com/Vikot10/viarticles/internal/database"
	"github.com/Vikot10/viarticles/internal/storage"
)

var version = "undefined"

func main() {
	cfg := config.Load()
	logger := createLogger(cfg.IsDebug)

	errRun := run(cfg, logger)
	if errRun != nil {
		log.Printf("run error: %v", errRun)
		os.Exit(1)
	}
}

func run(cfg *config.Config, logger *zerolog.Logger) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	logger.Info().Str("version", version).Msg("start")

	ln, errLn := net.Listen("tcp", cfg.Address)
	if errLn != nil {
		return fmt.Errorf("listen error: %w", errLn)
	}
	defer ln.Close()

	dbPool, errDB := connectionPostgres(ctx, cfg.Postgres)
	if errDB != nil {
		return fmt.Errorf("db error: %w", errDB)
	}
	defer dbPool.Close()

	if cfg.Postgres.NeedMigrate {
		err := database.MakeMigration(buildPGConnectionString(cfg.Postgres), logger)
		if err != nil {
			return fmt.Errorf("migration error: %w", err)
		}
	}

	wg := sync.WaitGroup{}

	store := storage.New(dbPool)

	app := application.New(store, logger)
	wg.Add(1)
	go app.Run(ctx, cancel, &wg, ln)

	<-ctx.Done()

	wg.Wait()

	logger.Info().Msg("stop")

	return nil
}

func buildPGConnectionString(cfgPg config.Postgres) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&sslrootcert=%s",
		cfgPg.Username,
		cfgPg.Password,
		cfgPg.Host,
		cfgPg.Port,
		cfgPg.Database,
		cfgPg.SSLMode,
		cfgPg.SSLCertPath)
}

func connectionPostgres(ctx context.Context, cfgPg config.Postgres) (*pgxpool.Pool, error) {
	stringConnection := buildPGConnectionString(cfgPg)

	dbPool, errDB := pgxpool.New(ctx, stringConnection)
	if errDB != nil {
		return nil, fmt.Errorf("error create pg pool: %w", errDB)
	}

	errDB = dbPool.Ping(ctx)
	if errDB != nil {
		return nil, fmt.Errorf("error ping pg pool: %w", errDB)
	}

	return dbPool, nil
}
