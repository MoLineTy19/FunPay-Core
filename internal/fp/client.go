package fp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	throttler  *Throttler
	goldenKey  string
	sessionID  string
	goldenSeal string
}

func NewClient(goldenKey, sessionID string, goldenSeal string, minDelay, maxJitter time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{},
		throttler:  NewThrottler(minDelay, maxJitter),
		goldenKey:  goldenKey,
		sessionID:  sessionID,
		goldenSeal: goldenSeal,
	}
}

func (c *Client) do(ctx context.Context, method, url string, body io.Reader, contentType string) ([]byte, error) {
	if err := c.throttler.Wait(ctx); err != nil {
		return nil, fmt.Errorf("throttler: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	cookie := "golden_key=" + c.goldenKey + "; PHPSESSID=" + c.sessionID
	if c.goldenSeal != "" {
		cookie += "; golden_seal=" + c.goldenSeal
	}
	req.Header.Set("Cookie", cookie)

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36 OPR/122.0.0.0")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "ru,en-US;q=0.9,en;q=0.8,ko;q=0.7,de;q=0.6,it;q=0.5,ja;q=0.4,zh-TW;q=0.3,zh;q=0.2,sv;q=0.1,zh-CN;q=0.1")
	req.Header.Set("Sec-Ch-Ua", `"Not)A;Brand";v="8", "Chromium";v="138", "Opera GX";v="122"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", "Windows")

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("http status: %s; body: %s", resp.Status, string(data))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	return data, nil
}
