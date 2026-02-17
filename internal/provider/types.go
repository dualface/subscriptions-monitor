package provider

import "time"

type AuthType string

const (
	AuthAPIKey AuthType = "api_key"
	AuthOAuth  AuthType = "oauth"
	AuthCookie AuthType = "cookie"
)

type AuthConfig struct {
	Type  AuthType          `yaml:"type" json:"type"`
	Key   string            `yaml:"key" json:"-"`
	Extra map[string]string `yaml:"extra,omitempty" json:"-"`
}

type Status string

const (
	StatusOK           Status = "ok"
	StatusError        Status = "error"
	StatusUnauthorized Status = "unauthorized"
)

type UsageSnapshot struct {
	ProviderID  string         `json:"provider_id"`
	DisplayName string         `json:"display_name"`
	Name        string         `json:"name"`
	Timestamp   time.Time      `json:"timestamp"`
	Plan        *PlanInfo      `json:"plan,omitempty"`
	Metrics     []UsageMetric  `json:"metrics"`
	Cost        *CostBreakdown `json:"cost,omitempty"`
	Status      Status         `json:"status"`
	Error       string         `json:"error,omitempty"`
}

type UsageMetric struct {
	Name   string      `json:"name"`
	Window UsageWindow `json:"window"`
	Amount UsageAmount `json:"amount"`
}

type UsageWindow struct {
	ID       string     `json:"id"`
	Label    string     `json:"label"`
	ResetsAt *time.Time `json:"resets_at,omitempty"`
}

type UsageAmount struct {
	Used      *float64 `json:"used,omitempty"`
	Limit     *float64 `json:"limit,omitempty"`
	Remaining *float64 `json:"remaining,omitempty"`
	Unit      string   `json:"unit"`
}

type PlanInfo struct {
	Name     string     `json:"name"`
	Type     string     `json:"type"`
	RenewsAt *time.Time `json:"renews_at,omitempty"`
}

type CostBreakdown struct {
	Total    float64            `json:"total"`
	Currency string             `json:"currency"`
	ByModel  map[string]float64 `json:"by_model,omitempty"`
	ByDay    []DailyCost        `json:"by_day,omitempty"`
	Period   TimePeriod         `json:"period"`
}

type DailyCost struct {
	Date string  `json:"date"`
	Cost float64 `json:"cost"`
}

type TimePeriod struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type Capabilities struct {
	SupportsUsageMetrics  bool       `json:"supports_usage_metrics"`
	SupportsCostBreakdown bool       `json:"supports_cost_breakdown"`
	SupportsCostByModel   bool       `json:"supports_cost_by_model"`
	AuthTypes             []AuthType `json:"auth_types"`
}

// SubscriptionEntry represents a configured subscription in the config file
type SubscriptionEntry struct {
	Provider string     `yaml:"provider" json:"provider"`
	Name     string     `yaml:"name" json:"name"`
	Auth     AuthConfig `yaml:"auth" json:"auth"`
}

// Ptr is a helper to create pointer to a value
func Ptr[T any](v T) *T {
	return &v
}
