package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/subscriptions-monitor/internal/provider"
)

func TestClient_GetCosts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/models" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"data": []map[string]string{{"id": "gpt-4.1-mini"}}})
			return
		}

		if r.URL.Path != "/v1/organization/costs" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("unexpected authorization header")
		}

		if r.URL.Query().Get("start_time") == "" || r.URL.Query().Get("end_time") == "" {
			t.Errorf("missing start_time/end_time query")
		}

		resp := costsResponse{
			Data: []costBucket{
				{
					StartTime: 1730419200,
					EndTime:   1730505600,
					Results: []costResult{
						{Amount: costAmount{Value: 1.23, Currency: "usd"}},
						{Amount: costAmount{Value: 0.77, Currency: "usd"}},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test-api-key", "org_123", "proj_123")
	client.baseURL = server.URL
	client.Debug = false

	ctx := context.Background()
	result, err := client.GetCosts(ctx, time.Unix(1730419200, 0), time.Unix(1730505600, 0))
	if err != nil {
		t.Fatalf("GetCosts failed: %v", err)
	}

	if len(result.Data) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(result.Data))
	}
}

func TestAdapter_Interface(t *testing.T) {
	adapter := New()
	var _ provider.Provider = adapter

	if adapter.ID() != "openai" {
		t.Errorf("unexpected ID: %s", adapter.ID())
	}

	caps := adapter.Capabilities()
	if !caps.SupportsCostBreakdown {
		t.Error("openai adapter should support cost breakdown")
	}
}

func TestAdapter_ValidateAuth_MissingKey(t *testing.T) {
	adapter := New()
	auth := provider.AuthConfig{Type: provider.AuthAPIKey}

	err := adapter.ValidateAuth(context.Background(), auth)
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
}

func TestClient_GetWhamUsage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/backend-api/wham/usage" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-bearer" {
			t.Fatalf("unexpected authorization header")
		}
		if r.Header.Get("Cookie") != "foo=bar" {
			t.Fatalf("unexpected cookie header")
		}

		payload := map[string]any{
			"limits": []map[string]any{{
				"name":      "GPT-5",
				"used":      12,
				"limit":     200,
				"remaining": 188,
			}},
			"chatgpt_plan_type": "plus",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	client := NewWebClient("test-bearer", "foo=bar", map[string]string{})
	client.webBaseURL = server.URL
	client.Debug = false

	resp, err := client.GetWhamUsage(context.Background())
	if err != nil {
		t.Fatalf("GetWhamUsage failed: %v", err)
	}

	if resp.Raw["chatgpt_plan_type"] != "plus" {
		t.Fatalf("expected plan type plus, got %v", resp.Raw["chatgpt_plan_type"])
	}
}

func TestBuildSnapshotFromWham(t *testing.T) {
	raw := map[string]any{
		"plan_type": "plus",
		"rate_limit": map[string]any{
			"allowed":       true,
			"limit_reached": false,
			"primary_window": map[string]any{
				"used_percent":         7,
				"limit_window_seconds": 18000,
				"reset_after_seconds":  15855,
				"reset_at":             1771518425,
			},
			"secondary_window": map[string]any{
				"used_percent":         2,
				"limit_window_seconds": 604800,
				"reset_after_seconds":  602655,
				"reset_at":             1772105225,
			},
		},
	}

	snap := buildSnapshotFromWham(raw, "openai", "OpenAI")
	if snap.Status != provider.StatusOK {
		t.Fatalf("expected status OK, got %s", snap.Status)
	}
	if len(snap.Metrics) < 2 {
		t.Fatalf("expected 5H and 7D metrics, got %d", len(snap.Metrics))
	}

	var found5H bool
	var found7D bool
	for _, m := range snap.Metrics {
		if m.Name == "5 Hour Usage" {
			found5H = true
			if m.Amount.Used == nil || *m.Amount.Used != 7 {
				t.Fatalf("expected 5H used percent 7, got %+v", m.Amount.Used)
			}
			if m.Amount.Remaining == nil || *m.Amount.Remaining != 93 {
				t.Fatalf("expected 5H remaining percent 93, got %+v", m.Amount.Remaining)
			}
		}
		if m.Name == "7 Day Usage" {
			found7D = true
			if m.Amount.Used == nil || *m.Amount.Used != 2 {
				t.Fatalf("expected 7D used percent 2, got %+v", m.Amount.Used)
			}
		}
	}
	if !found5H || !found7D {
		t.Fatal("expected both 5H and 7D window metrics")
	}

	if snap.Plan == nil || snap.Plan.Name == "" {
		t.Fatal("expected plan info from wham payload")
	}
}

func TestBuildSnapshotFromWham_LeafFallback(t *testing.T) {
	raw := map[string]any{
		"message_cap": map[string]any{
			"gpt5_remaining":   32,
			"gpt5_total_limit": 80,
		},
		"chatgpt_plan_type": "plus",
	}

	snap := buildSnapshotFromWham(raw, "openai", "OpenAI")
	if len(snap.Metrics) == 0 {
		t.Fatal("expected fallback leaf metrics")
	}

	foundRemaining := false
	for _, m := range snap.Metrics {
		if m.Amount.Used != nil && *m.Amount.Used == 32 {
			foundRemaining = true
			break
		}
	}
	if !foundRemaining {
		t.Fatal("expected metric parsed from leaf numeric usage fields")
	}
}
