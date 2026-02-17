package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/user/subscriptions-monitor/internal/provider"
)

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) usageHandler(w http.ResponseWriter, r *http.Request) {
	providerFilter := r.URL.Query().Get("provider")
	nameFilter := r.URL.Query().Get("name")

	if data, ok := s.cache.Get(); ok {
		filtered := s.filterSnapshots(data, providerFilter, nameFilter)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		json.NewEncoder(w).Encode(filtered)
		return
	}

	var filteredSubs []provider.SubscriptionEntry
	for _, sub := range s.config.Subscriptions {
		if providerFilter != "" && sub.Provider != providerFilter {
			continue
		}
		if nameFilter != "" && sub.Name != nameFilter {
			continue
		}
		filteredSubs = append(filteredSubs, sub)
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.config.Settings.Timeout)
	defer cancel()

	snapshots := s.registry.FetchAll(ctx, filteredSubs)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	json.NewEncoder(w).Encode(snapshots)
}

func (s *Server) filterSnapshots(snapshots []provider.UsageSnapshot, providerFilter, nameFilter string) []provider.UsageSnapshot {
	if providerFilter == "" && nameFilter == "" {
		return snapshots
	}

	var filtered []provider.UsageSnapshot
	for _, snap := range snapshots {
		if providerFilter != "" && snap.ProviderID != providerFilter {
			continue
		}
		if nameFilter != "" && snap.Name != nameFilter {
			continue
		}
		filtered = append(filtered, snap)
	}
	return filtered
}

func (s *Server) providersHandler(w http.ResponseWriter, r *http.Request) {
	providers := s.registry.All()
	providerInfo := make([]map[string]interface{}, len(providers))

	for i, p := range providers {
		providerInfo[i] = map[string]interface{}{
			"id":           p.ID(),
			"display_name": p.DisplayName(),
			"capabilities": p.Capabilities(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(providerInfo)
}

func (s *Server) registerHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/health", s.healthHandler)
	mux.HandleFunc("/api/v1/usage", s.usageHandler)
	mux.HandleFunc("/api/v1/providers", s.providersHandler)
}
