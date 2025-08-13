# 参考walmart官方网站的文档，用golang编写一段程序，演示walmart api的调用方法，包括如何设置请求header, 如何取得token，如何生成签名，完成一次请求发送和接受的过程

以下是基于 Walmart 官方开发者文档（developer.walmart.com）整理的 **Go 语言完整示例**。代码涵盖了：OAuth 2.0 Token 获取、请求头规范设置、HMAC 签名生成（注：官方 v3/v4 API 已全面转向 OAuth 2.0，签名仅用于兼容旧接口或特定场景）、以及完整的请求/响应处理流程。

### 📦 完整 Go 代码示例

```go
package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TokenResponse 接收 Walmart OAuth 2.0 Token 响应
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// getAccessToken 通过 Client Credentials 流程获取 Access Token
func getAccessToken(ctx context.Context, clientID, clientSecret string) (string, error) {
	tokenURL := "https://marketplace.walmartapis.com/v3/token"
	payload := []byte("grant_type=client_credentials")

	// 官方要求：Authorization: Basic <base64(client_id:client_secret)>
	auth := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("创建 token 请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送 token 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token 获取失败 (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var tResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tResp); err != nil {
		return "", fmt.Errorf("解析 token 响应失败: %w", err)
	}

	return tResp.AccessToken, nil
}

// generateSignature 生成 HMAC-SHA256 签名（适用于部分旧版/特定接口）
// 官方旧规范：签名 = Base64(HMAC-SHA256("METHOD\nURL\nTIMESTAMP", PrivateKey))
func generateSignature(method, fullURL, timestamp, privateKey string) (string, error) {
	signString := fmt.Sprintf("%s\n%s\n%s", method, fullURL, timestamp)
	mac := hmac.New(sha256.New, []byte(privateKey))
	mac.Write([]byte(signString))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}

// generateCorrelationID 生成符合规范的 UUID v4 关联ID（不依赖第三方库）
func generateCorrelationID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func main() {
	// ⚠️ 请替换为 Walmart Developer Portal 实际颁发的凭证
	clientID := "YOUR_CLIENT_ID"
	clientSecret := "YOUR_CLIENT_SECRET"
	privateKey := "YOUR_PRIVATE_KEY" // 仅部分接口需要，现代 OAuth 流程可忽略
	endpoint := "https://marketplace.walmartapis.com/v3/orders" // 示例：获取订单列表

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 1️⃣ 获取 Access Token
	token, err := getAccessToken(ctx, clientID, clientSecret)
	if err != nil {
		panic(err)
	}
	fmt.Println("✅ 成功获取 Access Token")

	// 2️⃣ 生成签名（如需兼容旧接口或特定 endpoint 要求）
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	signature, err := generateSignature(http.MethodGet, endpoint, timestamp, privateKey)
	if err != nil {
		panic(err)
	}

	// 3️⃣ 构建 API 请求
	correlationID := generateCorrelationID()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		panic(err)
	}

	// 4️⃣ 设置 Walmart 官方要求的请求头
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("WM_SVC.NAME", "Walmart Marketplace")
	req.Header.Set("WM_QOS.CORRELATION_ID", correlationID)
	req.Header.Set("WM_SEC.ACCESS_TOKEN", token)           // 部分旧接口要求重复携带
	req.Header.Set("WM_SEC.ACCESS_SIGNATURE", signature)   // 签名头（按需启用）
	req.Header.Set("Accept", "application/json")
	// 如为 POST/PUT 请求，还需添加: req.Header.Set("Content-Type", "application/json")

	// 5️⃣ 发送请求并接收响应
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		panic(fmt.Errorf("API 请求失败: %w", err))
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("\n📡 请求状态: %d\n", resp.StatusCode)
	fmt.Printf("📄 响应体:\n%s\n", string(respBody))
}
```

### 🔍 核心步骤解析（对照官方文档）

| 步骤 | 官方规范说明 | 代码实现 |
|------|----------------|----------|
| **1. 获取 Token** | 使用 `POST /v3/token`，`grant_type=client_credentials`，Header 传 `Basic <base64(id:secret)>` | `getAccessToken` 函数，严格遵循 OAuth 2.0 Client Credentials 流程 |
| **2. 设置 Header** | 必须包含 `Authorization: Bearer <token>`、`WM_SVC.NAME: Walmart Marketplace`、`WM_QOS.CORRELATION_ID: <UUID>` | `main` 中统一设置，关联ID用于 Walmart 内部链路追踪 |
| **3. 生成签名** | 官方 v3/v4 已**弃用强制签名**，全面采用 OAuth 2.0。但部分供应商接口或旧版仍要求 `WM_SEC.ACCESS_SIGNATURE`，算法为 `HMAC-SHA256(METHOD\nURL\nTIMESTAMP, PrivateKey)` | `generateSignature` 函数，按需启用。现代集成可安全注释掉 |
| **4. 请求与响应** | 需处理 `200 OK` 及业务级错误码（Walmart 常在 400/500 响应体中返回 JSON 错误详情） | 使用 `io.ReadAll` 读取完整响应，便于调试错误结构 |

### 📌 生产环境重要建议

1. **Token 缓存**：Walmart Token 有效期通常为 `3600s`。生产环境务必使用 Redis/内存缓存，避免每次请求都重新申请（频繁请求会触发速率限制）。
2. **环境区分**：
   - 生产环境：`https://marketplace.walmartapis.com`
   - 沙箱测试：`https://sandbox.walmartapis.com`
3. **签名是否必要？**  
   截至 2024-2026 官方文档，**Marketplace v3/v4 API 已不再要求手动生成签名**。若对接时遇到 `Invalid Signature` 错误，请检查是否误用了旧版 v2 接口，或确认 Partner 门户是否强制开启了签名校验。
4. **速率限制**：Walmart 对 Token 端点和业务端点均有严格限流（通常 `10-20 req/sec`），建议实现指数退避重试（Exponential Backoff）。

### 📖 官方文档参考
- OAuth 2.0 认证指南：`https://developer.walmart.com/marketplace/reference/authentication`
- API Header 规范：`https://developer.walmart.com/marketplace/reference/headers`
- 错误码与限流说明：`https://developer.walmart.com/marketplace/reference/errors`

如需对接具体业务接口（如 `POST /v3/orders` 创建订单、`GET /v3/items` 商品同步等），只需替换 `endpoint`、调整 `Method` 与 `Payload`，并补充对应的请求体序列化即可。


# 把Walmart API封装成golang包

将 Walmart API 封装为 Go 包时，需要遵循 **高内聚、线程安全、上下文传递、自动 Token 缓存、优雅错误处理** 等现代 Go 开发规范。以下是生产可用的完整包结构与实现。

### 📁 推荐目录结构
```
walmart/
├── go.mod
├── client.go      # 客户端初始化、配置、Token 缓存
├── auth.go        # OAuth 2.0 凭证获取逻辑
├── request.go     # HTTP 请求构建、Header 注入、响应解析
├── errors.go      # 自定义错误类型
├── util.go        # 签名生成、CorrelationID 工具
└── example/
    └── main.go    # 使用示例
```

---

### 📦 核心代码实现

#### 1. `client.go`（客户端与 Token 缓存）
```go
package walmart

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Config 客户端配置
type Config struct {
	ClientID     string
	ClientSecret string
	BaseURL      string // 生产: https://marketplace.walmartapis.com / 沙箱: https://sandbox.walmartapis.com
	PrivateKey   string // 仅旧版/特定接口需要，现代 OAuth 流程可留空
	HTTPClient   *http.Client
}

// Client Walmart API 客户端
type Client struct {
	config     Config
	httpClient *http.Client
	tokenMu    sync.Mutex
	tokenCache tokenCache
}

// NewClient 初始化客户端
func NewClient(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://marketplace.walmartapis.com"
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{
		config:     cfg,
		httpClient: cfg.HTTPClient,
	}
}

type tokenCache struct {
	mu        sync.RWMutex
	token     string
	expiresAt time.Time
}

func (tc *tokenCache) Get() (string, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	if time.Now().Before(tc.expiresAt) {
		return tc.token, true
	}
	return "", false
}

func (tc *tokenCache) Set(token string, ttl time.Duration) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.token = token
	// 预留 30 秒缓冲，避免临界点请求失败
	tc.expiresAt = time.Now().Add(ttl - 30*time.Second)
}

// GetAccessToken 获取有效 Token（自动缓存与并发安全）
func (c *Client) GetAccessToken(ctx context.Context) (string, error) {
	if token, ok := c.tokenCache.Get(); ok {
		return token, nil
	}

	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	// 双重检查
	if token, ok := c.tokenCache.Get(); ok {
		return token, nil
	}

	token, expires, err := c.fetchToken(ctx)
	if err != nil {
		return "", fmt.Errorf("fetch token: %w", err)
	}
	c.tokenCache.Set(token, time.Duration(expires)*time.Second)
	return token, nil
}
```

#### 2. `auth.go`（Token 获取）
```go
package walmart

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (c *Client) fetchToken(ctx context.Context) (string, int, error) {
	tokenURL := c.config.BaseURL + "/v3/token"
	auth := base64.StdEncoding.EncodeToString([]byte(c.config.ClientID + ":" + c.config.ClientSecret))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL,
		bytes.NewBufferString("grant_type=client_credentials"))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("token request failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var tResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tResp); err != nil {
		return "", 0, fmt.Errorf("decode token response: %w", err)
	}
	return tResp.AccessToken, tResp.ExpiresIn, nil
}
```

#### 3. `request.go`（请求执行与响应解析）
```go
package walmart

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Do 执行 Walmart API 请求（自动处理认证、Header、序列化）
func (c *Client) Do(ctx context.Context, method, path string, reqBody, respBody interface{}) error {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return err
	}

	var bodyReader io.Reader
	var isJSON bool
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
		isJSON = true
	}

	// 安全拼接 URL
	baseURL := strings.TrimRight(c.config.BaseURL, "/")
	cleanPath := strings.TrimLeft(path, "/")
	reqURL := fmt.Sprintf("%s/%s", baseURL, cleanPath)

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return err
	}

	// Walmart 官方强制 Header
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("WM_SVC.NAME", "Walmart Marketplace")
	req.Header.Set("WM_QOS.CORRELATION_ID", generateCorrelationID())
	req.Header.Set("Accept", "application/json")
	if isJSON {
		req.Header.Set("Content-Type", "application/json")
	}

	// 兼容旧版签名（按需启用）
	if c.config.PrivateKey != "" {
		timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
		sig, err := generateSignature(method, reqURL, timestamp, c.config.PrivateKey)
		if err != nil {
			return err
		}
		req.Header.Set("WM_SEC.ACCESS_SIGNATURE", sig)
		req.Header.Set("WM_SEC.TIMESTAMP", timestamp)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	return handleResponse(resp, respBody)
}
```

#### 4. `errors.go` & `util.go`（错误处理与工具）
```go
package walmart

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// APIError 封装 Walmart 业务错误
type APIError struct {
	StatusCode int    `json:"-"`
	ErrorCode  string `json:"error_code,omitempty"`
	Message    string `json:"message,omitempty"`
	RawBody    string `json:"-"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("walmart api error (HTTP %d): %s [%s]", e.StatusCode, e.Message, e.ErrorCode)
	}
	return fmt.Sprintf("walmart api error: HTTP %d, body: %s", e.StatusCode, e.RawBody)
}

func generateSignature(method, fullURL, timestamp, privateKey string) (string, error) {
	signString := fmt.Sprintf("%s\n%s\n%s", method, fullURL, timestamp)
	mac := hmac.New(sha256.New, []byte(privateKey))
	mac.Write([]byte(signString))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}

