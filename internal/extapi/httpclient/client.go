package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const MaxResponseBytes = 16 << 20

type Client struct {
	HTTP    *http.Client
	allowed map[string]bool
}

func New(hosts []string) *Client {
	allowed := make(map[string]bool, len(hosts))
	for _, h := range hosts {
		allowed[strings.ToLower(h)] = true
	}
	return &Client{HTTP: &http.Client{Timeout: 15 * time.Second, Transport: &http.Transport{Proxy: nil, MaxIdleConns: 100, MaxIdleConnsPerHost: 10, IdleConnTimeout: 90 * time.Second, ResponseHeaderTimeout: 10 * time.Second}, CheckRedirect: func(*http.Request, []*http.Request) error { return fmt.Errorf("upstream redirects are disabled") }}, allowed: allowed}
}
func (c *Client) Validate(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.Fragment != "" {
		return nil, fmt.Errorf("invalid upstream URL")
	}
	if !c.allowed[strings.ToLower(u.Host)] {
		return nil, fmt.Errorf("upstream host is not in OUTBOUND_ALLOWED_HOSTS")
	}
	return u, nil
}
func (c *Client) Do(ctx context.Context, method, raw string, body any, headers map[string]string, out any) error {
	if _, err := c.Validate(raw); err != nil {
		return err
	}
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, raw, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	response, err := c.HTTP.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("upstream request failed")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("upstream returned HTTP %d", response.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(response.Body, MaxResponseBytes+1))
	if err != nil {
		return err
	}
	if len(b) > MaxResponseBytes {
		return fmt.Errorf("upstream response exceeds limit")
	}
	if out == nil || len(b) == 0 {
		return nil
	}
	if err = json.Unmarshal(b, out); err != nil {
		return fmt.Errorf("decode upstream response: %w", err)
	}
	return nil
}
