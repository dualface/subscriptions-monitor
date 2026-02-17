package minimax

// CurrentSubscribeResponse 当前订阅响应
type CurrentSubscribeResponse struct {
	CurrentSubscribe CurrentSubscribe `json:"current_subscribe"`
}

type CurrentSubscribe struct {
	ButtonText                string `json:"button_text"`
	CurrentSubscribeTitle     string `json:"current_subscribe_title"`
	CurrentSubscribePrice     string `json:"current_subscribe_price"`
	CurrentSubscribeCredit    int    `json:"current_subscribe_credit"`
	CurrentSubscribeEndTime   string `json:"current_subscribe_end_time"`
	CurrentSubscribeComboType int    `json:"current_subscribe_combo_type"`
	CurrSubscribeComboID      string `json:"curr_subscribe_combo_id"`
	CurrentSubscribeCycleType int    `json:"current_subscribe_cycle_type"`
	CurrentCreditReloadTime   string `json:"current_credit_reload_time"`
	NextSubscribeTitle        string `json:"next_subscribe_title"`
	RenewalDate               string `json:"renewal_date"`
	NextSubscribeComboType    int    `json:"next_subscribe_combo_type"`
	RenewalState              int    `json:"renewal_state"`
	NxtSubscribeComboID       string `json:"nxt_subscribe_combo_id"`
	NextSubscribeCycleType    int    `json:"next_subscribe_cycle_type"`
}

// RemainsResponse 用量响应
type RemainsResponse struct {
	ModelRemains []ModelRemain `json:"model_remains"`
	BaseResp     BaseResp      `json:"base_resp"`
}

type ModelRemain struct {
	StartTime                 int64  `json:"start_time"`
	EndTime                   int64  `json:"end_time"`
	RemainsTime               int64  `json:"remains_time"`
	CurrentIntervalTotalCount int    `json:"current_interval_total_count"`
	CurrentIntervalUsageCount int    `json:"current_interval_usage_count"`
	ModelName                 string `json:"model_name"`
}

type BaseResp struct {
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}
