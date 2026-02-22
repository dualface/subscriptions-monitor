package minimax

import (
	"context"
	"fmt"
	"time"

	"github.com/user/subscriptions-monitor/internal/provider"
)

type Adapter struct{}

func New() *Adapter {
	return &Adapter{}
}

func (a *Adapter) ID() string {
	return "minimax"
}

func (a *Adapter) DisplayName() string {
	return "MiniMax"
}

func (a *Adapter) Capabilities() provider.Capabilities {
	return provider.Capabilities{
		SupportsUsageMetrics:  true,
		SupportsCostBreakdown: false,
		SupportsCostByModel:   false,
		AuthTypes:             []provider.AuthType{provider.AuthCookie},
	}
}

func (a *Adapter) ValidateAuth(ctx context.Context, auth provider.AuthConfig) error {
	cookie := auth.Extra["cookie"]
	groupID := auth.Extra["group_id"]

	if cookie == "" || groupID == "" {
		return fmt.Errorf("minimax requires cookie and group_id in auth.extra")
	}

	client := NewClient(cookie, groupID)
	client.Debug = false

	_, err := client.GetCurrentSubscribe(ctx)
	return err
}

func (a *Adapter) FetchUsage(ctx context.Context, auth provider.AuthConfig) (*provider.UsageSnapshot, error) {
	cookie := auth.Extra["cookie"]
	groupID := auth.Extra["group_id"]

	if cookie == "" || groupID == "" {
		return nil, fmt.Errorf("minimax requires cookie and group_id in auth.extra")
	}

	client := NewClient(cookie, groupID)
	client.Debug = false

	subResp, err1 := client.GetCurrentSubscribe(ctx)
	remainsResp, err2 := client.GetRemains(ctx)

	snap := &provider.UsageSnapshot{
		ProviderID:  a.ID(),
		DisplayName: a.DisplayName(),
		Timestamp:   time.Now(),
		Status:      provider.StatusOK,
	}

	var errors []string
	if err1 != nil {
		errors = append(errors, fmt.Sprintf("get_subscribe: %v", err1))
	}
	if err2 != nil {
		errors = append(errors, fmt.Sprintf("get_remains: %v", err2))
	}

	if len(errors) == 2 {
		return nil, fmt.Errorf("all endpoints failed: %v", errors)
	}
	if len(errors) > 0 {
		snap.Status = provider.StatusError
		snap.Error = fmt.Sprintf("partial data: %v", errors)
	}

	if subResp != nil {
		snap.Plan = &provider.PlanInfo{
			Name: subResp.CurrentSubscribe.CurrentSubscribeTitle,
			Type: "subscription",
		}
	}

	var metrics []provider.UsageMetric

	if remainsResp != nil && len(remainsResp.ModelRemains) > 0 {
		for _, remain := range remainsResp.ModelRemains {
			remaining := remain.CurrentIntervalUsageCount
			total := remain.CurrentIntervalTotalCount
			used := total - remaining

			resetsAt := time.UnixMilli(remain.EndTime)

			metrics = append(metrics, provider.UsageMetric{
				Name: fmt.Sprintf("%s Usage", remain.ModelName),
				Window: provider.UsageWindow{
					ID:       "interval",
					Label:    "Current Interval",
					ResetsAt: &resetsAt,
				},
				Amount: provider.UsageAmount{
					Used:      provider.Ptr(float64(used)),
					Limit:     provider.Ptr(float64(total)),
					Remaining: provider.Ptr(float64(remaining)),
					Unit:      "requests",
				},
			})
		}
	}

	snap.Metrics = metrics
	return snap, nil
}
