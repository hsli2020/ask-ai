# 给出一个golang语言的例子，如何调用walmart的api获取订单信息

以下是一个使用 Go 语言调用 Walmart API 获取订单信息的基本示例。Walmart Marketplace API 使用基于 Consumer ID 和 Private Key 的签名认证机制（通常遵循 OAuth 1.0a 或自定义签名方式），因此你需要生成带有签名的请求头。

⚠️ 注意：你需要从 Walmart 开发者门户获取以下凭证：

- **Consumer ID**
- **Private Key**（用于签名）
- **Client ID / Client Secret**（部分新版本 API 可能使用 OAuth 2）
- 请求通常需要按时间戳、URL、方法等生成签名

Walmart 的 API 文档建议使用 **Walmart-specific HMAC-SHA256 签名**。

---

### 示例：使用 Go 调用 Walmart 订单 API（`/v3/orders`）

```go
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	consumerID = "YOUR_CONSUMER_ID"
	privateKey = "YOUR_PRIVATE_KEY" // 这是你的私钥字符串
	baseURL    = "https://marketplace.walmartapis.com/v3/orders"
)

// 生成 Walmart API 请求签名
func generateWalmartSignature(urlStr, method, consumerID, privateKey string, timestamp int64) string {
	// Walmart 要求签名内容为：consumerId + url + method + timestamp
	signatureData := consumerID + urlStr + method + strconv.FormatInt(timestamp, 10)
	
	mac := hmac.New(sha256.New, []byte(private privateKey))
	mac.Write([]byte(signatureData))
	return hex.EncodeToString(mac.Sum(nil))
}

func getOrders() error {
	// 构造请求参数（可选）
	params := url.Values{}
	params.Add("createdStartDate", "2025-12-01T00:00:00Z")
	params.Add("createdEndDate", "2025-12-17T23:59:59Z")
	params.Add("limit", "20")

	// 构造完整 URL（不包含 query 参数用于签名）
	u, _ := url.Parse(baseURL)
	queryURL := u.String() + "?" + params.Encode()

	// 获取当前时间戳（毫秒）
	timestamp := time.Now().UnixMilli()

	// 生成签名（注意：Walmart 要求签名时使用不带 query 的原始路径）
	// 根据官方文档，签名用的 URL 应该是 "https://marketplace.walmartapis.com/v3/orders"
	signature := generateWalmartSignature(u.String(), "GET", consumerID, privateKey, timestamp)

	// 创建请求
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return err
	}

	// 设置必要的头
	req.Header.Set("WM_CONSUMER.ID", consumerID)
	req.Header.Set("WM_CONSUMER.INTIMESTAMP", strconv.FormatInt(timestamp, 10))
	req.Header.Set("WM_SEC.AUTH_SIGNATURE", signature)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Response Body:\n%s\n", string(body))
	return nil
}

func main() {
	err := getOrders()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
```

---

### 重要说明：

1. **签名细节**：
   - Walmart 要求使用 `WM_SEC.AUTH_SIGNATURE`，其生成方式为：
     ```text
     HMAC-SHA256(privateKey, consumerId + url + httpMethod + timestamp)
     ```
   - `url` 必须是 **不带 query 参数** 的完整路径，例如 `https://marketplace.walmartapis.com/v3/orders`
   - `timestamp` 是 **毫秒级 Unix 时间戳**（13 位数字）

2. **时间同步**：
   - 服务器时间必须与 Walmart 服务器时间偏差在 ±15 分钟内，否则签名会失败。

3. **API 版本**：
   - 本文使用 v3 订单 API（`/v3/orders`），请确认你有访问权限。

4. **错误处理**：
   - 如果返回 `403 Forbidden`，通常是签名、时间戳或 Consumer ID 错误。
   - 建议使用 Postman 或 curl 先测试，再集成到 Go 代码中。

