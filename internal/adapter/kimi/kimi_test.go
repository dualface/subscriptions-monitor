package kimi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/subscriptions-monitor/internal/provider"
)

func TestClient_GetUsages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/kimi.gateway.billing.v1.BillingService/GetUsages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("unexpected method: %s", r.Method)
		}

		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-auth-token" {
			t.Errorf("unexpected auth header: %s", auth)
		}

		resp := UsagesResponse{
			Usages: []Usage{
				{
					Scope: "FEATURE_CODING",
					Detail: UsageDetail{
						Limit:     "100",
						Used:      "6",
						Remaining: "94",
						ResetTime: time.Now().Add(24 * time.Hour),
					},
					Limits: []LimitInfo{
						{
							Window: struct {
								Duration int    `json:"duration"`
								TimeUnit string `json:"timeUnit"`
							}{
								Duration: 300,
								TimeUnit: "TIME_UNIT_MINUTE",
							},
							Detail: UsageDetail{
								Limit:     "100",
								Used:      "17",
								Remaining: "83",
								ResetTime: time.Now().Add(5 * time.Minute),
							},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test-auth-token", "test-cookie")
	client.baseURL = server.URL
	client.Debug = false

	ctx := context.Background()
	result, err := client.GetUsages(ctx)

	if err != nil {
		t.Fatalf("GetUsages failed: %v", err)
	}

	if len(result.Usages) == 0 {
		t.Fatal("expected usages")
	}

	usage := result.Usages[0]
	if usage.Scope != "FEATURE_CODING" {
		t.Errorf("expected scope FEATURE_CODING, got %s", usage.Scope)
	}

	if usage.Detail.Used != "6" {
		t.Errorf("expected used 6, got %s", usage.Detail.Used)
	}

	t.Logf("Daily used: %s/%s", usage.Detail.Used, usage.Detail.Limit)
	t.Logf("Window used: %s/%s", usage.Limits[0].Detail.Used, usage.Limits[0].Detail.Limit)
}

func TestAdapter_Interface(t *testing.T) {
	adapter := New()

	var _ provider.Provider = adapter

	t.Logf("Adapter ID: %s", adapter.ID())
	t.Logf("Adapter DisplayName: %s", adapter.DisplayName())

	caps := adapter.Capabilities()
	if !caps.SupportsUsageMetrics {
		t.Error("Should support usage metrics")
	}
}

func TestAdapter_ValidateAuth_MissingCredentials(t *testing.T) {
	adapter := New()

	auth := provider.AuthConfig{
		Type:  provider.AuthCookie,
		Extra: map[string]string{},
	}

	ctx := context.Background()
	err := adapter.ValidateAuth(ctx, auth)

	if err == nil {
		t.Error("expected error for missing credentials")
	}
}
