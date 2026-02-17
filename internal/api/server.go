package api

import (
	"context"
	"net/http"

	"github.com/user/subscriptions-monitor/internal/config"
	"github.com/user/subscriptions-monitor/internal/provider"
)

type Server struct {
	registry *provider.Registry
	config   *config.Config
	server   *http.Server
}

func NewServer(registry *provider.Registry, cfg *config.Config, addr string) *Server {
	s := &Server{
		registry: registry,
		config:   cfg,
	}

	mux := http.NewServeMux()
	s.registerHandlers(mux)

	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return s
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
