package kimi

import (
	"context"
	"fmt"
	"strconv"
	"strings"
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
		if !isCodingScope(usage.Scope) {
			continue
		}

		limit, _ := strconv.Atoi(string(usage.Detail.Limit))
		used, _ := strconv.Atoi(string(usage.Detail.Used))
		remaining, _ := strconv.Atoi(string(usage.Detail.Remaining))

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

		for _, limitInfo := range usage.Limits {
			duration := limitInfo.Window.Duration
			timeUnit := normalizeTimeUnit(string(limitInfo.Window.TimeUnit))

			windowLabel := fmt.Sprintf("%d %s", duration, timeUnit)
			if timeUnit == "TIME_UNIT_MINUTE" {
				windowLabel = fmt.Sprintf("%dm", duration)
			}

			windowLimit, _ := strconv.Atoi(string(limitInfo.Detail.Limit))
			windowUsed, _ := strconv.Atoi(string(limitInfo.Detail.Used))
			windowRemaining, _ := strconv.Atoi(string(limitInfo.Detail.Remaining))

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

	snap.Metrics = metrics
	return snap, nil
}

func isCodingScope(scope FlexibleString) bool {
	s := strings.TrimSpace(string(scope))
	if s == "" {
		return true
	}
	if s == "FEATURE_CODING" {
		return true
	}
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func normalizeTimeUnit(unit string) string {
	u := strings.TrimSpace(unit)
	switch u {
	case "5":
		return "TIME_UNIT_MINUTE"
	default:
		return u
	}
}
