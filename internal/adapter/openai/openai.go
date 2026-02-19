package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
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
	return "openai"
}

func (a *Adapter) DisplayName() string {
	return "OpenAI"
}

func (a *Adapter) Capabilities() provider.Capabilities {
	return provider.Capabilities{
		SupportsUsageMetrics:  true,
		SupportsCostBreakdown: true,
		SupportsCostByModel:   false,
		AuthTypes:             []provider.AuthType{provider.AuthAPIKey, provider.AuthCookie},
	}
}

func (a *Adapter) ValidateAuth(ctx context.Context, auth provider.AuthConfig) error {
	if isChatGPTWebAuth(auth) {
		client := NewWebClient(strings.TrimSpace(auth.Extra["bearer_token"]), strings.TrimSpace(auth.Extra["cookie"]), auth.Extra)
		client.Debug = false
		_, err := client.GetWhamUsage(ctx)
		return err
	}

	apiKey := strings.TrimSpace(auth.Key)
	if apiKey == "" {
		return fmt.Errorf("openai requires either auth.key (API key) or auth.extra.bearer_token + auth.extra.cookie")
	}

	client := NewClient(apiKey, auth.Extra["organization"], auth.Extra["project"])
	client.Debug = false

	_, err := client.GetModels(ctx)
	return err
}

func (a *Adapter) FetchUsage(ctx context.Context, auth provider.AuthConfig) (*provider.UsageSnapshot, error) {
	if isChatGPTWebAuth(auth) {
		client := NewWebClient(strings.TrimSpace(auth.Extra["bearer_token"]), strings.TrimSpace(auth.Extra["cookie"]), auth.Extra)
		client.Debug = false

		whamResp, err := client.GetWhamUsage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch chatgpt usage: %v", err)
		}

		return buildSnapshotFromWham(whamResp.Raw, a.ID(), a.DisplayName()), nil
	}

	apiKey := strings.TrimSpace(auth.Key)
	if apiKey == "" {
		return nil, fmt.Errorf("openai requires either auth.key (API key) or auth.extra.bearer_token + auth.extra.cookie")
	}

	client := NewClient(apiKey, auth.Extra["organization"], auth.Extra["project"])
	client.Debug = false

	start := time.Now().UTC().Add(-30 * 24 * time.Hour)
	end := time.Now().UTC()

	_, modelErr := client.GetModels(ctx)
	costsResp, costErr := client.GetCosts(ctx, start, end)

	if modelErr != nil && costErr != nil {
		return nil, fmt.Errorf("all endpoints failed: get_models: %v; get_costs: %v", modelErr, costErr)
	}

	snap := &provider.UsageSnapshot{
		ProviderID:  a.ID(),
		DisplayName: a.DisplayName(),
		Timestamp:   time.Now(),
		Status:      provider.StatusOK,
	}

	if costErr != nil {
		snap.Status = provider.StatusError
		snap.Error = fmt.Sprintf("partial data: get_costs: %v", costErr)
		snap.Metrics = []provider.UsageMetric{}
		return snap, nil
	}

	total := 0.0
	currency := "USD"
	byDay := make([]provider.DailyCost, 0, len(costsResp.Data))
	for _, bucket := range costsResp.Data {
		dayCost := 0.0
		for _, r := range bucket.Results {
			dayCost += r.Amount.Value
			total += r.Amount.Value
			if r.Amount.Currency != "" {
				currency = strings.ToUpper(r.Amount.Currency)
			}
		}

		if bucket.StartTime > 0 {
			day := time.Unix(bucket.StartTime, 0).UTC().Format("2006-01-02")
			byDay = append(byDay, provider.DailyCost{Date: day, Cost: dayCost})
		}
	}

	snap.Cost = &provider.CostBreakdown{
		Total:    total,
		Currency: currency,
		ByDay:    byDay,
		Period: provider.TimePeriod{
			Start: start,
			End:   end,
		},
	}

	snap.Metrics = []provider.UsageMetric{
		{
			Name: "Total Spend",
			Window: provider.UsageWindow{
				ID:    "rolling_30d",
				Label: "Last 30 Days",
			},
			Amount: provider.UsageAmount{
				Used: provider.Ptr(total),
				Unit: strings.ToLower(currency),
			},
		},
	}

	return snap, nil
}

func isChatGPTWebAuth(auth provider.AuthConfig) bool {
	return strings.TrimSpace(auth.Extra["bearer_token"]) != "" && strings.TrimSpace(auth.Extra["cookie"]) != ""
}

func buildSnapshotFromWham(raw map[string]any, providerID, displayName string) *provider.UsageSnapshot {
	now := time.Now()
	snap := &provider.UsageSnapshot{
		ProviderID:  providerID,
		DisplayName: displayName,
		Timestamp:   now,
		Status:      provider.StatusOK,
	}

	metrics := extractRateLimitWindowMetrics(raw)
	if len(metrics) == 0 {
		metrics = extractUsageMetrics(raw)
	}
	if len(metrics) == 0 {
		metrics = []provider.UsageMetric{
			{
				Name: "Usage Endpoint Reachable",
				Window: provider.UsageWindow{
					ID:    "current",
					Label: "Current",
				},
				Amount: provider.UsageAmount{
					Used: provider.Ptr(1.0),
					Unit: "status",
				},
			},
		}
	}
	snap.Metrics = metrics

	if planName := extractPlanName(raw); planName != "" {
		snap.Plan = &provider.PlanInfo{Name: planName, Type: "subscription"}
	}

	return snap
}