func generateCorrelationID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func handleResponse(resp *http.Response, v interface{}) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{StatusCode: resp.StatusCode, RawBody: string(body)}
		// 尝试解析 Walmart 标准错误 JSON
		var errResp struct {
			ErrorCode string `json:"error_code"`
			Message   string `json:"message"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.ErrorCode != "" {
			apiErr.ErrorCode = errResp.ErrorCode
			apiErr.Message = errResp.Message
		}
		return apiErr
	}

	if v != nil {
		if err := json.Unmarshal(body, v); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}
	return nil
}
```

---

### 🚀 使用示例 (`example/main.go`)
```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"yourmodule/walmart" // 替换为实际 module 路径
)

func main() {
	cfg := walmart.Config{
		ClientID:     "YOUR_CLIENT_ID",
		ClientSecret: "YOUR_CLIENT_SECRET",
		BaseURL:      "https://sandbox.walmartapis.com", // 测试环境
		// PrivateKey: "", // 仅对接旧版接口时填写
	}

	client := walmart.NewClient(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 示例：查询订单列表
	type OrderList struct {
		Count int `json:"count"`
		// 根据实际响应补充字段
	}

	var orders OrderList
	err := client.Do(ctx, "GET", "/v3/orders", nil, &orders)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	fmt.Printf("✅ 成功获取订单，数量: %d\n", orders.Count)

	// 示例：创建商品（POST）
	newItem := map[string]interface{}{
		"sku": "SKU-001",
		"product": map[string]interface{}{
			"productName": "Test Product",
			"category":    {"Electronics"},
		},
	}
	var createResp map[string]interface{}
	err = client.Do(ctx, "POST", "/v3/items", newItem, &createResp)
	if err != nil {
		log.Printf("创建商品失败(符合预期): %v", err)
	}
}
```

---

### 🔍 核心设计亮点

| 特性 | 实现方式 | 优势 |
|------|----------|------|
| **自动 Token 缓存** | `sync.RWMutex` + 双重检查锁 + 30s 缓冲过期 | 避免重复请求触发限流，高并发安全 |
| **统一错误封装** | `APIError` 结构体 + JSON 智能解析 | 业务代码只需 `errors.Is(err, &walmart.APIError{})` 即可精准处理 |
| **上下文透传** | 所有网络操作强制 `context.Context` | 支持超时控制、优雅取消、链路追踪 |
| **签名兼容** | 可选 `PrivateKey` 字段注入 | 现代 OAuth 2.0 免签名，旧接口无缝切换 |
| **零外部依赖** | 纯标准库实现 | 轻量、易审计、无供应链风险 |

---

### 📌 生产环境建议

1. **速率限制处理**：Walmart 对 `/v3/token` 限流较严（约 `5 req/sec`）。本包已内置缓存，正常情况下不会重复请求。若触发 `429 Too Many Requests`，建议在外层包装指数退避重试。
2. **沙箱 vs 生产**：通过 `Config.BaseURL` 切换。沙箱环境响应结构可能与生产略有差异，建议用 `reflect` 或 JSON Tag 做兼容解析。
3. **日志与监控**：在 `Do` 方法中可注入 `zap` 或 `slog` 记录 `correlation_id`，便于与 Walmart 工单对齐排查。
4. **Go 版本**：使用标准库 `strings.Trim` 替代 `url.JoinPath`，兼容 `Go 1.18+`。

将此结构放入 `go mod init` 项目后，即可通过 `import "yourmodule/walmart"` 在其他业务中复用。如需对接特定接口（如 `GET /v3/reports`、`POST /v3/feeds`），只需封装为 `Client.GetOrders(ctx)` 等快捷方法，底层统一调用 `c.Do()` 即可。


# 请封装order相关的api

以下是针对 Walmart Order 相关 API 的完整封装。代码严格遵循官方 v3/v4 规范，包含**列表查询、详情获取、接单确认、发货、取消**五大核心接口，并与前文的 `Client` 无缝集成。

### 📁 新增文件：`orders.go`
将此文件放入你的 `walmart/` 包目录中。

```go
package walmart

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// ============================================================
// 📦 请求/响应结构体（按 Walmart 官方 JSON 规范定义）
// ============================================================

// OrderListParams 订单查询参数
type OrderListParams struct {
	CreatedStartDate time.Time `url:"createdStartDate,omitempty"` // 必须 UTC 时间
	CreatedEndDate   time.Time `url:"createdEndDate,omitempty"`   // 必须 UTC 时间
	Status           string    `url:"status,omitempty"`           // Created/ReadyToShip/InProgress/Shipped/Delivered/Canceled
	Limit            int       `url:"limit,omitempty"`            // 默认 20，最大 50
	NextCursor       string    `url:"nextCursor,omitempty"`       // 分页游标
}

// OrderListResponse 订单列表响应
type OrderListResponse struct {
	Meta struct {
		TotalCount int    `json:"totalCount"`
		NextCursor string `json:"nextCursor,omitempty"`
		Limit      int    `json:"limit"`
	} `json:"meta"`
	Elements []Order `json:"elements"`
}

// Order 订单核心结构（仅列出常用字段，可按需扩展）
type Order struct {
	PurchaseOrderID string    `json:"purchaseOrderId"`
	CustomerOrderID string    `json:"customerOrderId"`
	Status          string    `json:"orderStatus"`
	OrderDate       time.Time `json:"orderDate"`
	CustomerEmailID string    `json:"customerEmailID"`
	ShippingInfo    struct {
		PostalAddress struct {
			Name       string `json:"name"`
			Address1   string `json:"address1"`
			Address2   string `json:"address2,omitempty"`
			City       string `json:"city"`
			State      string `json:"state"`
			PostalCode string `json:"postalCode"`
			Country    string `json:"country"`
		} `json:"postalAddress"`
		EstimatedDeliveryDate string `json:"estimatedDeliveryDate"`
	} `json:"shippingInfo"`
	OrderLines []OrderLine `json:"orderLines"`
}

type OrderLine struct {
	LineNumber       int     `json:"lineNumber"`
	ItemID           string  `json:"itemId"`
	SKU              string  `json:"sku"`
	Status           string  `json:"orderLineStatus"`
	Quantity         int     `json:"quantity"`
	UnitPrice        float64 `json:"unitPrice"`
	ChargeAmount     float64 `json:"chargeAmount,omitempty"`
}

// ShipOrderRequest 发货请求体
type ShipOrderRequest struct {
	OrderLines []ShipOrderLine `json:"orderLines"`
}

type ShipOrderLine struct {
	LineNumber        int               `json:"lineNumber"`
	OrderLineStatuses []OrderLineStatus `json:"orderLineStatuses"`
}

type OrderLineStatus struct {
	Status       string       `json:"status"` // Shipped 或 Cancelled
	TrackingInfo TrackingInfo `json:"trackingInfo"`
}

type TrackingInfo struct {
	CarrierName    string `json:"carrierName"`    // FedEx, UPS, USPS, OnTrac 等
	MethodCode     string `json:"methodCode"`     // Standard, Expedited, Priority, Freight
	TrackingNumber string `json:"trackingNumber"`
	TrackingURL    string `json:"trackingURL"`
}

// CancelOrderRequest 取消订单请求体
type CancelOrderRequest struct {
	OrderLines []CancelOrderLine `json:"orderLines"`
}

type CancelOrderLine struct {
	LineNumber     int    `json:"lineNumber"`
	CancelReason   string `json:"cancelReason"`   // CUSTOMER_CANCEL, NO_INVENTORY, PRICE_ERROR, DEFECTIVE
	CancelQuantity int    `json:"cancelQuantity"`
}

// ============================================================
// 🛠️ 核心 API 方法
// ============================================================

// ListOrders 获取订单列表（支持过滤与分页）
func (c *Client) ListOrders(ctx context.Context, params *OrderListParams) (*OrderListResponse, error) {
	query := url.Values{}
	if !params.CreatedStartDate.IsZero() {
		query.Set("createdStartDate", params.CreatedStartDate.UTC().Format(time.RFC3339))
	}
	if !params.CreatedEndDate.IsZero() {
		query.Set("createdEndDate", params.CreatedEndDate.UTC().Format(time.RFC3339))
	}
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	if params.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.NextCursor != "" {
		query.Set("nextCursor", params.NextCursor)
	}

	path := "/v3/orders"
	if q := query.Encode(); q != "" {
		path += "?" + q
	}

	var resp OrderListResponse
	err := c.Do(ctx, "GET", path, nil, &resp)
	return &resp, err
}

// GetOrder 获取单个订单详情
func (c *Client) GetOrder(ctx context.Context, purchaseOrderID string) (*Order, error) {
	path := fmt.Sprintf("/v3/orders/%s", purchaseOrderID)
	var resp Order
	err := c.Do(ctx, "GET", path, nil, &resp)
	return &resp, err
}

// AcknowledgeOrder 确认接单（状态变为 ReadyToShip）
func (c *Client) AcknowledgeOrder(ctx context.Context, purchaseOrderID string) error {
	path := fmt.Sprintf("/v3/orders/%s/acknowledge", purchaseOrderID)
	// Walmart 要求发送空 JSON 对象 {}
	return c.Do(ctx, "POST", path, map[string]interface{}{}, nil)
}

// ShipOrder 订单发货
func (c *Client) ShipOrder(ctx context.Context, purchaseOrderID string, req *ShipOrderRequest) error {
	path := fmt.Sprintf("/v3/orders/%s/ship", purchaseOrderID)
	return c.Do(ctx, "POST", path, req, nil)
}

// CancelOrder 取消订单
func (c *Client) CancelOrder(ctx context.Context, purchaseOrderID string, req *CancelOrderRequest) error {
	path := fmt.Sprintf("/v3/orders/%s/cancel", purchaseOrderID)
	return c.Do(ctx, "POST", path, req, nil)
}
```

---

### 🚀 完整调用示例 (`example/order_main.go`)

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"yourmodule/walmart" // 替换为实际模块路径
)