5. **官方文档**：
   - 参考：[Walmart Developer Portal - Orders API](https://developer.walmart.com/api/us/mp/orders)

---

### 安全建议：

- 不要将 `privateKey` 硬编码在代码中，应使用环境变量或配置文件（如 `.env`）。
- 考虑使用 `godotenv` 加载环境变量。

```go
// 示例：从环境变量读取
consumerID := os.Getenv("WALMART_CONSUMER_ID")
privateKey := os.Getenv("WALMART_PRIVATE_KEY")
```

---

如果你使用的是 Walmart 的 **OAuth 2.0**（部分新 API），则需要先获取 `access_token`，再调用 API。但截至 2025 年，订单 API 仍主要使用上述签名方式。

如需 OAuth 2.0 示例，也可以告诉我，我可以补充。

# 请给出用OAuth 2.0 方式调用的示例

截至 2025 年，Walmart Marketplace API 的部分新接口（如商品、库存、价格等）已逐步迁移到 **OAuth 2.0（Client Credentials 流程）**，而订单 API（`/v3/orders`）在某些区域或新账户中也可能支持 OAuth 2.0。以下是一个使用 **OAuth 2.0 Client Credentials** 获取访问令牌（access token），然后调用 Walmart 订单 API 的完整 Go 示例。

> ✅ 注意：是否支持 OAuth 2.0 取决于你的 Walmart 开发者账户配置和 API 版本。请先在 [Walmart Developer Portal](https://developer.walmart.com/) 确认你是否启用了 OAuth 2.0，并获取 `Client ID` 和 `Client Secret`。

---

### ✅ 步骤概览：

1. 使用 `Client ID + Client Secret` 向 Walmart 的 OAuth 2.0 Token Endpoint 发起请求，获取 `access_token`。
2. 将 `access_token` 放在请求头 `Authorization: Bearer <token>` 中，调用订单 API。

---

### 🔐 Go 示例：使用 OAuth 2.0 调用 Walmart 订单 API

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// 配置（建议从环境变量读取）
const (
	clientID     = "YOUR_WALMART_CLIENT_ID"
	clientSecret = "YOUR_WALMART_CLIENT_SECRET"

	tokenURL = "https://api.care.walmart.com/v1/token"
	ordersURL = "https://marketplace.walmartapis.com/v3/orders"
)

// TokenResponse 定义 OAuth 2.0 响应结构
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

// OrdersResponse 是订单 API 的简化响应结构（实际结构更复杂）
type OrdersResponse struct {
	Orders []struct {
		PurchaseOrderID string `json:"purchaseOrderId"`
		CustomerOrderID string `json:"customerOrderId"`
		OrderDate       string `json:"orderDate"`
	} `json:"list"`
	TotalResults int `json:"totalResults"`
}

func getAccessToken() (string, error) {
	// 构造请求体
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	// 创建请求
	req, err := http.NewRequest("POST", tokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}

	// 设置 Basic Auth：Client ID + Client Secret
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get token: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

func getOrders(accessToken string) error {
	// 可选查询参数
	params := url.Values{}
	params.Add("createdStartDate", "2025-12-01T00:00:00Z")
	params.Add("createdEndDate", "2025-12-17T23:59:59Z")
	params.Add("limit", "10")

	fullURL := ordersURL + "?" + params.Encode()

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return err
	}

	// 使用 Bearer Token
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("WM_QOS.CORRELATION_ID", "12345") // 可选：用于跟踪请求

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("orders API failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var orders OrdersResponse
	if err := json.Unmarshal(body, &orders); err != nil {
		fmt.Printf("Failed to parse JSON, raw response:\n%s\n", string(body))
		return err
	}

	fmt.Printf("Fetched %d orders:\n", orders.TotalResults)
	for _, order := range orders.Orders {
		fmt.Printf("- PO: %s, Customer ID: %s, Date: %s\n",
			order.PurchaseOrderID, order.CustomerOrderID, order.OrderDate)
	}

	return nil
}

func main() {
	// 从环境变量读取更安全（示例中直接写明，实际请改用 os.Getenv）
	// clientID = os.Getenv("WALMART_CLIENT_ID")
	// clientSecret = os.Getenv("WALMART_CLIENT_SECRET")

	accessToken, err := getAccessToken()
	if err != nil {
		fmt.Printf("❌ Failed to get access token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Access token obtained successfully.")

	err = getOrders(accessToken)
	if err != nil {
		fmt.Printf("❌ Failed to fetch orders: %v\n", err)
		os.Exit(1)
	}
}
```

---

### 🔑 环境变量建议（生产环境）

```bash
export WALMART_CLIENT_ID="your_client_id"
export WALMART_CLIENT_SECRET="your_client_secret"
```

并在代码中使用：

```go
clientID := os.Getenv("WALMART_CLIENT_ID")
clientSecret := os.Getenv("WALMART_CLIENT_SECRET")
```

---

### 📌 注意事项

1. **Token 有效期**：
   - Walmart 的 OAuth 2.0 token 通常有效期为 **60 分钟**（`expires_in: 3600`）。
   - 生产环境中建议缓存 token 并在过期前刷新（或每次请求前检查是否过期）。

2. **Endpoint 差异**：
   - Token Endpoint 可能因地区不同而变化，例如：
     - 美国：`https://api.care.walmart.com/v1/token`
     - 加拿大：可能不同，请查阅官方文档。
   - 订单 API Endpoint 通常是 `https://marketplace.walmartapis.com/v3/orders`

3. **权限范围（Scope）**：
   - Walmart 的 OAuth 2.0 通常不需要显式指定 scope，权限由你在开发者门户中启用的 API 权限控制。

4. **调试技巧**：
   - 使用 `WM_QOS.CORRELATION_ID` 头可以帮助 Walmart 支持团队追踪你的请求。
   - 开启详细日志记录请求/响应头（尤其排查 401/403 错误）。

---

### 🔗 官方参考

- Walmart OAuth 2.0 文档（需登录）:  
  https://developer.walmart.com/doc/us/mp/us-mp-authentication/

---

如果你不确定自己是否支持 OAuth 2.0，可以先尝试这个方式；如果返回 `401` 或 `unsupported_grant_type`，则可能仍需使用传统 HMAC 签名方式。

需要我帮你封装一个自动管理 token 缓存的客户端吗？

# 封装一个自动管理 token 缓存的客户端

当然可以！下面是一个 **封装了自动管理 OAuth 2.0 Token 缓存** 的 Go 客户端，专为 Walmart API 设计。它会：

- 首次调用时获取 `access_token`
- 缓存 token 及其过期时间
- 后续请求自动复用有效 token
- token 过期时自动刷新（线程安全）

---

### ✅ 特性

- 使用 `sync.RWMutex` 保证并发安全
- 支持自动 token 刷新（基于 `expires_in`）
- 抽象出 `GetOrders` 方法，易于扩展其他 API
- 使用 `http.Client` 复用连接
- 错误处理清晰

---

### 📦 完整代码：`walmart_client.go`

```go
// Package walmart provides a client for Walmart Marketplace APIs using OAuth 2.0
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	defaultTokenURL = "https://api.care.walmart.com/v1/token"
	defaultOrdersURL = "https://marketplace.walmartapis.com/v3/orders"
)

// Config holds client credentials and endpoints
type Config struct {
	ClientID     string
	ClientSecret string
	TokenURL     string
	OrdersURL    string
	Timeout      time.Duration
}

// Token represents an OAuth 2.0 access token response
type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

// WalmartClient is a thread-safe client for Walmart APIs
type WalmartClient struct {
	config Config

	httpClient *http.Client

	mu        sync.RWMutex
	token     string
	expiresAt time.Time
}

// NewWalmartClient creates a new Walmart API client
func NewWalmartClient(cfg Config) *WalmartClient {
	if cfg.TokenURL == "" {
		cfg.TokenURL = defaultTokenURL
	}
	if cfg.OrdersURL == "" {
		cfg.OrdersURL = defaultOrdersURL
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &WalmartClient{
		config:     cfg,
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}
}

// fetchNewToken requests a new access token from Walmart
func (c *WalmartClient) fetchNewToken(ctx context.Context) (string, time.Time, error) {
	data := url.Values{"grant_type": {"client_credentials"}}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.TokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("create token request: %w", err)
	}

	req.SetBasicAuth(c.config.ClientID, c.config.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("send token request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, fmt.Errorf("token request failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var tokenResp Token
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", time.Time{}, fmt.Errorf("parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", time.Time{}, fmt.Errorf("empty access_token in response")
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return tokenResp.AccessToken, expiresAt, nil
}

// getValidToken returns a valid (not expired) access token
func (c *WalmartClient) getValidToken(ctx context.Context) (string, error) {
	c.mu.RLock()
	token := c.token
	expiresAt := c.expiresAt
	c.mu.RUnlock()

	// Add 10 seconds buffer to avoid edge expiration
	if token != "" && time.Now().Add(10*time.Second).Before(expiresAt) {
		return token, nil
	}

	// Token missing or expired: acquire write lock and refresh
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock (in case another goroutine refreshed it)
	if c.token != "" && time.Now().Add(10*time.Second).Before(c.expiresAt) {
		return c.token, nil
	}

	newToken, newExpiresAt, err := c.fetchNewToken(ctx)
	if err != nil {
		return "", err
	}

	c.token = newToken
	c.expiresAt = newExpiresAt
	return newToken, nil
}

// GetOrders calls the Walmart /v3/orders API
func (c *WalmartClient) GetOrders(ctx context.Context, params url.Values) ([]byte, error) {
	token, err := c.getValidToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}

	// Build full URL
	u, err := url.Parse(c.config.OrdersURL)
	if err != nil {
		return nil, err
	}
	if params != nil {
		u.RawQuery = params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create orders request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send orders request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("orders API error: %d %s", resp.StatusCode, string(body))
	}

	return body, nil
}
```

---

### 🧪 示例：主程序使用客户端

```go
// main.go
package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"
)

func main() {
	cfg := Config{
		ClientID:     os.Getenv("WALMART_CLIENT_ID"),
		ClientSecret: os.Getenv("WALMART_CLIENT_SECRET"),
		// TokenURL 和 OrdersURL 可选，留空使用默认值
		Timeout: 20 * time.Second,
	}

	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		panic("Please set WALMART_CLIENT_ID and WALMART_CLIENT_SECRET environment variables")
	}

	client := NewWalmartClient(cfg)

	// 设置查询参数
	params := url.Values{}
	params.Add("createdStartDate", "2025-12-01T00:00:00Z")
	params.Add("createdEndDate", "2025-12-17T23:59:59Z")
	params.Add("limit", "5")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	body, err := client.GetOrders(ctx, params)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Success! Orders response:\n%s\n", string(body))
}
```

---

### 🧼 环境变量设置（bash）

```bash
export WALMART_CLIENT_ID="your_client_id_here"
export WALMART_CLIENT_SECRET="your_client_secret_here"
go run .
```

---

### 🔄 扩展建议

你可以轻松为其他 Walmart API 添加方法，例如：

```go
func (c *WalmartClient) GetItems(ctx context.Context, ...) ([]byte, error) { ... }
func (c *WalmartClient) UpdateInventory(ctx context.Context, ...) ([]byte, error) { ... }
```

所有方法都复用同一个 token 缓存机制。

---

### ✅ 优势总结

- **线程安全**：多个 goroutine 可安全并发调用 `GetOrders`
- **自动刷新**：无需手动管理 token 生命周期
- **可扩展**：易于支持更多 Walmart API
- **超时控制**：通过 `context` 和 `http.Client.Timeout` 防止阻塞

---

如需支持 **自动重试**、**日志记录** 或 **结构化解析订单数据**，也可以进一步扩展。需要我帮你加上这些功能吗？

# 加上“自动重试、日志记录、结构化解析订单数据”功能

当然可以！下面是在你已有 **自动 Token 管理客户端** 基础上，进一步增强的完整实现，包含以下三大增强功能：

---

### ✅ 新增功能

1. **自动重试机制**  
   - 对网络错误或临时性 5xx 错误自动重试（可配置次数和退避策略）
2. **结构化日志记录**  
   - 使用 `log/slog`（Go 1.21+ 标准库）记录关键事件（token 获取、API 调用、重试等）
3. **结构化解析订单数据**  
   - 将 `/v3/orders` 响应反序列化为强类型 Go 结构体，便于业务逻辑处理

---

### 📦 完整增强版客户端（`walmart_client.go`）

```go
// Package walmart provides a robust, thread-safe client for Walmart Marketplace APIs.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	defaultTokenURL  = "https://api.care.walmart.com/v1/token"
	defaultOrdersURL = "https://marketplace.walmartapis.com/v3/orders"
)

// Config holds client configuration
type Config struct {
	ClientID     string
	ClientSecret string
	TokenURL     string
	OrdersURL    string
	Timeout      time.Duration
	RetryCount   int           // 默认重试 2 次（总共 3 次尝试）
	RetryDelay   time.Duration // 初始重试延迟（指数退避）
	Logger       *slog.Logger
}

// Token represents OAuth 2.0 token response
type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

// OrderItem represents a single order item
type OrderItem struct {
	ItemID            string `json:"itemID"`
	SellerSKU         string `json:"sellerSKU"`
	ProductName       string `json:"productName"`
	Quantity          int    `json:"quantity"`
	SellingPrice      struct {
		Amount   float64 `json:"amount"`
		Currency string  `json:"currency"`
	} `json:"sellingPrice"`
}

// Order represents a Walmart order
type Order struct {
	PurchaseOrderID string       `json:"purchaseOrderId"`
	CustomerOrderID string       `json:"customerOrderId"`
	OrderDate       string       `json:"orderDate"`
	OrderLines      []OrderLine  `json:"orderLines"`
	Status          string       `json:"orderStatus"`
}

type OrderLine struct {
	LineStatus string      `json:"lineStatus"`
	Items      []OrderItem `json:"item"`
}

// OrdersResponse is the top-level response for /v3/orders
type OrdersResponse struct {
	Orders       []Order `json:"list"`
	TotalResults int     `json:"totalResults"`
	Offset       int     `json:"offset,omitempty"`
	Limit        int     `json:"limit,omitempty"`
}

// WalmartClient is a robust, thread-safe client
type WalmartClient struct {
	config Config

	httpClient *http.Client
	limiter    *rate.Limiter // 可选：用于遵守 API 速率限制（Walmart 通常为 10 QPS）

	mu        sync.RWMutex
	token     string
	expiresAt time.Time

	logger *slog.Logger
}

// NewWalmartClient creates a new client with defaults and validation
func NewWalmartClient(cfg Config) *WalmartClient {
	if cfg.TokenURL == "" {
		cfg.TokenURL = defaultTokenURL
	}
	if cfg.OrdersURL == "" {
		cfg.OrdersURL = defaultOrdersURL
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.RetryCount == 0 {
		cfg.RetryCount = 2 // 总共尝试 3 次
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = 1 * time.Second
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	return &WalmartClient{
		config:     cfg,
		httpClient: &http.Client{Timeout: cfg.Timeout},
		// Walmart 限制约 10 QPS，这里设置 8 QPS 安全余量
		limiter: rate.NewLimiter(rate.Limit(8), 10),
		logger:  cfg.Logger,
	}
}

// fetchNewToken requests a new access token
func (c *WalmartClient) fetchNewToken(ctx context.Context) (string, time.Time, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return "", time.Time{}, err
	}

	data := url.Values{"grant_type": {"client_credentials"}}
	req, err := http.NewRequestWithContext(ctx, "POST", c.config.TokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("create token request: %w", err)
	}

	req.SetBasicAuth(c.config.ClientID, c.config.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("send token request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, fmt.Errorf("token request failed: status=%d", resp.StatusCode)
	}

	var tokenResp Token
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", time.Time{}, fmt.Errorf("parse token response: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return tokenResp.AccessToken, expiresAt, nil
}

// getValidToken returns a valid access token, refreshing if needed
func (c *WalmartClient) getValidToken(ctx context.Context) (string, error) {
	c.mu.RLock()
	token := c.token
	expiresAt := c.expiresAt
	c.mu.RUnlock()

	if token != "" && time.Now().Add(10*time.Second).Before(expiresAt) {
		return token, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Now().Add(10*time.Second).Before(c.expiresAt) {
		return c.token, nil
	}

	c.logger.Info("Refreshing Walmart access token")
	newToken, newExpiresAt, err := c.fetchNewToken(ctx)
	if err != nil {
		c.logger.Error("Failed to refresh token", "error", err)
		return "", err
	}

	c.token = newToken
	c.expiresAt = newExpiresAt
	c.logger.Info("Token refreshed successfully", "expires_in", time.Until(newExpiresAt))
	return newToken, nil
}

// retryDo executes an HTTP request with retry logic
func (c *WalmartClient) retryDo(ctx context.Context, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= c.config.RetryCount; attempt++ {
		if attempt > 0 {
			// 指数退避：delay = base * 2^(attempt-1)
			delay := c.config.RetryDelay * time.Duration(1<<uint(attempt-1))
			c.logger.Warn("Retrying request after failure", "attempt", attempt, "delay", delay)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		if err := c.limiter.Wait(ctx); err != nil {
			return nil, err
		}

		resp, err = c.httpClient.Do(req.Clone(ctx))
		if err != nil {
			c.logger.Warn("Request failed (network error)", "error", err, "attempt", attempt)
			continue // retry on network errors
		}

		// 不重试 4xx 错误（如 401, 403, 400）
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			resp.Body.Close()
			return resp, nil
		}

		// 重试 5xx 临时错误
		if resp.StatusCode >= 500 {
			c.logger.Warn("Received 5xx error, will retry", "status", resp.StatusCode, "attempt", attempt)
			resp.Body.Close()
			continue
		}

		// 2xx or 3xx: success
		return resp, nil
	}

	return resp, fmt.Errorf("request failed after %d retries: %w", c.config.RetryCount+1, err)
}

// GetOrders fetches and parses orders from Walmart API
func (c *WalmartClient) GetOrders(ctx context.Context, params url.Values) (*OrdersResponse, error) {
	token, err := c.getValidToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}

	u, err := url.Parse(c.config.OrdersURL)
	if err != nil {
		return nil, err
	}
	if params != nil {
		u.RawQuery = params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create orders request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("WM_QOS.CORRELATION_ID", "go-walmart-client-"+time.Now().Format("20060102150405"))

	c.logger.Info("Calling Walmart Orders API", "url", u.String())

	resp, err := c.retryDo(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("execute orders request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Orders API returned non-200 status", "status", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("orders API error: %d", resp.StatusCode)
	}

	var orders OrdersResponse
	if err := json.Unmarshal(body, &orders); err != nil {
		c.logger.Error("Failed to parse orders response", "error", err, "raw_body", string(body))
		return nil, fmt.Errorf("parse orders response: %w", err)
	}

	c.logger.Info("Successfully fetched orders", "total", orders.TotalResults, "returned", len(orders.Orders))
	return &orders, nil
}
```

---

### 🧪 示例：主程序使用增强客户端

```go
// main.go
package main

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	"time"
)

func main() {
	// 设置结构化日志（可输出 JSON）
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := Config{
		ClientID:     os.Getenv("WALMART_CLIENT_ID"),
		ClientSecret: os.Getenv("WALMART_CLIENT_SECRET"),
		RetryCount:   2,
		RetryDelay:   2 * time.Second,
		Logger:       logger,
	}

	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		slog.Error("Missing WALMART_CLIENT_ID or WALMART_CLIENT_SECRET")
		os.Exit(1)
	}

	client := NewWalmartClient(cfg)

	params := url.Values{}
	params.Add("createdStartDate", "2025-12-01T00:00:00Z")
	params.Add("createdEndDate", "2025-12-17T00:00:00Z")
	params.Add("limit", "3")

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	orders, err := client.GetOrders(ctx, params)
	if err != nil {
		logger.Error("Failed to get orders", "error", err)
		os.Exit(1)
	}

	logger.Info("Processing orders...")
	for _, order := range orders.Orders {
		logger.Info("Order",
			"purchase_order_id", order.PurchaseOrderID,
			"customer_order_id", order.CustomerOrderID,
			"status", order.Status,
			"items_count", len(order.OrderLines),
		)
	}
}
```

---

### 📝 日志输出示例（JSON 格式）

```json
{"time":"2025-12-17T10:00:00Z","level":"INFO","msg":"Calling Walmart Orders API","url":"https://marketplace.walmartapis.com/v3/orders?createdEndDate=2025-12-17T00%3A00%3A00Z\u0026createdStartDate=2025-12-01T00%3A00%3A00Z\u0026limit=3"}
{"time":"2025-12-17T10:00:01Z","level":"INFO","msg":"Successfully fetched orders","total":42,"returned":3}
{"time":"2025-12-17T10:00:01Z","level":"INFO","msg":"Order","purchase_order_id":"123456789","customer_order_id":"CUST98765","status":"Created","items_count":2}
```

---

### ✅ 优势总结

| 功能 | 说明 |
|------|------|
| **自动重试** | 对 5xx 和网络错误自动重试，支持指数退避 |
| **结构化日志** | 使用 `slog` 输出带字段的日志，便于监控和排查 |
| **类型安全** | 定义 `Order`, `OrderItem` 等结构体，避免 `map[string]interface{}` |
| **速率限制** | 内置 `rate.Limiter` 防止触发 Walmart API 限流（10 QPS） |
| **上下文取消** | 支持超时和取消 |
| **相关 ID** | 自动添加 `WM_QOS.CORRELATION_ID` 便于 Walmart 支持追踪 |

---

> 💡 **提示**：Walmart 官方建议 QPS ≤ 10。本客户端默认限制为 8 QPS，可根据需要调整 `rate.Limiter`。

---

如需支持 **分页自动遍历所有订单**、**导出 CSV**、或 **对接数据库**，也可以继续扩展！需要的话请告诉我。

# 加上支持 分页自动遍历所有订单、导出 CSV

非常棒的需求！下面我们在已有的 **自动 Token 管理 + 重试 + 日志 + 结构化解析** 基础上，进一步增强 Walmart 客户端，支持：

---

### ✅ 新增功能

1. **自动分页遍历所有订单**  
   - 基于 `limit` 和 `offset` 递增，直到获取全部订单
   - 支持最大订单数限制（防无限拉取）
2. **导出为 CSV 文件**  
   - 一行一个订单项（order line item）
   - 包含关键字段：订单号、商品 SKU、数量、价格、状态、下单时间等
   - 使用标准库 `encoding/csv`

---

### 📦 更新后的客户端（精简只展示新增方法）

> 为避免重复，以下只展示 **新增的 `FetchAllOrders` 和 `ExportOrdersToCSV` 方法**，其余代码（如 Token 管理、重试等）保持不变。

```go
// Add these methods to the existing WalmartClient struct

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"
)

// FetchAllOrders fetches all orders matching the params by automatically paginating.
// It respects maxOrders to prevent excessive API calls.
func (c *WalmartClient) FetchAllOrders(ctx context.Context, params url.Values, maxOrders int) ([]Order, error) {
	if params == nil {
		params = url.Values{}
	}

	limit := 200 // Walmart 最大 limit 为 200
	if l := params.Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	// Override limit to ensure we control pagination
	params.Set("limit", strconv.Itoa(limit))

	var allOrders []Order
	offset := 0

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		params.Set("offset", strconv.Itoa(offset))

		c.logger.Info("Fetching order page", "offset", offset, "limit", limit)

		resp, err := c.GetOrders(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("fetch page at offset %d: %w", offset, err)
		}

		if len(resp.Orders) == 0 {
			break
		}

		allOrders = append(allOrders, resp.Orders...)

		c.logger.Info("Accumulated orders", "total_so_far", len(allOrders))

		// Stop if we've reached maxOrders or fetched fewer than limit (last page)
		if len(allOrders) >= maxOrders || len(resp.Orders) < limit {
			break
		}

		offset += limit

		// Optional: add small delay between pages to be respectful
		time.Sleep(200 * time.Millisecond)
	}

	if maxOrders > 0 && len(allOrders) > maxOrders {
		allOrders = allOrders[:maxOrders]
	}

	c.logger.Info("Completed fetching all orders", "total", len(allOrders))
	return allOrders, nil
}

// ExportOrdersToCSV exports a slice of orders to a CSV file.
// Each row represents one order line item.
func (c *WalmartClient) ExportOrdersToCSV(orders []Order, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"PurchaseOrderID",
		"CustomerOrderID",
		"OrderDate",
		"OrderStatus",
		"ItemID",
		"SellerSKU",
		"ProductName",
		"Quantity",
		"UnitPrice",
		"Currency",
		"LineStatus",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("write CSV header: %w", err)
	}

	// Write rows
	for _, order := range orders {
		for _, line := range order.OrderLines {
			for _, item := range line.Items {
				row := []string{
					order.PurchaseOrderID,
					order.CustomerOrderID,
					order.OrderDate,
					order.Status,
					item.ItemID,
					item.SellerSKU,
					sanitizeCSVField(item.ProductName),
					strconv.Itoa(item.Quantity),
					strconv.FormatFloat(item.SellingPrice.Amount, 'f', 2, 64),
					item.SellingPrice.Currency,
					line.LineStatus,
				}
				if err := writer.Write(row); err != nil {
					return fmt.Errorf("write CSV row: %w", err)
				}
			}
		}
	}

	c.logger.Info("Orders exported to CSV", "file", filename, "rows", len(orders))
	return nil
}

// sanitizeCSVField removes or escapes problematic characters (like newlines, quotes)
func sanitizeCSVField(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, `"`, "'")
	return strings.TrimSpace(s)
}
```

---

### 🧪 主程序示例：拉取全部订单并导出 CSV

```go
// main.go
package main

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := Config{
		ClientID:     os.Getenv("WALMART_CLIENT_ID"),
		ClientSecret: os.Getenv("WALMART_CLIENT_SECRET"),
		RetryCount:   2,
		RetryDelay:   2 * time.Second,
		Logger:       logger,
	}

	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		logger.Error("Missing WALMART_CLIENT_ID or WALMART_CLIENT_SECRET")
		os.Exit(1)
	}

	client := NewWalmartClient(cfg)

	// 设置时间范围（注意：Walmart 限制最大时间窗口，通常 ≤ 180 天）
	params := url.Values{}
	params.Add("createdStartDate", "2025-06-01T00:00:00Z") // 6 months ago max
	params.Add("createdEndDate", "2025-12-17T00:00:00Z")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute) // 分页可能耗时较长
	defer cancel()

	// 拉取最多 10,000 个订单（可调整）
	allOrders, err := client.FetchAllOrders(ctx, params, 10000)
	if err != nil {
		logger.Error("Failed to fetch all orders", "error", err)
		os.Exit(1)
	}

	// 导出 CSV
	csvFile := "walmart_orders_" + time.Now().Format("20060102_150405") + ".csv"
	if err := client.ExportOrdersToCSV(allOrders, csvFile); err != nil {
		logger.Error("Failed to export CSV", "error", err)
		os.Exit(1)
	}

	logger.Info("✅ All done!", "orders_count", len(allOrders), "csv_file", csvFile)
}
```

---

### 📄 生成的 CSV 示例

```csv
PurchaseOrderID,CustomerOrderID,OrderDate,OrderStatus,ItemID,SellerSKU,ProductName,Quantity,UnitPrice,Currency,LineStatus
123456789,CUST98765,2025-12-05T14:30:00Z,Created,987654321,SKU-123,"Wireless Bluetooth Headphones",2,29.99,USD,Created
123456789,CUST98765,2025-12-05T14:30:00Z,Created,112233445,SKU-456,"USB-C Charging Cable",1,9.99,USD,Created
```

> ✅ 每个订单行项（item）独立一行，便于后续分析（如销量统计、SKU 报表等）

---

### ⚠️ 注意事项

1. **Walmart 分页限制**：
   - 最大 `limit=200`
   - 时间窗口不能超过 **180 天**（否则返回错误）
   - 建议按月分段拉取（如每月一个 CSV），避免单次请求过大

2. **性能与速率**：
   - 自动分页会发出多个 API 请求
   - 客户端内置 `rate.Limiter`（8 QPS）防止被限流
   - 每页之间加了 `200ms` 延迟，可调整或移除

3. **内存使用**：
   - 所有订单会加载到内存中
   - 如需处理 **超大订单量（>10万）**，建议流式写入 CSV（边拉取边写），可进一步优化

---

### 🚀 可选优化方向（如需）

- **流式 CSV 导出**：不缓存所有订单，逐页写入 CSV
- **按月自动分段导出**：循环处理多个时间窗口
- **压缩 CSV**：导出 `.csv.gz`
- **上传到 S3/Google Drive**：自动归档

---

需要我帮你实现 **流式导出** 或 **按月分段拉取** 吗？欢迎继续提出！

