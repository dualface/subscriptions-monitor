package provider

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

type mockProvider struct {
	id          string
	displayName string
	failFetch   bool
}

func (m *mockProvider) ID() string {
	return m.id
}

func (m *mockProvider) DisplayName() string {
	return m.displayName
}

func (m *mockProvider) Capabilities() Capabilities {
	return Capabilities{SupportsUsageMetrics: true}
}

func (m *mockProvider) ValidateAuth(ctx context.Context, auth AuthConfig) error {
	if m.failFetch {
		return errors.New("auth validation failed")
	}
	return nil
}

func (m *mockProvider) FetchUsage(ctx context.Context, auth AuthConfig) (*UsageSnapshot, error) {
	if m.failFetch {
		return nil, errors.New("failed to fetch usage")
	}
	return &UsageSnapshot{
		ProviderID:  m.id,
		DisplayName: m.displayName,
		Name:        "test-sub",
		Status:      StatusOK,
	}, nil
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()
	p1 := &mockProvider{id: "mock1"}

	err := r.Register(p1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = r.Register(p1)
	if err == nil {
		t.Fatal("expected error for duplicate registration, got nil")
	}
}

func TestRegistry_Get(t *testing.T) {
	r := NewRegistry()
	p1 := &mockProvider{id: "mock1"}
	r.Register(p1)

	retrieved, ok := r.Get("mock1")
	if !ok {
		t.Fatal("expected to find provider, but didn't")
	}
	if retrieved.ID() != "mock1" {
		t.Errorf("expected provider ID 'mock1', got '%s'", retrieved.ID())
	}

	_, ok = r.Get("non-existent")
	if ok {
		t.Fatal("expected not to find provider, but did")
	}
}

func TestRegistry_All(t *testing.T) {
	r := NewRegistry()
	p1 := &mockProvider{id: "mock-b"}
	p2 := &mockProvider{id: "mock-a"}
	r.Register(p1)
	r.Register(p2)

	all := r.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(all))
	}

	expectedOrder := []string{"mock-a", "mock-b"}
	ids := []string{all[0].ID(), all[1].ID()}
	if !reflect.DeepEqual(ids, expectedOrder) {
		t.Errorf("expected sorted IDs %v, got %v", expectedOrder, ids)
	}
}

func TestRegistry_Unregister(t *testing.T) {
	r := NewRegistry()
	p1 := &mockProvider{id: "mock1"}
	r.Register(p1)

	r.Unregister("mock1")

	_, ok := r.Get("mock1")
	if ok {
		t.Fatal("expected provider to be unregistered, but it was found")
	}
}

func TestRegistry_FetchAll(t *testing.T) {
	r := NewRegistry()
	pSuccess := &mockProvider{id: "success-provider", displayName: "Success!"}
	pFail := &mockProvider{id: "fail-provider", failFetch: true}
	r.Register(pSuccess)
	r.Register(pFail)

	entries := []SubscriptionEntry{
		{Provider: "success-provider", Name: "test-sub"},
		{Provider: "fail-provider", Name: "test-sub-fail"},
		{Provider: "not-registered", Name: "test-sub-not-found"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	snapshots := r.FetchAll(ctx, entries)

	// Should return one result per entry, preserving order
	if len(snapshots) != 3 {
		t.Fatalf("expected 3 snapshots, got %d", len(snapshots))
	}

	// First: success
	if snapshots[0].Status != StatusOK {
		t.Errorf("expected status OK for success-provider, got %s", snapshots[0].Status)
	}
	if snapshots[0].ProviderID != "success-provider" {
		t.Errorf("expected provider 'success-provider', got '%s'", snapshots[0].ProviderID)
	}
	if snapshots[0].Name != "test-sub" {
		t.Errorf("expected name 'test-sub', got '%s'", snapshots[0].Name)
	}

	// Second: fetch error
	if snapshots[1].Status != StatusError {
		t.Errorf("expected status Error for fail-provider, got %s", snapshots[1].Status)
	}
	if snapshots[1].Error == "" {
		t.Error("expected non-empty error for fail-provider")
	}

	// Third: not registered
	if snapshots[2].Status != StatusError {
		t.Errorf("expected status Error for not-registered, got %s", snapshots[2].Status)
	}
	if snapshots[2].Error == "" {
		t.Error("expected non-empty error for not-registered provider")
	}
}