func main() {
	cfg := walmart.Config{
		ClientID:     "YOUR_CLIENT_ID",
		ClientSecret: "YOUR_CLIENT_SECRET",
		BaseURL:      "https://sandbox.walmartapis.com", // 生产环境请改为 https://marketplace.walmartapis.com
	}
	client := walmart.NewClient(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 1️⃣ 查询最近 24 小时待发货订单
	now := time.Now().UTC()
	yesterday := now.Add(-24 * time.Hour)
	resp, err := client.ListOrders(ctx, &walmart.OrderListParams{
		CreatedStartDate: yesterday,
		CreatedEndDate:   now,
		Status:           "ReadyToShip",
		Limit:            20,
	})
	if err != nil {
		log.Fatalf("查询订单失败: %v", err)
	}
	fmt.Printf("✅ 找到 %d 个待发货订单，游标: %s\n", resp.Meta.TotalCount, resp.Meta.NextCursor)

	for _, order := range resp.Elements {
		poID := order.PurchaseOrderID

		// 2️⃣ 确认接单
		if err := client.AcknowledgeOrder(ctx, poID); err != nil {
			log.Printf("⚠️ 订单 %s 接单失败: %v", poID, err)
			continue
		}
		fmt.Printf("✅ 订单 %s 已接单\n", poID)

		// 3️⃣ 模拟发货（实际需从 WMS/物流商获取真实单号）
		shipReq := &walmart.ShipOrderRequest{
			OrderLines: []walmart.ShipOrderLine{
				{
					LineNumber: order.OrderLines[0].LineNumber,
					OrderLineStatuses: []walmart.OrderLineStatus{
						{
							Status: "Shipped",
							TrackingInfo: walmart.TrackingInfo{
								CarrierName:    "FedEx",
								MethodCode:     "Standard",
								TrackingNumber: "FX1234567890",
								TrackingURL:    "https://www.fedex.com/fedextrack/?tracknumber=FX1234567890",
							},
						},
					},
				},
			},
		}
		if err := client.ShipOrder(ctx, poID, shipReq); err != nil {
			log.Printf("⚠️ 订单 %s 发货失败: %v", poID, err)
		} else {
			fmt.Printf("📦 订单 %s 已发货，单号: FX1234567890\n", poID)
		}
	}
}
```

---

### 📌 Walmart 订单 API 关键注意事项

| 模块 | 官方规范 & 避坑指南 |
|------|---------------------|
| **时间格式** | 查询参数必须为 **UTC 时间** 的 `RFC3339` 格式（如 `2024-05-20T10:00:00Z`）。代码已内置 `.UTC()` 转换。 |
| **订单状态机** | `Created` → `ReadyToShip`（Acknowledge 后） → `InProgress`（部分发货） → `Shipped` → `Delivered` / `Canceled`。状态流转不可逆，违反会报 `400 Invalid Status Transition`。 |
| **物流商代码** | `carrierName` 必须使用 Walmart 枚举值：`FedEx`, `UPS`, `USPS`, `OnTrac`, `Lasership`, `DHL`, `WFS` 等。大小写敏感。 |
| **取消原因** | `cancelReason` 仅限：`CUSTOMER_CANCEL`, `NO_INVENTORY`, `PRICE_ERROR`, `DEFECTIVE`, `OTHER`。必填且需匹配业务场景，否则影响 Seller 绩效。 |
| **分页机制** | Walmart 使用 `NextCursor` 游标分页，**不支持 offset/limit 传统翻页**。必须使用上次响应的 `nextCursor` 继续请求。 |
| **幂等性** | `Acknowledge`, `Ship`, `Cancel` 均为幂等操作。重复调用相同 `purchaseOrderId` 不会报错，但需确保 `orderLineStatus` 与当前状态兼容。 |

---

### 🔧 生产环境扩展建议

1. **异步批量发货**：大卖家常需处理百级订单。建议结合 `sync.WaitGroup` + 信号量控制并发（Walmart 限流约 `10 req/sec`），失败订单落库重试。
2. **Webhook 替代轮询**：Walmart 支持 `Event Notifications`（订单状态变更推送）。可结合 `ListOrders` 做补偿查询，大幅降低 API 调用量。
3. **自定义 Time 解析**：若遇到 Walmart 返回的时间带毫秒（如 `2024-05-20T10:00:00.123Z`）导致 `time.Time` 解析报错，可替换为自定义类型：
   ```go
   type WMTime struct{ time.Time }
   func (t *WMTime) UnmarshalJSON(b []byte) error {
       s := strings.Trim(string(b), `"`)
       parsed, err := time.Parse("2006-01-02T15:04:05.000Z", s)
       if err != nil { parsed, err = time.Parse(time.RFC3339, s) }
       t.Time = parsed; return err
   }
   ```
4. **日志追踪**：在 `client.go` 的 `Do` 方法中注入 `slog` 或 `zap`，打印 `WM_QOS.CORRELATION_ID`，便于与 Walmart 工单团队对齐问题。

将此文件加入你的包后，即可通过 `client.ListOrders()`, `client.ShipOrder()` 等方法直接对接订单全生命周期。如需补充 `Returns`（退货）、`Reports`（报表）或 `Inventory`（库存）模块，结构可完全复用此设计模式。


# 请封装report相关的API

以下是 Walmart **Report 模块**的完整封装。与订单模块不同，Walmart 报表生成是**异步过程**（请求 → 轮询状态 → 完成后下载），代码已针对此特性优化，并采用流式下载避免大文件内存溢出。

### 📁 新增文件：`reports.go`
将此文件放入你的 `walmart/` 包目录中。

```go
package walmart

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// ============================================================
// 📦 常量与枚举（对照 Walmart 官方规范）
// ============================================================

// ReportType 支持的报表类型
const (
	ReportTypeOrder        = "ORDER_REPORT"
	ReportTypeInventory    = "INVENTORY_REPORT"
	ReportTypeItem         = "ITEM_REPORT"
	ReportTypeReturn       = "RETURN_REPORT"
	ReportTypePerformance  = "PERFORMANCE_REPORT"
	ReportTypeSettlement   = "WFS_SETTLEMENT_REPORT_V2"
)

// ReportFileFormat 报表导出格式
const (
	FileFormatCSV  = "CSV"
	FileFormatJSON = "JSON"
)

// ReportStatus 报表生命周期状态
const (
	ReportStatusCreated    = "CREATED"
	ReportStatusInProgress = "IN_PROGRESS"
	ReportStatusCompleted  = "COMPLETED"
	ReportStatusFailed     = "FAILED"
)

// ============================================================
// 📄 请求/响应结构体
// ============================================================

// ReportCreateRequest 创建报表请求体
type ReportCreateRequest struct {
	ReportType string                 `json:"reportType"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`   // 过滤条件，因报表类型而异
	FileFormat string                 `json:"fileFormat,omitempty"` // CSV 或 JSON，默认 CSV
}

// ReportCreateResponse 创建报表响应
type ReportCreateResponse struct {
	ReportID    string `json:"reportId"`
	ReportType  string `json:"reportType"`
	Status      string `json:"status"`
	CreatedDate string `json:"createdDate"`
}

// ReportListParams 报表查询参数
type ReportListParams struct {
	ReportType   string    `url:"reportType,omitempty"`
	Status       string    `url:"status,omitempty"`
	CreatedStart time.Time `url:"createdStartDate,omitempty"`
	CreatedEnd   time.Time `url:"createdEndDate,omitempty"`
	Limit        int       `url:"limit,omitempty"`
	NextCursor   string    `url:"nextCursor,omitempty"`
}

// ReportListResponse 报表列表响应
type ReportListResponse struct {
	Meta struct {
		TotalCount int    `json:"totalCount"`
		NextCursor string `json:"nextCursor,omitempty"`
	} `json:"meta"`
	Elements []ReportDetail `json:"elements"`
}

// ReportDetail 报表详情（时间字段保留 string 避免 Walmart 格式不一致导致解析失败）
type ReportDetail struct {
	ReportID     string `json:"reportId"`
	ReportType   string `json:"reportType"`
	Status       string `json:"status"`
	CreatedDate  string `json:"createdDate"`
	ModifiedDate string `json:"modifiedDate"`
	RecordCount  int    `json:"recordCount,omitempty"`
	DownloadURL  string `json:"downloadUrl,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"` // 仅 FAILED 状态有效
}

// ============================================================
// 🛠️ 核心 API 方法
// ============================================================

// CreateReport 请求生成报表（异步任务，立即返回 reportId）
func (c *Client) CreateReport(ctx context.Context, req *ReportCreateRequest) (*ReportCreateResponse, error) {
	if req.FileFormat == "" {
		req.FileFormat = FileFormatCSV
	}
	var resp ReportCreateResponse
	err := c.Do(ctx, "POST", "/v3/reports", req, &resp)
	return &resp, err
}

// ListReports 查询报表列表（支持类型、状态、时间范围过滤）
func (c *Client) ListReports(ctx context.Context, params *ReportListParams) (*ReportListResponse, error) {
	query := url.Values{}
	if params.ReportType != "" {
		query.Set("reportType", params.ReportType)
	}
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	if !params.CreatedStart.IsZero() {
		query.Set("createdStartDate", params.CreatedStart.UTC().Format(time.RFC3339))
	}
	if !params.CreatedEnd.IsZero() {
		query.Set("createdEndDate", params.CreatedEnd.UTC().Format(time.RFC3339))
	}
	if params.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.NextCursor != "" {
		query.Set("nextCursor", params.NextCursor)
	}

	path := "/v3/reports"
	if q := query.Encode(); q != "" {
		path += "?" + q
	}

	var resp ReportListResponse
	err := c.Do(ctx, "GET", path, nil, &resp)
	return &resp, err
}

// GetReportStatus 获取指定报表状态与详情
func (c *Client) GetReportStatus(ctx context.Context, reportID string) (*ReportDetail, error) {
	path := fmt.Sprintf("/v3/reports/%s", reportID)
	var resp ReportDetail
	err := c.Do(ctx, "GET", path, nil, &resp)
	return &resp, err
}

// DownloadReport 下载已完成的报表文件（流式写入，支持大文件）
// writer 通常传入 *os.File 或 bytes.Buffer。Walmart 可能返回 302 跳转至预签名 URL，http.Client 会自动跟随。
func (c *Client) DownloadReport(ctx context.Context, reportID string, writer io.Writer) error {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return err
	}

	reqURL := c.config.BaseURL + fmt.Sprintf("/v3/reports/%s/download", reportID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("WM_SVC.NAME", "Walmart Marketplace")
	req.Header.Set("WM_QOS.CORRELATION_ID", generateCorrelationID())
	req.Header.Set("Accept", "application/octet-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return &APIError{StatusCode: resp.StatusCode, RawBody: string(body)}
	}

	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return fmt.Errorf("write report stream failed: %w", err)
	}
	return nil
}
```

---

### 🚀 使用示例：异步生成 → 轮询状态 → 流式下载

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"yourmodule/walmart" // 替换为实际模块路径
)

func main() {
	cfg := walmart.Config{
		ClientID:     "YOUR_CLIENT_ID",
		ClientSecret: "YOUR_CLIENT_SECRET",
		BaseURL:      "https://sandbox.walmartapis.com",
	}
	client := walmart.NewClient(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute) // 报表生成可能较慢
	defer cancel()

	// 1️⃣ 请求生成订单报表
	createReq := &walmart.ReportCreateRequest{
		ReportType: walmart.ReportTypeOrder,
		Metadata: map[string]interface{}{
			"createdStartDate": time.Now().Add(-7 * 24 * time.Hour).UTC().Format(time.RFC3339),
			"createdEndDate":   time.Now().UTC().Format(time.RFC3339),
			"status":           "Shipped",
		},
		FileFormat: walmart.FileFormatCSV,
	}

	createResp, err := client.CreateReport(ctx, createReq)
	if err != nil {
		log.Fatalf("创建报表失败: %v", err)
	}
	reportID := createResp.ReportID
	fmt.Printf("✅ 报表请求已提交，ReportID: %s\n", reportID)

	// 2️⃣ 轮询状态（指数退避，避免触发限流）
	pollInterval := 2 * time.Second
	maxPolls := 60 // 最长等待 2分钟
	for i := 0; i < maxPolls; i++ {
		time.Sleep(pollInterval)
		if pollInterval < 15*time.Second {
			pollInterval *= 2 // 指数退避
		}

		status, err := client.GetReportStatus(ctx, reportID)
		if err != nil {
			log.Printf("⚠️ 查询状态失败: %v", err)
			continue
		}

		switch status.Status {
		case walmart.ReportStatusCompleted:
			fmt.Printf("📊 报表生成完成，共 %d 条记录，准备下载...\n", status.RecordCount)
			
			// 3️⃣ 流式下载至本地文件
			file, err := os.Create(fmt.Sprintf("walmart_report_%s.csv", reportID))
			if err != nil {
				log.Fatalf("创建本地文件失败: %v", err)
			}
			defer file.Close()

			if err := client.DownloadReport(ctx, reportID, file); err != nil {
				log.Fatalf("下载报表失败: %v", err)
			}
			fmt.Println("✅ 报表已成功下载至本地")
			return

		case walmart.ReportStatusFailed:
			log.Fatalf("❌ 报表生成失败: %s", status.ErrorMessage)
			
		default:
			fmt.Printf("⏳ 当前状态: %s，继续等待...\n", status.Status)
		}
	}
	log.Fatal("⏱️ 报表生成超时")
}
```

---

### 📌 Walmart Report API 核心注意事项

| 模块 | 官方规范 & 避坑指南 |
|------|---------------------|
| **异步机制** | `POST /v3/reports` 仅返回 `reportId`，**不会同步返回数据**。必须轮询 `GET /v3/reports/{id}` 直到 `status=COMPLETED`。 |
| **Metadata 差异** | 不同 `reportType` 要求的 `metadata` 字段完全不同：<br>• `ORDER_REPORT`: `createdStartDate`, `createdEndDate`, `status`<br>• `WFS_SETTLEMENT_REPORT_V2`: `reportingPeriod` (如 `2024-01`)<br>• `INVENTORY_REPORT`: 无需 metadata 或仅传 `marketplace` |
| **轮询策略** | Walmart 对报表状态接口有严格限流（约 `5 req/sec`）。示例已内置 **指数退避**，生产环境建议结合 Redis 记录 `nextCheckTime` 避免无效请求。 |
| **下载重定向** | `/v3/reports/{id}/download` 通常返回 `302 Found` 跳转至 AWS S3 预签名 URL。Go 标准 `http.Client` 默认自动跟随（最多 10 次），无需手动处理。 |
| **文件解析** | CSV 表头与列顺序可能随 Walmart 更新微调。建议使用 `encoding/csv` 按列名映射而非固定索引。 |
| **数据保留期** | 生成的报表文件在 Walmart 服务器仅保留 **30 天**，超时后 `DownloadURL` 失效。务必及时落盘或同步至 OSS。 |

---

### 🔧 生产环境扩展建议

1. **封装轮询器**：若业务需高频生成报表，建议将轮询逻辑抽离为独立协程池，结合 `channel` 通知完成状态。
2. **CSV 流式解析**：大报表（>100MB）不要全量读入内存。可使用 `csv.NewReader(reader).Read()` 逐行处理并直写数据库。
3. **失败重试**：`FAILED` 状态可能由瞬时网络波动或元数据校验失败引起。可记录 `reportId` 与 `metadata`，按业务规则决定是否重试。
4. **合规日志**：在 `DownloadReport` 中注入 `slog` 记录 `reportType`、`recordCount` 与耗时，便于审计与成本分摊。

将此文件加入包后，即可通过 `client.CreateReport()` → `client.GetReportStatus()` → `client.DownloadReport()` 完整覆盖 Walmart 报表生命周期。如需对接 `Returns`、`Inventory Sync` 或 `Performance Metrics`，可沿用相同异步模式封装。


# 请封装feed相关的API

以下是 Walmart **Feed 模块**的完整封装。Feed API 用于批量操作（商品上架、库存同步、价格更新、订单处理等），具有**强异步特性**（提交 → 轮询状态 → 下载报告）。代码已针对大文件上传、流式下载、智能轮询进行优化，并与前序 `Client` 无缝集成。

### 📁 新增文件：`feeds.go`
将此文件放入你的 `walmart/` 包目录中。

