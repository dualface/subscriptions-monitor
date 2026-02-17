package zenmux

import "time"

// AllPlansResponse
type AllPlansResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    []Plan `json:"data"`
}

type Plan struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	DisplayName string     `json:"display_name"`
	Price       float64    `json:"price"`
	Currency    string     `json:"currency"`
	Features    []string   `json:"features"`
	Limits      PlanLimits `json:"limits"`
}

type PlanLimits struct {
	FlowsPerWindow int `json:"flows_per_window"`
	WindowHours    int `json:"window_hours"`
}

// CurrentSubscriptionResponse - 实际 API 格式
type CurrentSubscriptionResponse struct {
	Success bool                     `json:"success"`
	Data    *CurrentSubscriptionData `json:"data"`
}

type CurrentSubscriptionData struct {
	Price              float64   `json:"price"`
	StartedAt          time.Time `json:"startedAt"`
	ExpiredAt          time.Time `json:"expiredAt"`
	Status             string    `json:"status"`
	PlanKey            string    `json:"planKey"`
	NextBillingPlanKey *string   `json:"nextBillingPlanKey"`
	EnableExtraUsage   int       `json:"enable_extra_usage"`
	ExtraAPIKey        *string   `json:"extra_api_key"`
	Name               string    `json:"name"`
	Desc               string    `json:"desc"`
}

// CurrentUsageResponse - 实际 API 格式
type CurrentUsageResponse struct {
	Success bool        `json:"success"`
	Data    []UsageItem `json:"data"`
}

type UsageItem struct {
	TierCode       string    `json:"tierCode"`
	PeriodType     string    `json:"periodType"`
	PeriodDuration string    `json:"periodDuration"`
	CycleStartTime time.Time `json:"cycleStartTime"`
	CycleEndTime   time.Time `json:"cycleEndTime"`
	UsedRate       float64   `json:"usedRate"`
	QuotaStatus    int       `json:"quotaStatus"`
	Status         int       `json:"status"`
	Ext            *string   `json:"ext"`
}

// SubscriptionSummaryResponse - 实际 API 格式 (字符串数字)
type SubscriptionSummaryResponse struct {
	Success bool         `json:"success"`
	Data    *SummaryData `json:"data"`
}

type SummaryData struct {
	TotalCost           string `json:"totalCost"`
	InputCost           string `json:"inputCost"`
	OutputCost          string `json:"outputCost"`
	OtherCost           string `json:"otherCost"`
	RequestCounts       string `json:"requestCounts"`
	RequestAvgCost      string `json:"requestAvgCost"`
	TotalTokens         string `json:"totalTokens"`
	MillionTokenAvgCost string `json:"millionTokenAvgCost"`
}

// 别名保持兼容性
type SubscriptionDetail = CurrentSubscriptionData
type UsageDetail = UsageItem
type SummaryDetail = SummaryData
type ModelSummary struct {
	Model        string  `json:"model"`
	Cost         float64 `json:"cost"`
	Tokens       int64   `json:"tokens"`
	InputTokens  int64   `json:"input_tokens"`
	OutputTokens int64   `json:"output_tokens"`
	Requests     int     `json:"requests"`
}
