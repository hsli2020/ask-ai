# 🤵 “短链接”网站的工作原理

“短链接”网站（如 bit.ly、t.cn、tinyurl.com 等）的工作原理，其实就是一个**URL 重定向服务**，核心思想是用一个很短的唯一 ID 来替代原本很长的 URL，用户访问短链接时，服务器再把用户引导到原来的长链接。

我帮你分步骤拆一下：

---

## 1. 用户生成短链接的过程

1. **提交长链接**

   * 用户在短链接网站上输入一个长链接，例如：

     ```
     https://www.example.com/article?id=12345&source=weibo&utm_campaign=summer
     ```

2. **生成唯一标识（短码）**

   * 网站会生成一个短的字符串（例如 `abc123`）。
   * 生成方法常见有几种：

     * **随机字符串**（防止被猜测，需要查重）
     * **顺序 ID + Base62 编码**（数字转成 0-9、a-z、A-Z，共 62 个字符，短而紧凑）
     * **哈希（Hash）截断**（如 MD5/SHA1 取前几位，但需避免冲突）

3. **存储映射关系**

   * 网站把 **短码 → 原始长链接** 的映射存到数据库，比如：

     | 短码     | 原链接                                                                                                             |
     | ------ | --------------------------------------------------------------------------------------------------------------- |
     | abc123 | [https://www.example.com/article?id=12345\&source=weibo](https://www.example.com/article?id=12345&source=weibo) |
   * 还可能记录：

     * 生成时间
     * 创建者
     * 有效期
     * 访问次数

4. **返回短链接**

   * 服务器返回一个短网址，例如：

     ```
     https://short.ly/abc123
     ```

---

## 2. 用户访问短链接的过程

1. **浏览器请求短链接**

   * 用户点击 `https://short.ly/abc123`。
2. **短链接服务器查找映射**

   * 根据 `abc123` 在数据库中查到对应的原链接。
3. **返回重定向**

   * 服务器用 **HTTP 301（永久重定向）或 302（临时重定向）**，告诉浏览器跳转到原始长链接：

     ```http
     HTTP/1.1 301 Moved Permanently
     Location: https://www.example.com/article?id=12345&source=weibo
     ```
4. **浏览器跳转**

   * 浏览器自动访问长链接目标地址。

---

## 3. 常见的附加功能

* **访问统计**：记录访问次数、来源、地理位置、浏览器类型等。
* **有效期**：短链接可能会在一定时间后失效。
* **自定义短码**：让用户自己指定短码，比如 `https://short.ly/summer2025`。
* **防钓鱼 / 黑名单**：检测目标链接是否合法、安全。
* **批量生成**：为营销或广告活动一次生成成百上千个短链。

---

## 4. 技术要点

* **数据库设计**：高并发场景需要缓存（Redis）来加速短码映射查询。
* **短码生成算法**：要平衡短小、唯一性、防冲突。
* **SEO 友好性**：使用 301 重定向可以把权重传递到原链接。
* **安全性**：防止恶意跳转、垃圾链接。

---

如果你愿意，我可以帮你画一张**短链接工作原理流程图**，一眼就能看懂整个过程。
这样你不仅能记住，还能用它给别人讲解。
你要我画吗？
--------------------------------------------------------------------------------

# 🤵 设计数据库表，记录短链接相关信息，生成sql语句

好的，我给你设计一个**短链接服务数据库表**，考虑常用功能（短码映射、访问统计、有效期、安全等），并生成可直接用的 SQL 语句。

---

## 1. 设计目标

* **高效查找短码 → 长链接**（主键或唯一索引）
* **记录访问统计**（访问次数、最后访问时间）
* **支持有效期**
* **可选自定义短码**
* **记录安全状态（黑名单、白名单）**

---

## 2. 数据表设计

### 表：`short_links`

| 字段名               | 类型                                           | 说明                |
| ----------------- | -------------------------------------------- | ----------------- |
| `id`              | BIGINT UNSIGNED PK AI                        | 内部主键              |
| `short_code`      | VARCHAR(16) NOT NULL                         | 短码，唯一索引           |
| `original_url`    | TEXT NOT NULL                                | 原始长链接             |
| `custom_code`     | TINYINT(1) DEFAULT 0                         | 是否用户自定义（0=否, 1=是） |
| `creator_id`      | BIGINT NULL                                  | 创建者用户ID（可选）       |
| `visit_count`     | BIGINT UNSIGNED DEFAULT 0                    | 访问次数              |
| `last_visited_at` | DATETIME NULL                                | 最后一次访问时间          |
| `created_at`      | DATETIME NOT NULL DEFAULT CURRENT\_TIMESTAMP | 创建时间              |
| `expires_at`      | DATETIME NULL                                | 过期时间（NULL 表示永久有效） |
| `status`          | TINYINT DEFAULT 1                            | 链接状态（1=正常, 0=禁用）  |

---

## 3. 建表 SQL 语句（MySQL 示例）

```sql
CREATE TABLE `short_links` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `short_code` VARCHAR(16) NOT NULL COMMENT '短码',
    `original_url` TEXT NOT NULL COMMENT '原始长链接',
    `custom_code` TINYINT(1) DEFAULT 0 COMMENT '是否用户自定义短码',
    `creator_id` BIGINT NULL COMMENT '创建者用户ID',
    `visit_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '访问次数',
    `last_visited_at` DATETIME NULL COMMENT '最后访问时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `expires_at` DATETIME NULL COMMENT '过期时间',
    `status` TINYINT DEFAULT 1 COMMENT '状态：1=正常, 0=禁用',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_short_code` (`short_code`),
    INDEX `idx_creator_id` (`creator_id`),
    INDEX `idx_expires_at` (`expires_at`),
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='短链接映射表';
```

---

## 4. 附加访问日志表（可选）

如果需要记录每次访问的来源、IP、UA，可以额外建一张表。

```sql
CREATE TABLE `short_link_visits` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '访问记录ID',
    `short_link_id` BIGINT UNSIGNED NOT NULL COMMENT '对应的短链接ID',
    `ip_address` VARCHAR(45) NOT NULL COMMENT '访问IP',
    `user_agent` VARCHAR(255) NULL COMMENT '浏览器UA',
    `referer` VARCHAR(255) NULL COMMENT '来源地址',
    `visited_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '访问时间',
    PRIMARY KEY (`id`),
    INDEX `idx_short_link_id` (`short_link_id`),
    CONSTRAINT `fk_visits_short_link` FOREIGN KEY (`short_link_id`) REFERENCES `short_links` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='短链接访问日志表';
```

---

如果你愿意，我可以帮你在这个设计的基础上，补一个**短码生成算法 + 插入 SQL 示例**，这样你就可以直接跑起来生成短链接服务。
你要我直接写这部分代码吗？
--------------------------------------------------------------------------------
请帮我实现“短码生成算法 + 插入 SQL 示例”，用golang语言

好，我来帮你用 **Golang** 实现一个完整的 **短码生成算法 + 插入 SQL 示例**，
包含：

1. 生成唯一短码（Base62 编码）
2. 检查数据库是否冲突
3. 插入数据库（MySQL 示例）

---

## 1. 短码生成思路

* **Base62**：字符集 `0-9a-zA-Z`，生成短而可读的短码
* **随机生成 + 数据库查重**：确保唯一
* 短码长度建议 6\~8 位（约有 62⁶ ≈ 57B 种组合）

---

## 2. Golang 示例代码

```go
package main

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Base62 字符集
const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// 生成短码（长度固定）
func generateShortCode(length int) (string, error) {
	code := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(base62Chars))))
		if err != nil {
			return "", err
		}
		code[i] = base62Chars[num.Int64()]
	}
	return string(code), nil
}