```go
package walmart

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
)

// ============================================================
// 📦 常量与枚举（对照 Walmart v3 Feed API 规范）
// ============================================================

// FeedType 支持的 Feed 类型
const (
	FeedTypeItem          = "item"
	FeedTypeInventory     = "inventory"
	FeedTypePrice         = "price"
	FeedTypeOrderAck      = "orderAcknowledgement"
	FeedTypeOrderShip     = "orderShipping"
	FeedTypeOrderCancel   = "orderCancellation"
	FeedTypeReturn        = "returnCenter"
	FeedTypeCarrier       = "carrierCode"
)

// FeedStatus Feed 生命周期状态
const (
	FeedStatusCreated    = "CREATED"
	FeedStatusInProgress = "IN_PROGRESS"
	FeedStatusCompleted  = "COMPLETED"
	FeedStatusError      = "ERROR"
)

// ContentType Feed 文件格式
const (
	FeedContentTypeXML  = "application/xml"
	FeedContentTypeJSON = "application/json"
)

// ============================================================
// 📄 请求/响应结构体
// ============================================================

// FeedSubmitRequest 提交 Feed 请求
type FeedSubmitRequest struct {
	FeedType    string      // 必填：如 item, inventory, price 等
	File        io.Reader   // 必填：Feed 文件内容（XML 或 JSON）
	FileName    string      // 必填：文件名，如 feed_item.xml
	ContentType string      // 可选：application/xml 或 application/json，默认 application/xml
}

// FeedSubmitResponse 提交响应
type FeedSubmitResponse struct {
	FeedID      string `json:"feedId"`
	FeedType    string `json:"feedType"`
	Status      string `json:"status"`
	CreatedDate string `json:"createdDate"`
}

// FeedListParams 查询参数
type FeedListParams struct {
	FeedType     string    `url:"feedType,omitempty"`
	Status       string    `url:"status,omitempty"`
	CreatedStart time.Time `url:"createdStartDate,omitempty"`
	CreatedEnd   time.Time `url:"createdEndDate,omitempty"`
	Limit        int       `url:"limit,omitempty"`
	NextCursor   string    `url:"nextCursor,omitempty"`
}

// FeedListResponse 列表响应
type FeedListResponse struct {
	Meta struct {
		TotalCount int    `json:"totalCount"`
		NextCursor string `json:"nextCursor,omitempty"`
	} `json:"meta"`
	Elements []FeedStatusDetail `json:"elements"`
}

// FeedStatusDetail Feed 状态详情
type FeedStatusDetail struct {
	FeedID       string `json:"feedId"`
	FeedType     string `json:"feedType"`
	Status       string `json:"status"`
	CreatedDate  string `json:"createdDate"`
	ModifiedDate string `json:"modifiedDate"`
	// 注意：详细成功/失败记录需通过 DownloadFeedResult 获取
}

// ============================================================
// 🛠️ 核心 API 方法
// ============================================================

// SubmitFeed 提交批量 Feed（使用 multipart/form-data 上传）
func (c *Client) SubmitFeed(ctx context.Context, req *FeedSubmitRequest) (*FeedSubmitResponse, error) {
	if req.FeedType == "" {
		return nil, fmt.Errorf("feedType is required")
	}
	if req.File == nil {
		return nil, fmt.Errorf("file is required")
	}
	if req.FileName == "" {
		req.FileName = "feed.xml"
	}
	if req.ContentType == "" {
		req.ContentType = FeedContentTypeXML
	}

	// 构建 multipart 表单
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", req.FileName)
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, req.File); err != nil {
		return nil, fmt.Errorf("copy file content: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	// 拼接 URL
	query := url.Values{"feedType": {req.FeedType}}
	reqURL := c.config.BaseURL + "/v3/feeds?" + query.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, body)
	if err != nil {
		return nil, err
	}

	// 注入认证与标准 Header
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("WM_SVC.NAME", "Walmart Marketplace")
	httpReq.Header.Set("WM_QOS.CORRELATION_ID", generateCorrelationID())
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	return parseFeedResponse(resp)
}

// GetFeedStatus 查询指定 Feed 状态
func (c *Client) GetFeedStatus(ctx context.Context, feedID string) (*FeedStatusDetail, error) {
	path := fmt.Sprintf("/v3/feeds/%s", feedID)
	var resp FeedStatusDetail
	err := c.Do(ctx, http.MethodGet, path, nil, &resp)
	return &resp, err
}

// ListFeeds 查询历史 Feed 列表
func (c *Client) ListFeeds(ctx context.Context, params *FeedListParams) (*FeedListResponse, error) {
	query := url.Values{}
	if params.FeedType != "" {
		query.Set("feedType", params.FeedType)
	}
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	if !params.CreatedStart.IsZero() {
		query.Set("createdStartDate", params.CreatedStart.UTC().Format(time.RFC3339))
	}
	if !params.CreatedEnd.IsZero() {
		query.Set("createdEndDate", params.CreatedEnd.UTC().Format(time.RFC3339))
	}
	if params.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.NextCursor != "" {
		query.Set("nextCursor", params.NextCursor)
	}

	path := "/v3/feeds"
	if q := query.Encode(); q != "" {
		path += "?" + q
	}

	var resp FeedListResponse
	err := c.Do(ctx, http.MethodGet, path, nil, &resp)
	return &resp, err
}

// DownloadFeedResult 下载 Feed 执行结果（成功明细或错误报告）
func (c *Client) DownloadFeedResult(ctx context.Context, feedID string, writer io.Writer) error {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return err
	}

	reqURL := c.config.BaseURL + fmt.Sprintf("/v3/feeds/%s/result", feedID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("WM_SVC.NAME", "Walmart Marketplace")
	req.Header.Set("WM_QOS.CORRELATION_ID", generateCorrelationID())
	req.Header.Set("Accept", "*/*") // 结果可能是 XML 或 JSON

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download result failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return &APIError{StatusCode: resp.StatusCode, RawBody: string(body)}
	}

	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return fmt.Errorf("write result stream failed: %w", err)
	}
	return nil
}

// PollFeedUntilComplete 阻塞等待 Feed 处理完成（内置指数退避）
// 返回最终状态。若为 ERROR 或上下文取消，返回对应错误。
func (c *Client) PollFeedUntilComplete(ctx context.Context, feedID string, initialInterval, maxDuration time.Duration) (*FeedStatusDetail, error) {
	if initialInterval <= 0 {
		initialInterval = 2 * time.Second
	}
	if maxDuration <= 0 {
		maxDuration = 10 * time.Minute
	}

	deadline := time.Now().Add(maxDuration)
	interval := initialInterval

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			status, err := c.GetFeedStatus(ctx, feedID)
			if err != nil {
				return nil, fmt.Errorf("poll feed status: %w", err)
			}

			switch status.Status {
			case FeedStatusCompleted, FeedStatusError:
				return status, nil
			}

			// 指数退避，但不超过最大间隔 30s
			if interval < 30*time.Second {
				interval = min(interval*2, 30*time.Second)
			}
			ticker.Reset(interval)

			if time.Now().After(deadline) {
				return status, fmt.Errorf("feed polling timeout after %v", maxDuration)
			}
		}
	}
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func parseFeedResponse(resp *http.Response) (*FeedSubmitResponse, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{StatusCode: resp.StatusCode, RawBody: string(body)}
		var errResp struct {
			ErrorCode string `json:"error_code"`
			Message   string `json:"message"`
		}
		// Walmart 提交失败时有时返回 XML 错误，此处尝试解析 JSON，失败则返回原始体
		if json := tryUnmarshalJSON(body, &errResp); json == nil && errResp.ErrorCode != "" {
			apiErr.ErrorCode = errResp.ErrorCode
			apiErr.Message = errResp.Message
		}
		return nil, apiErr
	}

	var result FeedSubmitResponse
	if err := tryUnmarshalJSON(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal submit response: %w", err)
	}
	return &result, nil
}

// 辅助：安全解析 JSON，失败返回错误但不中断流程
func tryUnmarshalJSON(data []byte, v interface{}) error {
	// 简单剔除 BOM 或首尾空白
	if len(data) > 0 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		data = data[3:]
	}
	return jsonUnmarshal(data, v)
}
```

---

### 🚀 使用示例：提交 → 智能轮询 → 解析结果

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"yourmodule/walmart" // 替换为实际模块路径
)

