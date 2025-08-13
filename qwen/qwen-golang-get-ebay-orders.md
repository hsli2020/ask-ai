# 用golang编写一个程序，获取我的ebay店铺最近几天的订单信息

要在 Go（Golang）中获取你 eBay 店铺最近几天的订单信息，你需要使用 eBay 的 **Trading API**（现在部分功能已迁移到 **Sell API**）。由于 eBay 要求使用 **OAuth 2.0** 进行身份验证，推荐使用较新的 **Sell Fulfillment API** 来获取订单。

下面是一个使用 **eBay Sell Fulfillment API** 获取最近 7 天订单的完整 Go 示例程序。

---

## ✅ 前提条件

1. **注册 eBay 开发者账号**：https://developer.ebay.com/
2. **创建应用**（获取 Client ID 和 Client Secret）
3. **获取 OAuth 2.0 Token**
   - 你需要通过 OAuth 流程（推荐使用 **Client Credentials Grant** 获取访问令牌，用于后端服务）
   - 或者使用 **User Access Token**（如果你要访问特定用户（你）的订单）

> 对于店铺订单，通常需要 **User Access Token**（因为订单属于特定卖家账户）。

---

## 📦 依赖

你需要安装：

```bash
go get golang.org/x/oauth2
```

---

## 🧾 示例代码（Go）

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

// eBay API 配置
const (
	ClientID     = "YOUR_CLIENT_ID"
	ClientSecret = "YOUR_CLIENT_SECRET"
	// 你的生产环境用户名 Token（通过 OAuth 获取）
	UserAccessToken = "YOUR_USER_ACCESS_TOKEN"
)

// Order 代表一个订单
type Order struct {
	OrderID      string    `json:"orderId"`
	CreationDate time.Time `json:"creationDate"`
	Buyer        struct {
		Username string `json:"username"`
	} `json:"buyer"`
	OrderFulfillmentStatus string `json:"orderFulfillmentStatus"`
}

// OrdersResponse 响应结构
type OrdersResponse struct {
	Orders []Order `json:"orders"`
	Total  int     `json:"total"`
}

func main() {
	ctx := context.Background()

	// 设置 OAuth2 配置（仅用于刷新 token，这里我们直接使用已有 token）
	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: UserAccessToken,
	})

	client := oauth2.NewClient(ctx, ts)

	// 计算最近 7 天的开始时间（eBay API 要求 ISO 8601 格式，UTC）
	now := time.Now().UTC()
	sevenDaysAgo := now.AddDate(0, 0, -7)
	// 格式: 2023-10-01T00:00:00.000Z
	startDate := sevenDaysAgo.Format("2006-01-02T15:04:05.000Z")

	url := fmt.Sprintf(
		"https://api.ebay.com/sell/fulfillment/v1/order?creationdate=%s&limit=100",
		startDate,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("创建请求失败:", err)
	}

	// eBay 要求指定 marketplace ID（如 EBAY_US）
	req.Header.Set("X-EBAY-C-MARKETPLACE-ID", "EBAY_US")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+UserAccessToken)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("请求失败:", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("读取响应失败:", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("eBay API 错误: %d, body: %s", resp.StatusCode, string(body))
	}

	var ordersResp OrdersResponse
	if err := json.Unmarshal(body, &ordersResp); err != nil {
		log.Fatal("解析 JSON 失败:", err)
	}

	fmt.Printf("共找到 %d 个订单:\n", ordersResp.Total)
	for _, order := range ordersResp.Orders {
		fmt.Printf("订单ID: %s, 创建时间: %s, 买家: %s, 状态: %s\n",
			order.OrderID,
			order.CreationDate.Format("2006-01-02 15:04:05"),
			order.Buyer.Username,
			order.OrderFulfillmentStatus)
	}
}
```

---

## 🔐 如何获取 `UserAccessToken`

你需要通过 **OAuth 2.0 Authorization Code Flow** 获取用户 Token：

1. 重定向用户（你自己）到 eBay 登录授权页：
   ```
   https://auth.ebay.com/oauth2/authorize?
     client_id=YOUR_CLIENT_ID&
     redirect_uri=YOUR_REDIRECT_URI&
     response_type=code&
     scope=https://api.ebay.com/oauth/api_scope/sell.fulfillment.readonly
   ```

2. 获取 `code` 后，用它换 `access_token`：
   ```bash
   curl -X POST \
     https://api.ebay.com/identity/v1/oauth2/token \
     -H 'Content-Type: application/x-www-form-urlencoded' \
     -H 'Authorization: Basic <Base64(ClientID:ClientSecret)>' \
     -d 'grant_type=authorization_code&code=AUTHORIZATION_CODE&redirect_uri=YOUR_REDIRECT_URI'
   ```

> ⚠️ 注意：Token 有有效期（通常 2 小时），你需要保存并刷新（使用 `refresh_token`）。

---

## 🌐 注意事项

- eBay API 是 **RESTful**，但部分旧功能仍在 Trading API（SOAP）。
- **Sell Fulfillment API** 是获取订单的推荐方式。
- 时区：eBay 使用 UTC，确保时间格式正确。
- 生产环境 URL 是 `https://api.ebay.com`，沙箱是 `https://api.sandbox.ebay.com`。

---

## ✅ 下一步

- 将 Token 管理封装（自动刷新）
- 支持分页（`href` + `limit`/`offset`）
- 写入数据库或生成报表

如果你提供你的 Client ID / redirect URI，我可以帮你生成授权 URL。

需要我帮你写一个自动获取 OAuth token 的 Go 程序吗？

# 写入数据库