// 检查短码是否已存在
func shortCodeExists(db *sql.DB, code string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM short_links WHERE short_code = ?)"
	err := db.QueryRow(query, code).Scan(&exists)
	return exists, err
}

// 生成唯一短码（自动重试直到不重复）
func generateUniqueShortCode(db *sql.DB, length int) (string, error) {
	for {
		code, err := generateShortCode(length)
		if err != nil {
			return "", err
		}
		exists, err := shortCodeExists(db, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
}

// 插入短链接到数据库
func insertShortLink(db *sql.DB, shortCode, originalURL string, creatorID *int64, expiresAt *time.Time) error {
	query := `
		INSERT INTO short_links (short_code, original_url, custom_code, creator_id, created_at, expires_at, status)
		VALUES (?, ?, 0, ?, NOW(), ?, 1)
	`
	_, err := db.Exec(query, shortCode, originalURL, creatorID, expiresAt)
	return err
}

func main() {
	// 连接 MySQL
	dsn := "root:password@tcp(127.0.0.1:3306)/shortlink?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 生成唯一短码
	shortCode, err := generateUniqueShortCode(db, 6)
	if err != nil {
		log.Fatalf("短码生成失败: %v", err)
	}

	// 原始长链接
	originalURL := "https://www.example.com/article?id=12345&utm_source=test"

	// 可选的创建者ID
	var creatorID *int64 = nil

	// 可选的过期时间
	var expiresAt *time.Time
	// t := time.Now().AddDate(0, 1, 0) // 1个月后过期
	// expiresAt = &t

	// 插入数据库
	err = insertShortLink(db, shortCode, originalURL, creatorID, expiresAt)
	if err != nil {
		log.Fatalf("插入数据库失败: %v", err)
	}

	fmt.Printf("短链接生成成功: https://short.ly/%s -> %s\n", shortCode, originalURL)
}
```

---

## 3. 使用说明

1. **建表**
   先用我之前给你的 `CREATE TABLE short_links` SQL 建表。

2. **安装依赖**

   ```bash
   go get github.com/go-sql-driver/mysql
   ```

3. **运行**

   * 修改 MySQL DSN（账号、密码、数据库名）
   * 执行：

     ```bash
     go run main.go
     ```
   * 输出示例：

     ```
     短链接生成成功: https://short.ly/xY8aZ2 -> https://www.example.com/article?id=12345&utm_source=test
     ```

---

我建议你如果打算做成一个**高并发的短链接服务**，可以加上 **Redis 缓存** 做短码查重和映射查询，这样速度会非常快。

如果你愿意，我可以帮你把这个代码升级成 **带 Redis 缓存 + 访问统计 + 自动过期** 的版本，
那就可以直接做一个线上短链接服务了。
你要我直接写这个升级版吗？

--------------------------------------------------------------------------------
将代码升级成 ”带 Redis 缓存 + 访问统计 + 自动过期“的版本

好，那我帮你把短链接服务的 Golang 版本升级为：

**✅ 支持 Redis 缓存**（加速短码 → 长链接查询、避免频繁访问 MySQL）
**✅ 访问统计**（访问次数、最后访问时间）
**✅ 自动过期**（短码到期自动失效）

我会分三步给你写：

---

## 1. 系统设计改动

* **Redis 结构**

  * Key: `shortlink:{code}` → Value: `originalURL`
  * Key 过期时间：`expires_at` - `created_at`（秒）
  * 访问时先查 Redis，没有再查 MySQL 并回填缓存
* **MySQL 表**

  * 依然使用之前的 `short_links` 表
  * `visit_count` 访问时自增
  * `last_visited_at` 更新为当前时间
* **逻辑流程**

  1. **生成短码** → 存 MySQL → 同步到 Redis（带 TTL）
  2. **访问短码** → 查 Redis → 查不到再查 MySQL
  3. **更新统计** → MySQL `visit_count++` + `last_visited_at`
  4. **过期控制** → MySQL 检查 `expires_at`，Redis TTL 自动控制缓存过期

---

## 2. 升级版 Golang 代码

```go
package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// ---------- 短码生成 ----------
func generateShortCode(length int) (string, error) {
	code := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(base62Chars))))
		if err != nil {
			return "", err
		}
		code[i] = base62Chars[num.Int64()]
	}
	return string(code), nil
}

func shortCodeExists(db *sql.DB, code string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM short_links WHERE short_code = ?)"
	err := db.QueryRow(query, code).Scan(&exists)
	return exists, err
}

func generateUniqueShortCode(db *sql.DB, length int) (string, error) {
	for {
		code, err := generateShortCode(length)
		if err != nil {
			return "", err
		}
		exists, err := shortCodeExists(db, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
}

// ---------- 短链接插入 ----------
func insertShortLink(db *sql.DB, redisClient *redis.Client, shortCode, originalURL string, creatorID *int64, expiresAt *time.Time) error {
	query := `
		INSERT INTO short_links (short_code, original_url, custom_code, creator_id, created_at, expires_at, status)
		VALUES (?, ?, 0, ?, NOW(), ?, 1)
	`
	_, err := db.Exec(query, shortCode, originalURL, creatorID, expiresAt)
	if err != nil {
		return err
	}

	// 写入 Redis 缓存
	ctx := context.Background()
	if expiresAt != nil {
		ttl := time.Until(*expiresAt)
		redisClient.Set(ctx, "shortlink:"+shortCode, originalURL, ttl)
	} else {
		redisClient.Set(ctx, "shortlink:"+shortCode, originalURL, 0)
	}

	return nil
}

// ---------- 访问短链接 ----------
func getOriginalURL(db *sql.DB, redisClient *redis.Client, shortCode string) (string, error) {
	ctx := context.Background()

	// 1. 先查 Redis
	val, err := redisClient.Get(ctx, "shortlink:"+shortCode).Result()
	if err == nil {
		// 更新访问统计
		updateVisitStats(db, shortCode)
		return val, nil
	}

	// 2. 查 MySQL
	var originalURL string
	var expiresAt *time.Time
	query := `
		SELECT original_url, expires_at
		FROM short_links
		WHERE short_code = ? AND status = 1
	`
	err = db.QueryRow(query, shortCode).Scan(&originalURL, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("短链接不存在或已失效")
		}
		return "", err
	}

	// 检查是否过期
	if expiresAt != nil && time.Now().After(*expiresAt) {
		return "", fmt.Errorf("短链接已过期")
	}

	// 回写 Redis 缓存
	if expiresAt != nil {
		ttl := time.Until(*expiresAt)
		redisClient.Set(ctx, "shortlink:"+shortCode, originalURL, ttl)
	} else {
		redisClient.Set(ctx, "shortlink:"+shortCode, originalURL, 0)
	}

	// 更新访问统计
	updateVisitStats(db, shortCode)

	return originalURL, nil
}

// ---------- 更新访问统计 ----------
func updateVisitStats(db *sql.DB, shortCode string) {
	query := `
		UPDATE short_links
		SET visit_count = visit_count + 1,
			last_visited_at = NOW()
		WHERE short_code = ?
	`
	_, err := db.Exec(query, shortCode)
	if err != nil {
		log.Printf("更新访问统计失败: %v", err)
	}
}

func main() {
	// 连接 MySQL
	dsn := "root:password@tcp(127.0.0.1:3306)/shortlink?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 连接 Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})

	// ========== 示例：生成短链接 ==========
	shortCode, err := generateUniqueShortCode(db, 6)
	if err != nil {
		log.Fatalf("短码生成失败: %v", err)
	}

	originalURL := "https://www.example.com/article?id=12345&utm_source=test"
	var creatorID *int64 = nil
	var expiresAt *time.Time
	t := time.Now().Add(24 * time.Hour) // 24 小时后过期
	expiresAt = &t

	err = insertShortLink(db, redisClient, shortCode, originalURL, creatorID, expiresAt)
	if err != nil {
		log.Fatalf("插入数据库失败: %v", err)
	}
	fmt.Printf("短链接生成成功: https://short.ly/%s -> %s\n", shortCode, originalURL)

	// ========== 示例：访问短链接 ==========
	url, err := getOriginalURL(db, redisClient, shortCode)
	if err != nil {
		log.Fatalf("访问失败: %v", err)
	}
	fmt.Printf("访问短码 %s 重定向到: %s\n", shortCode, url)
}
```

---

## 3. 运行步骤

1. **安装依赖**

   ```bash
   go get github.com/go-sql-driver/mysql
   go get github.com/redis/go-redis/v9
   ```

2. **启动 MySQL 和 Redis**

   ```bash
   redis-server
   ```

3. **建表**
   用之前的 `short_links` 表结构。

4. **运行**

   ```bash
   go run main.go
   ```

---

这样，你的短链接服务就已经具备了：

* **MySQL 持久化**
* **Redis 缓存**（加速查询）
* **访问次数统计**
* **自动过期功能**（Redis TTL + MySQL 校验）

---

如果你愿意，我还可以帮你加一个**HTTP 接口版本**，这样可以直接 `curl` 或浏览器访问短码实现跳转，相当于一个最小可用的短链接网站。
你要我帮你加这个 HTTP API 吗？这样就能直接部署成服务用。

--------------------------------------------------------------------------------
请添加”HTTP 接口”

好，我帮你在刚才的 **带 Redis + 访问统计 + 自动过期** 版本上，加上一个 **HTTP API 接口**，
这样你可以直接运行，就变成一个最小可用的短链接网站。

我会加两个 HTTP 接口：

1. `POST /create` → 生成短链接
2. `GET /{shortCode}` → 访问短链接并自动跳转

---

## 升级版：带 HTTP 接口的短链接服务

```go
package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var (
	db          *sql.DB
	redisClient *redis.Client
)

// ---------- 短码生成 ----------
func generateShortCode(length int) (string, error) {
	code := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(base62Chars))))
		if err != nil {
			return "", err
		}
		code[i] = base62Chars[num.Int64()]
	}
	return string(code), nil
}