func main() {
	cfg := walmart.Config{
		ClientID:     "YOUR_CLIENT_ID",
		ClientSecret: "YOUR_CLIENT_SECRET",
		BaseURL:      "https://sandbox.walmartapis.com",
	}
	client := walmart.NewClient(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 1️⃣ 构造 XML Feed（以库存同步为例）
	feedXML := `<?xml version="1.0" encoding="UTF-8"?>
<inventory xmlns="http://walmart.com/" version="1.5">
  <sku>TEST-SKU-001</sku>
  <quantity unit="EACH">
    <amount>50</amount>
  </quantity>
  <fulfillmentLagTime>1</fulfillmentLagTime>
</inventory>`

	req := &walmart.FeedSubmitRequest{
		FeedType:    walmart.FeedTypeInventory,
		File:        strings.NewReader(feedXML),
		FileName:    "inventory_update.xml",
		ContentType: walmart.FeedContentTypeXML,
	}

	resp, err := client.SubmitFeed(ctx, req)
	if err != nil {
		log.Fatalf("提交 Feed 失败: %v", err)
	}
	fmt.Printf("✅ Feed 已提交，ID: %s，初始状态: %s\n", resp.FeedID, resp.Status)

	// 2️⃣ 阻塞等待处理完成（自动指数退避）
	finalStatus, err := client.PollFeedUntilComplete(ctx, resp.FeedID, 2*time.Second, 3*time.Minute)
	if err != nil {
		log.Fatalf("轮询失败: %v", err)
	}

	if finalStatus.Status == walmart.FeedStatusError {
		log.Fatal("❌ Feed 处理失败，请下载报告查看详细错误")
	}
	fmt.Printf("📊 Feed 处理完成，最终状态: %s\n", finalStatus.Status)

	// 3️⃣ 下载结果报告（XML/JSON 格式，包含成功/失败明细）
	resultFile, err := os.Create(fmt.Sprintf("feed_result_%s.xml", resp.FeedID))
	if err != nil {
		log.Fatalf("创建结果文件失败: %v", err)
	}
	defer resultFile.Close()

	if err := client.DownloadFeedResult(ctx, resp.FeedID, resultFile); err != nil {
		log.Fatalf("下载报告失败: %v", err)
	}
	fmt.Println("✅ 结果报告已保存至本地")
}
```

---

### 📌 Walmart Feed API 核心注意事项

| 模块 | 官方规范 & 避坑指南 |
|------|---------------------|
| **异步处理** | `POST /v3/feeds` 仅返回 `feedId`。**必须轮询** `GET /v3/feeds/{id}` 直至 `COMPLETED` 或 `ERROR`。代码已内置 `PollFeedUntilComplete` 处理此逻辑。 |
| **Multipart 上传** | 表单字段名**必须为 `file`**。Walmart 不校验外层 `Content-Type`，但会解析文件内部格式。确保 XML 命名空间与版本匹配（如 `version="1.5"`）。 |
| **结果文件格式** | `/v3/feeds/{id}/result` 返回的格式与提交时一致（XML 提交返回 XML，JSON 提交返回 JSON）。包含 `<result>` 根节点，内含 `<success>` 与 `<error>` 明细。 |
| **限流策略** | Feed 提交限流约 `5 req/sec`，状态查询 `10 req/sec`。轮询已做指数退避，生产环境建议对同一 `feedID` 做状态缓存，避免并发重复查询。 |
| **错误定位** | `ERROR` 状态不直接在状态接口返回明细。必须调用 `DownloadFeedResult` 解析 `<error>` 节点中的 `code` 与 `description`（如 `SKU_NOT_FOUND`, `INVALID_QUANTITY`）。 |
| **大文件处理** | 若 Feed 文件 >50MB，建议分片或改用流式 `io.Reader`。当前实现支持任意大小流式上传，但需注意 `http.Client` 默认 `MaxIdleConns` 与超时设置。 |

---

### 🔧 生产环境扩展建议

1. **结果自动解析**：Walmart 返回的 XML 结果结构固定。可封装 `ParseFeedResult(reader io.Reader)` 方法，使用 `encoding/xml` 提取成功 SKU 与失败原因，直写数据库。
2. **失败重试队列**：`ERROR` 状态的 Feed 通常由数据校验失败引起。建议将失败 SKU 提取后修正，重新生成子 Feed 提交，避免全量重试。
3. **监控与告警**：在 `PollFeedUntilComplete` 中注入指标埋点（如 `feed_duration_seconds`, `feed_status`），配合 Prometheus 监控处理耗时与失败率。
4. **并发控制**：大卖家日均处理千级 Feed。建议使用 `semaphore` 限制并发提交数，结合 Redis 分布式锁防止重复提交同一 `feedType`。

将此文件加入包后，即可通过 `client.SubmitFeed()` → `client.PollFeedUntilComplete()` → `client.DownloadFeedResult()` 完整覆盖 Walmart 批量数据同步生命周期。如需对接 `Item`、`Price` 或 `Order` 类型的 Feed，只需更换 `FeedType` 与对应 XML/JSON 模板，底层流程完全一致。


# 请封装Inventory相关的API

以下是 Walmart **Inventory 模块**的完整封装。与 Feed 不同，Inventory REST API 适用于**实时/小规模**库存同步（单次查询或更新 ≤50 SKU），代码已针对查询参数拼接、批量限流、处理时间约束进行优化，并与前序 `Client` 无缝集成。

### 📁 新增文件：`inventory.go`
将此文件放入你的 `walmart/` 包目录中。

```go
package walmart

import (
	"context"
	"fmt"
	"net/url"
)

// ============================================================
// 📦 常量与状态枚举
// ============================================================

const (
	InventoryStatusInStock     = "IN_STOCK"
	InventoryStatusOutOfStock  = "OUT_OF_STOCK"
	InventoryStatusUnavailable = "UNAVAILABLE"
)

// ============================================================
// 📄 请求/响应结构体
// ============================================================

// InventoryDetail 库存详情响应
type InventoryDetail struct {
	SKU                string `json:"sku"`
	Quantity           int    `json:"quantity"`
	FulfillmentLagTime int    `json:"fulfillmentLagTime"` // 处理时效(天)，1-5
	Status             string `json:"status"`             // IN_STOCK / OUT_OF_STOCK / UNAVAILABLE
	LastUpdatedTime    string `json:"lastUpdatedTime"`
}

// InventoryUpdateRequest 库存更新请求
type InventoryUpdateRequest struct {
	SKU                string `json:"sku,omitempty"`          // Bulk 更新时必填
	Quantity           int    `json:"quantity"`               // 必填，0 表示缺货
	FulfillmentLagTime int    `json:"fulfillmentLagTime,omitempty"` // 可选，1-5 天，默认 1
}

// ============================================================
// 🛠️ 核心 API 方法
// ============================================================

// GetInventory 查询单个 SKU 库存
func (c *Client) GetInventory(ctx context.Context, sku string) (*InventoryDetail, error) {
	path := fmt.Sprintf("/v3/inventory?sku=%s", url.QueryEscape(sku))
	var resp InventoryDetail
	err := c.Do(ctx, "GET", path, nil, &resp)
	return &resp, err
}

// GetInventoryBulk 批量查询库存（REST API 建议单次 ≤20 SKU，更多请改用 Feed）
func (c *Client) GetInventoryBulk(ctx context.Context, skus []string) ([]InventoryDetail, error) {
	if len(skus) == 0 {
		return nil, fmt.Errorf("sku list cannot be empty")
	}
	// Walmart 支持多个 sku= 查询参数
	query := url.Values{}
	for _, s := range skus {
		query.Add("sku", s)
	}
	path := "/v3/inventory?" + query.Encode()

	var resp []InventoryDetail
	err := c.Do(ctx, "GET", path, nil, &resp)
	return resp, err
}

// UpdateInventory 更新单个 SKU 库存（PUT 请求）
func (c *Client) UpdateInventory(ctx context.Context, sku string, req *InventoryUpdateRequest) error {
	if sku == "" {
		return fmt.Errorf("sku cannot be empty")
	}
	if req.Quantity < 0 {
		return fmt.Errorf("quantity cannot be negative")
	}
	if req.FulfillmentLagTime > 0 && (req.FulfillmentLagTime < 1 || req.FulfillmentLagTime > 5) {
		return fmt.Errorf("fulfillmentLagTime must be between 1 and 5 days")
	}

	path := fmt.Sprintf("/v3/inventory?sku=%s", url.QueryEscape(sku))
	// PUT 请求体不包含 sku 字段（API 从 URL 获取）
	payload := map[string]interface{}{
		"quantity": req.Quantity,
	}
	if req.FulfillmentLagTime > 0 {
		payload["fulfillmentLagTime"] = req.FulfillmentLagTime
	}
	return c.Do(ctx, "PUT", path, payload, nil)
}

// UpdateInventoryBulk 批量更新库存（POST /v3/inventory/bulk）
// ⚠️ 严格限流：建议单次 ≤50 SKU，并发 ≤2。大批量更新强烈建议使用 Feed API (feedType="inventory")
func (c *Client) UpdateInventoryBulk(ctx context.Context, reqs []InventoryUpdateRequest) error {
	if len(reqs) == 0 {
		return fmt.Errorf("update request list cannot be empty")
	}
	return c.Do(ctx, "POST", "/v3/inventory/bulk", reqs, nil)
}
```

---

### 🚀 使用示例：查询 → 实时更新 → 批量同步

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"yourmodule/walmart" // 替换为实际模块路径
)

func main() {
	cfg := walmart.Config{
		ClientID:     "YOUR_CLIENT_ID",
		ClientSecret: "YOUR_CLIENT_SECRET",
		BaseURL:      "https://sandbox.walmartapis.com",
	}
	client := walmart.NewClient(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 1️⃣ 查询单个 SKU 库存
	inv, err := client.GetInventory(ctx, "SKU-001")
	if err != nil {
		log.Printf("⚠️ 查询 SKU-001 失败: %v", err)
	} else {
		fmt.Printf("📦 SKU-001 库存: %d, 状态: %s, 处理时效: %d天\n",
			inv.Quantity, inv.Status, inv.FulfillmentLagTime)
	}

	// 2️⃣ 实时更新库存（如 WMS 推送变更）
	err = client.UpdateInventory(ctx, "SKU-001", &walmart.InventoryUpdateRequest{
		Quantity:           150,
		FulfillmentLagTime: 2, // 2天内发货
	})
	if err != nil {
		log.Printf("⚠️ 更新 SKU-001 失败: %v", err)
	} else {
		fmt.Println("✅ SKU-001 库存已更新为 150")
	}

	// 3️⃣ 批量查询 & 批量更新（≤50 SKU）
	targetSKUs := []string{"SKU-002", "SKU-003", "SKU-004"}
	batchInv, err := client.GetInventoryBulk(ctx, targetSKUs)
	if err != nil {
		log.Fatalf("批量查询失败: %v", err)
	}

	// 模拟从 ERP/WMS 拉取最新库存并计算差异
	var updates []walmart.InventoryUpdateRequest
	for _, item := range batchInv {
		newQty := item.Quantity + 10 // 业务逻辑计算
		updates = append(updates, walmart.InventoryUpdateRequest{
			SKU:      item.SKU,
			Quantity: newQty,
		})
	}

	if err := client.UpdateInventoryBulk(ctx, updates); err != nil {
		log.Printf("⚠️ 批量更新失败(触发限流或数据错误): %v", err)
	} else {
		fmt.Printf("✅ 成功批量更新 %d 个 SKU 的库存\n", len(updates))
	}
}
```

---

### 📌 Walmart Inventory API 核心注意事项

| 模块 | 官方规范 & 避坑指南 |
|------|---------------------|
| **REST vs Feed 选择** | • **≤50 SKU / 实时同步**：使用本封装的 REST API<br>• **>50 SKU / 定时全量**：必须使用 `Feed API (feedType="inventory")`，否则触发 `429 Too Many Requests` 或数据覆盖失败 |
| **库存状态映射** | Walmart 根据 `quantity` 自动计算状态：<br>• `quantity > 0` → `IN_STOCK`<br>• `quantity = 0` → `OUT_OF_STOCK`<br>• 商品下架/删除 → `UNAVAILABLE`（需通过 Item Feed 操作） |
| **Fulfillment Lag Time** | 必填字段（默认 1 天），表示**接单到发货的缓冲天数**。范围 `1-5`。值过大影响搜索排名，值过小易触发履约考核扣分。 |
| **批量查询限制** | `GET /v3/inventory` 单次最多支持约 `20` 个 `sku` 参数。超过会返回 `400 Invalid Query Parameter`。本封装未硬编码限制，但生产环境建议加切片分段。 |
| **幂等性** | 库存更新为**幂等操作**。重复提交相同 `sku` 与 `quantity` 不会报错，但会刷新 `lastUpdatedTime`。 |
| **延迟生效** | 库存变更后，前台搜索与购物车通常有 **1-3 分钟延迟**。高并发促销期间可能延长至 5 分钟。 |

---

### 🔧 生产环境扩展建议

1. **分段批量处理**：若 ERP 推送超 50 SKU，建议封装切片器：
   ```go
   func chunkSKUs(skus []string, size int) [][]string {
       var chunks [][]string
       for i := 0; i < len(skus); i += size {
           end := i + size
           if end > len(skus) { end = len(skus) }
           chunks = append(chunks, skus[i:end])
       }
       return chunks
   }
   ```
2. **限流与重试**：Inventory REST 接口限流约 `10 req/sec`。建议在 `client.go` 外层包装 `retry` 逻辑（`429` 状态码时 `sleep(2s)` 重试）。
3. **差异同步优化**：避免全量推送。维护本地缓存 `map[string]int`，仅当 `newQty != cachedQty` 时调用 `UpdateInventory`，大幅降低 API 调用量。
4. **监控埋点**：在 `UpdateInventory` 前后记录 `sku`, `oldQty`, `newQty`, `latency`。若 `status` 长期为 `OUT_OF_STOCK` 但 `quantity > 0`，需排查 Walmart 内部状态机延迟。
5. **与 Feed 模块联动**：当 `UpdateInventoryBulk` 连续返回 `429` 或业务量 >100 SKU/min 时，自动降级切换至 `client.SubmitFeed(&FeedSubmitRequest{FeedType: FeedTypeInventory, ...})`，保证数据最终一致性。

将此文件加入包后，即可通过 `client.GetInventory()`, `client.UpdateInventory()`, `client.UpdateInventoryBulk()` 覆盖 Walmart 实时库存同步场景。如需对接 `Price`、`Item` 或 `Carrier` 模块，可完全复用此设计模式。


# 请封装Items相关的API

以下是 Walmart **Items 模块**的完整封装。与 Feed 模块的批量异步特性不同，Items REST API 适用于**单商品查询、创建、更新与下架**。考虑到 Walmart 商品目录结构（Catalog Schema）极其庞大且动态，本封装采用**强类型元数据 + 灵活 Payload 设计**，并与前序 `Client` 保持完全一致。

### 📁 新增文件：`items.go`
将此文件放入你的 `walmart/` 包目录中。

```go
package walmart

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// ============================================================
// 📦 常量与状态枚举
// ============================================================

const (
	ItemStatusPublished   = "PUBLISHED"
	ItemStatusUnpublished = "UNPUBLISHED"
	ItemStatusRetired     = "RETIRED"
	ItemStatusError       = "ERROR"
	ItemStatusPending     = "PENDING"
)

// ============================================================
// 📄 请求/响应结构体
// ============================================================

// ItemListParams 商品查询参数
type ItemListParams struct {
	Status         string    `url:"status,omitempty"`          // PUBLISHED/UNPUBLISHED/RETIRED/ERROR
	SKU            string    `url:"sku,omitempty"`             // 精确匹配
	UpdatedStart   time.Time `url:"updatedStartDate,omitempty"`
	UpdatedEnd     time.Time `url:"updatedEndDate,omitempty"`
	Limit          int       `url:"limit,omitempty"`           // 默认 20，最大 50
	NextCursor     string    `url:"nextCursor,omitempty"`
	IncludeDetails bool      `url:"includeDetails,omitempty"`  // true 返回完整属性树
}

// ItemListResponse 商品列表响应
type ItemListResponse struct {
	Meta struct {
		TotalCount int    `json:"totalCount"`
		NextCursor string `json:"nextCursor,omitempty"`
		Limit      int    `json:"limit"`
	} `json:"meta"`
	Elements []ItemDetail `json:"elements"`
}

// ItemDetail 商品详情核心字段
type ItemDetail struct {
	SKU            string `json:"sku"`
	Status         string `json:"status"`
	Published      bool   `json:"published"`
	LastUpdated    string `json:"lastUpdatedTime"`
	LifecycleStatus string `json:"lifecycleStatus"`
	// Walmart 商品属性树极深，此处保留通用映射。实际业务可按需解析或直连 JSON 解析器。
	Attributes     map[string]interface{} `json:"attributes,omitempty"`
	VariantGroupID string                 `json:"variantGroupId,omitempty"`
}

// ItemCreateRequest 商品创建/更新请求体
// ⚠️ 结构高度依赖 Walmart Catalog Schema，使用 interface{} 保持灵活性
// 推荐通过 JSON 模板或外部配置生成，避免硬编码深层嵌套
type ItemCreateRequest struct {
	SKU         string                 `json:"sku"`
	Product     map[string]interface{} `json:"product"`
	Price       map[string]interface{} `json:"price,omitempty"`
	Inventory   map[string]interface{} `json:"inventory,omitempty"`
}

// ItemStatusCheck 商品上架状态快速查询
type ItemStatusCheck struct {
	SKU             string `json:"sku"`
	Status          string `json:"status"`
	Published       bool   `json:"published"`
	ProcessingSteps []struct {
		Step  string `json:"step"`
		State string `json:"state"` // COMPLETED/IN_PROGRESS/ERROR
	} `json:"processingSteps"`
}

// ============================================================
// 🛠️ 核心 API 方法
// ============================================================

// ListItems 查询商品列表（支持状态、更新时间、游标分页）
func (c *Client) ListItems(ctx context.Context, params *ItemListParams) (*ItemListResponse, error) {
	query := url.Values{}
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	if params.SKU != "" {
		query.Set("sku", params.SKU)
	}
	if !params.UpdatedStart.IsZero() {
		query.Set("updatedStartDate", params.UpdatedStart.UTC().Format(time.RFC3339))
	}
	if !params.UpdatedEnd.IsZero() {
		query.Set("updatedEndDate", params.UpdatedEnd.UTC().Format(time.RFC3339))
	}
	if params.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.NextCursor != "" {
		query.Set("nextCursor", params.NextCursor)
	}
	if params.IncludeDetails {
		query.Set("includeDetails", "true")
	}

	path := "/v3/items"
	if q := query.Encode(); q != "" {
		path += "?" + q
	}

	var resp ItemListResponse
	err := c.Do(ctx, "GET", path, nil, &resp)
	return &resp, err
}

// GetItem 获取单个商品完整详情
func (c *Client) GetItem(ctx context.Context, sku string) (*ItemDetail, error) {
	path := fmt.Sprintf("/v3/items/%s", url.QueryEscape(sku))
	var resp ItemDetail
	err := c.Do(ctx, "GET", path, nil, &resp)
	return &resp, err
}

// CreateItem 创建新商品（同步返回，后台异步审核）
func (c *Client) CreateItem(ctx context.Context, req *ItemCreateRequest) error {
	if req == nil || req.SKU == "" {
		return fmt.Errorf("sku is required")
	}
	return c.Do(ctx, "POST", "/v3/items", req, nil)
}

// UpdateItem 更新已有商品（PUT 请求，支持局部更新）
func (c *Client) UpdateItem(ctx context.Context, sku string, payload interface{}) error {
	if sku == "" {
		return fmt.Errorf("sku cannot be empty")
	}
	path := fmt.Sprintf("/v3/items/%s", url.QueryEscape(sku))
	return c.Do(ctx, "PUT", path, payload, nil)
}

// ArchiveItem 下架/归档商品（软删除，状态变为 RETIRED）
func (c *Client) ArchiveItem(ctx context.Context, sku string) error {
	if sku == "" {
		return fmt.Errorf("sku cannot be empty")
	}
	path := fmt.Sprintf("/v3/items/%s", url.QueryEscape(sku))
	return c.Do(ctx, "DELETE", path, nil, nil)
}

// CheckItemStatus 获取商品处理进度与审核状态
func (c *Client) CheckItemStatus(ctx context.Context, sku string) (*ItemStatusCheck, error) {
	path := fmt.Sprintf("/v3/items/%s/status", url.QueryEscape(sku))
	var resp ItemStatusCheck
	err := c.Do(ctx, "GET", path, nil, &resp)
	return &resp, err
}
```

---

### 🚀 使用示例：创建 → 状态轮询 → 更新 → 查询

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"yourmodule/walmart" // 替换为实际模块路径
)

