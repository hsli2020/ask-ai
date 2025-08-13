# JSON-RPC 2.0 支持一次发送一个请求，也支持一次发送多个请求，如
[
  {
    "jsonrpc": "2.0",
    "method": "getBalance",
    "params": ["user1"],
    "id": 1
  },
  {
    "jsonrpc": "2.0",
    "method": "getTransactions",
    "params": ["user1"],
    "id": 2
  },
  {
    "jsonrpc": "2.0",
    "method": "logActivity",
    "params": ["login"],
    "id": null  // 通知，无响应
  }
]
我用golang语言时遇到了麻烦，如何优雅地处理这种情况呢，即如何发送一个或多个请求，处理一个或多个应答


在 Go 中优雅地处理 JSON-RPC 2.0 的**单请求**和**批量请求（数组）**，关键在于：

1. **统一序列化/反序列化结构**：使用 `[]Request` 和 `Request` 共享结构。
2. **区分“通知”与“有响应请求”**：`id == nil` 是通知，无需等待响应。
3. **并发处理多个请求**：使用 goroutine + channel 或 sync.WaitGroup。
4. **返回结构清晰**：支持返回单个响应或响应数组。

---

## ✅ 推荐方案：通用、优雅、可扩展

### 1. 定义 JSON-RPC 2.0 结构体

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Request 表示一个 JSON-RPC 2.0 请求
type Request struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	Id      interface{} `json:"id"` // 可以是 int, string, null
}

// Response 表示一个 JSON-RPC 2.0 响应
type Response struct {
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
	Id      interface{} `json:"id"`
}

// Error 是 JSON-RPC 错误对象
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// BatchResponse 是批量响应的集合
type BatchResponse []Response
```

> 注意：`Id` 使用 `interface{}` 是为了兼容 `int`, `string`, `null`。

---

### 2. 发送单个请求（封装为通用函数）

```go
func sendRPCRequest(client *http.Client, url string, requests ...Request) (interface{}, error) {
	var body []byte
	var err error

	if len(requests) == 1 {
		// 单请求
		body, err = json.Marshal(requests[0])
	} else {
		// 批量请求
		body, err = json.Marshal(requests)
	}
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if len(requests) == 1 {
		// 单请求：直接解析为 Response
		var singleResp Response
		if err := json.Unmarshal(respBody, &singleResp); err != nil {
			return nil, fmt.Errorf("unmarshal single response: %w", err)
		}
		return singleResp, nil
	} else {
		// 批量请求：解析为 []Response
		var batchResp BatchResponse
		if err := json.Unmarshal(respBody, &batchResp); err != nil {
			return nil, fmt.Errorf("unmarshal batch response: %w", err)
		}
		return batchResp, nil
	}
}
```

---

### 3. 高级用法：异步发送 + 并发处理（推荐用于高并发场景）

如果你希望**并行处理多个请求**（比如每个请求独立耗时），可以这样设计：

```go
type RPCClient struct {
	client *http.Client
	url    string
}

func NewRPCClient(url string) *RPCClient {
	return &RPCClient{
		client: &http.Client{},
		url:    url,
	}
}

// SendBatch 发送批量请求，返回响应通道，支持异步消费
func (c *RPCClient) SendBatch(requests ...Request) (<-chan Response, error) {
	if len(requests) == 0 {
		return nil, fmt.Errorf("no requests")
	}

	ch := make(chan Response, len(requests)) // 缓冲通道避免阻塞

	go func() {
		defer close(ch)

		// 构造批量请求体
		body, err := json.Marshal(requests)
		if err != nil {
			for _, req := range requests {
				if req.Id != nil { // 只对有 ID 的请求返回错误
					ch <- Response{Id: req.Id, Error: &Error{Code: -32700, Message: "encode error"}}
				}
			}
			return
		}

		resp, err := c.client.Post(c.url, "application/json", bytes.NewBuffer(body))
		if err != nil {
			for _, req := range requests {
				if req.Id != nil {
					ch <- Response{Id: req.Id, Error: &Error{Code: -32000, Message: "network error"}}
				}
			}
			return
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			for _, req := range requests {
				if req.Id != nil {
					ch <- Response{Id: req.Id, Error: &Error{Code: -32000, Message: "read response error"}}
				}
			}
			return
		}

		var batch BatchResponse
		if err := json.Unmarshal(respBody, &batch); err != nil {
			for _, req := range requests {
				if req.Id != nil {
					ch <- Response{Id: req.Id, Error: &Error{Code: -32700, Message: "decode response error"}}
				}
			}
			return
		}

		// 按 id 匹配响应（注意：服务器可能乱序）
		respMap := make(map[interface{}]Response)
		for _, r := range batch {
			if r.Id != nil {
				respMap[r.Id] = r
			}
		}

		// 按原始请求顺序输出响应（重要！）
		for _, req := range requests {
			if req.Id == nil {
				continue // 通知，不返回响应
			}
			if resp, ok := respMap[req.Id]; ok {
				ch <- resp
			} else {
				ch <- Response{Id: req.Id, Error: &Error{Code: -32603, Message: "internal error: no matching response"}}
			}
		}
	}()

	return ch, nil
}
```

---

### 4. 使用示例

#### ✅ 场景1：单请求

```go
client := NewRPCClient("http://localhost:8080/rpc")