func shortCodeExists(code string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM short_links WHERE short_code = ?)"
	err := db.QueryRow(query, code).Scan(&exists)
	return exists, err
}

func generateUniqueShortCode(length int) (string, error) {
	for {
		code, err := generateShortCode(length)
		if err != nil {
			return "", err
		}
		exists, err := shortCodeExists(code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
}

// ---------- 数据库操作 ----------
func insertShortLink(shortCode, originalURL string, expiresAt *time.Time) error {
	query := `
		INSERT INTO short_links (short_code, original_url, custom_code, created_at, expires_at, status)
		VALUES (?, ?, 0, NOW(), ?, 1)
	`
	_, err := db.Exec(query, shortCode, originalURL, expiresAt)
	if err != nil {
		return err
	}

	// 写 Redis
	ctx := context.Background()
	if expiresAt != nil {
		ttl := time.Until(*expiresAt)
		redisClient.Set(ctx, "shortlink:"+shortCode, originalURL, ttl)
	} else {
		redisClient.Set(ctx, "shortlink:"+shortCode, originalURL, 0)
	}
	return nil
}

func getOriginalURL(shortCode string) (string, error) {
	ctx := context.Background()

	// 1. 先查 Redis
	val, err := redisClient.Get(ctx, "shortlink:"+shortCode).Result()
	if err == nil {
		updateVisitStats(shortCode)
		return val, nil
	}

	// 2. 查 MySQL
	var originalURL string
	var expiresAt *time.Time
	query := `
		SELECT original_url, expires_at
		FROM short_links
		WHERE short_code = ? AND status = 1
	`
	err = db.QueryRow(query, shortCode).Scan(&originalURL, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("短链接不存在或已失效")
		}
		return "", err
	}

	// 检查过期
	if expiresAt != nil && time.Now().After(*expiresAt) {
		return "", fmt.Errorf("短链接已过期")
	}

	// 回写 Redis
	if expiresAt != nil {
		ttl := time.Until(*expiresAt)
		redisClient.Set(ctx, "shortlink:"+shortCode, originalURL, ttl)
	} else {
		redisClient.Set(ctx, "shortlink:"+shortCode, originalURL, 0)
	}

	updateVisitStats(shortCode)
	return originalURL, nil
}

func updateVisitStats(shortCode string) {
	query := `
		UPDATE short_links
		SET visit_count = visit_count + 1,
			last_visited_at = NOW()
		WHERE short_code = ?
	`
	_, err := db.Exec(query, shortCode)
	if err != nil {
		log.Printf("更新访问统计失败: %v", err)
	}
}

// ---------- HTTP 处理 ----------
type CreateRequest struct {
	URL       string `json:"url"`
	ExpiresIn int64  `json:"expires_in"` // 秒，为 0 表示永久
}

type CreateResponse struct {
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "请求格式错误", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL 不能为空", http.StatusBadRequest)
		return
	}

	// 生成短码
	shortCode, err := generateUniqueShortCode(6)
	if err != nil {
		http.Error(w, "生成短码失败", http.StatusInternalServerError)
		return
	}

	// 过期时间
	var expiresAt *time.Time
	if req.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(req.ExpiresIn) * time.Second)
		expiresAt = &t
	}

	// 插入
	if err := insertShortLink(shortCode, req.URL, expiresAt); err != nil {
		http.Error(w, "保存短链接失败", http.StatusInternalServerError)
		return
	}

	resp := CreateResponse{
		ShortCode: shortCode,
		ShortURL:  fmt.Sprintf("http://localhost:8080/%s", shortCode),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["code"]

	originalURL, err := getOriginalURL(shortCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

// ---------- 主程序 ----------
func main() {
	var err error

	// MySQL
	dsn := "root:password@tcp(127.0.0.1:3306)/shortlink?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("MySQL 连接失败: %v", err)
	}
	defer db.Close()

	// Redis
	redisClient = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})

	// 路由
	r := mux.NewRouter()
	r.HandleFunc("/create", createHandler).Methods("POST")
	r.HandleFunc("/{code}", redirectHandler).Methods("GET")

	fmt.Println("短链接服务已启动: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
```

---

## 使用方法

### 1. 安装依赖

```bash
go get github.com/go-sql-driver/mysql
go get github.com/redis/go-redis/v9
go get github.com/gorilla/mux
```

### 2. 启动 MySQL + Redis

确保表结构是我们前面设计的 `short_links` 表。

### 3. 启动服务

```bash
go run main.go
```

### 4. 创建短链接

```bash
curl -X POST http://localhost:8080/create \
    -H "Content-Type: application/json" \
    -d '{"url":"https://www.example.com/article?id=123","expires_in":3600}'
```

返回：

```json
{
  "short_code": "xY8aZ2",
  "short_url": "http://localhost:8080/xY8aZ2"
}
```

### 5. 访问短链接

```
http://localhost:8080/xY8aZ2
```

浏览器会自动跳转到原链接，同时数据库会记录访问次数和最后访问时间。

---

如果你愿意，我可以帮你在这个版本里加一个**管理接口**，支持：

* 查询某个短码的访问统计
* 列出所有短码
* 删除失效短码

这样你就可以做一个完整的短链接管理后台。
要我帮你加吗？这样就直接是一个可用的生产级系统了。

--------------------------------------------------------------------------------
添加“管理接口”

好，我帮你把 **短链接系统 + 管理接口** 整合起来，形成一个带：

1. **短码生成**
2. **MySQL 存储**
3. **Redis 缓存**
4. **访问统计**
5. **自动过期**
6. **HTTP 接口（跳转 / 创建）**
7. **管理接口（查看 / 删除 / 更新过期时间）**

的 Golang 版本。

---

## 数据库表

```sql
CREATE TABLE short_urls (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(10) NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    visit_count BIGINT DEFAULT 0,
    expire_at DATETIME DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## Golang 实现（管理接口）

```go
package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "math/rand"
    "net/http"
    "strings"
    "time"

    _ "github.com/go-sql-driver/mysql"
    "github.com/redis/go-redis/v9"
    "golang.org/x/net/context"
)

var (
    db    *sql.DB
    rdb   *redis.Client
    ctx   = context.Background()
    letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

// 短码生成
func generateCode(n int) string {
    rand.Seed(time.Now().UnixNano())
    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

// 创建短链接
func createShortLink(w http.ResponseWriter, r *http.Request) {
    type Req struct {
        URL      string `json:"url"`
        ExpireIn int64  `json:"expire_in"` // 秒
    }
    var req Req
    json.NewDecoder(r.Body).Decode(&req)

    code := generateCode(6)
    expireAt := time.Now().Add(time.Duration(req.ExpireIn) * time.Second)

    // MySQL 插入
    _, err := db.Exec("INSERT INTO short_urls (code, original_url, expire_at) VALUES (?, ?, ?)",
        code, req.URL, expireAt)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // Redis 缓存
    rdb.Set(ctx, "short:"+code, req.URL, time.Duration(req.ExpireIn)*time.Second)

    json.NewEncoder(w).Encode(map[string]string{
        "short_url": fmt.Sprintf("http://localhost:8080/%s", code),
    })
}

// 访问短链接
func redirectHandler(w http.ResponseWriter, r *http.Request) {
    code := strings.TrimPrefix(r.URL.Path, "/")

    // 从 Redis 读取
    url, err := rdb.Get(ctx, "short:"+code).Result()
    if err == redis.Nil {
        // Redis 没有 → MySQL 查找
        var original string
        var expireAt sql.NullTime
        err := db.QueryRow("SELECT original_url, expire_at FROM short_urls WHERE code = ?", code).
            Scan(&original, &expireAt)
        if err != nil {
            http.NotFound(w, r)
            return
        }
        // 检查是否过期
        if expireAt.Valid && expireAt.Time.Before(time.Now()) {
            http.Error(w, "link expired", 410)
            return
        }
        // 回写 Redis
        ttl := time.Until(expireAt.Time)
        rdb.Set(ctx, "short:"+code, original, ttl)
        url = original
    } else if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // 增加访问统计
    db.Exec("UPDATE short_urls SET visit_count = visit_count + 1 WHERE code = ?", code)

    http.Redirect(w, r, url, http.StatusFound)
}

// 管理接口：查看详情
func adminGetLink(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Query().Get("code")
    var original string
    var visitCount int64
    var expireAt sql.NullTime
    err := db.QueryRow("SELECT original_url, visit_count, expire_at FROM short_urls WHERE code = ?", code).
        Scan(&original, &visitCount, &expireAt)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    json.NewEncoder(w).Encode(map[string]interface{}{
        "code":        code,
        "original":    original,
        "visit_count": visitCount,
        "expire_at":   expireAt.Time,
    })
}

// 管理接口：删除
func adminDeleteLink(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Query().Get("code")
    db.Exec("DELETE FROM short_urls WHERE code = ?", code)
    rdb.Del(ctx, "short:"+code)
    w.Write([]byte("deleted"))
}

// 管理接口：更新过期时间
func adminUpdateExpire(w http.ResponseWriter, r *http.Request) {
    type Req struct {
        Code     string `json:"code"`
        ExpireIn int64  `json:"expire_in"`
    }
    var req Req
    json.NewDecoder(r.Body).Decode(&req)

    expireAt := time.Now().Add(time.Duration(req.ExpireIn) * time.Second)
    db.Exec("UPDATE short_urls SET expire_at = ? WHERE code = ?", expireAt, req.Code)
    // 更新 Redis TTL
    url, _ := rdb.Get(ctx, "short:"+req.Code).Result()
    if url != "" {
        rdb.Set(ctx, "short:"+req.Code, url, time.Duration(req.ExpireIn)*time.Second)
    }
    w.Write([]byte("expire updated"))
}

func main() {
    var err error
    db, err = sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/shortdb?charset=utf8mb4&parseTime=True")
    if err != nil {
        log.Fatal(err)
    }
    rdb = redis.NewClient(&redis.Options{
        Addr: "127.0.0.1:6379",
    })

    // 用户接口
    http.HandleFunc("/", redirectHandler)
    http.HandleFunc("/create", createShortLink)

    // 管理接口
    http.HandleFunc("/admin/get", adminGetLink)
    http.HandleFunc("/admin/delete", adminDeleteLink)
    http.HandleFunc("/admin/update_expire", adminUpdateExpire)

    fmt.Println("Server running on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

---

### 接口示例

#### 1. 创建短链接

```bash
curl -X POST http://localhost:8080/create \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","expire_in":3600}'
```

#### 2. 管理接口 - 查看

```bash
curl "http://localhost:8080/admin/get?code=abc123"
```

#### 3. 管理接口 - 删除

```bash
curl "http://localhost:8080/admin/delete?code=abc123"
```

#### 4. 管理接口 - 更新过期时间

```bash
curl -X POST http://localhost:8080/admin/update_expire \
  -H "Content-Type: application/json" \
  -d '{"code":"abc123","expire_in":7200}'
```

---

如果你需要，我还可以帮你把这个版本改造成 **支持 JWT 管理员认证**，避免随便就能访问管理接口。这样就可以变成一个安全的 SaaS 短链接平台。
你要我帮你加上这个吗？

--------------------------------------------------------------------------------
添加“支持 JWT 管理员认证”

好——我把 JWT 管理员认证**完整地**加到你的短链接服务里：

* 支持管理员通过 `/admin/login`（用户名+密码）获取 JWT；
* 所有 `/admin/*` 管理接口都被 JWT 中间件保护；
* 在启动时若数据库中没有管理员账户，可用环境变量自动创建管理员（方便首次部署）；
* 使用 `bcrypt` 存储密码、`HS256` 签发/验证 JWT；
* 附带创建管理员、登录、使用 token 调用管理接口的示例。

下面给出完整的 Go 程序（基于你已有的 MySQL + Redis + 短链接逻辑）——把它保存为 `main.go`，按说明安装依赖并运行即可。

---

### 1) 依赖

```bash
go get github.com/go-sql-driver/mysql
go get github.com/redis/go-redis/v9
go get github.com/gorilla/mux
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt
```

---

### 2) 建表 SQL

（如果你用之前的表名不同，可调整代码或 SQL）

```sql
-- 短链接表（与你之前的略有差别字段名要一致）
CREATE TABLE IF NOT EXISTS short_links (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    short_code VARCHAR(64) NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    visit_count BIGINT DEFAULT 0,
    last_visited_at DATETIME NULL,
    expires_at DATETIME NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status TINYINT DEFAULT 1
);

-- 管理员表（保存 bcrypt 哈希）
CREATE TABLE IF NOT EXISTS admins (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(128) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

### 3) 完整 `main.go`

```go
package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var (
	db          *sql.DB
	redisClient *redis.Client
	ctx         = context.Background()
	jwtSecret   []byte
)

// -------------------- helpers --------------------
func generateShortCode(length int) (string, error) {
	code := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(base62Chars))))
		if err != nil {
			return "", err
		}
		code[i] = base62Chars[num.Int64()]
	}
	return string(code), nil
}

// -------------------- DB helpers --------------------
func shortCodeExists(code string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM short_links WHERE short_code = ?)", code).Scan(&exists)
	return exists, err
}

func insertShortLink(shortCode, originalURL string, expiresAt *time.Time) error {
	_, err := db.Exec(`
		INSERT INTO short_links (short_code, original_url, created_at, expires_at, status)
		VALUES (?, ?, NOW(), ?, 1)
	`, shortCode, originalURL, expiresAt)
	if err != nil {
		return err
	}
	// cache
	key := "shortlink:" + shortCode
	if expiresAt != nil {
		redisClient.Set(ctx, key, originalURL, time.Until(*expiresAt))
	} else {
		redisClient.Set(ctx, key, originalURL, 0)
	}
	return nil
}

func getOriginalURL(shortCode string) (string, error) {
	// try redis
	key := "shortlink:" + shortCode
	val, err := redisClient.Get(ctx, key).Result()
	if err == nil {
		// update stats async
		go updateVisitStats(shortCode)
		return val, nil
	}

	// redis miss -> mysql
	var originalURL string
	var expiresAt sql.NullTime
	err = db.QueryRow("SELECT original_url, expires_at FROM short_links WHERE short_code = ? AND status = 1", shortCode).
		Scan(&originalURL, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("短链接不存在")
		}
		return "", err
	}

	if expiresAt.Valid && time.Now().After(expiresAt.Time) {
		return "", errors.New("短链接已过期")
	}

	// backfill redis
	if expiresAt.Valid {
		redisClient.Set(ctx, key, originalURL, time.Until(expiresAt.Time))
	} else {
		redisClient.Set(ctx, key, originalURL, 0)
	}

	go updateVisitStats(shortCode)
	return originalURL, nil
}

func updateVisitStats(shortCode string) {
	_, err := db.Exec("UPDATE short_links SET visit_count = visit_count + 1, last_visited_at = NOW() WHERE short_code = ?", shortCode)
	if err != nil {
		log.Printf("updateVisitStats error: %v", err)
	}
}

// -------------------- Admin (bcrypt + jwt) --------------------

func ensureAdminFromEnv() error {
	// If env ADMIN_USERNAME and ADMIN_PASSWORD set, create admin if not exists.
	username := os.Getenv("ADMIN_USERNAME")
	password := os.Getenv("ADMIN_PASSWORD")
	if username == "" || password == "" {
		return nil // nothing to ensure
	}

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM admins WHERE username = ?)", username).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	// create
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO admins (username, password_hash) VALUES (?, ?)", username, string(hash))
	return err
}

func authenticateAdmin(username, password string) (bool, error) {
	var hash string
	err := db.QueryRow("SELECT password_hash FROM admins WHERE username = ?", username).Scan(&hash)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return false, nil
	}
	return true, nil
}

func issueJWT(username string, ttl time.Duration) (string, error) {
	// claims
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   username,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func parseAndValidateJWT(tokenString string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		// verify alg
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// -------------------- HTTP handlers --------------------

// LoginRequest {username, password}
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse {token, expires_in_seconds}
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expires_in"`
}

// POST /admin/login
func adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ok, err := authenticateAdmin(req.Username, req.Password)
	if err != nil {
		log.Printf("auth error: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	ttl := 24 * time.Hour
	token, err := issueJWT(req.Username, ttl)
	if err != nil {
		log.Printf("issue token err: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	resp := LoginResponse{Token: token, ExpiresIn: int64(ttl.Seconds())}
	json.NewEncoder(w).Encode(resp)
}

// Auth middleware for admin endpoints
func adminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ah := r.Header.Get("Authorization")
		if ah == "" {
			http.Error(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}
		parts := strings.SplitN(ah, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "invalid Authorization header", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]
		claims, err := parseAndValidateJWT(tokenString)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		// optional: can add claims.Subject into context for handler use
		ctx := context.WithValue(r.Context(), "admin_username", claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Example admin handler: GET /admin/get?code=xxx
func adminGetLinkHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}
	var original string
	var visitCount int64
	var expiresAt sql.NullTime
	err := db.QueryRow("SELECT original_url, visit_count, expires_at FROM short_links WHERE short_code = ?", code).
		Scan(&original, &visitCount, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	resp := map[string]interface{}{
		"code":        code,
		"original":    original,
		"visit_count": visitCount,
		"expires_at":  nil,
	}
	if expiresAt.Valid {
		resp["expires_at"] = expiresAt.Time
	}
	json.NewEncoder(w).Encode(resp)
}

// Admin delete: DELETE /admin/delete?code=xxx
func adminDeleteHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}
	_, err := db.Exec("DELETE FROM short_links WHERE short_code = ?", code)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	redisClient.Del(ctx, "shortlink:"+code)
	w.Write([]byte("deleted"))
}

// Admin update expire: POST /admin/update_expire  body: {"code":"xxx","expire_in":3600}
type UpdateExpireReq struct {
	Code     string `json:"code"`
	ExpireIn int64  `json:"expire_in"`
}
func adminUpdateExpireHandler(w http.ResponseWriter, r *http.Request) {
	var req UpdateExpireReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}
	var expireAt *time.Time
	if req.ExpireIn > 0 {
		t := time.Now().Add(time.Duration(req.ExpireIn) * time.Second)
		expireAt = &t
	}
	_, err := db.Exec("UPDATE short_links SET expires_at = ? WHERE short_code = ?", expireAt, req.Code)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	// update redis ttl if present
	key := "shortlink:" + req.Code
	orig, err := redisClient.Get(ctx, key).Result()
	if err == nil {
		if expireAt != nil {
			redisClient.Set(ctx, key, orig, time.Until(*expireAt))
		} else {
			redisClient.Set(ctx, key, orig, 0)
		}
	}
	w.Write([]byte("ok"))
}

// Public endpoints: create + redirect
type CreateReq struct {
	URL       string `json:"url"`
	ExpiresIn int64  `json:"expires_in"` // seconds
}

type CreateResp struct {
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.URL == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}
	// generate unique code
	var code string
	for {
		c, err := generateShortCode(6)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		exists, err := shortCodeExists(c)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		if !exists {
			code = c
			break
		}
	}
	var expiresAt *time.Time
	if req.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(req.ExpiresIn) * time.Second)
		expiresAt = &t
	}
	if err := insertShortLink(code, req.URL, expiresAt); err != nil {
		http.Error(w, "save error", http.StatusInternalServerError)
		return
	}
	resp := CreateResp{
		ShortCode: code,
		ShortURL:  fmt.Sprintf("http://%s/%s", r.Host, code),
	}
	json.NewEncoder(w).Encode(resp)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code := vars["code"]
	orig, err := getOriginalURL(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	http.Redirect(w, r, orig, http.StatusFound)
}

// -------------------- main --------------------
func main() {
	var err error
	// read envs
	dsn := os.Getenv("DSN")
	if dsn == "" {
		// default: user:pass@tcp(127.0.0.1:3306)/shortlink?charset=utf8mb4&parseTime=True&loc=Local
		dsn = "root:password@tcp(127.0.0.1:3306)/shortlink?charset=utf8mb4&parseTime=True&loc=Local"
	}
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("mysql open: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("mysql ping: %v", err)
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379"
	}
	redisClient = redis.NewClient(&redis.Options{Addr: redisAddr})
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatalf("redis ping: %v", err)
	}

	// JWT secret
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// fallback (development only)
		secret = "please_change_this_secret"
		log.Printf("WARNING: using default JWT secret; set JWT_SECRET in production")
	}
	jwtSecret = []byte(secret)

	// ensure admin from env (optional)
	if err := ensureAdminFromEnv(); err != nil {
		log.Fatalf("ensureAdminFromEnv err: %v", err)
	}

	// router
	r := mux.NewRouter()
	// public
	r.HandleFunc("/create", createHandler).Methods("POST")
	r.HandleFunc("/{code}", redirectHandler).Methods("GET")

	// admin auth endpoints
	r.HandleFunc("/admin/login", adminLoginHandler).Methods("POST")

	// admin protected endpoints - wrap with middleware
	adminRouter := r.PathPrefix("/admin").Subrouter()
	adminRouter.Use(func(next http.Handler) http.Handler { return adminAuthMiddleware(next) })
	adminRouter.HandleFunc("/get", adminGetLinkHandler).Methods("GET")
	adminRouter.HandleFunc("/delete", adminDeleteHandler).Methods("DELETE")
	adminRouter.HandleFunc("/update_expire", adminUpdateExpireHandler).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
```

---

### 4) 如何使用（示例）

1. 启动时可通过环境变量自动创建管理员（可选）：

```bash
export ADMIN_USERNAME=admin
export ADMIN_PASSWORD=Secret123
export JWT_SECRET="a-very-secret-key"
export DSN="root:password@tcp(127.0.0.1:3306)/shortlink?charset=utf8mb4&parseTime=True&loc=Local"
go run main.go
```

> 程序启动后，如果 `admins` 表里没有 `admin` 这个用户名，会自动插入（密码为 `ADMIN_PASSWORD` 的 bcrypt 哈希）。

2. 登录获取 JWT：

```bash
curl -s -X POST http://localhost:8080/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Secret123"}' | jq
```

返回类似：

```json
{
  "token": "eyJhbGciOiJIU... (JWT)",
  "expires_in": 86400
}
```

3. 使用 JWT 调用管理接口（示例：查看短链信息）：

```bash
TOKEN="eyJhbGciOiJI..."
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/admin/get?code=abc123"
```

4. 创建短链接（公开接口）：

```bash
curl -X POST http://localhost:8080/create \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","expires_in":3600}'
```

---

### 5) 安全与部署建议（简短）

* **强烈**在生产设置 `JWT_SECRET` 为强随机字符串并保存在安全位置（例如 Vault / env via secrets manager）。
* 把 `ADMIN_PASSWORD` 设置成强密码或使用外部身份提供者（OAuth, LDAP）。
* 使用 HTTPS（反代 Nginx / Traefik）以保护 token 在传输中的安全。
* 考虑 token 黑名单（登出/撤销）或缩短 token 有效期并使用刷新 token 方案。
* 在高并发场景，把更新访问统计改为批量/异步写入以减轻 DB 压力；或使用 Redis INCR 计数周期性回写。

---

如果你愿意，我可以：

* 帮你把 `admin` 登录改为带 **refresh token** 的更完整方案，或
* 在管理接口上加上分页 / 列表 / 搜索，或
* 把 Admin UI（简单网页）做成一个小前端（React / plain HTML）以便管理。

你想先做哪一个？

--------------------------------------------------------------------------------
把 admin 登录改为带 refresh token 的更完整方案

好的，我会帮你把当前的管理接口登录改成一个**带 Refresh Token 的完整 JWT 认证方案**，这样可以更安全地管理会话，并且避免频繁要求管理员重新登录。

下面我会给你一个 **Go + Redis + JWT + Refresh Token** 的示例，包含以下功能：

1. **登录接口**（生成 access token + refresh token）
2. **刷新 access token 接口**（用 refresh token 换新 access token）
3. **退出登录接口**（使 refresh token 失效）
4. **带认证的管理接口**（需要有效 access token 才能访问）

---

## 核心思路

* **Access Token**：短时有效（如 15 分钟），放在 HTTP 请求头中访问管理接口。
* **Refresh Token**：长时有效（如 7 天），仅用于刷新 Access Token，不直接访问接口。
* **Redis** 用来存储 refresh token（保证可控的过期和失效）。
* **JWT** 用于生成 access token（存储用户信息，不需要持久化）。

---

## 完整示例代码（Go）

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

var jwtSecret = []byte("super_secret_key")
var adminUser = "admin"
var adminPass = "123456"
var redisClient *redis.Client

const (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func main() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	r := mux.NewRouter()

	// 登录 & token 管理
	r.HandleFunc("/admin/login", AdminLogin).Methods("POST")
	r.HandleFunc("/admin/refresh", RefreshToken).Methods("POST")
	r.HandleFunc("/admin/logout", AdminLogout).Methods("POST")

	// 需要 Access Token 认证的管理接口
	r.Handle("/admin/stats", AuthMiddleware(http.HandlerFunc(AdminStats))).Methods("GET")

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// AdminLogin 登录生成 Access & Refresh Token
func AdminLogin(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	pass := r.URL.Query().Get("pass")

	if user != adminUser || pass != adminPass {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	accessToken, err := generateAccessToken(user)
	if err != nil {
		http.Error(w, "token error", http.StatusInternalServerError)
		return
	}

	refreshToken := fmt.Sprintf("refresh_%d", time.Now().UnixNano())
	ctx := context.Background()
	redisClient.Set(ctx, "refresh:"+refreshToken, user, RefreshTokenTTL)

	fmt.Fprintf(w, `{"access_token":"%s","refresh_token":"%s"}`, accessToken, refreshToken)
}

// RefreshToken 用 Refresh Token 换新的 Access Token
func RefreshToken(w http.ResponseWriter, r *http.Request) {
	refresh := r.URL.Query().Get("refresh_token")
	ctx := context.Background()

	user, err := redisClient.Get(ctx, "refresh:"+refresh).Result()
	if err != nil {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}

	newAccessToken, err := generateAccessToken(user)
	if err != nil {
		http.Error(w, "token error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, `{"access_token":"%s"}`, newAccessToken)
}

// AdminLogout 注销（删除 Refresh Token）
func AdminLogout(w http.ResponseWriter, r *http.Request) {
	refresh := r.URL.Query().Get("refresh_token")
	ctx := context.Background()
	redisClient.Del(ctx, "refresh:"+refresh)
	fmt.Fprintln(w, "logout ok")
}

// AdminStats 管理员接口（示例）
func AdminStats(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "short link stats: {...}")
}

// AuthMiddleware 验证 Access Token
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// 生成 Access Token
func generateAccessToken(username string) (string, error) {
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenTTL)),
			Issuer:    "shortlink-admin",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
```

---

## 使用方式示例

### 登录获取 token

```
POST /admin/login?user=admin&pass=123456
返回：
{
  "access_token": "xxxxx",
  "refresh_token": "refresh_169167..."
}
```

### 刷新 Access Token

```
POST /admin/refresh?refresh_token=refresh_169167...
返回：
{
  "access_token": "new_xxxxx"
}
```

### 管理接口访问

```
GET /admin/stats
Header: Authorization: <access_token>
```

### 登出

```
POST /admin/logout?refresh_token=refresh_169167...
```

---

如果你愿意，我可以帮你**把这个 JWT + Refresh Token 登录系统直接整合进之前的“短链接系统”代码**，这样整个项目就会变成一个支持短链接、Redis 缓存、访问统计、自动过期、带管理员后台的成品系统。

你是想让我帮你做这个整合版吗？这样就可以直接跑了。

--------------------------------------------------------------------------------
做这个整合版

太好了 —— 我把之前的短链接服务与 **带 Refresh Token 的完整 JWT 管理认证** 做了一个整合。
下面的 `main.go` 包含如下功能（开箱即用）：

* 短链接：创建、重定向（MySQL 持久化 + Redis 缓存 + 自动过期）
* 访问统计：`visit_count`、`last_visited_at`
* Admin 用户：`admins` 表（bcrypt 存密码）
* Admin 登录：返回 **Access Token (短期)** + **Refresh Token (长期)**，Refresh Token 存 Redis 并有 TTL
* Refresh 流程：用 refresh token 换取新的 access token，并**轮换**（原 refresh 失效，发新 token）
* Logout：删除 refresh token（使其失效）
* Admin 接口受 Access Token 保护（`Authorization: Bearer <access_token>`）
* 启动时可通过 `ADMIN_USERNAME`/`ADMIN_PASSWORD` 环境变量自动创建管理员（方便首次部署）

---

## 使用说明（先读这段再运行）

1. 需要 MySQL 与 Redis；执行建表 SQL（见代码注释）。
2. 设置环境变量（示例）：

   ```bash
   export DSN="root:password@tcp(127.0.0.1:3306)/shortlink?charset=utf8mb4&parseTime=True&loc=Local"
   export REDIS_ADDR="127.0.0.1:6379"
   export JWT_SECRET="replace_with_strong_secret"
   export ADMIN_USERNAME="admin"
   export ADMIN_PASSWORD="Secret123"   # 可选：首次启动会自动创建 admin 帐号
   export PORT=8080
   ```
3. 安装依赖：

   ```bash
   go get github.com/go-sql-driver/mysql
   go get github.com/redis/go-redis/v9
   go get github.com/gorilla/mux
   go get github.com/golang-jwt/jwt/v5
   go get golang.org/x/crypto/bcrypt
   ```
4. 运行：

   ```bash
   go run main.go
   ```
5. 常用 API：

   * `POST /create` — 创建短链接（公开）
   * `GET /{code}` — 重定向到原始 URL（自动统计）
   * `POST /admin/login` — 管理员登录，返回 `{access_token, refresh_token, expires_in}`
   * `POST /admin/refresh` — 用 refresh token 换新 access & refresh（返回新两者）
   * `POST /admin/logout` — 注销（删除 refresh token）
   * 受保护的 admin 接口举例：`GET /admin/get?code=xxx`、`DELETE /admin/delete?code=xxx`、`POST /admin/update_expire`

---

## 建表 SQL（MySQL）

```sql
CREATE TABLE IF NOT EXISTS short_links (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  short_code VARCHAR(64) NOT NULL UNIQUE,
  original_url TEXT NOT NULL,
  visit_count BIGINT DEFAULT 0,
  last_visited_at DATETIME NULL,
  expires_at DATETIME NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  status TINYINT DEFAULT 1
);

CREATE TABLE IF NOT EXISTS admins (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(128) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## 完整 `main.go`

```go
package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var (
	db          *sql.DB
	redisClient *redis.Client
	ctx         = context.Background()
	jwtSecret   []byte

	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
)

// ---------- util ----------
func randBase62(n int) (string, error) {
	buf := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(base62Chars))))
		if err != nil {
			return "", err
		}
		buf[i] = base62Chars[num.Int64()]
	}
	return string(buf), nil
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ---------- short link logic ----------
func insertShortLink(shortCode, originalURL string, expiresAt *time.Time) error {
	_, err := db.Exec(`
		INSERT INTO short_links (short_code, original_url, created_at, expires_at, status)
		VALUES (?, ?, NOW(), ?, 1)
	`, shortCode, originalURL, expiresAt)
	if err != nil {
		return err
	}
	// write to redis
	key := "shortlink:" + shortCode
	if expiresAt != nil {
		redisClient.Set(ctx, key, originalURL, time.Until(*expiresAt))
	} else {
		redisClient.Set(ctx, key, originalURL, 0)
	}
	return nil
}

func shortCodeExists(code string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM short_links WHERE short_code=?)", code).Scan(&exists)
	return exists, err
}

func getOriginalURL(shortCode string) (string, error) {
	key := "shortlink:" + shortCode
	val, err := redisClient.Get(ctx, key).Result()
	if err == nil {
		// async stats update
		go updateVisitStats(shortCode)
		return val, nil
	}
	if err != nil && err != redis.Nil {
		// continue to mysql even if redis error
		log.Printf("redis get err: %v", err)
	}

	var original string
	var expiresAt sql.NullTime
	err = db.QueryRow("SELECT original_url, expires_at FROM short_links WHERE short_code=? AND status=1", shortCode).
		Scan(&original, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("短链接不存在")
		}
		return "", err
	}
	if expiresAt.Valid && time.Now().After(expiresAt.Time) {
		return "", errors.New("短链接已过期")
	}
	// backfill
	if expiresAt.Valid {
		redisClient.Set(ctx, key, original, time.Until(expiresAt.Time))
	} else {
		redisClient.Set(ctx, key, original, 0)
	}
	go updateVisitStats(shortCode)
	return original, nil
}

func updateVisitStats(shortCode string) {
	_, err := db.Exec("UPDATE short_links SET visit_count = visit_count + 1, last_visited_at = NOW() WHERE short_code = ?", shortCode)
	if err != nil {
		log.Printf("updateVisitStats err: %v", err)
	}
}

// ---------- admin auth (bcrypt + jwt + refresh in redis) ----------
func ensureAdminFromEnv() error {
	username := os.Getenv("ADMIN_USERNAME")
	password := os.Getenv("ADMIN_PASSWORD")
	if username == "" || password == "" {
		return nil
	}
	var exists bool
	if err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM admins WHERE username=?)", username).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO admins (username, password_hash) VALUES (?, ?)", username, string(hash))
	return err
}

func authenticateAdmin(username, password string) (bool, error) {
	var hash string
	err := db.QueryRow("SELECT password_hash FROM admins WHERE username=?", username).Scan(&hash)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return false, nil
	}
	return true, nil
}

func issueAccessToken(username string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   username,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenTTL)),
		Issuer:    "shortlink-admin",
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(jwtSecret)
}

// create refresh token and store in redis with TTL
func storeRefreshToken(username string) (string, error) {
	tok, err := generateRefreshToken()
	if err != nil {
		return "", err
	}
	key := "refresh:" + tok
	if err := redisClient.Set(ctx, key, username, RefreshTokenTTL).Err(); err != nil {
		return "", err
	}
	return tok, nil
}

// rotate: delete old refresh token and create new one
func rotateRefreshToken(old string) (string, error) {
	keyOld := "refresh:" + old
	username, err := redisClient.Get(ctx, keyOld).Result()
	if err != nil {
		return "", err
	}
	// delete old
	redisClient.Del(ctx, keyOld)
	// store new
	return storeRefreshToken(username)
}

// validate refresh token exists and return username
func validateRefreshToken(tok string) (string, error) {
	username, err := redisClient.Get(ctx, "refresh:"+tok).Result()
	if err != nil {
		return "", err
	}
	return username, nil
}

func revokeRefreshToken(tok string) error {
	_, err := redisClient.Del(ctx, "refresh:"+tok).Result()
	return err
}

// ---------- HTTP handlers ----------

// Request/Response structs
type CreateReq struct {
	URL       string `json:"url"`
	ExpiresIn int64  `json:"expires_in"` // seconds, 0 = never
}
type CreateResp struct {
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
}

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type LoginResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // access token ttl in seconds
}

// public: create short link
func createHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.URL == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}
	// generate unique code
	var code string
	for {
		c, err := randBase62(6)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		exists, err := shortCodeExists(c)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		if !exists {
			code = c
			break
		}
	}
	var expiresAt *time.Time
	if req.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(req.ExpiresIn) * time.Second)
		expiresAt = &t
	}
	if err := insertShortLink(code, req.URL, expiresAt); err != nil {
		http.Error(w, "save error", http.StatusInternalServerError)
		return
	}
	resp := CreateResp{
		ShortCode: code,
		ShortURL:  fmt.Sprintf("http://%s/%s", r.Host, code),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// public: redirect
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code := vars["code"]
	orig, err := getOriginalURL(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	http.Redirect(w, r, orig, http.StatusFound)
}

// admin: login -> issue access + refresh
func adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ok, err := authenticateAdmin(req.Username, req.Password)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	access, err := issueAccessToken(req.Username)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	refresh, err := storeRefreshToken(req.Username)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	resp := LoginResp{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(AccessTokenTTL.Seconds()),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// admin: refresh -> rotate refresh token + new access token
type RefreshReq struct {
	RefreshToken string `json:"refresh_token"`
}
type RefreshResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}
func adminRefreshHandler(w http.ResponseWriter, r *http.Request) {
	var req RefreshReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	username, err := validateRefreshToken(req.RefreshToken)
	if err != nil {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}
	// rotate
	newRefresh, err := rotateRefreshToken(req.RefreshToken)
	if err != nil {
		http.Error(w, "failed rotate refresh", http.StatusInternalServerError)
		return
	}
	access, err := issueAccessToken(username)
	if err != nil {
		http.Error(w, "failed create access", http.StatusInternalServerError)
		return
	}
	resp := RefreshResp{
		AccessToken:  access,
		RefreshToken: newRefresh,
		ExpiresIn:    int64(AccessTokenTTL.Seconds()),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// admin: logout -> revoke refresh token
type LogoutReq struct {
	RefreshToken string `json:"refresh_token"`
}
func adminLogoutHandler(w http.ResponseWriter, r *http.Request) {
	var req LogoutReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := revokeRefreshToken(req.RefreshToken); err != nil {
		// still return ok to avoid leaking info
		log.Printf("revoke err: %v", err)
	}
	w.Write([]byte("ok"))
}

// middleware: validate access token
func adminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ah := r.Header.Get("Authorization")
		if ah == "" {
			http.Error(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}
		parts := strings.SplitN(ah, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "invalid Authorization header", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]
		claims := &jwt.RegisteredClaims{}
		_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			// enforce HS256
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return jwtSecret, nil
		})
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		// attach admin username
		ctx := context.WithValue(r.Context(), "admin_username", claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// admin: protected endpoints
func adminGetLinkHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}
	var original string
	var visitCount int64
	var expiresAt sql.NullTime
	err := db.QueryRow("SELECT original_url, visit_count, expires_at FROM short_links WHERE short_code = ?", code).
		Scan(&original, &visitCount, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	resp := map[string]interface{}{
		"code":        code,
		"original":    original,
		"visit_count": visitCount,
		"expires_at":  nil,
	}
	if expiresAt.Valid {
		resp["expires_at"] = expiresAt.Time
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func adminDeleteHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}
	if _, err := db.Exec("DELETE FROM short_links WHERE short_code = ?", code); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	redisClient.Del(ctx, "shortlink:"+code)
	w.Write([]byte("deleted"))
}

type UpdateExpireReq struct {
	Code     string `json:"code"`
	ExpireIn int64  `json:"expire_in"`
}
func adminUpdateExpireHandler(w http.ResponseWriter, r *http.Request) {
	var req UpdateExpireReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	var expiresAt *time.Time
	if req.ExpireIn > 0 {
		t := time.Now().Add(time.Duration(req.ExpireIn) * time.Second)
		expiresAt = &t
	}
	if _, err := db.Exec("UPDATE short_links SET expires_at = ? WHERE short_code = ?", expiresAt, req.Code); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	// update redis TTL if present
	key := "shortlink:" + req.Code
	orig, err := redisClient.Get(ctx, key).Result()
	if err == nil {
		if expiresAt != nil {
			redisClient.Set(ctx, key, orig, time.Until(*expiresAt))
		} else {
			redisClient.Set(ctx, key, orig, 0)
		}
	}
	w.Write([]byte("ok"))
}

// ---------- main ----------
func main() {
	// envs
	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = "root:password@tcp(127.0.0.1:3306)/shortlink?charset=utf8mb4&parseTime=True&loc=Local"
	}
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379"
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Println("WARNING: JWT_SECRET not set. Using insecure default.")
		secret = "please_change_me"
	}
	jwtSecret = []byte(secret)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("mysql open err: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("mysql ping err: %v", err)
	}

	redisClient = redis.NewClient(&redis.Options{Addr: redisAddr})
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatalf("redis ping err: %v", err)
	}

	// optional auto-create admin
	if err := ensureAdminFromEnv(); err != nil {
		log.Fatalf("ensure admin err: %v", err)
	}

	r := mux.NewRouter()

	// public
	r.HandleFunc("/create", createHandler).Methods("POST")
	r.HandleFunc("/{code}", redirectHandler).Methods("GET")

	// auth endpoints (admin)
	r.HandleFunc("/admin/login", adminLoginHandler).Methods("POST")
	r.HandleFunc("/admin/refresh", adminRefreshHandler).Methods("POST")
	r.HandleFunc("/admin/logout", adminLogoutHandler).Methods("POST")

	// protected admin endpoints
	admin := r.PathPrefix("/admin").Subrouter()
	admin.Use(adminAuthMiddleware)
	admin.HandleFunc("/get", adminGetLinkHandler).Methods("GET")
	admin.HandleFunc("/delete", adminDeleteHandler).Methods("DELETE")
	admin.HandleFunc("/update_expire", adminUpdateExpireHandler).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
```

---

## 示例：登录 / 刷新 / 登出 使用

### 登录（得到 access+refresh）

```bash
curl -s -X POST http://localhost:8080/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Secret123"}'
# 返回 JSON: { "access_token":"...", "refresh_token":"...", "expires_in":900 }
```

### 使用 access 访问管理接口

```bash
curl -H "Authorization: Bearer <access_token>" "http://localhost:8080/admin/get?code=abc123"
```

### 刷新（轮换 refresh token）

```bash
curl -s -X POST http://localhost:8080/admin/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'
# 返回新的 access_token 和新的 refresh_token（旧的 refresh 被删除）
```

### 登出（撤销 refresh token）

```bash
curl -s -X POST http://localhost:8080/admin/logout \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'
```

---

## 安全与改进建议（简短）

* 在生产 **必须** 使用强 `JWT_SECRET`，并通过 HTTPS 暴露服务。
* 可扩展：实现 refresh token 黑名单（若不做轮换），或在 Redis 中记录 refresh token 的元数据（IP、UA）以防盗用。
* 可增加短链管理的分页/搜索接口、和 Admin 前端（React/Vue）配合使用。
* 高并发场景：将访问统计从同步 `UPDATE` 改为 Redis `INCR` 聚合、周期性回写 MySQL。

---

我已把整合版的完整代码给你了。要不要我现在帮你：

1. 把访问统计升级为 Redis `INCR` + 周期回写（高并发更稳），或
2. 加一个简单的 Admin 前端页面（单文件 HTML）用于登录 / 列表 / 删除？

你选一个，我马上继续写。

--------------------------------------------------------------------------------

--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