func main() {
	cfg := walmart.Config{
		ClientID:     "YOUR_CLIENT_ID",
		ClientSecret: "YOUR_CLIENT_SECRET",
		BaseURL:      "https://sandbox.walmartapis.com",
	}
	client := walmart.NewClient(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1️⃣ 构造商品 Payload（符合 Walmart Catalog Schema）
	newItem := &walmart.ItemCreateRequest{
		SKU: "WAL-TEST-SKU-001",
		Product: map[string]interface{}{
			"productName": "Wireless Bluetooth Headphones",
			"brand":       "AudioTech",
			"category":    []string{"Electronics", "Headphones"},
			"shortDescription": "High-quality noise cancelling headphones.",
			"mainImageUrl":    "https://example.com/image.jpg",
			"attributes": map[string]interface{}{
				"color": "Black",
				"weight": map[string]interface{}{
					"value": "250",
					"unit":  "GRAM",
				},
			},
		},
		Price: map[string]interface{}{
			"amount": 49.99,
			"currency": "USD",
		},
	}

	// 2️⃣ 创建商品
	if err := client.CreateItem(ctx, newItem); err != nil {
		log.Fatalf("创建商品失败: %v", err)
	}
	fmt.Println("✅ 商品已提交，等待 Walmart 审核...")

	// 3️⃣ 轮询审核状态（建议间隔 5-10 秒）
	for i := 0; i < 10; i++ {
		time.Sleep(5 * time.Second)
		status, err := client.CheckItemStatus(ctx, newItem.SKU)
		if err != nil {
			log.Printf("⚠️ 查询状态失败: %v", err)
			continue
		}
		fmt.Printf("🔄 SKU: %s | 状态: %s | 已发布: %v\n", status.SKU, status.Status, status.Published)
		if status.Status == walmart.ItemStatusPublished || status.Status == walmart.ItemStatusError {
			break
		}
	}

	// 4️⃣ 更新商品价格
	priceUpdate := map[string]interface{}{
		"price": map[string]interface{}{
			"amount":   39.99,
			"currency": "USD",
		},
	}
	if err := client.UpdateItem(ctx, newItem.SKU, priceUpdate); err != nil {
		log.Printf("⚠️ 更新价格失败: %v", err)
	} else {
		fmt.Println("💰 商品价格已更新为 $39.99")
	}

	// 5️⃣ 查询商品完整详情
	item, err := client.GetItem(ctx, newItem.SKU)
	if err != nil {
		log.Printf("⚠️ 查询详情失败: %v", err)
	} else {
		fmt.Printf("📦 商品 %s 当前状态: %s, 最后更新: %s\n", item.SKU, item.Status, item.LastUpdated)
	}

	// 6️⃣ 下架商品（测试清理）
	if err := client.ArchiveItem(ctx, newItem.SKU); err != nil {
		log.Printf("⚠️ 下架失败: %v", err)
	} else {
		fmt.Println("🗑️ 商品已归档下架")
	}
}
```

---

### 📌 Walmart Items API 核心注意事项

| 模块 | 官方规范 & 避坑指南 |
|------|---------------------|
| **Catalog Schema 复杂性** | Walmart 商品属性树包含 `attributes`、`variantGroupIds`、`shippingWeight` 等数百个字段。**强烈建议**使用官方提供的 JSON Schema 模板或动态生成，避免硬编码。 |
| **异步审核机制** | `POST /v3/items` 仅提交数据，**不会立即上架**。需轮询 `/v3/items/{sku}/status` 直到 `published=true`。审核耗时通常 `5分钟~24小时`。 |
| **PUT 局部更新** | `PUT` 仅更新 Payload 中提供的字段。若未传 `price`，则价格不变。删除字段需显式传 `null` 或使用 Feed API 全量覆盖。 |
| **状态枚举** | `PUBLISHED`（可售）、`UNPUBLISHED`（草稿/审核中）、`RETIRED`（已下架）、`ERROR`（校验失败）。`ERROR` 时需查看 `/status` 返回的 `processingSteps` 定位错误字段。 |
| **REST vs Feed 选择** | • **≤50 SKU / 实时操作**：使用本封装 REST API<br>• **>50 SKU / 全量同步**：必须使用 `Feed API (feedType="item")`，否则易触发限流或数据覆盖冲突 |
| **SKU 唯一性** | SKU 在 Seller 账户内全局唯一。创建失败常因 `SKU already exists`。更新时需确保 `sku` 与路径参数一致。 |

---

### 🔧 生产环境扩展建议

1. **Schema 校验前置**：在调用 `CreateItem` 前，使用 `github.com/xeipuuv/gojsonschema` 校验 Payload 是否符合 Walmart 最新 Catalog Schema，避免无效请求消耗限流配额。
2. **状态机缓存**：维护本地 `map[string]ItemStatus`，仅当 `CheckItemStatus` 返回状态变更时才更新业务数据库，降低轮询压力。
3. **变体商品处理**：若商品含多规格（颜色/尺寸），需通过 `variantGroupId` 关联。建议封装 `CreateVariantGroup` 辅助方法，统一生成 Group ID 并绑定子 SKU。
4. **错误明细解析**：`ERROR` 状态下，`processingSteps` 会返回类似：
   ```json
   {"step": "VALIDATION", "state": "ERROR", "error": "Missing required attribute: brand"}
   ```
   建议封装解析函数，自动提取错误字段并告警至运营看板。
5. **限流降级**：Items REST 接口限流约 `5-10 req/sec`。连续 `429` 时自动切换至 `Feed` 模块提交，保证数据不丢失。

将此文件加入包后，即可通过 `client.CreateItem()`, `client.CheckItemStatus()`, `client.UpdateItem()` 等方法覆盖 Walmart 商品全生命周期管理。结合已封装的 `Inventory`、`Feed`、`Order` 模块，即可构建完整的企业级 Walmart Marketplace 集成方案。

# 请封装Fulfillment相关的API

以下是 Walmart **Fulfillment (WFS 入仓)** 模块的完整封装。该模块核心围绕 **履约中心查询、入仓单创建、状态跟踪、物流信息更新、标签下载** 展开。代码已针对二进制标签下载、异步状态机、分页游标进行优化，并与前序 `Client` 保持架构一致。

### 📁 新增文件：`fulfillment.go`
将此文件放入你的 `walmart/` 包目录中。

```go
package walmart

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// ============================================================
// 📦 常量与状态枚举（对照 Walmart WFS 规范）
// ============================================================

// InboundStatus 入仓单生命周期状态
const (
	InboundStatusCreated    = "CREATED"
	InboundStatusInProgress = "IN_PROGRESS"
	InboundStatusInTransit  = "IN_TRANSIT"
	InboundStatusReceived   = "RECEIVED"
	InboundStatusClosed     = "CLOSED"
	InboundStatusCancelled  = "CANCELLED"
)

// LabelType 标签类型（用于下载箱标/托盘标）
const (
	LabelTypePallet = "PALLET"
	LabelTypeCase   = "CASE"
	LabelTypeItem   = "ITEM"
)

// FacilityType 履约中心类型
const (
	FacilityTypeFC = "FC" // Fulfillment Center
	FacilityTypeDS = "DS" // Delivery Station
)

// ============================================================
// 📄 请求/响应结构体
// ============================================================

// Facility 履约中心信息
type Facility struct {
	FacilityID   string `json:"facilityId"`
	FacilityName string `json:"facilityName"`
	Type         string `json:"facilityType"` // FC / DS
	Address      struct {
		Address1   string `json:"address1"`
		Address2   string `json:"address2,omitempty"`
		City       string `json:"city"`
		State      string `json:"state"`
		PostalCode string `json:"postalCode"`
		Country    string `json:"country"`
	} `json:"address"`
}

// InboundCreateRequest 创建入仓单请求
type InboundCreateRequest struct {
	FacilityID  string        `json:"facilityId"`  // 必填：目标FC ID
	ReferenceID string        `json:"referenceId"` // 必填：卖家内部单号（需全局唯一）
	Items       []InboundItem `json:"items"`       // 必填：商品明细
	// 其他动态字段可通过 embedding 或 map 补充
}

// InboundItem 入仓商品明细
type InboundItem struct {
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
}

// InboundShipment 入仓单详情
type InboundShipment struct {
	InboundID      string        `json:"inboundId"`
	FacilityID     string        `json:"facilityId"`
	ReferenceID    string        `json:"referenceId"`
	Status         string        `json:"status"`
	CreatedDate    string        `json:"createdDate"`
	ModifiedDate   string        `json:"modifiedDate"`
	Items          []InboundItem `json:"items"`
	TrackingInfo   []TrackingInfo `json:"trackingInfo,omitempty"`
}

// InboundListParams 入仓单查询参数
type InboundListParams struct {
	Status       string    `url:"status,omitempty"`
	CreatedStart time.Time `url:"createdStartDate,omitempty"`
	CreatedEnd   time.Time `url:"createdEndDate,omitempty"`
	FacilityID   string    `url:"facilityId,omitempty"`
	Limit        int       `url:"limit,omitempty"`
	NextCursor   string    `url:"nextCursor,omitempty"`
}

// InboundListResponse 入仓单列表响应
type InboundListResponse struct {
	Meta struct {
		TotalCount int    `json:"totalCount"`
		NextCursor string `json:"nextCursor,omitempty"`
	} `json:"meta"`
	Elements []InboundShipment `json:"elements"`
}

// InboundUpdateTrackingRequest 更新入仓物流信息
type InboundUpdateTrackingRequest struct {
	CarrierName    string `json:"carrierName"`
	MethodCode     string `json:"methodCode"` // LTL, FTL, Parcel 等
	TrackingNumber string `json:"trackingNumber"`
	ShipDate       string `json:"shipDate"` // RFC3339 格式
}

// ============================================================
// 🛠️ 核心 API 方法
// ============================================================

// ListFacilities 查询可用履约中心列表
func (c *Client) ListFacilities(ctx context.Context, facilityType string) ([]Facility, error) {
	query := url.Values{}
	if facilityType != "" {
		query.Set("facilityType", facilityType)
	}
	path := "/v3/facilities"
	if q := query.Encode(); q != "" {
		path += "?" + q
	}

	var resp []Facility
	err := c.Do(ctx, http.MethodGet, path, nil, &resp)
	return resp, err
}

// CreateInbound 创建入仓单（同步返回 inboundId）
func (c *Client) CreateInbound(ctx context.Context, req *InboundCreateRequest) (*InboundShipment, error) {
	if req == nil || req.FacilityID == "" || req.ReferenceID == "" {
		return nil, fmt.Errorf("facilityId and referenceId are required")
	}
	var resp InboundShipment
	err := c.Do(ctx, http.MethodPost, "/v3/facilities/inbounds", req, &resp)
	return &resp, err
}

// GetInbound 获取入仓单详情与当前状态
func (c *Client) GetInbound(ctx context.Context, inboundID string) (*InboundShipment, error) {
	path := fmt.Sprintf("/v3/facilities/inbounds/%s", inboundID)
	var resp InboundShipment
	err := c.Do(ctx, http.MethodGet, path, nil, &resp)
	return &resp, err
}

// ListInbounds 查询历史入仓单（支持状态、时间、FC过滤）
func (c *Client) ListInbounds(ctx context.Context, params *InboundListParams) (*InboundListResponse, error) {
	query := url.Values{}
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	if !params.CreatedStart.IsZero() {
		query.Set("createdStartDate", params.CreatedStart.UTC().Format(time.RFC3339))
	}
	if !params.CreatedEnd.IsZero() {
		query.Set("createdEndDate", params.CreatedEnd.UTC().Format(time.RFC3339))
	}
	if params.FacilityID != "" {
		query.Set("facilityId", params.FacilityID)
	}
	if params.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.NextCursor != "" {
		query.Set("nextCursor", params.NextCursor)
	}

	path := "/v3/facilities/inbounds"
	if q := query.Encode(); q != "" {
		path += "?" + q
	}

	var resp InboundListResponse
	err := c.Do(ctx, http.MethodGet, path, nil, &resp)
	return &resp, err
}

// UpdateInboundTracking 更新入仓单物流信息（状态转为 IN_TRANSIT）
func (c *Client) UpdateInboundTracking(ctx context.Context, inboundID string, req *InboundUpdateTrackingRequest) error {
	path := fmt.Sprintf("/v3/facilities/inbounds/%s/tracking", inboundID)
	return c.Do(ctx, http.MethodPut, path, req, nil)
}

// DownloadInboundLabels 下载入仓标签（PDF 二进制流）
// labelType: PALLET / CASE / ITEM
func (c *Client) DownloadInboundLabels(ctx context.Context, inboundID, labelType string, writer io.Writer) error {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return err
	}

	query := url.Values{"labelType": {labelType}}
	reqURL := c.config.BaseURL + fmt.Sprintf("/v3/facilities/inbounds/%s/labels?%s", inboundID, query.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("WM_SVC.NAME", "Walmart Marketplace")
	req.Header.Set("WM_QOS.CORRELATION_ID", generateCorrelationID())
	req.Header.Set("Accept", "application/pdf") // 标签固定为 PDF

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download label request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return &APIError{StatusCode: resp.StatusCode, RawBody: string(body)}
	}

	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return fmt.Errorf("write label stream failed: %w", err)
	}
	return nil
}
```

---

### 🚀 使用示例：查FC → 创建入仓 → 更新物流 → 下载标签

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"yourmodule/walmart" // 替换为实际模块路径
)

func main() {
	cfg := walmart.Config{
		ClientID:     "YOUR_CLIENT_ID",
		ClientSecret: "YOUR_CLIENT_SECRET",
		BaseURL:      "https://sandbox.walmartapis.com",
	}
	client := walmart.NewClient(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 1️⃣ 查询可用 FC
	fcs, err := client.ListFacilities(ctx, walmart.FacilityTypeFC)
	if err != nil {
		log.Fatalf("查询履约中心失败: %v", err)
	}
	if len(fcs) == 0 {
		log.Fatal("未找到可用 FC")
	}
	targetFC := fcs[0]
	fmt.Printf("📍 目标履约中心: %s (%s)\n", targetFC.FacilityName, targetFC.FacilityID)

	// 2️⃣ 创建入仓单
	createReq := &walmart.InboundCreateRequest{
		FacilityID:  targetFC.FacilityID,
		ReferenceID: fmt.Sprintf("REF-%d", time.Now().Unix()),
		Items: []walmart.InboundItem{
			{SKU: "SKU-001", Quantity: 100},
			{SKU: "SKU-002", Quantity: 50},
		},
	}
	inbound, err := client.CreateInbound(ctx, createReq)
	if err != nil {
		log.Fatalf("创建入仓单失败: %v", err)
	}
	fmt.Printf("✅ 入仓单创建成功, ID: %s, 初始状态: %s\n", inbound.InboundID, inbound.Status)

	// 3️⃣ 安排物流后更新 Tracking（状态变为 IN_TRANSIT）
	trackingReq := &walmart.InboundUpdateTrackingRequest{
		CarrierName:    "FedEx",
		MethodCode:     "LTL",
		TrackingNumber: "FX-987654321",
		ShipDate:       time.Now().UTC().Format(time.RFC3339),
	}
	if err := client.UpdateInboundTracking(ctx, inbound.InboundID, trackingReq); err != nil {
		log.Printf("⚠️ 更新物流信息失败: %v", err)
	} else {
		fmt.Println("🚚 物流信息已更新，入仓单状态转为 IN_TRANSIT")
	}

	// 4️⃣ 下载箱标（CASE）用于贴标发货
	labelFile, err := os.Create(fmt.Sprintf("inbound_%s_case_labels.pdf", inbound.InboundID))
	if err != nil {
		log.Fatalf("创建标签文件失败: %v", err)
	}
	defer labelFile.Close()

	if err := client.DownloadInboundLabels(ctx, inbound.InboundID, walmart.LabelTypeCase, labelFile); err != nil {
		log.Fatalf("下载标签失败: %v", err)
	}
	fmt.Println("🏷️ 箱标已下载至本地，请打印并粘贴至外箱")
}
```

