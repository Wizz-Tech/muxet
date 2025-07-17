# muxet

A **lightweight**, **testable**, and **extensible** HTTP client for Go with support for:

- ✅ Request timeouts and retry logic with exponential backoff
- 🌐 Optional base URL
- 🔌 Middleware-style hooks (BeforeRequest / AfterResponse)
- 📦 Response wrapping with helpers (e.g. `.JSON()`)
- 🧪 Mockable HTTP transport via `HTTPDoer` for easy testing

---

## 🚀 Getting Started

```go
import "https://github.com/Wizz-Tech/muxet"
```

### Creating a client

```go
client := muxet.NewClient().
    SetBaseURL("https://api.example.com").
    SetHeader("Authorization", "Bearer TOKEN").
    SetTimeout(10 * time.Second).
    SetMaxRetries(3).
    SetBackoff(500 * time.Millisecond).
    SetLogger(logger) // optional
```

---

## 🔧 Requests

### GET

```go
var data MyData
resp, err := client.Get(context.Background(), "/items/1", &data, nil)
if err != nil {
    log.Fatal(err)
}
```

### POST with payload

```go
payload := map[string]string{"name": "foo"}
var result MyResult
resp, err := client.Post(context.Background(), "/items", payload, &result, nil)
```

Use `.Put(...)` or `.Delete(...)` similarly.

---

## ⚙️ Middleware Hooks

### BeforeRequest

Inspect or mutate the outgoing request:

```go
client.SetBeforeRequestHook(func(r *muxet.Request) error {
    r.Headers["X-Trace-ID"] = generateTraceID()
    return nil
})
```

### AfterResponse

Handle or inspect the response before it's decoded:

```go
client.SetAfterResponseHook(func(r *muxet.Response) error {
    if r.StatusCode == 429 {
        return fmt.Errorf("rate limited")
    }
    return nil
})
```

Access response body:

```go
var data MyRespType
if r.StatusCode == 200 {
    r.JSON(&data)
}
```

---

## 🔃 Retry Logic

Configure automatic retries on network or HTTP errors:

```go
client.SetMaxRetries(2).
       SetBackoff(200 * time.Millisecond)
```

Retries use **exponential backoff**: `backoff * 2^attempt`.

---

## 🧪 Testability

Easily inject a stubbed HTTP client:

```go
type MockDoer struct{}

func (m *MockDoer) Do(req *http.Request) (*http.Response, error) {
    // return a fake response
}

client := muxet.NewClient()
client.client = &MockDoer{}
```

---

## 🤩 Types Overview

```go
type Request struct {
    Method  string
    URL     string
    Headers map[string]string
    Body    any
    Context context.Context
}

type Response struct {
    StatusCode int
    Headers    map[string][]string
    Body       []byte
    Raw        *http.Response
}

func (r *Response) JSON(out any) error
```

---

## 📄 API Reference

### `func NewClient() *Client`

Creates a new `Client` with default settings.

### Fluent setters on `*Client`

```go
SetTimeout(d time.Duration)       *Client
SetHeader(key, value string)     *Client
SetLogger(l Logger)              *Client
SetBaseURL(base string)          *Client
SetMaxRetries(n int)             *Client
SetBackoff(d time.Duration)      *Client
SetBeforeRequestHook(fn func(*Request) error)
SetAfterResponseHook(fn func(*Response) error)
```

### Request execution

```go
DoRequest(ctx context.Context, method, url string, body any, out any, headers map[string]string) (*http.Response, error)
Get(ctx, url string, out any, headers map[string]string)
Post(ctx, url string, body any, out any, headers map[string]string)
Put(ctx, url string, body any, out any, headers map[string]string)
Delete(ctx, url string, out any, headers map[string]string)
```

---

## ✅ Examples

```go
client := muxet.NewClient().
    SetBaseURL("https://api.example.com").
    SetHeader("Authorization", "Bearer TOKEN").
    SetMaxRetries(2).
    SetBackoff(300 * time.Millisecond).
    SetBeforeRequestHook(func(r *muxet.Request) error {
        r.Headers["X-Request-ID"] = uuid.New().String()
        return nil
    }).
    SetAfterResponseHook(func(r *muxet.Response) error {
        if r.StatusCode >= 500 {
            return fmt.Errorf("server error: %d", r.StatusCode)
        }
        return nil
    })

type User struct { ID int; Name string }
var user User
if _, err := client.Get(context.Background(), "/users/42", &user, nil); err != nil {
    log.Fatal(err)
}
```

---

## 📄 License

MIT License. See [LICENSE](./LICENCE) for details.

