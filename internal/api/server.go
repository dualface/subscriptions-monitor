package api

import (
	"context"
	"net/http"
	"time"

	"github.com/user/subscriptions-monitor/internal/config"
	"github.com/user/subscriptions-monitor/internal/provider"
)

const (
	cacheTTL       = 90 * time.Second
	refreshInterval = 60 * time.Second
)

type Server struct {
	registry *provider.Registry
	config   *config.Config
	server   *http.Server
	cache    *Cache
	stopChan chan struct{}
}

func NewServer(registry *provider.Registry, cfg *config.Config, addr string) *Server {
	s := &Server{
		registry: registry,
		config:   cfg,
		cache:    NewCache(cacheTTL),
		stopChan: make(chan struct{}),
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
	ctx := context.Background()
	s.refreshCache(ctx)

	go s.startBackgroundRefresh()

	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	close(s.stopChan)
	return s.server.Shutdown(ctx)
}

func (s *Server) startBackgroundRefresh() {
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			s.refreshCache(ctx)
		case <-s.stopChan:
			return
		}
	}
}

func (s *Server) refreshCache(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, s.config.Settings.Timeout)
	defer cancel()

	snapshots := s.registry.FetchAll(ctx, s.config.Subscriptions)
	s.cache.Set(snapshots)
}
