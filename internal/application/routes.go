package application

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Vikot10/viarticles/internal/api/handlers"
)

func (app *Application) registerRoutes(r *chi.Mux) {
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(corsMiddleware)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy"}`))
	})

	articleHandler := handlers.NewArticleHandler(app.as)
	categoryHandler := handlers.NewCategoryHandler(app.as)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/articles", func(r chi.Router) {
			r.Get("/", articleHandler.List)
			r.Post("/", articleHandler.Create)
			r.Get("/{id}", articleHandler.Get)
			r.Put("/{id}", articleHandler.Update)
			r.Delete("/{id}", articleHandler.Delete)
		})

		r.Route("/categories", func(r chi.Router) {
			r.Get("/", categoryHandler.List)
			r.Post("/", categoryHandler.Create)
			r.Get("/{id}", categoryHandler.Get)
			r.Put("/{id}", categoryHandler.Update)
			r.Delete("/{id}", categoryHandler.Delete)
		})

		r.Route("/sync", func(r chi.Router) {
			r.Post("/vk", app.syncVK)
			r.Post("/telegram", app.syncTelegram)
			r.Post("/habr", app.syncHabr)
		})
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *Application) syncVK(w http.ResponseWriter, r *http.Request) {
	go func() {
		if err := app.vk.SynchronizeArticles(context.Background()); err != nil {
			app.logger.Error().Err(err).Msg("vk sync failed")
		}
	}()
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"sync started"}`))
}

func (app *Application) syncTelegram(w http.ResponseWriter, r *http.Request) {
	go func() {
		if err := app.tg.SynchronizeArticles(context.Background()); err != nil {
			app.logger.Error().Err(err).Msg("telegram sync failed")
		}
	}()
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"sync started"}`))
}

func (app *Application) syncHabr(w http.ResponseWriter, r *http.Request) {
	go func() {
		if err := app.habr.SynchronizeArticles(context.Background()); err != nil {
			app.logger.Error().Err(err).Msg("habr sync failed")
		}
	}()
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"sync started"}`))
}
