# 用golang编写一个通用的发送http请求的函数，能使用不同的http方法，能设置头，设置token，设置cookie，能将请求信息存入日志文件，返回响应

下面是一个用 Go 编写的通用 HTTP 请求函数，支持：

- 不同的 HTTP 方法（GET、POST、PUT、DELETE 等）
- 自定义请求头
- 设置 Token（自动添加到 `Authorization: Bearer <token>`）
- 设置 Cookie
- 请求信息记录到日志文件
- 返回响应体和错误

---

✅ **功能亮点：**

- 使用 `net/http` 标准库
- 支持设置 `Header`、`Token`、`Cookie`
- 日志记录到文件（使用 `log` 包）
- 返回 `[]byte` 响应体和 `error`

---

📁 **先创建日志文件目录（如 logs/）或确保可写入当前目录**

---

### ✅ 完整代码：

```go
package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// HTTPClient 封装客户端和日志器
type HTTPClient struct {
	client *http.Client
	logger *log.Logger
}

// NewHTTPClient 创建带日志记录的 HTTP 客户端
func NewHTTPClient(logFile string) (*HTTPClient, error) {
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	logger := log.New(f, "HTTP_LOG: ", log.Ldate|log.Ltime|log.Lshortfile)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &HTTPClient{
		client: client,
		logger: logger,
	}, nil
}

// RequestOptions 请求配置选项
type RequestOptions struct {
	Method  string            // HTTP 方法：GET, POST, PUT, DELETE 等
	URL     string            // 请求 URL
	Headers map[string]string // 自定义头
	Token   string            // Bearer Token，自动添加到 Authorization 头
	Cookies []*http.Cookie    // Cookie 列表
	Body    []byte            // 请求体（如 POST/PUT 的 JSON）
}

// DoRequest 发送 HTTP 请求
func (hc *HTTPClient) DoRequest(opts RequestOptions) ([]byte, error) {
	// 创建请求
	req, err := http.NewRequest(opts.Method, opts.URL, bytes.NewBuffer(opts.Body))
	if err != nil {
		hc.logger.Printf("[ERROR] 创建请求失败: %v", err)
		return nil, err
	}

	// 设置 Header
	for key, value := range opts.Headers {
		req.Header.Set(key, value)
	}

	// 设置 Token（Bearer）
	if opts.Token != "" {
		req.Header.Set("Authorization", "Bearer "+opts.Token)
	}

	// 设置 Cookie
	for _, cookie := range opts.Cookies {
		req.AddCookie(cookie)
	}

	// 记录请求日志
	hc.logger.Printf("[REQUEST] %s %s", opts.Method, opts.URL)
	hc.logger.Printf("[HEADERS] %v", req.Header)
	if len(opts.Body) > 0 {
		hc.logger.Printf("[BODY] %s", string(opts.Body))
	}
	if len(opts.Cookies) > 0 {
		hc.logger.Printf("[COOKIES] %+v", opts.Cookies)
	}

	// 发送请求
	resp, err := hc.client.Do(req)
	if err != nil {
		hc.logger.Printf("[ERROR] 请求失败: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		hc.logger.Printf("[ERROR] 读取响应失败: %v", err)
		return nil, err
	}

	// 记录响应日志
	hc.logger.Printf("[RESPONSE] 状态码: %d", resp.StatusCode)
	hc.logger.Printf("[RESPONSE HEADERS] %v", resp.Header)
	hc.logger.Printf("[RESPONSE BODY] %s", string(body))

	return body, nil
}

// 示例使用
func main() {
	// 初始化客户端，日志写入 http_requests.log
	client, err := NewHTTPClient("http_requests.log")
	if err != nil {
		log.Fatal("初始化客户端失败:", err)
	}

	// 示例：发送 GET 请求
	opts := RequestOptions{
		Method: "GET",
		URL:    "https://httpbin.org/get",
		Headers: map[string]string{
			"User-Agent": "Go-HTTP-Client/1.0",
		},
		Token: "your-access-token-here",
		Cookies: []*http.Cookie{
			{Name: "session_id", Value: "abc123xyz"},
		},
	}

	respBody, err := client.DoRequest(opts)
	if err != nil {
		log.Fatal("请求失败:", err)
	}

	log.Println("响应:", string(respBody))
}
```

---

### ✅ 功能说明：

- **日志记录**：所有请求/响应信息（URL、方法、头、Body、状态码）都会写入指定日志文件。
- **Token 支持**：只需传入 `Token` 字段，自动设置 `Authorization: Bearer xxx`。
- **Cookie 支持**：传入 `[]*http.Cookie`，会自动附加。
- **灵活 Body**：支持任意 `[]byte`，可传入 JSON、表单等。
- **超时控制**：默认 30 秒，可自定义修改 `client.Timeout`。
- **错误处理**：记录错误日志并返回给调用者。

