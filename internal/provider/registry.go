package provider

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
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
				errMsg := ctx.Err().Error()
				results[idx] = UsageSnapshot{
					ProviderID: e.Provider,
					Name:       e.Name,
					Metrics:    []UsageMetric{},
					Status:     StatusError,
					Error:      errMsg,
				}
				logFetchWarning(e.Provider, e.Name, errMsg)
				return
			}

			p, ok := r.Get(e.Provider)
			if !ok {
				errMsg := fmt.Sprintf("provider %q not registered", e.Provider)
				results[idx] = UsageSnapshot{
					ProviderID: e.Provider,
					Name:       e.Name,
					Metrics:    []UsageMetric{},
					Status:     StatusError,
					Error:      errMsg,
				}
				logFetchWarning(e.Provider, e.Name, errMsg)
				return
			}

			snap, err := p.FetchUsage(ctx, e.Auth)
			if err != nil {
				errMsg := fmt.Sprintf("provider %q fetch failed: %v", e.Provider, err)
				results[idx] = UsageSnapshot{
					ProviderID:  e.Provider,
					DisplayName: p.DisplayName(),
					Name:        e.Name,
					Metrics:     []UsageMetric{},
					Status:      StatusError,
					Error:       errMsg,
				}
				logFetchWarning(e.Provider, e.Name, errMsg)
				return
			}
			snap.Name = e.Name
			if snap.Status == StatusError && snap.Error != "" {
				snap.Error = fmt.Sprintf("provider %q fetch failed: %s", e.Provider, snap.Error)
				logFetchWarning(e.Provider, e.Name, snap.Error)
			}
			if snap.Status == StatusError && snap.Metrics == nil {
				snap.Metrics = []UsageMetric{}
			}
			results[idx] = *snap
		}(i, entry)
	}

	wg.Wait()
	return results
}

func logFetchWarning(providerID, name, errMsg string) {
	errMsg = strings.TrimSpace(errMsg)
	if errMsg == "" {
		return
	}

	if name != "" {
		fmt.Fprintf(os.Stderr, "Warning: provider %q (%s) fetch failed: %s\n", providerID, name, errMsg)
		return
	}

	fmt.Fprintf(os.Stderr, "Warning: provider %q fetch failed: %s\n", providerID, errMsg)
}
