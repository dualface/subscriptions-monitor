package openai

type whamUsageResponse struct {
	Raw map[string]any
}

type modelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

type costsResponse struct {
	Data []costBucket `json:"data"`
}

type costBucket struct {
	StartTime int64        `json:"start_time"`
	EndTime   int64        `json:"end_time"`
	Results   []costResult `json:"results"`
}

type costResult struct {
	Amount costAmount `json:"amount"`
}

type costAmount struct {
	Value    float64 `json:"value"`
	Currency string  `json:"currency"`
}