---

### ✅ 扩展建议：

- 支持 Context 控制取消/超时
- 支持重试机制
- 支持自动解码 JSON 响应
- 支持 TLS 配置 / 代理

---

### 📌 示例日志输出：

```
HTTP_LOG: 2025/04/05 10:30:25 main.go:88: [REQUEST] GET https://httpbin.org/get
HTTP_LOG: 2025/04/05 10:30:25 main.go:89: [HEADERS] map[Authorization:[Bearer your-access-token-here] User-Agent:[Go-HTTP-Client/1.0]]
HTTP_LOG: 2025/04/05 10:30:25 main.go:93: [COOKIES] [Name:session_id Value:abc123xyz ...]
HTTP_LOG: 2025/04/05 10:30:26 main.go:108: [RESPONSE] 状态码: 200
HTTP_LOG: 2025/04/05 10:30:26 main.go:109: [RESPONSE HEADERS] map[Content-Type:[application/json] ...]
HTTP_LOG: 2025/04/05 10:30:26 main.go:110: [RESPONSE BODY] { ... }
```

---

✅ 你可以将此函数封装成包供项目复用，非常灵活通用！

如需支持 `context`、`JSON 自动 Marshal/Unmarshal`、`重试` 等，可进一步扩展。欢迎提出需求！


# 如何使用这个函数上传文件呢？

非常好的问题！上面的通用 HTTP 函数目前支持普通 `[]byte` 请求体（如 JSON、表单数据），但**上传文件通常需要使用 `multipart/form-data` 编码**，这需要特殊构造请求体。

---

## ✅ 解决方案：扩展 `DoRequest` 支持文件上传

我们可以在原有结构基础上，**新增一个字段 `Files map[string]string`**（字段名 → 文件路径），并在函数内部构造 `multipart/form-data` 请求体。

---

## ✅ 修改后的完整代码（支持文件上传）

