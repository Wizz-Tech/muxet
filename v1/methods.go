package v1

import (
	"context"
	"net/http"
)

func (c *Client) Get(ctx context.Context, url string, out any, headers map[string]string) (*http.Response, error) {
	return c.DoRequest(ctx, http.MethodGet, url, nil, out, headers)
}

func (c *Client) Post(ctx context.Context, url string, body any, out any, headers map[string]string) (*http.Response, error) {
	return c.DoRequest(ctx, http.MethodPost, url, body, out, headers)
}

func (c *Client) Put(ctx context.Context, url string, body any, out any, headers map[string]string) (*http.Response, error) {
	return c.DoRequest(ctx, http.MethodPut, url, body, out, headers)
}

func (c *Client) Delete(ctx context.Context, url string, out any, headers map[string]string) (*http.Response, error) {
	return c.DoRequest(ctx, http.MethodDelete, url, nil, out, headers)
}
