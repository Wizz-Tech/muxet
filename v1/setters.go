package v1

import "time"

func (c *Client) SetTimeout(d time.Duration) *Client {
	c.timeout = d
	return c
}

func (c *Client) SetHeader(key, value string) *Client {
	c.headers[key] = value
	return c
}

func (c *Client) SetLogger(l Logger) *Client {
	c.logger = l
	return c
}

func (c *Client) SetBaseURL(base string) *Client {
	c.BaseURL = base
	return c
}

func (c *Client) SetMaxRetries(n int) *Client {
	c.maxRetries = n
	return c
}

func (c *Client) SetBackoff(d time.Duration) *Client {
	c.backoff = d
	return c
}

func (c *Client) SetBeforeRequestHook(fn func(*Request) error) *Client {
	c.BeforeRequest = fn
	return c
}

func (c *Client) SetAfterResponseHook(fn func(*Response) error) *Client {
	c.AfterResponse = fn
	return c
}