resp, err := sendRPCRequest(client.client, client.url, Request{
	Jsonrpc: "2.0",
	Method:  "getBalance",
	Params:  []string{"user1"},
	Id:      1,
})
if err != nil {
	log.Fatal(err)
}

if singleResp, ok := resp.(Response); ok {
	if singleResp.Error != nil {
		fmt.Printf("Error: %v\n", singleResp.Error)
	} else {
		fmt.Printf("Result: %+v\n", singleResp.Result)
	}
}
```

#### ✅ 场景2：批量请求（同步）

```go
requests := []Request{
	{
		Jsonrpc: "2.0",
		Method:  "getBalance",
		Params:  []string{"user1"},
		Id:      1,
	},
	{
		Jsonrpc: "2.0",
		Method:  "getTransactions",
		Params:  []string{"user1"},
		Id:      2,
	},
	{
		Jsonrpc: "2.0",
		Method:  "logActivity",
		Params:  []string{"login"},
		Id:      nil, // 通知，无响应
	},
}

resp, err := sendRPCRequest(client.client, client.url, requests...)
if err != nil {
	log.Fatal(err)
}

if batch, ok := resp.(BatchResponse); ok {
	for _, r := range batch {
		if r.Id == nil {
			continue // 跳过通知
		}
		if r.Error != nil {
			fmt.Printf("ID %v Error: %v\n", r.Id, r.Error.Message)
		} else {
			fmt.Printf("ID %v Result: %+v\n", r.Id, r.Result)
		}
	}
}
```

#### ✅ 场景3：异步批量请求（高性能）

```go
requests := []Request{
	{Jsonrpc: "2.0", Method: "getBalance", Params: []string{"user1"}, Id: 1},
	{Jsonrpc: "2.0", Method: "getTransactions", Params: []string{"user1"}, Id: 2},
	{Jsonrpc: "2.0", Method: "logActivity", Params: []string{"login"}, Id: nil},
}

ch, err := client.SendBatch(requests...)
if err != nil {
	log.Fatal(err)
}

for resp := range ch {
	if resp.Error != nil {
		fmt.Printf("❌ ID %v: %s\n", resp.Id, resp.Error.Message)
	} else {
		fmt.Printf("✅ ID %v: %+v\n", resp.Id, resp.Result)
	}
}
```

---

## ✅ 关键设计优点

| 特性 | 实现 |
|------|------|
| **统一结构** | `Request` 和 `Response` 支持单个和批量 |
| **通知忽略** | `Id == nil` 自动跳过响应处理 |
| **并发安全** | 异步版本使用 channel，适合高并发 |
| **顺序保证** | 响应按请求顺序输出（即使服务端乱序） |
| **错误处理** | 网络/编码/解码错误都包装成标准 JSON-RPC 错误 |
| **类型安全** | 使用接口 `interface{}` 处理灵活 ID 类型 |

---

## 🚀 进阶建议

- 使用 `github.com/gorilla/rpc/v2` 或 `github.com/json-iterator/go` 提升性能（可选）。
- 如果是 WebSocket，可以用 `gorilla/websocket` 替代 HTTP。
- 对于大型系统，考虑引入 `context.Context` 控制超时。

---

## ✅ 总结

> **Go 中优雅处理 JSON-RPC 2.0 批量请求的核心是：**
>
> 1. 用 `[]Request` 统一表示批量；
> 2. 用 `Id == nil` 标识通知；
> 3. 用 `channel` 实现异步非阻塞；
> 4. 用 `interface{}` 支持灵活 ID 类型；
> 5. 返回值根据输入类型动态返回 `Response` 或 `[]Response`。

这样你就能写出既简洁又健壮的 JSON-RPC 客户端了！👍
