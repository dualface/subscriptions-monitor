package provider

import "context"

type Provider interface {
	ID() string
	DisplayName() string
	Capabilities() Capabilities
	ValidateAuth(ctx context.Context, auth AuthConfig) error
	FetchUsage(ctx context.Context, auth AuthConfig) (*UsageSnapshot, error)
}

type CostProvider interface {
	Provider
	FetchCosts(ctx context.Context, auth AuthConfig, period TimePeriod) (*CostBreakdown, error)
}