func extractRateLimitWindowMetrics(raw map[string]any) []provider.UsageMetric {
	rateLimit, ok := raw["rate_limit"].(map[string]any)
	if !ok {
		return nil
	}

	metrics := make([]provider.UsageMetric, 0, 2)
	metrics = append(metrics, buildRateLimitMetrics(rateLimit, "primary_window", "5h", "5 Hour")...)
	metrics = append(metrics, buildRateLimitMetrics(rateLimit, "secondary_window", "7d", "7 Day")...)

	return metrics
}

func buildRateLimitMetrics(rateLimit map[string]any, windowKey, windowID, windowLabel string) []provider.UsageMetric {
	window, ok := rateLimit[windowKey].(map[string]any)
	if !ok {
		return nil
	}

	usedPercent, ok := toFloat(window["used_percent"])
	if !ok {
		return nil
	}

	remainingPercent := 100.0 - usedPercent
	if remainingPercent < 0 {
		remainingPercent = 0
	}

	metric := provider.UsageMetric{
		Name: fmt.Sprintf("%s Usage", windowLabel),
		Window: provider.UsageWindow{
			ID:    windowID,
			Label: windowLabel,
		},
		Amount: provider.UsageAmount{
			Used:      provider.Ptr(usedPercent),
			Limit:     provider.Ptr(100.0),
			Remaining: provider.Ptr(remainingPercent),
			Unit:      "percent",
		},
	}

	if resetAt, ok := toFloat(window["reset_at"]); ok {
		resetTime := time.Unix(int64(resetAt), 0).UTC()
		metric.Window.ResetsAt = &resetTime
	} else if resetAfter, ok := toFloat(window["reset_after_seconds"]); ok {
		resetTime := time.Now().UTC().Add(time.Duration(resetAfter) * time.Second)
		metric.Window.ResetsAt = &resetTime
	}

	return []provider.UsageMetric{metric}
}

func extractUsageMetrics(raw map[string]any) []provider.UsageMetric {
	metrics := make([]provider.UsageMetric, 0)
	seen := map[string]bool{}

	var walk func(path []string, v any)
	walk = func(path []string, v any) {
		switch node := v.(type) {
		case map[string]any:
			used := findNumeric(node, "used", "usage", "consumed", "current_usage")
			limit := findNumeric(node, "limit", "quota", "max", "cap")
			remaining := findNumeric(node, "remaining", "left", "available")
			if used != nil || limit != nil || remaining != nil {
				name := findString(node, "name", "label", "title", "model", "plan")
				if name == "" {
					name = strings.TrimSpace(strings.Join(path, " / "))
				}
				if name == "" {
					name = "Usage"
				}
				id := normalizeID(name)
				if !seen[id] {
					seen[id] = true
					metric := provider.UsageMetric{
						Name: name,
						Window: provider.UsageWindow{
							ID:    id,
							Label: "Current",
						},
						Amount: provider.UsageAmount{
							Used:      used,
							Limit:     limit,
							Remaining: remaining,
							Unit:      strings.ToLower(valueOrDefault(findString(node, "unit", "currency"), "count")),
						},
					}
					if reset := findResetTime(node); reset != nil {
						metric.Window.ResetsAt = reset
					}
					metrics = append(metrics, metric)
				}
			}

			for k, child := range node {
				walk(append(path, k), child)
			}
		case []any:
			for idx, child := range node {
				walk(append(path, strconv.Itoa(idx)), child)
			}
		}
	}

	walk(nil, raw)

	if len(metrics) == 0 {
		metrics = append(metrics, extractLeafMetrics(raw, seen)...)
	}
	return metrics
}

func extractLeafMetrics(raw map[string]any, seen map[string]bool) []provider.UsageMetric {
	metrics := make([]provider.UsageMetric, 0)

	var walk func(path []string, v any)
	walk = func(path []string, v any) {
		switch node := v.(type) {
		case map[string]any:
			for k, child := range node {
				walk(append(path, k), child)
			}
		case []any:
			for idx, child := range node {
				walk(append(path, strconv.Itoa(idx)), child)
			}
		default:
			n, ok := toFloat(node)
			if !ok {
				return
			}

			key := ""
			if len(path) > 0 {
				key = strings.ToLower(path[len(path)-1])
			}
			joined := strings.ToLower(strings.Join(path, "/"))
			if !isUsageLikePath(joined, key) || isTimestampLikeKey(key) {
				return
			}

			name := prettifyPath(path)
			if name == "" {
				name = "Usage"
			}
			id := normalizeID(name)
			if seen[id] {
				return
			}
			seen[id] = true

			unit := "count"
			if strings.Contains(key, "cost") || strings.Contains(key, "spend") {
				unit = "usd"
			} else if strings.Contains(key, "token") {
				unit = "tokens"
			} else if strings.Contains(key, "request") || strings.Contains(key, "message") {
				unit = "requests"
			}

			metrics = append(metrics, provider.UsageMetric{
				Name: name,
				Window: provider.UsageWindow{
					ID:    id,
					Label: "Current",
				},
				Amount: provider.UsageAmount{
					Used: provider.Ptr(n),
					Unit: unit,
				},
			})
		}
	}

	walk(nil, raw)
	if len(metrics) > 8 {
		return metrics[:8]
	}
	return metrics
}

