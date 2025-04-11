package main

import (
	"crypto-trading-platform/api"
	"crypto-trading-platform/internal/kraken" // Use internal Kraken package
	"crypto-trading-platform/service"
	"database/sql"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

func main() {
	print("--------------0.1\n")

	// Initialize the server with configuration
	server := rest.MustNewServer(rest.RestConf{
		Host: "0.0.0.0",
		Port: 8888,
	})
	defer server.Stop()

	// Initialize SQLite database
	db, err := sql.Open("sqlite3", "./trading_platform.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Initialize Kraken client using internal/kraken
	krakenClient := kraken.NewClient("your-kraken-api-key", "your-kraken-api-secret") // Adjust based on actual constructor

	// Initialize services with required dependencies
	userService := service.NewUserService(db, "your-salt-or-secret", 3600) // e.g., 3600 seconds timeout
	walletService := service.NewWalletService(db, krakenClient.NewClientAPI())
	tradingService := service.NewTradingService(db, krakenClient, walletService)

	// JWT secret
	jwtSecret := "your-jwt-secret-here" // Replace with your actual JWT secret

	// Create the handler
	handler := api.NewHandler(userService, walletService, tradingService, jwtSecret)

	// Middleware for CORS (to allow requests from your frontend)
	server.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173") // Fixed typo
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next(w, r)
		}
	})

	// Register all routes
	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/api/user/register",
		Handler: handler.Register,
	})
	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/api/user/login",
		Handler: handler.Login,
	})
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/api/market/price/:pair",
		Handler: handler.GetPrice,
	})
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/api/wallet/balance/:currency",
		Handler: handler.GetWalletBalance,
	})
	// Example for a protected route with JWT
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/api/user/info",
		Handler: handler.GetUserInfo,
	}, rest.WithJwt(jwtSecret))

	// Start the server
	logx.Info("Starting server on :8888...")
	server.Start()
}
