package provider

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

func (r *Registry) Register(p Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[p.ID()]; exists {
		return fmt.Errorf("provider with ID '%s' already registered", p.ID())
	}
	r.providers[p.ID()] = p
	return nil
}

func (r *Registry) Unregister(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.providers, id)
}

func (r *Registry) Get(id string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[id]
	return p, ok
}

func (r *Registry) All() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ps := make([]Provider, 0, len(r.providers))
	for _, p := range r.providers {
		ps = append(ps, p)
	}

	sort.Slice(ps, func(i, j int) bool {
		return ps[i].ID() < ps[j].ID()
	})

	return ps
}

func (r *Registry) FetchAll(ctx context.Context, entries []SubscriptionEntry) []UsageSnapshot {
	results := make([]UsageSnapshot, len(entries))
	var wg sync.WaitGroup

	for i, entry := range entries {
		wg.Add(1)
		go func(idx int, e SubscriptionEntry) {
			defer wg.Done()

			if ctx.Err() != nil {
				results[idx] = UsageSnapshot{
					ProviderID: e.Provider,
					Name:       e.Name,
					Status:     StatusError,
					Error:      ctx.Err().Error(),
				}
				return
			}

			p, ok := r.Get(e.Provider)
			if !ok {
				results[idx] = UsageSnapshot{
					ProviderID: e.Provider,
					Name:       e.Name,
					Status:     StatusError,
					Error:      fmt.Sprintf("provider %q not registered", e.Provider),
				}
				return
			}

			snap, err := p.FetchUsage(ctx, e.Auth)
			if err != nil {
				results[idx] = UsageSnapshot{
					ProviderID:  e.Provider,
					DisplayName: p.DisplayName(),
					Name:        e.Name,
					Status:      StatusError,
					Error:       err.Error(),
				}
				return
			}
			snap.Name = e.Name
			results[idx] = *snap
		}(i, entry)
	}

	wg.Wait()
	return results
}
