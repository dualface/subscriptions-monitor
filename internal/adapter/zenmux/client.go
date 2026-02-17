package zenmux

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

var (
	baseURL = "https://zenmux.ai"
)

type Client struct {
	httpClient   *http.Client
	baseURL      string
	ctoken       string
	sessionID    string
	sessionIDSig string
	Debug        bool
}

func NewClient(ctoken, sessionID string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
		ctoken:     ctoken,
		sessionID:  sessionID,
		Debug:      true,
	}
}

func NewClientWithSig(ctoken, sessionID, sessionIDSig string) *Client {
	return &Client{
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		baseURL:      baseURL,
		ctoken:       ctoken,
		sessionID:    sessionID,
		sessionIDSig: sessionIDSig,
		Debug:        true,
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
	fmt.Printf("Headers:\n")
	for k, v := range resp.Header {
		fmt.Printf("  %s: %v\n", k, v)
	}
	fmt.Printf("Body:\n%s\n", string(body))
	fmt.Println("====================")
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Referer", "https://zenmux.ai/platform/subscription")
}

func (c *Client) setCookies(req *http.Request) {
	req.AddCookie(&http.Cookie{Name: "ctoken", Value: c.ctoken})
	req.AddCookie(&http.Cookie{Name: "sessionId", Value: c.sessionID})
	if c.sessionIDSig != "" {
		req.AddCookie(&http.Cookie{Name: "sessionId.sig", Value: c.sessionIDSig})
	}
}

func (c *Client) doRequest(ctx context.Context, method, urlStr string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, urlStr, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	c.setCookies(req)
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

	return body, nil
}

func (c *Client) GetCurrentSubscription(ctx context.Context) (*CurrentSubscriptionResponse, error) {
	u, _ := url.Parse(c.baseURL + "/api/subscription/get_current")
	q := u.Query()
	q.Set("ctoken", c.ctoken)
	u.RawQuery = q.Encode()

	body, err := c.doRequest(ctx, "GET", u.String())
	if err != nil {
		return nil, err
	}

	var result CurrentSubscriptionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetCurrentUsage(ctx context.Context) (*CurrentUsageResponse, error) {
	u, _ := url.Parse(c.baseURL + "/api/subscription/get_current_usage")
	q := u.Query()
	q.Set("ctoken", c.ctoken)
	u.RawQuery = q.Encode()

	body, err := c.doRequest(ctx, "GET", u.String())
	if err != nil {
		return nil, err
	}

	var result CurrentUsageResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetSubscriptionSummary(ctx context.Context) (*SubscriptionSummaryResponse, error) {
	u, _ := url.Parse(c.baseURL + "/api/dashboard/cost/query/subscription_summary")
	q := u.Query()
	q.Set("ctoken", c.ctoken)
	u.RawQuery = q.Encode()

	body, err := c.doRequest(ctx, "GET", u.String())
	if err != nil {
		return nil, err
	}

	var result SubscriptionSummaryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
