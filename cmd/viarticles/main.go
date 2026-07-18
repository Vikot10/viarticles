package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Vikot10/viarticles/internal/config"
	"github.com/Vikot10/viarticles/internal/storage"
	"github.com/Vikot10/viarticles/internal/telegram"
	"github.com/Vikot10/viarticles/internal/webapp"
)

var version = "undefined"

func main() {
	if err := run(); err != nil {
		slog.Error("application stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store, err := storage.Open(ctx, cfg.DatabasePath)
	if err != nil {
		return err
	}
	defer store.Close()

	web, err := webapp.New(store)
	if err != nil {
		return err
	}
	bot := telegram.New(cfg.TelegramBotToken, cfg.TelegramAllowedChatID, store)
	go bot.Run(ctx)

	server := &http.Server{
		Addr:              cfg.Address,
		Handler:           web.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("viarticles started", "version", version, "address", cfg.Address, "database", cfg.DatabasePath)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
