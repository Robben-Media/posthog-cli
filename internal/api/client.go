package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	projectID  string
	apiKey     string
	userAgent  string
}

type ClientOption func(*Client)

func WithBaseURL(u string) ClientOption {
	return func(c *Client) {
		c.baseURL = u
	}
}

func WithProjectID(id string) ClientOption {
	return func(c *Client) {
		c.projectID = id
	}
}

func WithUserAgent(ua string) ClientOption {
	return func(c *Client) {
		c.userAgent = ua
	}
}

func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

func NewClient(apiKey string, opts ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		apiKey:    apiKey,
		userAgent: "posthog-cli/dev",
		baseURL:   "https://us.posthog.com",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// ProjectID returns the configured project ID.
func (c *Client) ProjectID() string {
	return c.projectID
}

// BaseURL returns the configured base URL.
func (c *Client) BaseURL() string {
	return c.baseURL
}

type Request struct {
	Method  string
	Path    string
	Body    any
	Params  url.Values
	Headers map[string]string
}

func (c *Client) Do(ctx context.Context, req Request) (*http.Response, error) {
	var bodyReader io.Reader

	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}

		bodyReader = bytes.NewReader(bodyBytes)
	}

	u := c.baseURL + req.Path

	if len(req.Params) > 0 {
		u += "?" + req.Params.Encode()
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, u, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent)

	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}

	return resp, nil
}

func (c *Client) Get(ctx context.Context, path string, params url.Values, result any) error {
	return c.doJSON(ctx, Request{Method: http.MethodGet, Path: path, Params: params}, result)
}

func (c *Client) Post(ctx context.Context, path string, body, result any) error {
	return c.doJSON(ctx, Request{Method: http.MethodPost, Path: path, Body: body}, result)
}

// GetRaw performs a GET and returns the raw response bytes.
func (c *Client) GetRaw(ctx context.Context, path string, params url.Values) ([]byte, error) {
	resp, err := c.Do(ctx, Request{Method: http.MethodGet, Path: path, Params: params})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp)
	}

	return io.ReadAll(resp.Body)
}

// PostRaw performs a POST and returns the raw response bytes.
func (c *Client) PostRaw(ctx context.Context, path string, body any) ([]byte, error) {
	resp, err := c.Do(ctx, Request{Method: http.MethodPost, Path: path, Body: body})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp)
	}

	return io.ReadAll(resp.Body)
}

// Query executes a HogQL query and returns raw response bytes.
func (c *Client) Query(ctx context.Context, hogql string) ([]byte, error) {
	path := fmt.Sprintf("/api/projects/%s/query/", c.projectID)
	body := map[string]any{
		"query": map[string]any{
			"kind":  "HogQLQuery",
			"query": hogql,
		},
	}

	return c.PostRaw(ctx, path, body)
}

func (c *Client) doJSON(ctx context.Context, req Request, result any) error {
	resp, err := c.Do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return parseAPIError(resp)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.Message)
}

func parseAPIError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var apiErr struct {
		Message string `json:"message"`
		Detail  string `json:"detail"`
		Err     string `json:"error"`
	}

	if json.Unmarshal(body, &apiErr) == nil {
		msg := apiErr.Detail
		if msg == "" {
			msg = apiErr.Message
		}

		if msg == "" {
			msg = apiErr.Err
		}

		if msg != "" {
			return &APIError{StatusCode: resp.StatusCode, Message: msg}
		}
	}

	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    http.StatusText(resp.StatusCode),
	}
}