```go
package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"
	"log"
)

// HTTPClient 封装客户端和日志器
type HTTPClient struct {
	client *http.Client
	logger *log.Logger
}

// NewHTTPClient 创建带日志记录的 HTTP 客户端
func NewHTTPClient(logFile string) (*HTTPClient, error) {
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	logger := log.New(f, "HTTP_LOG: ", log.Ldate|log.Ltime|log.Lshortfile)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &HTTPClient{
		client: client,
		logger: logger,
	}, nil
}

// RequestOptions 请求配置选项
type RequestOptions struct {
	Method  string            // HTTP 方法
	URL     string            // 请求 URL
	Headers map[string]string // 自定义头（上传时一般不手动设 Content-Type）
	Token   string            // Bearer Token
	Cookies []*http.Cookie    // Cookie 列表
	Body    []byte            // 普通请求体（JSON / 表单等）
	Files   map[string]string // 文件字段名 -> 文件路径，如 {"file": "./photo.jpg"}
	// 可选：普通表单字段
	FormData map[string]string // 普通字段，如 {"title": "My Photo"}
}

// DoRequest 发送 HTTP 请求（支持文件上传）
func (hc *HTTPClient) DoRequest(opts RequestOptions) ([]byte, error) {
	var bodyBuffer *bytes.Buffer
	var contentType string

	// 如果有文件，构造 multipart/form-data
	if len(opts.Files) > 0 {
		bodyBuffer = &bytes.Buffer{}
		writer := multipart.NewWriter(bodyBuffer)

		// 添加普通表单字段
		if opts.FormData != nil {
			for key, val := range opts.FormData {
				_ = writer.WriteField(key, val)
			}
		}

		// 添加文件
		for fieldName, filePath := range opts.Files {
			file, err := os.Open(filePath)
			if err != nil {
				hc.logger.Printf("[ERROR] 打开文件失败 %s: %v", filePath, err)
				return nil, err
			}
			defer file.Close()

			// 获取文件名
			filename := filepath.Base(filePath)
			// 创建表单文件头
			h := make(textproto.MIMEHeader)
			h.Set("Content-Disposition",
				fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, filename))
			h.Set("Content-Type", getContentType(filePath))

			// 创建 multipart 部分
			part, err := writer.CreatePart(h)
			if err != nil {
				hc.logger.Printf("[ERROR] 创建 multipart 部分失败: %v", err)
				return nil, err
			}

			// 拷贝文件内容
			_, err = io.Copy(part, file)
			if err != nil {
				hc.logger.Printf("[ERROR] 写入文件内容失败: %v", err)
				return nil, err
			}
		}

		// 关闭 writer，写入 boundary
		err := writer.Close()
		if err != nil {
			hc.logger.Printf("[ERROR] 关闭 multipart writer 失败: %v", err)
			return nil, err
		}

		contentType = writer.FormDataContentType()
	} else {
		// 没有文件，使用普通 Body
		bodyBuffer = bytes.NewBuffer(opts.Body)
	}

	// 创建请求
	req, err := http.NewRequest(opts.Method, opts.URL, bodyBuffer)
	if err != nil {
		hc.logger.Printf("[ERROR] 创建请求失败: %v", err)
		return nil, err
	}

	// 设置 Header（如果未设置 Content-Type，则使用 multipart 的）
	for key, value := range opts.Headers {
		req.Header.Set(key, value)
	}
	if len(opts.Files) > 0 && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", contentType)
	}

	// 设置 Token（Bearer）
	if opts.Token != "" {
		req.Header.Set("Authorization", "Bearer "+opts.Token)
	}

	// 设置 Cookie
	for _, cookie := range opts.Cookies {
		req.AddCookie(cookie)
	}

	// 记录请求日志（不记录文件内容，避免日志过大）
	hc.logger.Printf("[REQUEST] %s %s", opts.Method, opts.URL)
	hc.logger.Printf("[HEADERS] %v", req.Header)
	if len(opts.Body) > 0 && len(opts.Files) == 0 {
		hc.logger.Printf("[BODY] %s", string(opts.Body))
	}
	if len(opts.Files) > 0 {
		hc.logger.Printf("[UPLOADING FILES] %v", opts.Files)
		if opts.FormData != nil {
			hc.logger.Printf("[FORM DATA] %v", opts.FormData)
		}
	}
	if len(opts.Cookies) > 0 {
		hc.logger.Printf("[COOKIES] %+v", opts.Cookies)
	}

	// 发送请求
	resp, err := hc.client.Do(req)
	if err != nil {
		hc.logger.Printf("[ERROR] 请求失败: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		hc.logger.Printf("[ERROR] 读取响应失败: %v", err)
		return nil, err
	}

	// 记录响应日志
	hc.logger.Printf("[RESPONSE] 状态码: %d", resp.StatusCode)
	hc.logger.Printf("[RESPONSE HEADERS] %v", resp.Header)
	hc.logger.Printf("[RESPONSE BODY] %s", string(body))

	return body, nil
}

// 辅助函数：根据文件扩展名简单判断 Content-Type
func getContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".zip":
		return "application/zip"
	case ".mp4":
		return "video/mp4"
	case ".mp3":
		return "audio/mpeg"
	default:
		return "application/octet-stream"
	}
}

// 示例：上传文件
func main() {
	client, err := NewHTTPClient("http_requests.log")
	if err != nil {
		log.Fatal("初始化客户端失败:", err)
	}

	// 示例：上传文件到 https://httpbin.org/post
	opts := RequestOptions{
		Method: "POST",
		URL:    "https://httpbin.org/post",
		Token:  "your-token-here",
		Headers: map[string]string{
			"User-Agent": "Go-File-Uploader/1.0",
		},
		FormData: map[string]string{
			"title":       "My Vacation Photo",
			"description": "Photo taken in Hawaii",
		},
		Files: map[string]string{
			"file1": "./photo.jpg",     // 表单字段名 "file1" 对应本地文件
			"avatar": "./avatar.png",   // 另一个文件，字段名 "avatar"
		},
	}

	respBody, err := client.DoRequest(opts)
	if err != nil {
		log.Fatal("上传失败:", err)
	}

	log.Println("服务器响应:", string(respBody))
}
```

---

## ✅ 使用说明

你只需在 `RequestOptions` 中设置：

```go
opts := RequestOptions{
    Method: "POST",
    URL:    "https://your-upload-api.com/upload",
    Files: map[string]string{
        "file": "./document.pdf",      // 表单字段名为 "file"
        "image": "./screenshot.png",   // 表单字段名为 "image"
    },
    FormData: map[string]string{       // 可选：其他表单字段
        "user_id": "123",
        "category": "profile",
    },
    Token: "xxx",
}
```

> ⚠️ 注意：
> - 不要手动设置 `Content-Type`，函数会自动生成正确的 `multipart/form-data; boundary=...`
> - 文件路径必须存在且可读
> - 日志中不会记录文件内容，避免日志爆炸

---

## ✅ 服务器端如何接收？

以 `multipart/form-data` 上传，服务器（如 PHP、Node.js、Go、Python Flask/Django）需用对应方式解析：

- Go: `r.ParseMultipartForm()` + `r.FormFile("file")`
- Node.js (Express): `multer`
- Python: `request.files['file']`
- PHP: `$_FILES['file']`

---

## ✅ 扩展建议

- 支持上传 `io.Reader`（内存中的文件，如从 S3 读取）
- 支持设置每个文件的自定义 `Content-Type`
- 支持上传进度回调
- 支持并发上传多个文件

---

✅ 现在这个通用函数既能发普通请求，又能上传文件，非常实用！欢迎根据项目需求进一步定制。
