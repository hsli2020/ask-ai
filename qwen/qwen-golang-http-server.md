# 用golang编写一个httpserver，能用来作为静态网站的服务

这是一个使用 Go 标准库编写的生产级静态文件服务器。包含 **命令行参数解析**、**优雅关闭**、以及 **SPA（单页应用）路由回退** 支持。

### 📦 完整代码 (`main.go`)

```go
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	// 命令行参数定义
	port := flag.String("port", "8080", "HTTP 服务监听端口")
	dir := flag.String("dir", "./public", "静态文件根目录")
	spa := flag.Bool("spa", false, "开启 SPA 模式：未匹配到静态文件时回退到 index.html")
	flag.Parse()

	// 检查目录是否存在
	if _, err := os.Stat(*dir); os.IsNotExist(err) {
		log.Fatalf("❌ 错误: 目录 %s 不存在", *dir)
	}

	// 基础文件服务器
	fs := http.FileServer(http.Dir(*dir))
	var handler http.Handler = fs

	// 如果开启 SPA 模式，包装处理器实现路由回退
	if *spa {
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 尝试打开请求路径对应的文件/目录
			if f, err := http.Dir(*dir).Open(r.URL.Path); err == nil {
				f.Close() // 能打开说明是静态资源，交由 FileServer 处理
				fs.ServeHTTP(w, r)
			} else {
				// 文件不存在，返回 index.html（交由前端路由接管）
				http.ServeFile(w, r, filepath.Join(*dir, "index.html"))
			}
		})
	}

	// 配置 HTTP 服务器
	server := &http.Server{
		Addr:    ":" + *port,
		Handler: handler,
	}

	// 启动服务（非阻塞）
	go func() {
		log.Printf("🚀 静态文件服务器已启动 | 端口: %s | 目录: %s | SPA: %v", *port, *dir, *spa)
		log.Printf("🌐 访问地址: http://localhost:%s", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ 服务器启动失败: %v", err)
		}
	}()

	// 监听系统信号，实现优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("⏳ 正在优雅关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("❌ 服务器强制关闭失败: %v", err)
	}
	log.Println("✅ 服务器已安全退出")
}
```

### 🛠️ 编译与运行

```bash
# 1. 编译（生成独立可执行文件，无外部依赖）
go build -o static-server main.go

# 2. 基础用法（默认监听 8080，服务 ./public 目录）
./static-server

# 3. 自定义端口与目录
./static-server -port 3000 -dir /var/www/html

# 4. 开启 SPA 模式（适用于 Vue/React/Angular 打包产物）
./static-server -spa
```

### 💡 核心特性说明

| 特性 | 说明 |
|------|------|
| **零依赖** | 仅使用 Go 标准库，编译后仅一个可执行文件（通常 5~10MB） |
| **自动防目录遍历** | `http.FileServer` 内置安全校验，自动拦截 `../` 等恶意路径 |
| **SPA 路由支持** | 开启 `-spa` 后，访问 `/about` 等前端路由不会报 404，而是返回 `index.html` |
| **优雅关闭** | 捕获 `Ctrl+C` 或 `SIGTERM`，等待正在处理的请求完成后安全退出 |
| **MIME 自动识别** | 根据文件扩展名自动设置 `Content-Type`（js/css/png/html 等） |

### 📝 生产环境建议

1. **静态资源缓存**：前端打包时建议给文件名加 Hash，并在 Nginx/CDN 层设置长期缓存。
2. **HTTPS**：Go 原生支持 TLS，可改用 `server.ListenAndServeTLS(cert, key)`，但生产环境通常推荐使用 **Caddy / Nginx** 反向代理并自动续签证书。
3. **压缩支持**：如需 Gzip/Brotli，可引入 `github.com/klauspost/compress` 中间件包装 `handler`。
4. **权限控制**：运行服务时建议使用非 root 用户，并确保静态目录权限为 `755`。

如需添加 **日志记录中间件**、**跨域头配置** 或 **Docker 部署示例**，可告诉我具体需求。
