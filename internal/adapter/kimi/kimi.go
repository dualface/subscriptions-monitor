package kimi

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/user/subscriptions-monitor/internal/provider"
)

type Adapter struct{}

func New() *Adapter {
	return &Adapter{}
}

func (a *Adapter) ID() string {
	return "kimi"
}

func (a *Adapter) DisplayName() string {
	return "Kimi Code"
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
	authToken := auth.Extra["auth_token"]
	cookie := auth.Extra["cookie"]

	if authToken == "" || cookie == "" {
		return fmt.Errorf("kimi requires auth_token and cookie in auth.extra")
	}

	client := NewClient(authToken, cookie)
	client.Debug = false

	_, err := client.GetUsages(ctx)
	return err
}

func (a *Adapter) FetchUsage(ctx context.Context, auth provider.AuthConfig) (*provider.UsageSnapshot, error) {
	authToken := auth.Extra["auth_token"]
	cookie := auth.Extra["cookie"]

	if authToken == "" || cookie == "" {
		return nil, fmt.Errorf("kimi requires auth_token and cookie in auth.extra")
	}

	client := NewClient(authToken, cookie)
	client.Debug = false

	usagesResp, usagesErr := client.GetUsages(ctx)
	subResp, subErr := client.GetSubscription(ctx)

	snap := &provider.UsageSnapshot{
		ProviderID:  a.ID(),
		DisplayName: a.DisplayName(),
		Timestamp:   time.Now(),
		Status:      provider.StatusOK,
	}

	if usagesErr != nil {
		return nil, fmt.Errorf("failed to fetch kimi usages: %v", usagesErr)
	}

	if subErr != nil {
		snap.Status = provider.StatusError
		snap.Error = fmt.Sprintf("failed to fetch subscription: %v", subErr)
	} else if subResp != nil {
		snap.Plan = &provider.PlanInfo{
			Name: subResp.Subscription.Goods.Title,
			Type: "subscription",
		}
	}

	var metrics []provider.UsageMetric

	for _, usage := range usagesResp.Usages {
		// 主要额度（日额度）
		if usage.Scope == "FEATURE_CODING" {
			limit, _ := strconv.Atoi(usage.Detail.Limit)
			used, _ := strconv.Atoi(usage.Detail.Used)
			remaining, _ := strconv.Atoi(usage.Detail.Remaining)

			metrics = append(metrics, provider.UsageMetric{
				Name: "Daily Requests",
				Window: provider.UsageWindow{
					ID:       "daily",
					Label:    "Daily",
					ResetsAt: &usage.Detail.ResetTime,
				},
				Amount: provider.UsageAmount{
					Used:      provider.Ptr(float64(used)),
					Limit:     provider.Ptr(float64(limit)),
					Remaining: provider.Ptr(float64(remaining)),
					Unit:      "requests",
				},
			})

			// 窗口限制（5分钟窗口）
			for _, limitInfo := range usage.Limits {
				duration := limitInfo.Window.Duration
				timeUnit := limitInfo.Window.TimeUnit

				windowLabel := fmt.Sprintf("%d %s", duration, timeUnit)
				if timeUnit == "TIME_UNIT_MINUTE" {
					windowLabel = fmt.Sprintf("%dm", duration)
				}

				windowLimit, _ := strconv.Atoi(limitInfo.Detail.Limit)
				windowUsed, _ := strconv.Atoi(limitInfo.Detail.Used)
				windowRemaining, _ := strconv.Atoi(limitInfo.Detail.Remaining)

				metrics = append(metrics, provider.UsageMetric{
					Name: fmt.Sprintf("Window (%s)", windowLabel),
					Window: provider.UsageWindow{
						ID:       "window",
						Label:    windowLabel,
						ResetsAt: &limitInfo.Detail.ResetTime,
					},
					Amount: provider.UsageAmount{
						Used:      provider.Ptr(float64(windowUsed)),
						Limit:     provider.Ptr(float64(windowLimit)),
						Remaining: provider.Ptr(float64(windowRemaining)),
						Unit:      "requests",
					},
				})
			}
		}
	}

	snap.Metrics = metrics
	return snap, nil
}
