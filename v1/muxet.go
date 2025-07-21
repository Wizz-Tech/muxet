package v1

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

// Request net/http wrapper passed to hooks
type Request struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    any
	Context context.Context
}

// Response net/http wrapper passed to hooks
type Response struct {
	StatusCode int
	Headers    map[string][]string
	Body       []byte
	Raw        *http.Response
}

func (r *Response) JSON(out any) error {
	return json.Unmarshal(r.Body, out)
}

// Logger logger interface
type Logger interface {
	Logf(format string, args ...any)
}

// HTTPDoer
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is a reusable HTTP client with timeouts, base URL, retry logic, and hooks
type Client struct {
	client        HTTPDoer
	headers       map[string]string
	timeout       time.Duration
	BaseURL       string
	logger        Logger
	maxRetries    int
	backoff       time.Duration
	BeforeRequest func(*Request) error
	AfterResponse func(*Response) error
}

// NewClient creates a new HTTP client with default settings
func NewClient() *Client {
	return &Client{
		client:     &http.Client{},
		headers:    make(map[string]string),
		timeout:    5 * time.Second,
		maxRetries: 0,
		backoff:    0,
	}
}

func (c *Client) DoRequest(ctx context.Context, method, rawURL string, body any, out any, headers map[string]string) (*http.Response, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), c.timeout)
		defer cancel()
	}

	fullURL, err := c.resolveURL(rawURL)
	if err != nil {
		return nil, err
	}

	// Merge headers
	hdr := make(map[string]string)
	for k, v := range c.headers {
		hdr[k] = v
	}
	for k, v := range headers {
		hdr[k] = v
	}

	muxReq := &Request{
		Method:  method,
		URL:     fullURL,
		Headers: hdr,
		Body:    body,
		Context: ctx,
	}

	if c.BeforeRequest != nil {
		if err := c.BeforeRequest(muxReq); err != nil {
			return nil, fmt.Errorf("before request hook failed: %w", err)
		}
	}

	var origBody []byte
	if muxReq.Body != nil {
		origBody, err = json.Marshal(muxReq.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
	}

	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		var reqBody io.Reader
		if muxReq.Body != nil {
			reqBody = bytes.NewReader(origBody)
		}

		req, err := http.NewRequestWithContext(muxReq.Context, muxReq.Method, muxReq.URL, reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		for k, v := range muxReq.Headers {
			req.Header.Set(k, v)
		}

		if muxReq.Body != nil && req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}

		if c.logger != nil {
			c.logger.Logf("Request: %s %s (attempt %d)", muxReq.Method, muxReq.URL, attempt+1)
		}

		resp, err = c.client.Do(req)
		if err != nil {
			lastErr = err
			if c.logger != nil {
				c.logger.Logf("Request failed: %v", err)
			}
			time.Sleep(c.backoff * time.Duration(1<<attempt))
			continue
		}

		rawBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp, fmt.Errorf("failed to read response body: %w", err)
		}

		muxResp := &Response{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header.Clone(),
			Body:       rawBody,
			Raw:        resp,
		}

		if c.AfterResponse != nil {
			if err := c.AfterResponse(muxResp); err != nil {
				return resp, fmt.Errorf("after response hook failed: %w", err)
			}
		}

		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
			time.Sleep(c.backoff * time.Duration(1<<attempt))
			continue
		}

		if out != nil {
			if s, ok := out.(*string); ok {
				*s = string(rawBody)
			} else {
				// decode json from rawBody bytes instead of resp.Body
				if err := json.Unmarshal(rawBody, out); err != nil {
					return resp, fmt.Errorf("failed to decode response: %w", err)
				}
			}
		}

		return resp, nil
	}

	return resp, fmt.Errorf("request failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

func (c *Client) resolveURL(input string) (string, error) {
	u, err := url.Parse(input)
	if err != nil {
		return "", fmt.Errorf("invalid request URL: %w", err)
	}
	if u.IsAbs() || c.BaseURL == "" {
		return input, nil
	}
	base, err := url.Parse(c.BaseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}
	return base.ResolveReference(u).String(), nil
}