好的！下面是在之前获取 eBay 订单的基础上，**将订单数据写入 SQLite 数据库**的完整 Go 程序（使用 `gorm` ORM，轻量且支持 SQLite、MySQL、PostgreSQL 等）。

---

## ✅ 功能说明

1. 使用 eBay Sell Fulfillment API 获取最近 7 天订单
2. 将订单信息（ID、创建时间、买家、状态）存入本地 SQLite 数据库
3. 避免重复插入（通过订单 ID 去重）
4. 使用 `GORM` 管理数据库

---

## 📦 依赖安装

```bash
go mod init ebay-order-sync
go get -u gorm.io/gorm
go get -u gorm.io/driver/sqlite
go get golang.org/x/oauth2
```

---

## 🧾 完整 Go 代码

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// === eBay 配置 ===
const (
	ClientID       = "YOUR_CLIENT_ID"
	ClientSecret   = "YOUR_CLIENT_SECRET"
	UserAccessToken = "YOUR_USER_ACCESS_TOKEN" // 有效 User Access Token
)

// === 数据库模型 ===
type Order struct {
	ID           uint      `gorm:"primaryKey"`
	OrderID      string    `gorm:"uniqueIndex;not null"`
	CreationDate time.Time
	BuyerUsername string    `gorm:"column:buyer_username"`
	Status       string    `gorm:"column:fulfillment_status"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TableName 指定表名
func (Order) TableName() string {
	return "ebay_orders"
}

// === eBay API 响应结构 ===
type EbayOrder struct {
	OrderID      string    `json:"orderId"`
	CreationDate time.Time `json:"creationDate"`
	Buyer        struct {
		Username string `json:"username"`
	} `json:"buyer"`
	OrderFulfillmentStatus string `json:"orderFulfillmentStatus"`
}

type OrdersResponse struct {
	Orders []EbayOrder `json:"orders"`
	Total  int         `json:"total"`
}

func main() {
	// 1. 初始化数据库
	db, err := gorm.Open(sqlite.Open("ebay_orders.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 自动迁移表结构
	err = db.AutoMigrate(&Order{})
	if err != nil {
		log.Fatal("迁移数据库失败:", err)
	}

	// 2. 获取 eBay 订单
	ebayOrders, err := fetchEbayOrders()
	if err != nil {
		log.Fatal("获取 eBay 订单失败:", err)
	}

	// 3. 转换并保存到数据库（避免重复）
	var ordersToSave []Order
	for _, eo := range ebayOrders {
		ordersToSave = append(ordersToSave, Order{
			OrderID:       eo.OrderID,
			CreationDate:  eo.CreationDate,
			BuyerUsername: eo.Buyer.Username,
			Status:        eo.OrderFulfillmentStatus,
		})
	}

	// 使用 ON CONFLICT (OrderID) 忽略重复（SQLite 支持）
	// GORM 的 clause.OnConflict 用于 upsert 或忽略
	result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "order_id"}},
		DoNothing: true,
	}).Create(&ordersToSave)

	fmt.Printf("成功插入 %d 条新订单（跳过重复）\n", result.RowsAffected)
}

// fetchEbayOrders 从 eBay API 获取最近 7 天订单
func fetchEbayOrders() ([]EbayOrder, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: UserAccessToken,
	})
	client := oauth2.NewClient(ctx, ts)

	// 最近 7 天（UTC）
	sevenDaysAgo := time.Now().UTC().AddDate(0, 0, -7)
	startDate := sevenDaysAgo.Format("2006-01-02T15:04:05.000Z")

	url := fmt.Sprintf("https://api.ebay.com/sell/fulfillment/v1/order?creationdate=%s&limit=100", startDate)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("构造请求失败: %w", err)
	}

	req.Header.Set("X-EBAY-C-MARKETPLACE-ID", "EBAY_US")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+UserAccessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("eBay API 返回错误: %d, body: %s", resp.StatusCode, string(body))
	}

	var ordersResp OrdersResponse
	if err := json.Unmarshal(body, &ordersResp); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	return ordersResp.Orders, nil
}
```

---

## 🗃️ 生成的数据库表结构

运行后会生成 `ebay_orders.db` 文件，表结构如下（SQLite）：

```sql
CREATE TABLE `ebay_orders` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `order_id` text UNIQUE NOT NULL,
  `creation_date` datetime,
  `buyer_username` text,
  `fulfillment_status` text,
  `created_at` datetime,
  `updated_at` datetime
);
```

- `order_id` 是唯一索引，防止重复插入
- 自动记录 `created_at` / `updated_at`

---

## 🔐 安全建议

- **不要将 Token 硬编码在代码中**！建议使用环境变量：
  ```go
  import "os"

  UserAccessToken = os.Getenv("EBAY_USER_ACCESS_TOKEN")
  ```
- 使用 `.env` 文件 + `godotenv` 库管理配置（可选）

---

## 🔄 扩展建议

1. **定时任务**：用 `cron` 每天自动同步
2. **支持 MySQL/PostgreSQL**：只需改一行数据库连接
   ```go
   // MySQL 示例
   gorm.Open(mysql.Open("user:pass@tcp(127.0.0.1:3306)/dbname"), &gorm.Config{})
   ```
3. **记录日志到文件**
4. **处理分页**（eBay 最多返回 100 条，需循环 `href`）

---

## ✅ 运行

```bash
go run main.go
```

输出示例：
```
成功插入 5 条新订单（跳过重复）
```

---

如果你希望我帮你：
- 自动生成 OAuth 授权流程（获取 Token）
- 支持分页获取全部订单
- 导出为 CSV
- 部署为定时服务（systemd / cron）

请告诉我！
