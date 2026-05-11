package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Add 业务逻辑保持不变
func Add(a, b int) int {
	return a + b
}

// 增加一个简单的日志中间件，用来生成 RequestID 和记录耗时
func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// 模拟生成 RequestID（实际可用 UUID）
		reqID := fmt.Sprintf("%d", start.UnixNano())
		w.Header().Set("X-Request-ID", reqID)

		next.ServeHTTP(w, r)

		// 打印类似你之前看到的日志格式
		log.Printf("[GET] %s | ID: %s | Time: %v | Size: %s",
			r.URL.Path, reqID, time.Since(start), w.Header().Get("Content-Length"))
	})
}

func handler(w http.ResponseWriter, r *http.Request) {
	result := Add(1, 2)
	fmt.Fprintf(w, "result: %d", result)
}

func main() {
	mux := http.NewServeMux()
	// 注册路由，并包装中间件
	mux.Handle("/", loggerMiddleware(http.HandlerFunc(handler)))

	// --- 生产级自定义 Server 配置 ---
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
		// 读取客户端 Header 的超时，防止慢连接攻击 (Slowloris)
		ReadHeaderTimeout: 2 * time.Second,
		// 读取整个请求体的超时
		ReadTimeout: 5 * time.Second,
		// 响应写入超时
		WriteTimeout: 10 * time.Second,
		// 关键：控制长连接（Keep-Alive）的空闲时间
		IdleTimeout: 30 * time.Second,
		// 限制 Header 的大小，防止恶意大报文
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// 使用协程启动服务，方便后续平滑关闭
	go func() {
		fmt.Println("🚀 生产级服务启动，监听端口 :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %s\n", err)
		}
	}()

	// --- 平滑关闭（Graceful Shutdown） ---
	// 监听系统中断信号（如 Ctrl+C 或 Render 的停止信号）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n⚠️ 正在关闭服务器...")

	// 给现有的请求 5 秒的处理宽限期，处理完再关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("服务器强制关闭:", err)
	}

	fmt.Println("✅ 服务已平稳退出")
}
