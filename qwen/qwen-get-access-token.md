# 我用golang语言调用某网站的api，首先要请求token，网站返回如下信息：
{
  "access_token": "IwEBIL1Cuc5QOpIKO8uOkoFXBQGK",
  "refresh_token": "RedoNxDMbqgnMnbWQUaidOjmKQV",
  "token_type": "bearer",
  "expires_in": 3600
}
以后每次调用api，都要带上这个token，这个token的有效期是3600秒，请帮我编写检查token是否过期，若过期，就申请新token的过程


你可以使用 Go 语言编写一个结构体来管理 Token 的获取和刷新逻辑。下面是一个完整的示例，包含：

- 存储 token 及其过期时间
- 检查 token 是否即将过期（比如提前 60 秒刷新）
- 自动请求新 token
- 使用 HTTP 客户端调用 API 时自动携带 token

---

### ✅ 示例代码：Token 管理器

```go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// TokenResponse 是从认证接口返回的 JSON 结构
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // 过期时间（秒）
}

// TokenManager 负责管理 token 的获取和刷新
type TokenManager struct {
	mu          sync.RWMutex
	token       string
	refreshToken string
	expiresAt   time.Time // token 的过期时间点
	client      *http.Client
	tokenURL    string
}

// NewTokenManager 创建一个新的 TokenManager
func NewTokenManager(tokenURL string) *TokenManager {
	return &TokenManager{
		client:     &http.Client{Timeout: 10 * time.Second},
		tokenURL:   tokenURL,
		expiresAt:  time.Now().Add(-time.Second), // 初始设为已过期，触发首次获取
	}
}

// GetToken 获取有效的 access token（自动刷新如果需要）
func (tm *TokenManager) GetToken() (string, error) {
	tm.mu.RLock()
	shouldRefresh := time.Now().After(tm.expiresAt.Add(-60 * time.Second)) // 提前 60 秒刷新
	tm.mu.RUnlock()

	if shouldRefresh {
		tm.mu.Lock()
		defer tm.Unlock()

		// 再次检查，防止多个 goroutine 重复刷新
		if time.Now().After(tm.expiresAt.Add(-60 * time.Second)) {
			if err := tm.refreshTokenNow(); err != nil {
				return "", fmt.Errorf("failed to refresh token: %w", err)
			}
		}
	}

	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.token, nil
}

// refreshTokenNow 实际请求新 token
func (tm *TokenManager) refreshTokenNow() error {
	req, err := http.NewRequest("POST", tm.tokenURL, nil)
	if err != nil {
		return err
	}

	// 如果有 refresh_token，可以带上；否则首次请求不需要
	if tm.refreshToken != "" {
		req.Header.Set("Authorization", "Bearer "+tm.refreshToken)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := tm.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token request failed: %d %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return err
	}

	// 更新 token 和过期时间
	tm.token = tokenResp.AccessToken
	tm.refreshToken = tokenResp.RefreshToken
	tm.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	fmt.Printf("Token refreshed: expires at %v\n", tm.expiresAt)
	return nil
}

// 示例：使用 token 调用 API
func callAPI(tm *TokenManager, apiURL string) error {
	token, err := tm.GetToken()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("API Response (%d): %s\n", resp.StatusCode, string(body))
	return nil
}

// 主函数示例
func main() {
	tokenURL := "https://your-api.com/oauth/token" // 替换为实际 token 接口
	apiURL := "https://your-api.com/api/data"       // 替换为你要调用的 API

	tm := NewTokenManager(tokenURL)

	// 首次调用会自动获取 token
	if err := callAPI(tm, apiURL); err != nil {
		fmt.Printf("Call failed: %v\n", err)
		return
	}

	// 模拟后续调用（token 有效期内）
	time.Sleep(2 * time.Second)
	if err := callAPI(tm, apiURL); err != nil {
		fmt.Printf("Call failed: %v\n", err)
	}

	// 模拟等待快过期（实际中你可能长时间运行）
	// time.Sleep(3550 * time.Second) // 接近过期，下次调用会自动刷新
}
```

---

### 🔍 说明

1. **线程安全**：使用 `sync.RWMutex` 保证多 goroutine 下安全读写 token。
2. **提前刷新**：在过期前 60 秒就刷新，避免刚好过期导致请求失败。
3. **自动刷新**：`GetToken()` 检查是否需要刷新，自动调用 `refreshTokenNow()`。
4. **refresh_token 支持**：如果你的 API 支持使用 `refresh_token` 刷新，可以在请求中带上它（注意有些 API 首次获取不需要 `refresh_token`，仅后续刷新需要）。
5. **错误处理**：网络失败、HTTP 错误、JSON 解析都做了处理。

---

### 🔄 注意事项

- 根据你的 API 文档，`/oauth/token` 接口可能是 `POST` 表单或 JSON，可能需要传 `grant_type` 等参数。上面示例假设是简单接口，你可能需要调整请求体：
  ```go
  data := url.Values{}
  data.Set("grant_type", "client_credentials")
  req, _ := http.NewRequest("POST", tm.tokenURL, strings.NewReader(data.Encode()))
  ```

- 如果你使用 client_id/client_secret，也需要添加。

---

如果你提供具体的 token 请求方式（比如是否需要 `client_id`、`client_secret`、`grant_type` 等），我可以进一步帮你完善 `refreshTokenNow()` 中的请求构造。

需要我帮你加上 `client_credentials` 模式吗？