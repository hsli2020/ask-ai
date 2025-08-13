# 请用golang编写一个函数，调用coinbase 的 API下单，买入或卖出加密货币

以下是一个使用 Go 语言调用 Coinbase API 进行下单（买入或卖出加密货币）的示例函数。该函数使用 **Coinbase Advanced Trade API**（原 Pro API），支持身份验证（HMAC 签名）。

---

### ✅ 前提条件

1. **API Key**：在 [Coinbase Advanced Trade](https://www.coinbase.com/advanced-trade/settings/api) 页面创建 API Key，并保存：
   - API Key
   - Secret Key
   - Passphrase

2. **权限**：确保 API Key 有 **交易权限（Trade）**。

3. **依赖库**：
   ```bash
   go get github.com/google/uuid
   ```

---

### ✅ Go 代码实现

```go
package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

const (
	baseURL = "https://api.coinbase.com/api/v3/brokerage"
)

// OrderRequest 定义下单请求结构
type OrderRequest struct {
	ClientOrderID string  `json:"client_order_id"`
	ProductID     string  `json:"product_id"`     // 例如: "BTC-USD"
	Side          string  `json:"side"`           // "BUY" 或 "SELL"
	Type          string  `json:"order_type"`     // "MARKET", "LIMIT" 等
	Amount        string  `json:"size"`           // 数量（字符串，避免精度丢失）
	LimitPrice    *string `json:"limit_price,omitempty"` // 限价单时使用
}

// OrderResponse 定义响应结构（简化版）
type OrderResponse struct {
	OrderID       string `json:"order_id"`
	ClientOrderID string `json:"client_order_id"`
	ProductID     string `json:"product_id"`
	Side          string `json:"side"`
	Status        string `json:"status"`
}

// CoinbaseClient 封装客户端
type CoinbaseClient struct {
	APIKey    string
	SecretKey string
	Passphrase string
}

// NewCoinbaseClient 创建新客户端
func NewCoinbaseClient(apiKey, secretKey, passphrase string) *CoinbaseClient {
	return &CoinbaseClient{
		APIKey:     apiKey,
		SecretKey:  secretKey,
		Passphrase: passphrase,
	}
}

// PlaceOrder 下单（买入或卖出）
func (c *CoinbaseClient) PlaceOrder(req OrderRequest) (*OrderResponse, error) {
	// 设置 client_order_id（可选，但推荐）
	if req.ClientOrderID == "" {
		req.ClientOrderID = uuid.New().String()
	}

	// 序列化请求体
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// 构建请求
	url := baseURL + "/orders"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// 设置必要头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("CB-ACCESS-KEY", c.APIKey)
	httpReq.Header.Set("CB-ACCESS-PASSPHRASE", c.Passphrase)

	// 生成时间戳
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	httpReq.Header.Set("CB-ACCESS-TIMESTAMP", timestamp)

	// 生成签名
	message := timestamp + "POST" + "/api/v3/brokerage/orders" + string(body)
	signature := c.generateSignature(message)
	httpReq.Header.Set("CB-ACCESS-SIGN", signature)

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %d, body: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var orderResp OrderResponse
	if err := json.Unmarshal(respBody, &orderResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &orderResp, nil
}

// generateSignature 生成 HMAC-SHA256 签名
func (c *CoinbaseClient) generateSignature(message string) string {
	key, err := hex.DecodeString(c.SecretKey)
	if err != nil {
		panic("invalid secret key (must be hex string)")
	}
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// 示例用法
func main() {
	// 替换为你的实际密钥（建议从环境变量读取）
	apiKey := "your-api-key"
	secretKey := "your-secret-key" // 注意：这是十六进制字符串
	passphrase := "your-passphrase"

	client := NewCoinbaseClient(apiKey, secretKey, passphrase)

	// 示例：市价买入 0.001 BTC
	orderReq := OrderRequest{
		ProductID: "BTC-USD",
		Side:      "BUY",
		Type:      "MARKET",
		Amount:    "0.001", // 注意：某些交易对使用 size 表示数量
	}

	resp, err := client.PlaceOrder(orderReq)
	if err != nil {
		fmt.Printf("下单失败: %v\n", err)
		return
	}

	fmt.Printf("下单成功! 订单ID: %s, 状态: %s\n", resp.OrderID, resp.Status)
}
```

---

### 🔐 安全提示

- **不要硬编码密钥**：建议从环境变量或安全的密钥管理服务中读取。
  ```go
  apiKey := os.Getenv("COINBASE_API_KEY")
  secretKey := os.Getenv("COINBASE_SECRET_KEY")
  passphrase := os.Getenv("COINBASE_PASSPHRASE")
  ```
- **Secret Key 是十六进制字符串**：在 Coinbase 控制台生成的 Secret Key 是 base64 编码的，但 API 要求是 **hex 解码后的字节**。不过，Coinbase Advanced Trade API 的文档说明 Secret Key 应作为 **原始字节** 使用，而控制台显示的是 base64。因此，你可能需要先 base64 解码，再用于 HMAC。

> ⚠️ **重要修正**：如果你的 Secret Key 是 base64 格式（通常如此），请先解码：

```go
// 在 generateSignature 中替换：
keyBytes, err := base64.StdEncoding.DecodeString(c.SecretKey)
if err != nil {
    panic("invalid base64 secret key")
}
h := hmac.New(sha256.New, keyBytes)
```

并添加 `import "encoding/base64"`

---

### 📚 参考文档

- [Coinbase Advanced Trade API 文档](https://docs.cdp.coinbase.com/advanced-trade/docs/rest-api-auth)
- 支持的订单类型：`MARKET`, `LIMIT`, `STOP_LIMIT` 等
- 产品 ID 列表：如 `BTC-USD`, `ETH-USDT` 等

---

如需支持限价单、止损单等，只需在 `OrderRequest` 中设置 `LimitPrice` 字段即可。

需要我帮你添加限价单示例或错误重试逻辑吗？


# 后续除了调用PlaceOrder，可能还需要调用其它的API，请把目前PlaceOrder代码中通用的部分提取出来，方便不同API调用时重用代码

好的！为了提高代码复用性，我们可以将 **通用的 HTTP 请求逻辑、签名生成、错误处理** 等部分提取到一个通用的 `doRequest` 方法中。这样，无论是下单、查询订单、获取账户信息等 API 调用，都可以复用这套认证和请求机制。

---

### ✅ 重构后的代码（支持多 API 复用）

```go
package coinbase

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	BaseURL = "https://api.coinbase.com/api/v3/brokerage"
)

// Client 封装 Coinbase API 客户端
type Client struct {
	APIKey     string
	SecretKey  string // base64-encoded
	Passphrase string
	HTTPClient *http.Client
}

// NewClient 创建新客户端
func NewClient(apiKey, secretKey, passphrase string) *Client {
	return &Client{
		APIKey:     apiKey,
		SecretKey:  secretKey,
		Passphrase: passphrase,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// doRequest 是通用的带认证的 HTTP 请求方法
func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
	}

	url := BaseURL + path
	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}

	// 设置通用 Header
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("CB-ACCESS-KEY", c.APIKey)
	req.Header.Set("CB-ACCESS-PASSPHRASE", c.Passphrase)

	// 时间戳
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	req.Header.Set("CB-ACCESS-TIMESTAMP", timestamp)

	// 构造签名消息
	message := timestamp + method + path + string(bodyBytes)
	signature, err := c.generateSignature(message)
	if err != nil {
		return nil, fmt.Errorf("generate signature: %w", err)
	}
	req.Header.Set("CB-ACCESS-SIGN", signature)

	// 发送请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error: %d, response: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// generateSignature 生成 HMAC-SHA256 签名（支持 base64 SecretKey）
func (c *Client) generateSignature(message string) (string, error) {
	// Coinbase 返回的 Secret Key 是 base64 编码的，需先解码为原始字节
	secretBytes, err := base64.StdEncoding.DecodeString(c.SecretKey)
	if err != nil {
		return "", fmt.Errorf("decode base64 secret key: %w", err)
	}

	h := hmac.New(sha256.New, secretBytes)
	h.Write([]byte(message))
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// -----------------------------
// 具体 API 方法（可扩展）
// -----------------------------

// OrderRequest 定义下单请求
type OrderRequest struct {
	ClientOrderID string  `json:"client_order_id,omitempty"`
	ProductID     string  `json:"product_id"`
	Side          string  `json:"side"`           // "BUY" or "SELL"
	Type          string  `json:"order_type"`     // e.g., "MARKET", "LIMIT"
	Amount        string  `json:"size"`           // 数量（字符串）
	LimitPrice    *string `json:"limit_price,omitempty"`
}

// OrderResponse 定义下单响应
type OrderResponse struct {
	OrderID       string `json:"order_id"`
	ClientOrderID string `json:"client_order_id"`
	ProductID     string `json:"product_id"`
	Side          string `json:"side"`
	Status        string `json:"status"`
}

// PlaceOrder 下单（买入/卖出）
func (c *Client) PlaceOrder(req OrderRequest) (*OrderResponse, error) {
	respBody, err := c.doRequest("POST", "/orders", req)
	if err != nil {
		return nil, err
	}

	var resp OrderResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal PlaceOrder response: %w", err)
	}
	return &resp, nil
}

// GetAccounts 获取账户列表（示例：展示如何复用 doRequest）
func (c *Client) GetAccounts() ([]Account, error) {
	respBody, err := c.doRequest("GET", "/accounts", nil)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Accounts []Account `json:"accounts"`
	}
	if err := json.Unmarshal(respBody, &wrapper); err != nil {
		return nil, fmt.Errorf("unmarshal GetAccounts response: %w", err)
	}
	return wrapper.Accounts, nil
}

// Account 账户结构（简化）
type Account struct {
	UUID         string `json:"uuid"`
	Currency     string `json:"currency"`
	Available    string `json:"available_balance"`
	Hold         string `json:"hold"`
}

// GetOrder 获取单个订单详情
func (c *Client) GetOrder(orderID string) (*OrderResponse, error) {
	respBody, err := c.doRequest("GET", "/orders/"+orderID, nil)
	if err != nil {
		return nil, err
	}

	var resp OrderResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal GetOrder response: %w", err)
	}
	return &resp, nil
}
```

---

### ✅ 使用示例

```go
package main

import (
	"fmt"
	"os"
	"your-project/coinbase" // 替换为你的模块路径
)

func main() {
	apiKey := os.Getenv("COINBASE_API_KEY")
	secretKey := os.Getenv("COINBASE_SECRET_KEY")     // base64 格式
	passphrase := os.Getenv("COINBASE_PASSPHRASE")

	client := coinbase.NewClient(apiKey, secretKey, passphrase)

	// 1. 下单
	order, err := client.PlaceOrder(coinbase.OrderRequest{
		ProductID: "BTC-USD",
		Side:      "BUY",
		Type:      "MARKET",
		Amount:    "0.001",
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("下单成功: %s\n", order.OrderID)

	// 2. 查询账户
	accounts, err := client.GetAccounts()
	if err != nil {
		panic(err)
	}
	for _, acc := range accounts {
		fmt.Printf("账户: %s, 可用: %s %s\n", acc.UUID, acc.Available, acc.Currency)
	}

	// 3. 查询订单
	detail, err := client.GetOrder(order.OrderID)
	if err != nil {
		panic(err)
	}
	fmt.Printf("订单状态: %s\n", detail.Status)
}
```

---

### ✅ 优势

- ✅ **签名逻辑复用**：所有 API 共享同一套认证机制。
- ✅ **错误统一处理**：HTTP 错误、JSON 解析错误集中处理。
- ✅ **易于扩展**：新增 API 只需写一个简单方法调用 `doRequest`。
- ✅ **支持 base64 SecretKey**：符合 Coinbase 实际返回格式。
- ✅ **结构清晰**：`Client` 负责通用逻辑，具体方法专注业务。

---

### 🔒 安全建议

- 使用环境变量管理密钥：
  ```bash
  export COINBASE_API_KEY="..."
  export COINBASE_SECRET_KEY="..."  # base64 string from Coinbase
  export COINBASE_PASSPHRASE="..."
  ```

---

现在你可以轻松添加更多 API，比如取消订单、获取行情、列出产品等，只需在 `Client` 上新增方法并调用 `c.doRequest(...)` 即可。需要我帮你实现某个具体 API 吗？


# 帮我实现取消订单、获取行情、列出产品等API

当然可以！以下是基于我们之前重构的通用 `Client`，新增 **取消订单**、**获取行情（ticker）**、**列出产品（products）** 等常用 API 的完整实现。

---

### ✅ 新增 API 方法（接续之前的 `coinbase` 包）

```go
// CancelOrderRequest 取消订单请求（可选参数）
type CancelOrderRequest struct {
	OrderIDs []string `json:"order_ids"` // 最多 100 个
}

// CancelOrderResponse 取消订单响应
type CancelOrderResponse struct {
	Results []struct {
		OrderID string `json:"order_id"`
		Success bool   `json:"success"`
		Error   string `json:"error,omitempty"`
	} `json:"results"`
}

// CancelOrders 批量取消订单（最多100个）
func (c *Client) CancelOrders(orderIDs []string) (*CancelOrderResponse, error) {
	if len(orderIDs) == 0 {
		return nil, fmt.Errorf("at least one order ID is required")
	}
	if len(orderIDs) > 100 {
		return nil, fmt.Errorf("maximum 100 order IDs allowed")
	}

	req := CancelOrderRequest{OrderIDs: orderIDs}
	respBody, err := c.doRequest("POST", "/orders/batch_cancel", req)
	if err != nil {
		return nil, err
	}

	var resp CancelOrderResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal CancelOrders response: %w", err)
	}
	return &resp, nil
}

// CancelOrder 取消单个订单（便捷方法）
func (c *Client) CancelOrder(orderID string) (bool, error) {
	resp, err := c.CancelOrders([]string{orderID})
	if err != nil {
		return false, err
	}
	if len(resp.Results) == 0 {
		return false, fmt.Errorf("no result returned for order %s", orderID)
	}
	result := resp.Results[0]
	if !result.Success {
		return false, fmt.Errorf("cancel failed: %s", result.Error)
	}
	return true, nil
}

// -----------------------------
// 获取行情（Ticker）
// -----------------------------

// Ticker 行情数据
type Ticker struct {
	ProductID string `json:"product_id"`
	Price     string `json:"price"`
	Bid       string `json:"bid"`
	Ask       string `json:"ask"`
	Volume    string `json:"volume_24h"`
	Time      string `json:"time"` // ISO 8601
}

// GetTicker 获取指定交易对的最新行情
func (c *Client) GetTicker(productID string) (*Ticker, error) {
	respBody, err := c.doRequest("GET", "/best_bid_ask?product_ids="+productID, nil)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		PriceBooks []struct {
			ProductID string `json:"product_id"`
			Bid       []struct {
				Price string `json:"price"`
			} `json:"bids"`
			Ask []struct {
				Price string `json:"price"`
			} `json:"asks"`
		} `json:"pricebooks"`
	}

	if err := json.Unmarshal(respBody, &wrapper); err != nil {
		return nil, fmt.Errorf("unmarshal GetTicker response: %w", err)
	}

	if len(wrapper.PriceBooks) == 0 {
		return nil, fmt.Errorf("no pricebook found for %s", productID)
	}

	pb := wrapper.PriceBooks[0]
	ticker := &Ticker{
		ProductID: pb.ProductID,
		Bid:       "",
		Ask:       "",
	}

	if len(pb.Bids) > 0 {
		ticker.Bid = pb.Bids[0].Price
	}
	if len(pb.Asks) > 0 {
		ticker.Ask = pb.Asks[0].Price
	}
	// 注意：Coinbase Advanced Trade 的 /best_bid_ask 不直接返回最新成交价
	// 如果你需要最新成交价，请使用 /ticker（但该端点可能不开放）
	// 或通过 WebSocket 获取。此处我们用 bid/ask 近似

	// 若需最新成交价，可考虑调用 /products/{product_id}/ticker（如果可用）
	// 但根据当前文档，建议使用 WebSocket。此处先返回 bid/ask

	return ticker, nil
}

// -----------------------------
// 列出所有可交易产品（Products）
// -----------------------------

// Product 产品信息
type Product struct {
	ProductID        string `json:"product_id"`
	BaseCurrency     string `json:"base_currency"`
	QuoteCurrency    string `json:"quote_currency"`
	BaseMinSize      string `json:"base_min_size"`
	BaseMaxSize      string `json:"base_max_size"`
	QuoteIncrement   string `json:"quote_increment"`
	Status           string `json:"status"` // "online", "offline", etc.
	StatusMessage    string `json:"status_message,omitempty"`
	PostOnly         bool   `json:"post_only"`
	LimitOnly        bool   `json:"limit_only"`
	CancelOnly       bool   `json:"cancel_only"`
	TradingDisabled  bool   `json:"trading_disabled"`
}

// ListProducts 获取所有可交易产品列表
func (c *Client) ListProducts() ([]Product, error) {
	respBody, err := c.doRequest("GET", "/products", nil)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Products []Product `json:"products"`
	}

	if err := json.Unmarshal(respBody, &wrapper); err != nil {
		return nil, fmt.Errorf("unmarshal ListProducts response: %w", err)
	}

	return wrapper.Products, nil
}

// GetProduct 获取单个产品详情
func (c *Client) GetProduct(productID string) (*Product, error) {
	respBody, err := c.doRequest("GET", "/products/"+productID, nil)
	if err != nil {
		return nil, err
	}

	var product Product
	if err := json.Unmarshal(respBody, &product); err != nil {
		return nil, fmt.Errorf("unmarshal GetProduct response: %w", err)
	}
	return &product, nil
}
```

---

### 📌 补充说明

#### 1. **行情 API 的注意事项**
- Coinbase Advanced Trade REST API **不直接提供 `/ticker` 端点**（如旧版 Pro API 那样）。
- 推荐使用 `/best_bid_ask` 获取最优买卖价（如上实现）。
- 若需**最新成交价（last price）**，官方建议使用 **WebSocket feed**。
- 如果你确实需要成交价，可考虑调用非认证的公开端点（如有）：
  ```go
  // 示例：使用公开 API（无需认证）
  func (c *Client) GetPublicTicker(productID string) (string, error) {
      resp, err := http.Get("https://api.coinbase.com/api/v3/brokerage/products/" + productID + "/ticker")
      // 注意：该端点可能不存在或需不同路径
  }
  ```
  但根据 [官方文档](https://docs.cdp.coinbase.com/advanced-trade/docs/rest-api-overview)，目前 REST API 未提供公开 ticker。因此，**最优 bid/ask 是最接近的替代方案**。

#### 2. **取消订单**
- 使用 `/orders/batch_cancel` 支持批量取消（最多 100 个）。
- 单个取消通过封装提供便捷方法。

#### 3. **产品列表**
- `/products` 返回所有可交易对及其规则（最小下单量、价格精度等）。
- 对交易前做参数校验非常有用。

---

### ✅ 使用示例

```go
func main() {
	client := coinbase.NewClient(
		os.Getenv("COINBASE_API_KEY"),
		os.Getenv("COINBASE_SECRET_KEY"),
		os.Getenv("COINBASE_PASSPHRASE"),
	)

	// 1. 列出产品
	products, _ := client.ListProducts()
	fmt.Printf("共有 %d 个交易对\n", len(products))
	for _, p := range products[:3] { // 打印前3个
		fmt.Printf("- %s (%s/%s), min: %s\n", p.ProductID, p.BaseCurrency, p.QuoteCurrency, p.BaseMinSize)
	}

	// 2. 获取 BTC-USD 行情
	ticker, _ := client.GetTicker("BTC-USD")
	fmt.Printf("BTC-USD: Bid=%s, Ask=%s\n", ticker.Bid, ticker.Ask)

	// 3. 下单（假设之前已下单）
	order, _ := client.PlaceOrder(coinbase.OrderRequest{
		ProductID: "ETH-USD",
		Side:      "BUY",
		Type:      "MARKET",
		Amount:    "0.01",
	})

	// 4. 取消订单
	success, err := client.CancelOrder(order.OrderID)
	if err != nil {
		fmt.Printf("取消失败: %v\n", err)
	} else if success {
		fmt.Println("订单已取消")
	}
}
```

---

### 📚 参考文档

- [Products](https://docs.cdp.coinbase.com/advanced-trade/reference/retailbrokerageapi_getproducts)
- [Best Bid/Ask](https://docs.cdp.coinbase.com/advanced-trade/reference/retailbrokerageapi_getbestbidask)
- [Batch Cancel Orders](https://docs.cdp.coinbase.com/advanced-trade/reference/retailbrokerageapi_batchcancelorders)

---

现在你的 `coinbase.Client` 已支持：
- ✅ 下单（`PlaceOrder`）
- ✅ 取消订单（`CancelOrder`, `CancelOrders`）
- ✅ 获取行情（`GetTicker`）
- ✅ 列出/获取产品（`ListProducts`, `GetProduct`）
- ✅ 获取账户（`GetAccounts`）
- ✅ 查询订单（`GetOrder`）

如需添加 **订单簿（order book）**、**历史订单**、**资金划转** 等功能，也可以按同样模式扩展。需要我继续实现吗？

# 添加 订单簿（order book）、历史订单、资金划转 等功能

当然可以！以下是基于我们已有的通用 `Client`，继续扩展 **订单簿（Order Book）**、**历史订单（Fills / Historical Orders）**、**资金划转（Transfers）** 等功能的完整实现。

> 💡 注意：Coinbase Advanced Trade API 中：
> - “历史订单”通常指 **成交记录（Fills）** 或 **订单历史（Orders）**
> - “资金划转”包括 **充值（deposit）** 和 **提现（withdraw）**，但 REST API 仅支持部分操作（如提现需额外权限）

---

### ✅ 新增结构体与方法（接续 `coinbase` 包）

```go
import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// -----------------------------
// 1. 订单簿（Order Book）
// -----------------------------

// OrderBookLevel 订单簿档位
type OrderBookLevel struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// OrderBook 订单簿
type OrderBook struct {
	ProductID string            `json:"product_id"`
	Bids      []OrderBookLevel  `json:"bids"`
	Asks      []OrderBookLevel  `json:"asks"`
}

// GetOrderBook 获取订单簿（默认深度 100）
func (c *Client) GetOrderBook(productID string, limit int) (*OrderBook, error) {
	if limit <= 0 || limit > 100 {
		limit = 100 // API 最大支持 100
	}

	path := fmt.Sprintf("/product_book?product_id=%s&limit=%d", url.QueryEscape(productID), limit)
	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Pricebook struct {
			ProductID string `json:"product_id"`
			Bids      []struct {
				Price string `json:"price"`
				Size  string `json:"size"`
			} `json:"bids"`
			Asks []struct {
				Price string `json:"price"`
				Size  string `json:"size"`
			} `json:"asks"`
		} `json:"pricebook"`
	}

	if err := json.Unmarshal(respBody, &wrapper); err != nil {
		return nil, fmt.Errorf("unmarshal GetOrderBook response: %w", err)
	}

	// 转换 bids/asks
	bids := make([]OrderBookLevel, len(wrapper.Pricebook.Bids))
	for i, b := range wrapper.Pricebook.Bids {
		bids[i] = OrderBookLevel{Price: b.Price, Size: b.Size}
	}
	asks := make([]OrderBookLevel, len(wrapper.Pricebook.Asks))
	for i, a := range wrapper.Pricebook.Asks {
		asks[i] = OrderBookLevel{Price: a.Price, Size: a.Size}
	}

	return &OrderBook{
		ProductID: wrapper.Pricebook.ProductID,
		Bids:      bids,
		Asks:      asks,
	}, nil
}

// -----------------------------
// 2. 历史订单（Orders）与成交记录（Fills）
// -----------------------------

// Order 历史订单详情
type Order struct {
	OrderID           string `json:"order_id"`
	ClientOrderID     string `json:"client_order_id,omitempty"`
	ProductID         string `json:"product_id"`
	Side              string `json:"side"` // BUY / SELL
	Type              string `json:"order_type"`
	Size              string `json:"size"`
	Status            string `json:"status"` // OPEN, FILLED, CANCELLED, etc.
	TimeInForce       string `json:"time_in_force,omitempty"`
	CreatedTime       string `json:"created_time"` // ISO 8601
	FilledSize        string `json:"filled_size"`
	AverageFilledPrice string `json:"average_filled_price,omitempty"`
}

// ListOrdersParams 查询历史订单参数
type ListOrdersParams struct {
	ProductID string
	OrderIDs  []string
	Status    []string // e.g., ["OPEN", "FILLED"]
	Limit     int      // max 100
	StartDate string   // ISO 8601
	EndDate   string   // ISO 8601
}

// ListOrders 获取历史订单列表
func (c *Client) ListOrders(params ListOrdersParams) ([]Order, error) {
	v := url.Values{}

	if params.ProductID != "" {
		v.Set("product_id", params.ProductID)
	}
	if len(params.OrderIDs) > 0 {
		v.Set("order_ids", strings.Join(params.OrderIDs, ","))
	}
	if len(params.Status) > 0 {
		v.Set("order_status", strings.Join(params.Status, ","))
	}
	if params.Limit > 0 && params.Limit <= 100 {
		v.Set("limit", strconv.Itoa(params.Limit))
	} else {
		v.Set("limit", "100")
	}
	if params.StartDate != "" {
		v.Set("start_date", params.StartDate)
	}
	if params.EndDate != "" {
		v.Set("end_date", params.EndDate)
	}

	path := "/orders"
	if len(v) > 0 {
		path += "?" + v.Encode()
	}

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Orders []Order `json:"orders"`
	}

	if err := json.Unmarshal(respBody, &wrapper); err != nil {
		return nil, fmt.Errorf("unmarshal ListOrders response: %w", err)
	}

	return wrapper.Orders, nil
}

// Fill 成交记录
type Fill struct {
	FillID        string `json:"fill_id"`
	OrderID       string `json:"order_id"`
	TradeID       string `json:"trade_id"`
	ProductID     string `json:"product_id"`
	Side          string `json:"side"`
	Size          string `json:"size"`
	Price         string `json:"price"`
	CreatedTime   string `json:"created_time"` // ISO 8601
	Fee           string `json:"fee"`
	FeeCurrency   string `json:"fee_currency"`
}

// ListFillsParams 查询成交记录参数
type ListFillsParams struct {
	ProductID string
	OrderID   string
	Limit     int
	StartDate string // ISO 8601
	EndDate   string // ISO 8601
}

// ListFills 获取成交记录（历史交易）
func (c *Client) ListFills(params ListFillsParams) ([]Fill, error) {
	v := url.Values{}

	if params.ProductID != "" {
		v.Set("product_id", params.ProductID)
	}
	if params.OrderID != "" {
		v.Set("order_id", params.OrderID)
	}
	if params.Limit > 0 && params.Limit <= 100 {
		v.Set("limit", strconv.Itoa(params.Limit))
	} else {
		v.Set("limit", "100")
	}
	if params.StartDate != "" {
		v.Set("start_date", params.StartDate)
	}
	if params.EndDate != "" {
		v.Set("end_date", params.EndDate)
	}

	path := "/fills"
	if len(v) > 0 {
		path += "?" + v.Encode()
	}

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Fills []Fill `json:"fills"`
	}

	if err := json.Unmarshal(respBody, &wrapper); err != nil {
		return nil, fmt.Errorf("unmarshal ListFills response: %w", err)
	}

	return wrapper.Fills, nil
}

// -----------------------------
// 3. 资金划转（Transfers）
// -----------------------------

// TransferRequest 资金划转请求（仅支持提现到 Coinbase 账户）
type TransferRequest struct {
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`    // e.g., "BTC", "USD"
	TransferType string `json:"transfer_type"` // "INTERNAL_WITHDRAW" or "INTERNAL_DEPOSIT"
	// 注意：Advanced Trade API 通常只允许 INTERNAL 转账（主账户 ↔ 交易账户）
}

// TransferResponse 划转响应
type TransferResponse struct {
	TransferID string `json:"transfer_id"`
	Amount     string `json:"amount"`
	Currency   string `json:"currency"`
	Type       string `json:"type"` // e.g., "INTERNAL_WITHDRAW"
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

// CreateTransfer 创建资金划转（仅限内部转账）
func (c *Client) CreateTransfer(req TransferRequest) (*TransferResponse, error) {
	respBody, err := c.doRequest("POST", "/transfers", req)
	if err != nil {
		return nil, err
	}

	var resp TransferResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal CreateTransfer response: %w", err)
	}
	return &resp, nil
}

// ListTransfersParams 查询划转记录参数
type ListTransfersParams struct {
	TransferType string // e.g., "INTERNAL_WITHDRAW"
	Limit        int
}

// Transfer 划转记录
type Transfer struct {
	TransferID string `json:"transfer_id"`
	Amount     string `json:"amount"`
	Currency   string `json:"currency"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

// ListTransfers 获取资金划转历史
func (c *Client) ListTransfers(params ListTransfersParams) ([]Transfer, error) {
	v := url.Values{}
	if params.TransferType != "" {
		v.Set("transfer_type", params.TransferType)
	}
	if params.Limit > 0 && params.Limit <= 100 {
		v.Set("limit", strconv.Itoa(params.Limit))
	} else {
		v.Set("limit", "100")
	}

	path := "/transfers"
	if len(v) > 0 {
		path += "?" + v.Encode()
	}

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Transfers []Transfer `json:"transfers"`
	}

	if err := json.Unmarshal(respBody, &wrapper); err != nil {
		return nil, fmt.Errorf("unmarshal ListTransfers response: %w", err)
	}

	return wrapper.Transfers, nil
}
```

---

### 📌 重要说明

#### 🔹 **资金划转限制**
- Coinbase Advanced Trade API **仅支持内部转账**（`INTERNAL_DEPOSIT` / `INTERNAL_WITHDRAW`），即：
  - **交易账户 ↔ 主 Coinbase 账户**
- **不支持直接提现到外部钱包**（如 BTC 地址）通过此 API。外部提现需通过 [Coinbase Prime](https://prime.coinbase.com/) 或网页操作。
- 示例转账方向：
  ```go
  // 从主账户转入交易账户（充值）
  req := TransferRequest{
      Amount:      "100",
      Currency:    "USD",
      TransferType: "INTERNAL_DEPOSIT",
  }

  // 从交易账户转出到主账户（提现）
  req := TransferRequest{
      Amount:      "0.1",
      Currency:    "BTC",
      TransferType: "INTERNAL_WITHDRAW",
  }
  ```

#### 🔹 **时间格式**
- 所有时间参数（如 `StartDate`）需为 **ISO 8601 格式**，例如：
  ```go
  start := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
  ```

#### 🔹 **分页**
- 当前实现仅支持单页（最多 100 条）。如需完整分页，可解析响应中的 `cursor` 字段（部分端点支持），但为简化暂未实现。

---

### ✅ 使用示例

```go
func main() {
	client := coinbase.NewClient(
		os.Getenv("COINBASE_API_KEY"),
		os.Getenv("COINBASE_SECRET_KEY"),
		os.Getenv("COINBASE_PASSPHRASE"),
	)

	// 1. 获取订单簿
	ob, _ := client.GetOrderBook("BTC-USD", 10)
	fmt.Printf("订单簿深度: Bids=%d, Asks=%d\n", len(ob.Bids), len(ob.Asks))

	// 2. 查询最近成交
	fills, _ := client.ListFills(coinbase.ListFillsParams{
		ProductID: "BTC-USD",
		Limit:     5,
	})
	for _, f := range fills {
		fmt.Printf("成交: %s %s @ %s\n", f.Side, f.Size, f.Price)
	}

	// 3. 查询历史订单
	orders, _ := client.ListOrders(coinbase.ListOrdersParams{
		Status: []string{"FILLED"},
		Limit:  5,
	})
	for _, o := range orders {
		fmt.Printf("订单: %s, 状态: %s, 已成交: %s\n", o.OrderID, o.Status, o.FilledSize)
	}

	// 4. 内部转账（示例：提 BTC 到主账户）
	// transfer, err := client.CreateTransfer(coinbase.TransferRequest{
	//     Amount:      "0.01",
	//     Currency:    "BTC",
	//     TransferType: "INTERNAL_WITHDRAW",
	// })
	// if err != nil {
	//     fmt.Printf("转账失败: %v\n", err)
	// } else {
	//     fmt.Printf("转账成功 ID: %s\n", transfer.TransferID)
	// }
}
```

---

### 📚 参考文档

- [Order Book](https://docs.cdp.coinbase.com/advanced-trade/reference/retailbrokerageapi_getproductbook)
- [List Orders](https://docs.cdp.coinbase.com/advanced-trade/reference/retailbrokerageapi_gethistoricalorders)
- [List Fills](https://docs.cdp.coinbase.com/advanced-trade/reference/retailbrokerageapi_getfills)
- [Transfers](https://docs.cdp.coinbase.com/advanced-trade/reference/retailbrokerageapi_createtransfer)

---

现在你的 `coinbase.Client` 已非常完整，支持：
- ✅ 下单 / 取消订单
- ✅ 行情 / 订单簿
- ✅ 产品列表
- ✅ 账户 / 订单历史 / 成交记录
- ✅ 内部资金划转

如需 **WebSocket 实时行情**、**更复杂的分页** 或 **错误重试机制**，也可以进一步扩展。需要我帮你实现吗？


# 用WebSocket 接收实时行情

当然可以！Coinbase 提供了 **WebSocket Feed** 用于接收实时行情、订单簿、交易等数据。下面我将为你实现一个 **Go 语言的 WebSocket 客户端**，用于连接 Coinbase 的公开 WebSocket 服务（无需认证），并订阅实时行情（ticker）、交易（matches）、订单簿（level2）等频道。

> ✅ **无需 API Key**：行情类 WebSocket 是公开的，不涉及账户操作。

---

### ✅ 1. 安装依赖

```bash
go get github.com/gorilla/websocket
```

---

### ✅ 2. 定义消息结构体

```go
// coinbase/ws.go

package coinbase

import (
	"encoding/json"
	"time"
)

const (
	WebSocketURL = "wss://advanced-trade-ws.coinbase.com"

	// 频道类型
	ChannelTicker = "ticker"
	ChannelMatches = "matches"
	ChannelLevel2  = "level2"
	ChannelHeartbeat = "heartbeat"
)

// WebSocket 消息通用结构
type WSMessage struct {
	Type      string          `json:"type"`
	ProductID string          `json:"product_id,omitempty"`
	Channel   string          `json:"channel,omitempty"`
	Time      *time.Time      `json:"time,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// TickerData 实时行情数据
type TickerData struct {
	Price       string `json:"price"`
	Bid         string `json:"bid"`
	Ask         string `json:"ask"`
	Volume24H   string `json:"volume_24h"`
	TradeID     string `json:"trade_id,omitempty"`
	Side        string `json:"side,omitempty"` // for matches
	Size        string `json:"size,omitempty"` // for matches
}

// Level2Snapshot 订单簿快照（首次订阅时）
type Level2Snapshot struct {
	Bids [][]string `json:"bids"` // [price, size]
	Asks [][]string `json:"asks"`
}

// Level2Update 订单簿更新
type Level2Update struct {
	Bids [][]string `json:"bids"`
	Asks [][]string `json:"asks"`
}
```

---

### ✅ 3. 实现 WebSocket 客户端

```go
// coinbase/ws_client.go

package coinbase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// WSClient WebSocket 客户端
type WSClient struct {
	conn      *websocket.Conn
	done      chan struct{}
	msgChan   chan WSMessage
	errorChan chan error
}

// NewWSClient 创建新的 WebSocket 客户端
func NewWSClient() *WSClient {
	return &WSClient{
		done:      make(chan struct{}),
		msgChan:   make(chan WSMessage, 100),
		errorChan: make(chan error, 10),
	}
}

// Connect 连接到 Coinbase WebSocket
func (c *WSClient) Connect(ctx context.Context) error {
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.DialContext(ctx, WebSocketURL, nil)
	if err != nil {
		return fmt.Errorf("dial websocket: %w", err)
	}
	c.conn = conn

	// 启动读取消息的 goroutine
	go c.readLoop()
	return nil
}

// Subscribe 订阅指定产品和频道
func (c *WSClient) Subscribe(productIDs []string, channels []string) error {
	subscribeMsg := map[string]interface{}{
		"type":       "subscribe",
		"product_ids": productIDs,
		"channels":   channels,
	}

	return c.conn.WriteJSON(subscribeMsg)
}

// Unsubscribe 取消订阅
func (c *WSClient) Unsubscribe(productIDs []string, channels []string) error {
	unsubscribeMsg := map[string]interface{}{
		"type":       "unsubscribe",
		"product_ids": productIDs,
		"channels":   channels,
	}
	return c.conn.WriteJSON(unsubscribeMsg)
}

// Messages 返回接收到的消息通道
func (c *WSClient) Messages() <-chan WSMessage {
	return c.msgChan
}

// Errors 返回错误通道
func (c *WSClient) Errors() <-chan error {
	return c.errorChan
}

// Close 关闭连接
func (c *WSClient) Close() error {
	close(c.done)
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// readLoop 读取消息循环
func (c *WSClient) readLoop() {
	defer close(c.msgChan)
	defer close(c.errorChan)

	for {
		select {
		case <-c.done:
			return
		default:
			// 设置读取超时（可选）
			_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

			var msg WSMessage
			err := c.conn.ReadJSON(&msg)
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Println("WebSocket closed normally")
				} else {
					c.errorChan <- fmt.Errorf("read websocket message: %w", err)
				}
				return
			}

			// 将消息发送到通道
			select {
			case c.msgChan <- msg:
			case <-c.done:
				return
			}
		}
	}
}
```

---

### ✅ 4. 使用示例

```go
// main.go

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"your-module/coinbase" // 替换为你的模块路径
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client := coinbase.NewWSClient()
	if err := client.Connect(ctx); err != nil {
		log.Fatal("连接失败:", err)
	}
	defer client.Close()

	// 订阅 BTC-USD 和 ETH-USD 的 ticker 和 matches
	productIDs := []string{"BTC-USD", "ETH-USD"}
	channels := []string{coinbase.ChannelTicker, coinbase.ChannelMatches}

	if err := client.Subscribe(productIDs, channels); err != nil {
		log.Fatal("订阅失败:", err)
	}

	fmt.Println("已订阅，等待消息...")

	// 处理消息
	for {
		select {
		case msg := <-client.Messages():
			switch msg.Type {
			case "ticker":
				var data coinbase.TickerData
				if err := json.Unmarshal(msg.Data, &data); err != nil {
					log.Printf("解析 ticker 失败: %v", err)
					continue
				}
				fmt.Printf("[%s] 最新价: %s, 买一: %s, 卖一: %s\n",
					msg.ProductID, data.Price, data.Bid, data.Ask)

			case "match":
				var data coinbase.TickerData
				if err := json.Unmarshal(msg.Data, &data); err != nil {
					log.Printf("解析 match 失败: %v", err)
					continue
				}
				fmt.Printf("[%s] 成交: %s %s @ %s\n",
					msg.ProductID, data.Side, data.Size, data.Price)

			case "subscriptions":
				fmt.Println("订阅确认:", string(msg.Data))

			case "error":
				log.Printf("服务端错误: %s", msg.Data)

			default:
				// 忽略 heartbeat 等
			}

		case err := <-client.Errors():
			log.Printf("WebSocket 错误: %v", err)
			return

		case <-ctx.Done():
			fmt.Println("超时退出")
			return
		}
	}
}
```

---

### 📌 支持的频道说明

| 频道 | 说明 |
|------|------|
| `ticker` | 每次有新成交时推送最新价格、买卖一价 |
| `matches` | 每笔成交详情（价格、数量、方向） |
| `level2` | 实时订单簿更新（需处理快照 + 增量更新） |
| `heartbeat` | 心跳包（每秒一次） |

> 💡 **订单簿（level2）使用提示**：
> - 首次订阅会收到 `l2update` 类型的 **快照（snapshot）**
> - 后续收到的是 **增量更新（update）**
> - 你需要维护本地订单簿状态

---

### 🛡️ 注意事项

1. **不要高频订阅/取消**：避免被限流。
2. **重连机制**：生产环境建议添加自动重连逻辑。
3. **消息顺序**：WebSocket 保证消息顺序，但网络抖动可能导致延迟。
4. **产品 ID**：必须使用 Coinbase 官方格式，如 `"BTC-USD"`、`"ETH-USDT"`。

---

### ✅ 扩展建议（生产环境）

- 添加 **自动重连 + 指数退避**
- 实现 **订单簿状态管理器**
- 使用 **context 控制生命周期**
- 添加 **日志和监控**

需要我帮你实现 **自动重连** 或 **订单簿状态管理** 吗？

