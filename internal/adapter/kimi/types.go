package kimi

import "time"

// UsagesResponse Kimi Code 用量响应
type UsagesResponse struct {
	Usages []Usage `json:"usages"`
}

type Usage struct {
	Scope  string      `json:"scope"`
	Detail UsageDetail `json:"detail"`
	Limits []LimitInfo `json:"limits"`
}

type UsageDetail struct {
	Limit     string    `json:"limit"`
	Used      string    `json:"used"`
	Remaining string    `json:"remaining"`
	ResetTime time.Time `json:"resetTime"`
}

type LimitInfo struct {
	Window struct {
		Duration int    `json:"duration"`
		TimeUnit string `json:"timeUnit"`
	} `json:"window"`
	Detail UsageDetail `json:"detail"`
}

// BalanceResponse 保留用于兼容性
type BalanceResponse struct {
	Data struct {
		TotalBalance    float64 `json:"total_balance"`
		GrantedBalance  float64 `json:"granted_balance"`
		ToppedUpBalance float64 `json:"topped_up_balance"`
	} `json:"data"`
	Object string `json:"object"`
}

type SubscriptionResponse struct {
	Subscription Subscription `json:"subscription"`
}

type Subscription struct {
	Goods SubscriptionGoods `json:"goods"`
}

type SubscriptionGoods struct {
	Title string `json:"title"`
}
