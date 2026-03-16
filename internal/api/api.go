package api

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/zneix/dink-filter/internal/config"
)

type API struct {
	router *chi.Mux
	cfg    *config.Config
}

func New(cfg *config.Config) *API {
	router := chi.NewRouter()

	router.Use(middleware.StripPrefix(cfg.RoutePrefix))
	router.Use(middleware.StripSlashes)

	router.Get("/", handleRoot)
	router.Post("/filter", func(w http.ResponseWriter, r *http.Request) {
		handleFilter(w, r, cfg)
	})

	return &API{
		router: router,
		cfg:    cfg,
	}
}

// Listen listens on cfg.BindAddress (blocking)
func (api *API) Listen() {
	srv := &http.Server{
		Handler: api.router,
		Addr:    api.cfg.BindAddress,
	}

	log.Printf("[API] Listening on %q (prefix=%q)\n", api.cfg.BindAddress, api.cfg.RoutePrefix)
	err := srv.ListenAndServe()
	log.Println("[API] Failed to listen:", err)
}
