package main

import (
	"crypto-trading-platform/api"
	"crypto-trading-platform/config"
	"crypto-trading-platform/internal/database"
	"log"
	"net/http"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

// CorsMiddleware 是一个处理 CORS 头的中间件
func CorsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 调试：打印请求信息
		logx.Infof("Handling request: %s %s from %s", r.Method, r.URL.Path, r.Header.Get("Origin"))

		// 设置 CORS 头
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// 调试：打印设置的 CORS 头
		logx.Infof("Set CORS headers: Access-Control-Allow-Origin=%s", w.Header().Get("Access-Control-Allow-Origin"))

		// 处理 OPTIONS 预检请求
		if r.Method == http.MethodOptions {
			logx.Info("Handling OPTIONS preflight request")
			w.WriteHeader(http.StatusOK)
			return
		}

		// 继续处理其他请求
		next(w, r)
	}
}

func main() {
	// 初始化日志
	logx.MustSetup(logx.LogConf{ServiceName: "crypto-trading-platform", Mode: "console"})
	logx.Info("Starting application...")

	// 加载配置
	var c config.Config
	conf.MustLoad("config.yaml", &c)
	logx.Infof("Loaded config: %+v", c)

	// 初始化数据库
	db, err := database.InitDB(c.Database.Path)
	if err != nil {
		logx.Errorf("Failed to initialize database: %v", err)
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	logx.Info("Database initialized successfully")

	// 创建服务器，禁用默认 CORS 处理
	server := rest.MustNewServer(c.RestConf, rest.WithCustomCors(nil, nil))
	defer server.Stop()
	logx.Info("HTTP server created")

	// 应用 CORS 中间件
	server.Use(CorsMiddleware)
	logx.Info("CORS middleware applied")

	// 注册路由
	api.RegisterHandlers(server, db, c)
	logx.Info("Routes registered")

	// 启动服务
	logx.Infof("Starting server at %s:%d...", c.RestConf.Host, c.RestConf.Port)
	server.Start()
}
