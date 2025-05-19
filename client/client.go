// client/client.go
package client

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

// StoreClient defines the operations supported by the data store client.
type StoreClient interface {
	SetString(ctx context.Context, key, value string, ttl time.Duration) error
	GetString(ctx context.Context, key string) (string, error)
	DeleteString(ctx context.Context, key string) error

	LPush(ctx context.Context, key string, items ...string) error
	RPop(ctx context.Context, key string) (string, error)
}

// Client implements StoreClient over HTTP.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	token      string
}

// NewClient creates a new Client with the given base URL.
// e.g. "http://localhost:8080"
func NewClient(rawBaseURL string, token string) (StoreClient, error) {
	u, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL %q: %w", rawBaseURL, err)
	}
	return &Client{baseURL: u, httpClient: http.DefaultClient, token: token}, nil
}

// doRequest builds the full URL, performs the HTTP call,
// returns an HTTPError for status >= 400, and optionally unmarshals into outObj.
func (c *Client) doRequest(
	ctx context.Context,
	method, endpoint string,
	reqObj interface{},
	outObj interface{},
) error {
	// Resolve full URL
	ref, _ := url.Parse(endpoint)
	fullURL := c.baseURL.ResolveReference(ref).String()

	// Marshal request body if present
	var body io.Reader
	if reqObj != nil {
		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(reqObj); err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		body = buf
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	// Do request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP error status
	if resp.StatusCode >= 400 {
		// Attempt to decode structured JSON error
		var he HTTPError
		he.Code = resp.StatusCode
		if err := json.NewDecoder(resp.Body).Decode(&he); err == nil && he.Message != "" {
			return &he
		}
		// Fallback to plain-text
		data, _ := io.ReadAll(resp.Body)
		msg := strings.TrimSpace(string(data))
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, msg)
	}

	// Decode successful response if needed
	if outObj != nil {
		if err := json.NewDecoder(resp.Body).Decode(outObj); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// SetString sets a string value with an optional TTL (0 = server default).
func (c *Client) SetString(ctx context.Context, key, value string, ttl time.Duration) error {
	req := stringRequest{Value: value, TTLSeconds: int(ttl.Seconds())}
	endpoint := fmt.Sprintf("/v1/string/%s", url.PathEscape(key))
	return c.doRequest(ctx, http.MethodPost, endpoint, req, nil)
}

// GetString retrieves a string value by key.
func (c *Client) GetString(ctx context.Context, key string) (string, error) {
	var resp stringResponse
	endpoint := fmt.Sprintf("/v1/string/%s", url.PathEscape(key))
	if err := c.doRequest(ctx, http.MethodGet, endpoint, nil, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// DeleteString deletes a string value by key.
func (c *Client) DeleteString(ctx context.Context, key string) error {
	endpoint := fmt.Sprintf("/v1/string/%s", url.PathEscape(key))
	return c.doRequest(ctx, http.MethodDelete, endpoint, nil, nil)
}

// LPush pushes items onto the head of the list at key.
func (c *Client) LPush(ctx context.Context, key string, items ...string) error {
	req := listRequest{Items: items}
	endpoint := fmt.Sprintf("/v1/list/%s/push", url.PathEscape(key))
	return c.doRequest(ctx, http.MethodPost, endpoint, req, nil)
}

// RPop pops an item from the tail of the list at key.
func (c *Client) RPop(ctx context.Context, key string) (string, error) {
	var resp stringResponse
	endpoint := fmt.Sprintf("/v1/list/%s/pop", url.PathEscape(key))
	if err := c.doRequest(ctx, http.MethodPost, endpoint, nil, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}
