package application

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	getArticle "github.com/Vikot10/viarticles/internal/api/article/get"
)

func (app *Application) registerRoutes(r *chi.Mux) {
	r.Use(middleware.Recoverer)
	r.Get("/", func(http.ResponseWriter, *http.Request) {})

	r.Route("articles", func(r chi.Router) {
		//r.Get("/", func(writer http.ResponseWriter, request *http.Request) {
		//	app.ListArticles(writer, request)
		//})
		r.Get("/{id}", func(writer http.ResponseWriter, request *http.Request) {
			getArticle.Handler(writer, request, app.as)
		})
		//r.Post("/", app.CreateArticle)
		//r.Put("/{id}", func(writer http.ResponseWriter, request *http.Request) {
		//
		//})
		//r.Delete("/{id}", app.DeleteArticle)
	})
}
