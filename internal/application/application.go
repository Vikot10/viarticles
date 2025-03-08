package application

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/Vikot10/viarticles/internal/service/articleservice"
	"github.com/Vikot10/viarticles/internal/service/vkservice"
	"github.com/Vikot10/viarticles/internal/storage"
)

type Application struct {
	logger *zerolog.Logger

	as *articleservice.ArticleService
	vk *vkservice.VkService
}

func New(store *storage.Storage, logger *zerolog.Logger) *Application {
	app := &Application{}

	app.as = articleservice.New(logger, store)
	app.vk = vkservice.New(logger, "")

	l := logger.With().Str("component", "application").Logger()
	app.logger = &l

	return app
}

func (app *Application) Run(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup, ln net.Listener) {
	defer wg.Done()

	r := chi.NewRouter()
	app.registerRoutes(r)

	server := http.Server{
		Handler: r,
	}

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		<-ctx.Done()

		//logger.Info("server shutdown")
		errShutdown := server.Shutdown(context.Background())
		if errShutdown != nil {
			//logger.Error("server shutdown error", zap.Error(errShutdown))
		}
	}(wg)

	//logger.Info("server start", zap.String("address", ln.Addr().String()))
	errServe := server.Serve(ln)
	if errServe != nil {
		if !errors.Is(errServe, http.ErrServerClosed) {
			//logger.Error("server serve error", zap.Error(errServe))
		}
		cancel()
	}

}