func extractCostSummary(raw map[string]any) (float64, string) {
	total := 0.0
	currency := "USD"
	var walk func(path []string, v any)
	walk = func(path []string, v any) {
		switch node := v.(type) {
		case map[string]any:
			for k, child := range node {
				if n, ok := toFloat(child); ok {
					lowerK := strings.ToLower(k)
					if strings.Contains(lowerK, "cost") || strings.Contains(lowerK, "spend") {
						total += n
					}
				}
				walk(append(path, k), child)
			}
			if c := findString(node, "currency", "unit"); c != "" && len(c) <= 4 {
				currency = strings.ToUpper(c)
			}
		case []any:
			for _, child := range node {
				walk(path, child)
			}
		}
	}
	walk(nil, raw)

	if math.IsNaN(total) || math.IsInf(total, 0) {
		total = 0
	}
	return total, currency
}

func extractPlanName(raw map[string]any) string {
	if v := findString(raw, "plan", "plan_type", "chatgpt_plan_type", "subscription", "tier"); v != "" {
		return formatPlanName(v)
	}

	if authMap, ok := raw["auth"].(map[string]any); ok {
		if v := findString(authMap, "chatgpt_plan_type"); v != "" {
			return formatPlanName(v)
		}
	}

	buf, err := json.Marshal(raw)
	if err != nil {
		return ""
	}
	var deep map[string]any
	if err := json.Unmarshal(buf, &deep); err != nil {
		return ""
	}
	return findStringDeep(deep, "chatgpt_plan_type")
}

func findNumeric(m map[string]any, keys ...string) *float64 {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if n, ok := toFloat(v); ok {
				return provider.Ptr(n)
			}
		}
	}
	return nil
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(n), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func findString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k].(string); ok {
			v = strings.TrimSpace(v)
			if v != "" {
				return v
			}
		}
	}
	return ""
}

func findStringDeep(v any, key string) string {
	switch node := v.(type) {
	case map[string]any:
		if s, ok := node[key].(string); ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
		for _, child := range node {
			if s := findStringDeep(child, key); s != "" {
				return s
			}
		}
	case []any:
		for _, child := range node {
			if s := findStringDeep(child, key); s != "" {
				return s
			}
		}
	}
	return ""
}

func findResetTime(m map[string]any) *time.Time {
	for _, key := range []string{"reset_at", "resets_at", "resetAt", "resetsAt", "reset_time", "resetTime"} {
		v, ok := m[key]
		if !ok {
			continue
		}
		if ts, ok := toFloat(v); ok {
			t := time.Unix(int64(ts), 0).UTC()
			return &t
		}
		if s, ok := v.(string); ok {
			if t, err := time.Parse(time.RFC3339, s); err == nil {
				return &t
			}
		}
	}
	return nil
}

func normalizeID(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "-", "_")
	if s == "" {
		return "usage"
	}
	return s
}

func isTimestampLikeKey(key string) bool {
	for _, kw := range []string{"time", "timestamp", "date", "expires", "expiry", "reset", "created", "updated"} {
		if strings.Contains(key, kw) {
			return true
		}
	}
	return false
}

func isUsageLikePath(joinedPath, key string) bool {
	keywords := []string{
		"usage", "used", "remaining", "remain", "left", "available", "quota", "limit", "cap",
		"count", "requests", "messages", "tokens", "spend", "cost", "credits", "balance", "percent",
	}
	for _, kw := range keywords {
		if strings.Contains(key, kw) || strings.Contains(joinedPath, kw) {
			return true
		}
	}
	return false
}

func prettifyPath(path []string) string {
	filtered := make([]string, 0, len(path))
	for _, p := range path {
		if _, err := strconv.Atoi(p); err == nil {
			continue
		}
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		filtered = append(filtered, p)
	}
	if len(filtered) == 0 {
		return ""
	}

	if len(filtered) > 3 {
		filtered = filtered[len(filtered)-3:]
	}

	for i, p := range filtered {
		p = strings.ReplaceAll(p, "_", " ")
		p = strings.ReplaceAll(p, "-", " ")
		parts := strings.Fields(strings.ToLower(p))
		for j, part := range parts {
			if len(part) > 0 {
				parts[j] = strings.ToUpper(part[:1]) + part[1:]
			}
		}
		filtered[i] = strings.Join(parts, " ")
	}

	return strings.Join(filtered, " / ")
}

func formatPlanName(v string) string {
	v = strings.ReplaceAll(strings.TrimSpace(v), "_", " ")
	parts := strings.Fields(strings.ToLower(v))
	for i, p := range parts {
		if len(p) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}
