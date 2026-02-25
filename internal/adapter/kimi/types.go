package kimi

import (
	"encoding/json"
	"strconv"
	"time"
)

type FlexibleString string

func (f *FlexibleString) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*f = ""
		return nil
	}

	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*f = FlexibleString(s)
		return nil
	}

	var num json.Number
	if err := json.Unmarshal(data, &num); err == nil {
		*f = FlexibleString(num.String())
		return nil
	}

	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		*f = FlexibleString(strconv.FormatBool(b))
		return nil
	}

	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*f = FlexibleString("")
	return nil
}

// UsagesResponse Kimi Code 用量响应
type UsagesResponse struct {
	Usages []Usage `json:"usages"`
}

type Usage struct {
	Scope  FlexibleString `json:"scope"`
	Detail UsageDetail    `json:"detail"`
	Limits []LimitInfo    `json:"limits"`
}

type UsageDetail struct {
	Limit     FlexibleString `json:"limit"`
	Used      FlexibleString `json:"used"`
	Remaining FlexibleString `json:"remaining"`
	ResetTime time.Time      `json:"resetTime"`
}

type LimitInfo struct {
	Window struct {
		Duration int            `json:"duration"`
		TimeUnit FlexibleString `json:"timeUnit"`
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