---

### 📌 Walmart Fulfillment API 核心注意事项

| 模块 | 官方规范 & 避坑指南 |
|------|---------------------|
| **状态机流转** | `CREATED` → `IN_PROGRESS`（仓库接单） → `IN_TRANSIT`（更新Tracking后） → `RECEIVED`（FC签收） → `CLOSED`（上架完成）。**不可跳跃**，违反会报 `400 Invalid State Transition`。 |
| **ReferenceID 唯一性** | `referenceId` 是卖家内部单号，**必须全局唯一且不可重复使用**。建议格式：`WFS-YYYYMMDD-SEQ`。重复提交会直接报 `409 Conflict`。 |
| **标签下载时机** | 仅当状态为 `IN_PROGRESS` 或 `IN_TRANSIT` 时允许下载 `CASE`/`PALLET` 标签。`CREATED` 阶段下载可能返回空或报错。 |
| **物流承运商代码** | `carrierName` 必须使用 Walmart 白名单值：`FedEx`, `UPS`, `USPS`, `XPO`, `YRC`, `Estes`, `Saia` 等。大小写敏感，拼写错误会导致路由失败。 |
| **REST vs Feed 选择** | • **日常入仓操作**：使用本封装 REST API<br>• **大批量/自动化预约**：Walmart 推荐结合 `WFS Appointment API` 或 `Feed` 模块提交预入库计划 |
| **上架延迟** | FC 签收 (`RECEIVED`) 到前台可售 (`IN_STOCK`) 通常有 **24-48 小时** 上架处理期。大促期间可能延长至 3-5 天。 |

---

### 🔧 生产环境扩展建议

1. **状态机本地同步**：维护 `map[string]InboundStatus` 缓存，结合 Webhook（若开通）或定时轮询，自动触发下游动作（如：状态变 `RECEIVED` 时通知财务结算预估费用）。
2. **标签自动打印集成**：下载的二进制 PDF 可直接对接 `CUPS`、`Zebra ZPL` 或云服务（如 `Lob`, `Shippo`）。建议在 `DownloadInboundLabels` 后封装 `PrintLabel(pdf []byte, printerIP string)`。
3. **异常拦截**：若 `GetInbound` 返回 `CANCELLED`，立即检查 `ReferenceID` 是否被重复使用、FC 是否爆仓、或 SKU 是否被 WFS 拒收。记录日志并告警至仓储 WMS。
4. **限流与重试**：Fulfillment 接口限流约 `5 req/sec`。连续 `429` 时采用指数退避，并在重试间加入 `time.Sleep` 避免雪崩。
5. **多仓路由策略**：Walmart WFS 支持智能分仓。可封装 `SelectOptimalFacility(sku, qty, region)` 方法，结合 `ListFacilities` 响应与历史上架时效数据，动态选择最优 FC。

将此文件加入包后，即可通过 `client.CreateInbound()`, `client.UpdateInboundTracking()`, `client.DownloadInboundLabels()` 完整覆盖 Walmart 履约入仓核心流程。结合已封装的 `Inventory`、`Items`、`Orders`、`Feeds` 模块，你已具备构建 **企业级 Walmart Marketplace + WFS 全链路集成系统** 的全部基础设施。


# 请封装Prices相关的API

以下是 Walmart **Prices 模块**的完整封装。价格同步属于高频核心操作，Walmart 对价格接口有严格的**MAP 政策校验、竞争性定价规则与限流策略**。本封装采用**指针类型处理可选字段**、**前置参数校验**，并与前序 `Client` 保持架构一致。

### 📁 新增文件：`prices.go`
将此文件放入你的 `walmart/` 包目录中。

```go
package walmart

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// ============================================================
// 📦 常量与策略枚举
// ============================================================

const (
	CurrencyUSD = "USD"
	CurrencyCAD = "CAD"
	CurrencyMXN = "MXN"
)

// PriceUpdateStrategy 价格更新策略（影响 Buy Box 与搜索排名）
const (
	StrategyCompetitive = "COMPETITIVE" // 匹配竞品价（需开通 Price Match）
	StrategyStatic      = "STATIC"      // 固定价格
)

// ============================================================
// 📄 请求/响应结构体
// ============================================================

// PriceDetail 商品价格详情响应
type PriceDetail struct {
	SKU             string    `json:"sku"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	ComparisonPrice float64   `json:"comparisonPrice,omitempty"` // “原价”或 MAP 价
	EffectiveDate   string    `json:"effectiveDate,omitempty"`   // RFC3339，为空表示立即生效
	Retire          bool      `json:"retire,omitempty"`          // 是否下架
	ShippingCost    float64   `json:"shippingCost,omitempty"`
	LastUpdatedTime string    `json:"lastUpdatedTime"`
}

// PriceUpdateRequest 价格更新请求体
// 使用指针区分“未设置”与“零值”，避免误传 0 或空时间
type PriceUpdateRequest struct {
	Amount          float64    `json:"amount"`
	Currency        string     `json:"currency"`
	ComparisonPrice *float64   `json:"comparisonPrice,omitempty"`
	EffectiveDate   *time.Time `json:"effectiveDate,omitempty"`
	Retire          *bool      `json:"retire,omitempty"`
	ShippingCost    *float64   `json:"shippingCost,omitempty"`
}

// Validate 前置校验，避免无效请求消耗限流配额
func (r *PriceUpdateRequest) Validate() error {
	if r.Amount <= 0 {
		return fmt.Errorf("price amount must be greater than 0")
	}
	if r.Currency == "" {
		r.Currency = CurrencyUSD
	}
	if r.ComparisonPrice != nil && *r.ComparisonPrice < r.Amount {
		return fmt.Errorf("comparisonPrice must be >= amount")
	}
	if r.EffectiveDate != nil && r.EffectiveDate.Before(time.Now().UTC()) {
		return fmt.Errorf("effectiveDate cannot be in the past")
	}
	return nil
}

// ============================================================
// 🛠️ 核心 API 方法
// ============================================================

// GetPrice 查询单个 SKU 当前价格与生效策略
func (c *Client) GetPrice(ctx context.Context, sku string) (*PriceDetail, error) {
	path := fmt.Sprintf("/v3/items/%s/price", url.QueryEscape(sku))
	var resp PriceDetail
	err := c.Do(ctx, "GET", path, nil, &resp)
	return &resp, err
}

// UpdatePrice 更新单个 SKU 价格（PUT 请求，支持部分字段更新）
func (c *Client) UpdatePrice(ctx context.Context, sku string, req *PriceUpdateRequest) error {
	if sku == "" {
		return fmt.Errorf("sku cannot be empty")
	}
	if req == nil {
		return fmt.Errorf("price update request cannot be nil")
	}
	if err := req.Validate(); err != nil {
		return err
	}

	path := fmt.Sprintf("/v3/items/%s/price", url.QueryEscape(sku))
	return c.Do(ctx, "PUT", path, req, nil)
}

// UpdatePriceBulk 批量更新价格（REST 接口，单次建议 ≤20 SKU）
// ⚠️ 超过 20 SKU 或高频调用强烈建议降级至 Feed API (feedType="price")
func (c *Client) UpdatePriceBulk(ctx context.Context, updates map[string]PriceUpdateRequest) error {
	if len(updates) == 0 {
		return fmt.Errorf("updates map cannot be empty")
	}

	// Walmart REST 批量接口：POST /v3/prices
	payload := make([]map[string]interface{}, 0, len(updates))
	for sku, req := range updates {
		if err := req.Validate(); err != nil {
			return fmt.Errorf("invalid price for sku %s: %w", sku, err)
		}
		item := map[string]interface{}{
			"sku":      sku,
			"amount":   req.Amount,
			"currency": req.Currency,
		}
		if req.ComparisonPrice != nil {
			item["comparisonPrice"] = *req.ComparisonPrice
		}
		if req.EffectiveDate != nil {
			item["effectiveDate"] = req.EffectiveDate.UTC().Format(time.RFC3339)
		}
		if req.Retire != nil {
			item["retire"] = *req.Retire
		}
		if req.ShippingCost != nil {
			item["shippingCost"] = *req.ShippingCost
		}
		payload = append(payload, item)
	}

	return c.Do(ctx, "POST", "/v3/prices", payload, nil)
}
```

---

### 🚀 使用示例：查询 → 校验 → 实时更新 → 批量同步

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"yourmodule/walmart" // 替换为实际模块路径
)

func main() {
	cfg := walmart.Config{
		ClientID:     "YOUR_CLIENT_ID",
		ClientSecret: "YOUR_CLIENT_SECRET",
		BaseURL:      "https://sandbox.walmartapis.com",
	}
	client := walmart.NewClient(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1️⃣ 查询当前价格
	price, err := client.GetPrice(ctx, "SKU-001")
	if err != nil {
		log.Printf("⚠️ 查询价格失败: %v", err)
	} else {
		fmt.Printf("🏷️ SKU-001 现价: %.2f %s (对比价: %.2f)\n", price.Amount, price.Currency, price.ComparisonPrice)
	}

	// 2️⃣ 更新单个价格（含前置校验）
	compPrice := 29.99
	now := time.Now().UTC().Add(1 * time.Hour) // 1小时后生效
	updateReq := &walmart.PriceUpdateRequest{
		Amount:          19.99,
		Currency:        walmart.CurrencyUSD,
		ComparisonPrice: &compPrice,
		EffectiveDate:   &now,
	}

	if err := client.UpdatePrice(ctx, "SKU-001", updateReq); err != nil {
		log.Printf("⚠️ 更新价格失败: %v", err)
	} else {
		fmt.Println("✅ SKU-001 价格已提交，将于 1 小时后生效")
	}

	// 3️⃣ 批量更新价格（≤20 SKU）
	batchUpdates := map[string]walmart.PriceUpdateRequest{
		"SKU-002": {Amount: 45.50, Currency: walmart.CurrencyUSD},
		"SKU-003": {Amount: 12.00, Currency: walmart.CurrencyUSD, ComparisonPrice: &compPrice},
	}

	if err := client.UpdatePriceBulk(ctx, batchUpdates); err != nil {
		log.Printf("⚠️ 批量更新失败(建议改用 Feed API): %v", err)
	} else {
		fmt.Printf("✅ 成功提交 %d 个 SKU 的价格变更请求\n", len(batchUpdates))
	}
}
```

---

### 📌 Walmart Prices API 核心注意事项

| 模块 | 官方规范 & 避坑指南 |
|------|---------------------|
| **MAP 与合规校验** | Walmart 严格执行 **Minimum Advertised Price (MAP)**。若 `amount` 低于品牌方设定的 MAP 价，接口返回 `400 MAP_VIOLATION`。需提前在 Partner Portal 配置豁免或同步合法定价。 |
| **ComparisonPrice 规则** | “对比价”必须 `≥ amount`，且需有历史销售记录支撑。频繁设置虚高对比价会触发算法降权或 `PRICE_MANIPULATION` 风控。 |
| **生效时间** | `effectiveDate` 必须为 **UTC 时间且 ≥ 当前时间**。若省略则立即生效。修改已生效价格需确保新价不违反价格保护期政策。 |
| **REST vs Feed 选择** | • **≤20 SKU / 实时调价**：使用本封装 REST API<br>• **>20 SKU / 定时全量/促销价**：必须使用 `Feed API (feedType="price")`，REST 会返回 `429` 或 `413 Payload Too Large` |
| **Buy Box 影响** | 价格变动后，Walmart 算法需 `5-15 分钟` 重新计算购物车资格。频繁调价（<1次/小时）可能导致排名波动或临时失去 Buy Box。 |
| **货币与税费** | `currency` 必须与站点匹配（US=`USD`, CA=`CAD`, MX=`MXN`）。价格不含税，Walmart 根据收货地址自动计算 Sales Tax。 |

---

### 🔧 生产环境扩展建议

1. **价格缓存与差异同步**：维护本地 `map[string]float64` 缓存。仅当 `newPrice != cachedPrice` 时调用 API，减少 `90%+` 无效请求。
2. **MAP 合规中间件**：在 `UpdatePrice` 前接入品牌方 MAP 表或外部合规 API，拦截违规价格并告警至运营看板。
3. **限流与降级策略**：
   ```go
   // 伪代码：连续 429 时自动降级至 Feed
   if err == walmart.ErrRateLimited && len(updates) > 10 {
       return client.SubmitFeed(ctx, &walmart.FeedSubmitRequest{
           FeedType: walmart.FeedTypePrice, File: generatePriceFeedXML(updates), ...
       })
   }
   ```
4. **促销价生命周期管理**：结合 `EffectiveDate` 实现定时生效/恢复原价。建议封装 `SchedulePrice(sku, promoPrice, start, end, originalPrice)` 自动维护价格状态机。
5. **监控埋点**：记录 `price_change_latency`, `map_violation_count`, `effective_date_drift`。若 `lastUpdatedTime` 与提交时间差 >10 分钟，触发异步补偿查询。

将此文件加入包后，即可通过 `client.GetPrice()`, `client.UpdatePrice()`, `client.UpdatePriceBulk()` 覆盖 Walmart 实时价格同步场景。结合已封装的 `Inventory`、`Items`、`Orders`、`Feeds`、`Fulfillment` 模块，你已具备构建 **企业级 Walmart Marketplace 全链路自动化运营系统** 的完整基础设施。


