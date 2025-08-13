# golang语言的标准库http使用起来不是很方便，要写很多代码，有没有第三方库提供简单易用的接口？

Go 标准库 `net/http` 确实偏向底层和灵活，导致日常开发中需要编写大量样板代码（如创建请求、设置超时、解析 JSON、处理错误等）。社区有多个优秀的第三方 HTTP 客户端库，提供了**链式调用、自动序列化、内置重试/超时、中间件**等开箱即用的能力。以下是目前最主流且易用的两个推荐：

### 🥇 1. `resty`（最成熟、生态最广）
**仓库**：`github.com/go-resty/resty/v2`  
**特点**：企业级首选，API 直观，支持 JSON/XML 自动解析、请求/响应钩子、重试、代理、文件上传等，文档完善，社区活跃。

```go
package main

import (
	"fmt"
	"github.com/go-resty/resty/v2"
)

func main() {
	client := resty.New().
		SetTimeout(10 * time.Second).
		SetRetryCount(2)

	var users []User
	resp, err := client.R().
		SetHeader("Authorization", "Bearer xxx").
		SetQueryParams(map[string]string{
			"page": "1",
			"size": "20",
		}).
		SetResult(&users). // 自动 JSON 反序列化到 users
		Get("https://api.example.com/users")

	if err != nil {
		// 处理网络/解析错误
		return
	}
	fmt.Printf("状态码: %d, 数据: %+v\n", resp.StatusCode(), users)
}
```

---

### 🥈 2. `req`（现代 API、更贴近直觉）
**仓库**：`github.com/imroc/req/v3`  
**特点**：设计受 Python `requests` 启发，API 更简洁现代，底层性能优秀，原生支持 HTTP/2、上下文传递、细粒度重试/超时控制，近年增长迅速。

```go
package main

import (
	"fmt"
	"github.com/imroc/req/v3"
	"time"
)

func main() {
	client := req.C().
		SetTimeout(10 * time.Second).
		EnableDebugLog()

	var users []User
	resp, err := client.R().
		SetHeader("Authorization", "Bearer xxx").
		SetQueryParams(map[string]string{"page": "1"}).
		SetResult(&users). // 自动反序列化
		Get("https://api.example.com/users")

	if err != nil {
		return
	}
	fmt.Printf("状态码: %d, 数据: %+v\n", resp.GetStatusCode(), users)
}
```

---

### 📊 快速对比
| 特性                | `resty`                  | `req`                    |
|---------------------|--------------------------|--------------------------|
| 成熟度与稳定性       | ⭐⭐⭐⭐⭐（企业广泛使用）   | ⭐⭐⭐⭐（活跃，API 更现代）|
| 链式调用            | ✅                        | ✅                        |
| 自动 JSON 解析       | `SetResult(&v)`          | `SetResult(&v)`          |
| 重试/超时控制        | 支持                     | 支持（配置更细粒度）     |
| 中间件/钩子          | 请求/响应/错误钩子       | 支持，API 更简洁         |
| 学习成本            | 低                       | 极低                     |

---

### 📦 安装方式
```bash
# 二选一
go get github.com/go-resty/resty/v2
go get github.com/imroc/req/v3
```

---

### 💡 选型建议
- **求稳、团队熟悉、生产项目** → 选 `resty`
- **新项目、喜欢现代 API、追求简洁** → 选 `req`
- 两者都内置了连接池复用，**无需手动管理 `http.Client`**，开箱即用。

### ⚠️ 生产环境注意事项
1. **务必设置全局超时**（如 `SetTimeout`），避免 goroutine 泄漏
2. 敏感 Header/Token 建议通过中间件统一注入，而非硬编码
3. 复杂业务可封装一层业务客户端（如 `type APIClient struct { *resty.Client }`）
4. 若不想引入第三方依赖，也可用标准库 + 少量封装实现类似体验，但开发效率会打折扣。

需要我针对某个具体场景（如：文件上传、OAuth2 认证、流式响应、Mock 测试）提供完整示例吗？


# 提供 文件上传、OAuth2 认证、流式响应 的示例代码

以下基于 `resty`（`github.com/go-resty/resty/v2`）提供三个场景的**生产级示例**。`resty` 本身不实现 OAuth2，但能无缝对接 Go 官方 `golang.org/x/oauth2`；流式与上传也提供了零样板代码的封装。

---
### 📦 前置依赖
```bash
go get github.com/go-resty/resty/v2
go get golang.org/x/oauth2
```

---
### 1️⃣ 文件上传（Multipart / 表单+文件）
适用于上传单文件、多文件、附带额外表单字段。

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
)

