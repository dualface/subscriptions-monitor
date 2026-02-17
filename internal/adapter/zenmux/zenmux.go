package zenmux

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
	return "zenmux"
}

func (a *Adapter) DisplayName() string {
	return "ZenMux"
}

func (a *Adapter) Capabilities() provider.Capabilities {
	return provider.Capabilities{
		SupportsUsageMetrics:  true,
		SupportsCostBreakdown: true,
		SupportsCostByModel:   false,
		AuthTypes:             []provider.AuthType{provider.AuthCookie},
	}
}

func (a *Adapter) ValidateAuth(ctx context.Context, auth provider.AuthConfig) error {
	ctoken := auth.Extra["ctoken"]
	sessionID := auth.Extra["session_id"]
	sessionIDSig := auth.Extra["session_id_sig"]

	if ctoken == "" || sessionID == "" {
		return fmt.Errorf("zenmux requires ctoken and session_id in auth.extra")
	}

	client := NewClientWithSig(ctoken, sessionID, sessionIDSig)
	client.Debug = false

	_, err := client.GetCurrentSubscription(ctx)
	return err
}

func (a *Adapter) FetchUsage(ctx context.Context, auth provider.AuthConfig) (*provider.UsageSnapshot, error) {
	ctoken := auth.Extra["ctoken"]
	sessionID := auth.Extra["session_id"]
	sessionIDSig := auth.Extra["session_id_sig"]

	if ctoken == "" || sessionID == "" {
		return nil, fmt.Errorf("zenmux requires ctoken and session_id in auth.extra")
	}

	client := NewClientWithSig(ctoken, sessionID, sessionIDSig)
	client.Debug = false

	subResp, err1 := client.GetCurrentSubscription(ctx)
	usageResp, err2 := client.GetCurrentUsage(ctx)
	summaryResp, err3 := client.GetSubscriptionSummary(ctx)

	snap := &provider.UsageSnapshot{
		ProviderID:  a.ID(),
		DisplayName: a.DisplayName(),
		Timestamp:   time.Now(),
		Status:      provider.StatusOK,
	}

	var errors []string
	if err1 != nil {
		errors = append(errors, fmt.Sprintf("get_current: %v", err1))
	}
	if err2 != nil {
		errors = append(errors, fmt.Sprintf("get_current_usage: %v", err2))
	}
	if err3 != nil {
		errors = append(errors, fmt.Sprintf("subscription_summary: %v", err3))
	}

	if len(errors) == 3 {
		return nil, fmt.Errorf("all endpoints failed: %v", errors)
	}
	if len(errors) > 0 {
		snap.Status = provider.StatusError
		snap.Error = fmt.Sprintf("partial data: %v", errors)
	}

	if subResp != nil && subResp.Data != nil {
		snap.Plan = &provider.PlanInfo{
			Name: subResp.Data.Name,
			Type: "subscription",
		}
		if !subResp.Data.ExpiredAt.IsZero() {
			snap.Plan.RenewsAt = &subResp.Data.ExpiredAt
		}
	}

	var metrics []provider.UsageMetric

	if usageResp != nil {
		for _, item := range usageResp.Data {
			var metricName, windowID, windowLabel string

			switch item.PeriodType {
			case "hour_5":
				metricName = "5h Flows"
				windowID = "5h"
				windowLabel = "5 Hour"
			case "week":
				metricName = "7d Flows"
				windowID = "7d"
				windowLabel = "7 Day"
			default:
				metricName = item.PeriodType
				windowID = item.PeriodType
				windowLabel = item.PeriodType
			}

			quota := parseQuota(subResp.Data.Desc)

			used := float64(quota) * item.UsedRate
			remaining := float64(quota) - used

			metrics = append(metrics, provider.UsageMetric{
				Name: metricName,
				Window: provider.UsageWindow{
					ID:       windowID,
					Label:    windowLabel,
					ResetsAt: &item.CycleEndTime,
				},
				Amount: provider.UsageAmount{
					Used:      provider.Ptr(used),
					Limit:     provider.Ptr(float64(quota)),
					Remaining: provider.Ptr(remaining),
					Unit:      "flows",
				},
			})
		}
	}

	if summaryResp != nil && summaryResp.Data != nil {
		data := summaryResp.Data

		if totalTokens, err := strconv.ParseInt(data.TotalTokens, 10, 64); err == nil {
			metrics = append(metrics, provider.UsageMetric{
				Name: "Total Tokens",
				Window: provider.UsageWindow{
					ID:    "total",
					Label: "Total",
				},
				Amount: provider.UsageAmount{
					Used: provider.Ptr(float64(totalTokens)),
					Unit: "tokens",
				},
			})
		}

		if reqCounts, err := strconv.ParseInt(data.RequestCounts, 10, 64); err == nil {
			metrics = append(metrics, provider.UsageMetric{
				Name: "API Requests",
				Window: provider.UsageWindow{
					ID:    "total",
					Label: "Total",
				},
				Amount: provider.UsageAmount{
					Used: provider.Ptr(float64(reqCounts)),
					Unit: "requests",
				},
			})
		}

		snap.Cost = &provider.CostBreakdown{
			Currency: "USD",
		}
		if totalCost, err := strconv.ParseFloat(data.TotalCost, 64); err == nil {
			snap.Cost.Total = totalCost
		}
	}

	snap.Metrics = metrics
	return snap, nil
}

func parseQuota(desc string) int {
	var num int
	fmt.Sscanf(desc, "%d", &num)
	if num == 0 {
		return 1200
	}
	return num
}
