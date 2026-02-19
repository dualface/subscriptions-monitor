package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	baseURL    = "https://api.openai.com"
	webBaseURL = "https://chatgpt.com"
)

type Client struct {
	httpClient   *http.Client
	baseURL      string
	webBaseURL   string
	apiKey       string
	bearerToken  string
	cookie       string
	organization string
	project      string
	deviceID     string
	clientBuild  string
	clientVer    string
	language     string
	userAgent    string
	referer      string
	Debug        bool
}

func NewClient(apiKey, organization, project string) *Client {
	return &Client{
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		baseURL:      baseURL,
		webBaseURL:   webBaseURL,
		apiKey:       apiKey,
		organization: organization,
		project:      project,
		Debug:        true,
	}
}

func NewWebClient(bearerToken, cookie string, extra map[string]string) *Client {
	return &Client{
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		baseURL:     baseURL,
		webBaseURL:  webBaseURL,
		bearerToken: bearerToken,
		cookie:      cookie,
		deviceID:    extra["oai_device_id"],
		clientBuild: extra["oai_client_build_number"],
		clientVer:   extra["oai_client_version"],
		language:    valueOrDefault(extra["oai_language"], "en-US"),
		userAgent:   extra["user_agent"],
		referer:     valueOrDefault(extra["referer"], "https://chatgpt.com/codex/settings/usage"),
		Debug:       true,
	}
}

func (c *Client) logRequest(req *http.Request) {
	if !c.Debug {
		return
	}
	dump, _ := httputil.DumpRequestOut(req, false)
	fmt.Println("=== HTTP REQUEST ===")
	fmt.Println(string(dump))
	fmt.Println("===================")
}

func (c *Client) logResponse(resp *http.Response, body []byte) {
	if !c.Debug {
		return
	}
	fmt.Printf("=== HTTP RESPONSE ===\n")
	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Body:\n%s\n", string(body))
	fmt.Println("====================")
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	if c.organization != "" {
		req.Header.Set("OpenAI-Organization", c.organization)
	}
	if c.project != "" {
		req.Header.Set("OpenAI-Project", c.project)
	}
}

func (c *Client) setWebHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Cookie", c.cookie)
	req.Header.Set("Referer", c.referer)

	if c.language != "" {
		req.Header.Set("Accept-Language", c.language)
		req.Header.Set("oai-language", c.language)
	}
	if c.deviceID != "" {
		req.Header.Set("oai-device-id", c.deviceID)
	}
	if c.clientBuild != "" {
		req.Header.Set("oai-client-build-number", c.clientBuild)
	}
	if c.clientVer != "" {
		req.Header.Set("oai-client-version", c.clientVer)
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
}

func (c *Client) doRequest(ctx context.Context, method, urlStr string, headerSetter func(*http.Request)) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, urlStr, nil)
	if err != nil {
		return nil, err
	}
	headerSetter(req)
	c.logRequest(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	c.logResponse(resp, body)

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return body, nil
}

func (c *Client) GetModels(ctx context.Context) (*modelsResponse, error) {
	body, err := c.doRequest(ctx, http.MethodGet, c.baseURL+"/v1/models", c.setHeaders)
	if err != nil {
		return nil, err
	}

	var result modelsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetCosts(ctx context.Context, start, end time.Time) (*costsResponse, error) {
	u, _ := url.Parse(c.baseURL + "/v1/organization/costs")
	q := u.Query()
	q.Set("start_time", strconv.FormatInt(start.Unix(), 10))
	q.Set("end_time", strconv.FormatInt(end.Unix(), 10))
	u.RawQuery = q.Encode()

	body, err := c.doRequest(ctx, http.MethodGet, u.String(), c.setHeaders)
	if err != nil {
		return nil, err
	}

	var result costsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetWhamUsage(ctx context.Context) (*whamUsageResponse, error) {
	body, err := c.doRequest(ctx, http.MethodGet, c.webBaseURL+"/backend-api/wham/usage", c.setWebHeaders)
	if err != nil {
		return nil, err
	}

	var result whamUsageResponse
	if err := json.Unmarshal(body, &result.Raw); err != nil {
		return nil, err
	}
	return &result, nil
}

func valueOrDefault(v, fallback string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return fallback
	}
	return v
}