# 请封装Promotion相关的API

以下是 Walmart **Promotion 模块**的完整封装。促销活动涉及复杂的规则引擎（折扣类型、适用条件、时间窗口、优先级冲突），本封装采用**强类型基础字段 + 灵活规则映射**设计，内置前置校验与时间标准化，并与前序 `Client` 保持完全一致的架构风格。

### 📁 新增文件：`promotions.go`
将此文件放入你的 `walmart/` 包目录中。

```go
package walmart

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// ============================================================
// 📦 常量与枚举（对照 Walmart v3 Promotion API 规范）
// ============================================================

// PromotionStatus 促销生命周期状态
const (
	PromoStatusDraft     = "DRAFT"
	PromoStatusActive    = "ACTIVE"
	PromoStatusExpired   = "EXPIRED"
	PromoStatusCompleted = "COMPLETED"
	PromoStatusError     = "ERROR"
	PromoStatusSuspended = "SUSPENDED"
)

// DiscountType 折扣类型
const (
	DiscountTypePercent     = "PERCENTAGE"
	DiscountTypeFixedAmount = "FIXED_AMOUNT"
	DiscountTypeFreeShipping = "FREE_SHIPPING"
	DiscountTypeBOGO        = "BOGO" // Buy One Get One
)

// EligibilityType 适用条件类型
const (
	EligibilitySKU          = "SKU"
	EligibilityCategory     = "CATEGORY"
	EligibilityMinCartValue = "MIN_CART_VALUE"
	EligibilityCustomerType = "CUSTOMER_TYPE" // e.g., "WALMART_PLUS"
)

// ============================================================
// 📄 请求/响应结构体
// ============================================================

// PromotionCreateRequest 创建/更新促销请求体
// 注意：discount 与 eligibility 为动态结构，使用 map 保持灵活性
type PromotionCreateRequest struct {
	PromotionName string      `json:"promotionName"`
	Description   string      `json:"description,omitempty"`
	StartDate     time.Time   `json:"startDate"`
	EndDate       time.Time   `json:"endDate"`
	Channel       string      `json:"channel,omitempty"` // "WALMART_US", "WALMART_CA" 等
	Discount      interface{} `json:"discount"`          // 必填，见下方示例
	Eligibility   interface{} `json:"eligibility"`       // 必填，见下方示例
	Priority      int         `json:"priority,omitempty"` // 数值越小优先级越高 (1-100)
	Status        string      `json:"status,omitempty"`   // DRAFT / ACTIVE
}

// Validate 前置业务校验，避免无效请求触发 400 错误
func (r *PromotionCreateRequest) Validate() error {
	if r.PromotionName == "" {
		return fmt.Errorf("promotionName is required")
	}
	if r.Discount == nil {
		return fmt.Errorf("discount rule is required")
	}
	if r.Eligibility == nil {
		return fmt.Errorf("eligibility rule is required")
	}
	if !r.StartDate.Before(r.EndDate) {
		return fmt.Errorf("startDate must be before endDate")
	}
	if r.StartDate.Before(time.Now().UTC().Add(-24 * time.Hour)) {
		return fmt.Errorf("startDate cannot be in the past (max 24h grace period)")
	}
	if r.Status == "" {
		r.Status = PromoStatusDraft
	}
	if r.Priority == 0 {
		r.Priority = 50 // 默认中等优先级
	}
	return nil
}

// PromotionDetail 促销详情响应
type PromotionDetail struct {
	PromotionID   string                 `json:"promotionId"`
	PromotionName string                 `json:"promotionName"`
	Description   string                 `json:"description,omitempty"`
	Status        string                 `json:"status"`
	StartDate     string                 `json:"startDate"`
	EndDate       string                 `json:"endDate"`
	Discount      map[string]interface{} `json:"discount,omitempty"`
	Eligibility   map[string]interface{} `json:"eligibility,omitempty"`
	Priority      int                    `json:"priority"`
	Channel       string                 `json:"channel,omitempty"`
	CreatedDate   string                 `json:"createdDate"`
	ModifiedDate  string                 `json:"modifiedDate"`
	ErrorDetails  string                 `json:"errorDetails,omitempty"` // 仅 ERROR 状态有效
}

// PromotionListParams 促销查询参数
type PromotionListParams struct {
	Status       string    `url:"status,omitempty"`
	Channel      string    `url:"channel,omitempty"`
	CreatedStart time.Time `url:"createdStartDate,omitempty"`
	CreatedEnd   time.Time `url:"createdEndDate,omitempty"`
	Limit        int       `url:"limit,omitempty"`
	NextCursor   string    `url:"nextCursor,omitempty"`
}

// PromotionListResponse 促销列表响应
type PromotionListResponse struct {
	Meta struct {
		TotalCount int    `json:"totalCount"`
		NextCursor string `json:"nextCursor,omitempty"`
	} `json:"meta"`
	Elements []PromotionDetail `json:"elements"`
}

// ============================================================
// 🛠️ 核心 API 方法
// ============================================================

// CreatePromotion 创建新促销（同步返回 promotionId，后台异步校验）
func (c *Client) CreatePromotion(ctx context.Context, req *PromotionCreateRequest) (*PromotionDetail, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var resp PromotionDetail
	err := c.Do(ctx, "POST", "/v3/promotions", req, &resp)
	return &resp, err
}

// UpdatePromotion 更新已有促销（PUT 请求，支持全量/部分字段覆盖）
func (c *Client) UpdatePromotion(ctx context.Context, promotionID string, req *PromotionCreateRequest) error {
	if promotionID == "" {
		return fmt.Errorf("promotionID cannot be empty")
	}
	if err := req.Validate(); err != nil {
		return err
	}
	path := fmt.Sprintf("/v3/promotions/%s", promotionID)
	return c.Do(ctx, "PUT", path, req, nil)
}

// GetPromotion 获取单个促销详情与当前状态
func (c *Client) GetPromotion(ctx context.Context, promotionID string) (*PromotionDetail, error) {
	path := fmt.Sprintf("/v3/promotions/%s", promotionID)
	var resp PromotionDetail
	err := c.Do(ctx, "GET", path, nil, &resp)
	return &resp, err
}

// ListPromotions 查询历史促销列表（支持状态、渠道、时间过滤与游标分页）
func (c *Client) ListPromotions(ctx context.Context, params *PromotionListParams) (*PromotionListResponse, error) {
	query := url.Values{}
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	if params.Channel != "" {
		query.Set("channel", params.Channel)
	}
	if !params.CreatedStart.IsZero() {
		query.Set("createdStartDate", params.CreatedStart.UTC().Format(time.RFC3339))
	}
	if !params.CreatedEnd.IsZero() {
		query.Set("createdEndDate", params.CreatedEnd.UTC().Format(time.RFC3339))
	}
	if params.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.NextCursor != "" {
		query.Set("nextCursor", params.NextCursor)
	}

	path := "/v3/promotions"
	if q := query.Encode(); q != "" {
		path += "?" + q
	}

	var resp PromotionListResponse
	err := c.Do(ctx, "GET", path, nil, &resp)
	return &resp, err
}

// UpdatePromotionStatus 快捷方法：仅更新促销状态（ACTIVE/DRAFT/SUSPENDED）
func (c *Client) UpdatePromotionStatus(ctx context.Context, promotionID, newStatus string) error {
	if promotionID == "" || newStatus == "" {
		return fmt.Errorf("promotionID and newStatus are required")
	}
	payload := map[string]string{"status": newStatus}
	path := fmt.Sprintf("/v3/promotions/%s/status", promotionID)
	return c.Do(ctx, "POST", path, payload, nil)
}
```

---

### 🚀 使用示例：创建 → 激活 → 查询 → 状态管理

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"yourmodule/walmart" // 替换为实际模块路径
)

func main() {
	cfg := walmart.Config{
		ClientID:     "YOUR_CLIENT_ID",
		ClientSecret: "YOUR_CLIENT_SECRET",
		BaseURL:      "https://sandbox.walmartapis.com",
	}
	client := walmart.NewClient(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1️⃣ 构造促销规则（JSON 映射需严格符合 Walmart Schema）
	now := time.Now().UTC()
	start := now.Add(2 * time.Hour)
	end := now.Add(7 * 24 * time.Hour)

	// 折扣规则示例：满减 15%
	discountRule := map[string]interface{}{
		"type":  walmart.DiscountTypePercent,
		"value": 15.0,
	}
	// 适用条件示例：仅限指定 SKU 且购物车满 $50
	eligibilityRule := map[string]interface{}{
		"type": walmart.EligibilityMinCartValue,
		"value": 50.00,
		"scope": map[string]interface{}{
			"type":  walmart.EligibilitySKU,
			"value": []string{"SKU-001", "SKU-002"},
		},
	}

	promoReq := &walmart.PromotionCreateRequest{
		PromotionName: "Spring Sale - 15% OFF",
		Description:   "Limited time discount on select electronics",
		StartDate:     start,
		EndDate:       end,
		Channel:       "WALMART_US",
		Discount:      discountRule,
		Eligibility:   eligibilityRule,
		Priority:      10, // 高优先级
		Status:        walmart.PromoStatusDraft,
	}

	// 2️⃣ 创建促销
	promo, err := client.CreatePromotion(ctx, promoReq)
	if err != nil {
		log.Fatalf("创建促销失败: %v", err)
	}
	fmt.Printf("✅ 促销已提交, ID: %s, 状态: %s, 预计生效: %s\n", promo.PromotionID, promo.Status, promo.StartDate)

	// 3️⃣ 激活促销（状态转为 ACTIVE）
	if err := client.UpdatePromotionStatus(ctx, promo.PromotionID, walmart.PromoStatusActive); err != nil {
		log.Printf("⚠️ 激活促销失败: %v", err)
	} else {
		fmt.Println("🔥 促销已激活，前台将开始展示折扣")
	}

	// 4️⃣ 查询详情与状态
	detail, err := client.GetPromotion(ctx, promo.PromotionID)
	if err != nil {
		log.Printf("⚠️ 查询详情失败: %v", err)
	} else {
		fmt.Printf("📊 当前状态: %s | 优先级: %d | 错误信息: %s\n", detail.Status, detail.Priority, detail.ErrorDetails)
	}

	// 5️⃣ 暂停促销（如库存不足或策略调整）
	if err := client.UpdatePromotionStatus(ctx, promo.PromotionID, walmart.PromoStatusSuspended); err != nil {
		log.Printf("⚠️ 暂停促销失败: %v", err)
	} else {
		fmt.Println("⏸️ 促销已暂停")
	}
}
```

---

### 📌 Walmart Promotion API 核心注意事项

| 模块 | 官方规范 & 避坑指南 |
|------|---------------------|
| **规则引擎复杂性** | `discount` 与 `eligibility` 为嵌套 JSON，结构随类型变化。**必须**参考官方 Schema 生成。错误结构会直接返回 `400 INVALID_PROMOTION_RULE` 且不带详细提示。 |
| **时间窗口校验** | `startDate` 必须 `≥ 当前时间 - 24h`（允许极短回溯），`endDate` 必须 `> startDate`。跨时区促销需统一使用 **UTC**，系统自动按买家当地时区展示。 |
| **优先级与冲突** | `priority` 范围 `1-100`。若多个促销适用同一商品，**数值越小越优先**。 Walmart 不允许完全重叠的相同规则，提交会报 `409 CONFLICTING_PROMOTION`。 |
| **异步校验机制** | `Create/Update` 仅提交数据，后台需 `1-5 分钟` 进行规则校验、库存关联、价格冲突扫描。期间状态为 `DRAFT`，完成后转为 `ACTIVE` 或 `ERROR`。 |
| **状态流转** | `DRAFT` → `ACTIVE` → `EXPIRED/COMPLETED`。支持 `SUSPENDED` 临时暂停。已 `EXPIRED` 的促销不可修改，需新建。 |
| **前台生效延迟** | 促销激活后，搜索页、购物车、结算页通常有 **5-15 分钟缓存延迟**。大促期间可能延长。不可用 `GetPromotion` 状态替代前台实际展示校验。 |

---

### 🔧 生产环境扩展建议

1. **规则模板库**：将常用 `discount`/`eligibility` 结构封装为 Go 结构体或 JSON 模板，通过 `text/template` 动态渲染，避免手写 `map` 导致字段遗漏。
2. **冲突预检中间件**：在 `CreatePromotion` 前，调用 `ListPromotions(status=ACTIVE)` 拉取当前生效促销，本地比对 `eligibility` 与时间窗口，提前拦截 `409` 冲突。
3. **自动降级与告警**：若 `GetPromotion` 返回 `ERROR` 状态，解析 `errorDetails` 自动提取违规字段（如 `discount.value > 50` 触发风控），发送企飞/钉钉告警至运营。
4. **定时任务集成**：结合 `cron` 或消息队列，实现：
   ```go
   // 伪代码：到期前 2 小时自动检查库存，不足则暂停
   if inventoryCount < threshold {
       client.UpdatePromotionStatus(ctx, id, walmart.PromoStatusSuspended)
   }
   ```
5. **审计日志埋点**：记录 `promotion_id`, `action`, `old_status`, `new_status`, `trigger_by`（系统/人工/库存）。Walmart 对促销合规审查严格，完整日志可加速客诉处理。

将此文件加入包后，即可通过 `client.CreatePromotion()`, `client.UpdatePromotionStatus()`, `client.ListPromotions()` 完整覆盖 Walmart 促销活动全生命周期。结合已封装的 `Prices`、`Inventory`、`Items`、`Orders` 模块，你已具备构建 **企业级 Walmart Marketplace 智能营销与运营自动化平台** 的全部基础设施。


# 请封装Recommendations相关的API

