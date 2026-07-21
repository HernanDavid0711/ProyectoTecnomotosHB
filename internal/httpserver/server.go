package httpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"tecnomotos/internal/clientes"
	"tecnomotos/internal/config"
	"tecnomotos/internal/motos"
	"tecnomotos/internal/shared"
)

type Server struct {
	httpServer *http.Server
	db         *pgxpool.Pool
	cfg        *config.Config
}

func New(cfg *config.Config, db *pgxpool.Pool) *Server {
	mux := http.NewServeMux()

	s := &Server{
		db:  db,
		cfg: cfg,
	}

	mux.HandleFunc("GET /health", s.healthHandler)
	mux.HandleFunc("GET /", s.homeHandler)

	clienteRepo := clientes.NewRepository(db)
	clienteService := clientes.NewService(clienteRepo)
	clienteHandler := clientes.NewHandler(clienteService)
	clienteHandler.RegisterRoutes(mux)

	motoRepo := motos.NewRepository(db)
	motoService := motos.NewService(motoRepo)
	motoHandler := motos.NewHandler(motoService)
	motoHandler.RegisterRoutes(mux)

	s.httpServer = &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return s
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "API Tecnomotos funcionando",
	})
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := s.db.Ping(ctx); err != nil {
		shared.WriteJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status":   "error",
			"database": "down",
			"error":    err.Error(),
		})
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"status":   "ok",
		"database": "up",
	})
}
