package api

import (
	"crypto-trading-platform/config"
	"crypto-trading-platform/internal/kraken" // Corrected from krake to kraken
	"crypto-trading-platform/service"
	"database/sql"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

// RegisterHandlers 注册所有HTTP路由处理器
func RegisterHandlers(server *rest.Server, db *sql.DB, c config.Config) {
	// 初始化日志
	logx.MustSetup(logx.LogConf{ServiceName: "crypto-trading-api", Mode: "console"})
	logx.Info("Registering API routes...")

	// 创建Kraken客户端
	krakenClient := kraken.NewClient(c.Kraken.APIKey, c.Kraken.APISecret)

	// 创建服务
	userService := service.NewUserService(db, c.Auth.AccessSecret, c.Auth.AccessExpire)
	walletService := service.NewWalletService(db, krakenClient.NewClientAPI()) // Adjusted to use Api field
	tradingService := service.NewTradingService(db, krakenClient, walletService)

	// 创建处理器
	handler := NewHandler(userService, walletService, tradingService, c.Auth.AccessSecret)

	// 不需要认证的路由
	publicRoutes := []rest.Route{
		{
			Method:  http.MethodPost,
			Path:    "/api/user/register",
			Handler: handler.Register,
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/user/login",
			Handler: handler.Login,
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/market/price/:pair",
			Handler: handler.GetPrice,
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/market/pairs",
			Handler: handler.GetTradablePairs,
		},
	}
	server.AddRoutes(publicRoutes)
	logx.Info("Registered public routes", len(publicRoutes))

	// 需要认证的路由
	protectedRoutes := []rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/user/info",
			Handler: handler.GetUserInfo,
		},
		{
			Method:  http.MethodPut,
			Path:    "/api/user/password",
			Handler: handler.UpdateUserPassword,
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/wallet/list",
			Handler: handler.GetUserWallets,
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/wallet/balance/:currency",
			Handler: handler.GetWalletBalance,
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/wallet/sync",
			Handler: handler.SyncWalletBalances,
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/trade/order",
			Handler: handler.PlaceOrder,
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/trade/history",
			Handler: handler.GetUserTransactions,
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/trade/status/:id",
			Handler: handler.GetTransactionStatus,
		},
	}
	server.AddRoutes(protectedRoutes, rest.WithJwt(c.Auth.AccessSecret))
	logx.Info("Registered protected routes", len(protectedRoutes))
}
