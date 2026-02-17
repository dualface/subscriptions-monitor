package kimi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"time"
)

var (
	baseURL = "https://www.kimi.com/apiv2"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	authToken  string
	cookie     string
	Debug      bool
}

func NewClient(authToken, cookie string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
		authToken:  authToken,
		cookie:     cookie,
		Debug:      true,
	}
}

func (c *Client) logRequest(req *http.Request, body []byte) {
	if !c.Debug {
		return
	}
	dump, _ := httputil.DumpRequestOut(req, false)
	fmt.Println("=== HTTP REQUEST ===")
	fmt.Println(string(dump))
	if len(body) > 0 {
		fmt.Printf("Body: %s\n", string(body))
	}
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
	req.Header.Set("Authorization", "Bearer "+c.authToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Connect-Protocol-Version", "1")
	req.Header.Set("Origin", "https://www.kimi.com")
	req.Header.Set("Referer", "https://www.kimi.com/code/console")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("X-Language", "zh-CN")
	req.Header.Set("X-Msh-Platform", "web")
	req.Header.Set("X-Msh-Version", "1.0.0")
}

func (c *Client) setCookies(req *http.Request) {
	req.Header.Set("Cookie", c.cookie)
}

func (c *Client) doRequest(ctx context.Context, method, urlStr string, body []byte) ([]byte, error) {
	var bodyReader *bytes.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	c.setCookies(req)
	c.logRequest(req, body)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	c.logResponse(resp, respBody)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return respBody, nil
}

// GetUsages 查询用量信息
func (c *Client) GetUsages(ctx context.Context) (*UsagesResponse, error) {
	url := c.baseURL + "/kimi.gateway.billing.v1.BillingService/GetUsages"

	reqBody := []byte(`{"scope":["FEATURE_CODING"]}`)

	body, err := c.doRequest(ctx, "POST", url, reqBody)
	if err != nil {
		return nil, err
	}

	var result UsagesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetSubscription(ctx context.Context) (*SubscriptionResponse, error) {
	url := c.baseURL + "/kimi.gateway.order.v1.SubscriptionService/GetSubscription"

	body, err := c.doRequest(ctx, "POST", url, []byte("{}"))
	if err != nil {
		return nil, err
	}

	var result SubscriptionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
