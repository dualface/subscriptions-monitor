package minimax

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/subscriptions-monitor/internal/provider"
)

func TestClient_GetCurrentSubscribe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/charge/combo/cycle_audio_resource_package" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		query := r.URL.Query()
		if query.Get("GroupId") != "test-group-id" {
			t.Errorf("unexpected group_id: %s", query.Get("GroupId"))
		}

		resp := CurrentSubscribeResponse{
			CurrentSubscribe: CurrentSubscribe{
				CurrentSubscribeTitle:   "CodePlanStarter-月度会员",
				CurrentSubscribePrice:   "",
				CurrentSubscribeCredit:  0,
				CurrentSubscribeEndTime: "02/23/2026",
				CurrSubscribeComboID:    "310001",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test-cookie", "test-group-id")
	client.baseURL = server.URL
	client.Debug = false

	ctx := context.Background()
	result, err := client.GetCurrentSubscribe(ctx)

	if err != nil {
		t.Fatalf("GetCurrentSubscribe failed: %v", err)
	}

	if result.CurrentSubscribe.CurrentSubscribeTitle != "CodePlanStarter-月度会员" {
		t.Errorf("unexpected title: %s", result.CurrentSubscribe.CurrentSubscribeTitle)
	}

	t.Logf("Plan: %s", result.CurrentSubscribe.CurrentSubscribeTitle)
	t.Logf("End Time: %s", result.CurrentSubscribe.CurrentSubscribeEndTime)
}

func TestClient_GetRemains(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/coding_plan/remains" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := RemainsResponse{
			ModelRemains: []ModelRemain{
				{
					StartTime:                 1771293600000,
					EndTime:                   1771311600000,
					RemainsTime:               8717750,
					CurrentIntervalTotalCount: 600,
					CurrentIntervalUsageCount: 600,
					ModelName:                 "MiniMax-M2",
				},
			},
			BaseResp: BaseResp{
				StatusCode: 0,
				StatusMsg:  "success",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test-cookie", "test-group-id")
	client.baseURL = server.URL
	client.Debug = false

	ctx := context.Background()
	result, err := client.GetRemains(ctx)

	if err != nil {
		t.Fatalf("GetRemains failed: %v", err)
	}

	if len(result.ModelRemains) == 0 {
		t.Fatal("expected model remains")
	}

	remain := result.ModelRemains[0]
	if remain.ModelName != "MiniMax-M2" {
		t.Errorf("unexpected model name: %s", remain.ModelName)
	}

	t.Logf("Model: %s", remain.ModelName)
	t.Logf("Usage: %d/%d", remain.CurrentIntervalUsageCount, remain.CurrentIntervalTotalCount)
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