func main() {
	client := resty.New().SetTimeout(30 * time.Second)

	resp, err := client.R().
		// 1. 上传文件（字段名: "file"，本地路径）
		SetFile("file", "./report.pdf").
		// 2. 附带表单字段（可选）
		SetFormData(map[string]string{
			"title":       "Q3 Financial Report",
			"department":  "finance",
			"is_public":   "false",
		}).
		// 3. 发送请求
		Post("https://httpbin.org/post")

	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	if resp.IsError() {
		log.Fatalf("服务端错误: %d %s", resp.StatusCode(), string(resp.Body()))
	}

	fmt.Printf("上传成功, 响应: %s\n", resp.String())
}
```
🔹 **进阶技巧**：
- 动态内容（非本地文件）：使用 `SetFileReader("field", "filename", io.Reader)`
- 多文件：`SetFiles(map[string]string{"file1": "a.pdf", "file2": "b.docx"})`

---
### 2️⃣ OAuth2 认证（Client Credentials 模式）
Go 官方 `golang.org/x/oauth2` 已实现 Token 获取、缓存与自动刷新。只需将其生成的 `http.Client` 注入 `resty` 即可。

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

func main() {
	// 1. 配置 OAuth2 Client Credentials 模式（服务端对服务端）
	conf := &clientcredentials.Config{
		ClientID:     "your-client-id",
		ClientSecret: "your-client-secret",
		TokenURL:     "https://auth.example.com/oauth2/token",
		Scopes:       []string{"api:read", "api:write"},
	}

	// 2. 生成自动管理 Token 的 http.Client（自动刷新、线程安全）
	ctx := context.Background()
	oauthClient := conf.Client(ctx)

	// 3. 注入 resty
	client := resty.New().
		SetClient(oauthClient).
		SetTimeout(10 * time.Second)

	// 4. 直接请求受保护资源（无需手动加 Authorization Header）
	resp, err := client.R().Get("https://api.example.com/v1/users")
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	if resp.IsError() {
		log.Fatalf("API错误: %d %s", resp.StatusCode(), string(resp.Body()))
	}

	fmt.Printf("认证成功, 数据: %s\n", resp.String())
}
```
🔹 **其他 OAuth2 流程说明**：
- `Authorization Code` / `PKCE`（用户授权）：改用 `oauth2.Config` + `Exchange()` 获取 Token，再包装成 `conf.Client(ctx, token)`。
- `resty` 仅负责 HTTP 层，认证逻辑交给官方包是 Go 生态最佳实践。

---
### 3️⃣ 流式响应（Streaming / 大文件 / SSE / NDJSON）
默认 `resty` 会读取完整响应到内存。流式处理需关闭自动解析，直接操作 `Body`。

#### 📌 场景 A：逐行处理（日志流、SSE、JSON Lines）
```go
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
)

func main() {
	client := resty.New().SetTimeout(5 * time.Minute) // 长连接建议加大超时

	resp, err := client.R().
		SetDoNotParseResponse(true). // ⚠️ 关键：禁止自动读取 Body
		Get("https://httpbin.org/stream/100")
	if err != nil {
		log.Fatal(err)
	}
	// 必须显式关闭
	defer resp.RawResponse.Body.Close()

	// 使用 bufio.Scanner 逐行读取
	scanner := bufio.NewScanner(resp.RawResponse.Body)
	// 可选：调整单行最大长度（默认 64KB）
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	lineNum := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		fmt.Printf("[Line %d] %s\n", lineNum, line)
		// 可在此处做 JSON 反序列化、入库、推送到 channel 等
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("流读取中断: %v", err)
	}
}
```

#### 📌 场景 B：二进制流直传（下载大文件 / 管道转发）
```go
// 替换上面 for scanner.Scan() 部分：
import (
	"io"
	"os"
)

// 直接流式写入文件（不占内存）
out, _ := os.Create("./downloaded.dat")
defer out.Close()

_, err = io.Copy(out, resp.RawResponse.Body)
if err != nil {
	log.Fatal(err)
}
```
🔹 **流式关键提醒**：
1. 必须调用 `SetDoNotParseResponse(true)`，否则 `resp.Body()` 会提前消费流。
2. 长连接/流式接口务必配合 `context.WithTimeout` 或 `context.WithCancel` 控制生命周期。
3. `resp.RawResponse.Body` 是标准 `io.ReadCloser`，可直接对接 `bufio`、`io.Copy`、`json.Decoder` 等。

---
### 🛡️ 生产环境通用建议
| 场景         | 避坑指南                                                                 |
|--------------|--------------------------------------------------------------------------|
| 文件上传     | 超过 50MB 建议分片上传；服务端超时通常较长，客户端 `SetTimeout` 需匹配    |
| OAuth2       | 不要手动拼 `Authorization: Bearer xxx`，交给 `oauth2` 包自动刷新更安全    |
| 流式响应     | 务必 `defer body.Close()`；高并发下监控 goroutine 数量，防泄漏            |
| 全局配置     | 使用 `resty.New()` 创建单例 Client，复用连接池；避免每次请求都 `New()`    |

如果需要 `req` 库的等价写法、或具体对接某家云厂商 API（如阿里云 OSS 直传、微信 OAuth2、GitHub SSE），可提供具体场景，我给出针对性模板。
