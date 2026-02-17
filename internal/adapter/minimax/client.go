package minimax

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"time"
)

var (
	baseURL = "https://www.minimaxi.com/v1/api/openplatform"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	cookie     string
	groupID    string
	Debug      bool
}

func NewClient(cookie, groupID string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
		cookie:     cookie,
		groupID:    groupID,
		Debug:      true,
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
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Origin", "https://platform.minimaxi.com")
	req.Header.Set("Referer", "https://platform.minimaxi.com/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
}

func (c *Client) setCookies(req *http.Request) {
	req.Header.Set("Cookie", c.cookie)
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return body, nil
}

// GetCurrentSubscribe 获取当前订阅信息
func (c *Client) GetCurrentSubscribe(ctx context.Context) (*CurrentSubscribeResponse, error) {
	url := fmt.Sprintf("%s/charge/combo/cycle_audio_resource_package?biz_line=2&cycle_type=3&resource_package_type=7&GroupId=%s",
		c.baseURL, c.groupID)

	body, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, err
	}

	var result CurrentSubscribeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetRemains 获取当前用量
func (c *Client) GetRemains(ctx context.Context) (*RemainsResponse, error) {
	url := fmt.Sprintf("%s/coding_plan/remains?GroupId=%s", c.baseURL, c.groupID)

	body, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, err
	}

	var result RemainsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
