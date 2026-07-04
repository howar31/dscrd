// Package api is a thin Discord REST API HTTP client.
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Client is a thin Discord REST API HTTP client.
type Client struct {
	Token      string
	BaseURL    string // overridable for tests; default https://discord.com/api/v10
	HTTP       *http.Client
	MaxRetries int // rate-limit / index-warm-up retries
}

// New returns a Client for the given bot token.
func New(token string) *Client {
	return &Client{
		Token:      token,
		BaseURL:    "https://discord.com/api/v10",
		HTTP:       &http.Client{Timeout: 30 * time.Second},
		MaxRetries: 3,
	}
}

// Do invokes a Discord REST endpoint. method is an HTTP verb; path starts with
// "/" and may carry a query string. A non-nil body is marshaled as JSON.
// Returns the raw response bytes (empty for 204 No Content).
func (c *Client) Do(method, path string, body any) ([]byte, error) {
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("%s %s: encoding body: %w", method, path, err)
		}
	}
	return c.roundTrip(method, path, "application/json", func() (io.Reader, error) {
		if payload == nil {
			return nil, nil
		}
		return bytes.NewReader(payload), nil
	})
}

// roundTrip performs the request with rate-limit and search-index retries.
// makeBody is called per attempt so retries resend a fresh reader.
func (c *Client) roundTrip(method, path, contentType string, makeBody func() (io.Reader, error)) ([]byte, error) {
	endpoint := c.BaseURL + path

	for attempt := 0; ; attempt++ {
		bodyReader, err := makeBody()
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest(method, endpoint, bodyReader)
		if err != nil {
			return nil, err
		}
		if bodyReader != nil {
			req.Header.Set("Content-Type", contentType)
		}
		req.Header.Set("Authorization", "Bot "+c.Token)

		resp, err := c.HTTP.Do(req)
		if err != nil {
			return nil, err
		}
		raw, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("%s %s: reading response body: %w", method, path, err)
		}

		if wait, retry := retryDelay(resp, raw); retry {
			if attempt < c.MaxRetries {
				time.Sleep(wait)
				continue
			}
			return nil, newAPIError(method, path, resp.StatusCode, raw)
		}

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return nil, newAPIError(method, path, resp.StatusCode, raw)
		}
		return raw, nil
	}
}

// retryDelay reports whether the response is retryable (rate limit, or a
// message-search index that is still warming up) and for how long to wait.
func retryDelay(resp *http.Response, raw []byte) (time.Duration, bool) {
	warming := false
	if resp.StatusCode == http.StatusAccepted {
		var probe struct {
			Code int `json:"code"`
		}
		if json.Unmarshal(raw, &probe) == nil && probe.Code == codeIndexNotReady {
			warming = true
		}
	}
	if resp.StatusCode != http.StatusTooManyRequests && !warming {
		return 0, false
	}

	var body struct {
		RetryAfter float64 `json:"retry_after"`
	}
	if json.Unmarshal(raw, &body) == nil && body.RetryAfter > 0 {
		return time.Duration(body.RetryAfter * float64(time.Second)), true
	}
	if h := resp.Header.Get("Retry-After"); h != "" {
		if secs, err := strconv.ParseFloat(h, 64); err == nil && secs > 0 {
			return time.Duration(secs * float64(time.Second)), true
		}
	}
	return time.Second, true
}
