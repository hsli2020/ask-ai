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
	dir := flag.String("dir", "./", "静态文件根目录")
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