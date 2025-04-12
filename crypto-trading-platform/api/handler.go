package api

import (
	"crypto-trading-platform/service"
	"encoding/json"
	"fmt"
	"log" // 新增：用于打印日志
	"net/http"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type Handler struct {
	userService    *service.UserService
	walletService  *service.WalletService
	tradingService *service.TradingService
	jwtSecret      string
}

func NewHandler(userService *service.UserService, walletService *service.WalletService, tradingService *service.TradingService, jwtSecret string) *Handler {
	return &Handler{
		userService:    userService,
		walletService:  walletService,
		tradingService: tradingService,
		jwtSecret:      jwtSecret,
	}
}

func (h *Handler) getUserIDFromRequest(r *http.Request) (int64, error) {
	// 从请求上下文中获取 JWT token
	tokenVal := r.Context().Value("jwt")
	if tokenVal == nil {
		log.Printf("Error: JWT token not found in request context")
		return 0, fmt.Errorf("jwt token not found in request context")
	}

	// 类型断言：确保 tokenVal 是 *jwt.Token 类型
	token, ok := tokenVal.(*jwt.Token)
	if !ok {
		log.Printf("Error: Invalid JWT token type, got: %T", tokenVal)
		return 0, fmt.Errorf("invalid jwt token type, got: %T", tokenVal)
	}

	// 获取 token 的 claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Printf("Error: Invalid JWT claims type, got: %T", token.Claims)
		return 0, fmt.Errorf("invalid jwt claims type, got: %T", token.Claims)
	}

	// 从 claims 中获取 userId
	userIDVal, exists := claims["userId"]
	if !exists {
		log.Printf("Error: userId not found in JWT claims")
		return 0, fmt.Errorf("userId not found in jwt claims")
	}

	// 类型断言：确保 userId 是 float64 类型（JWT 通常用 float64 表示数字）
	userIDFloat, ok := userIDVal.(float64)
	if !ok {
		log.Printf("Error: Invalid userId type in JWT claims, got: %T", userIDVal)
		return 0, fmt.Errorf("invalid userId type in jwt claims, got: %T", userIDVal)
	}

	// 转换为 int64
	userID := int64(userIDFloat)
	log.Printf("Successfully extracted userID: %d", userID)
	return userID, nil
}

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func success(w http.ResponseWriter, data interface{}) {
	resp := Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	}
	httpx.OkJson(w, resp)
}

func fail(w http.ResponseWriter, code int, msg string) {
	resp := Response{
		Code: code,
		Msg:  msg,
	}
	httpx.OkJson(w, resp)
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	print("--------------0.4\n")

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fail(w, 400, "Invalid request: "+err.Error())
		return
	}

	userID, err := h.userService.Register(req.Username, req.Password, req.Email)

	if err != nil {
		fail(w, 500, "Registration failed: "+err.Error())
		return
	}

	print("--------------0.5\n")

	success(w, map[string]int64{"user_id": userID})
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fail(w, 400, "Invalid request: "+err.Error())
		return
	}

	token, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		fail(w, 401, "Login failed: "+err.Error())
		return
	}

	success(w, map[string]string{"token": token})
}

func (h *Handler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		fail(w, 401, "Unauthorized: "+err.Error())
		return
	}

	user, err := h.userService.GetUserInfo(userID)
	if err != nil {
		fail(w, 500, "Failed to get user info: "+err.Error())
		return
	}

	user.Password = ""
	success(w, user)
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *Handler) UpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		fail(w, 401, "Unauthorized: "+err.Error())
		return
	}

	var req UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fail(w, 400, "Invalid request: "+err.Error())
		return
	}

	err = h.userService.UpdateUserPassword(userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		fail(w, 500, "Failed to update password: "+err.Error())
		return
	}

	success(w, nil)
}

func (h *Handler) GetUserWallets(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		fail(w, 401, "Unauthorized: "+err.Error())
		return
	}

	wallets, err := h.walletService.GetUserWallets(userID)
	if err != nil {
		fail(w, 500, "Failed to get wallets: "+err.Error())
		return
	}

	success(w, wallets)
}

func (h *Handler) GetWalletBalance(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("----------------------   00001 ")

	userID, err := h.getUserIDFromRequest(r)

	if err != nil {
		fail(w, 401, "Unauthorized: "+err.Error())
		return
	}

	currency, ok := r.Context().Value("currency").(string)
	if !ok || currency == "" {
		fail(w, 400, "Currency is required")
		return
	}

	balance, err := h.walletService.GetWalletBalance(userID, currency)
	if err != nil {
		fail(w, 500, "Failed to get balance: "+err.Error())
		return
	}

	success(w, map[string]float64{"balance": balance})
}

func (h *Handler) SyncWalletBalances(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		fail(w, 401, "Unauthorized: "+err.Error())
		return
	}

	fmt.Println("------------------------   000")

	err = h.walletService.SyncWalletBalances(userID)
	if err != nil {
		fail(w, 500, "Failed to sync balances: "+err.Error())
		return
	}

	success(w, nil)
}

type PriceRequest struct {
	Pair string `path:"pair"`
}

func (h *Handler) GetPrice(w http.ResponseWriter, r *http.Request) {
	print("----------------- 000.1")

	var req PriceRequest
	if err := httpx.ParsePath(r, &req); err != nil {
		fail(w, 400, "Failed to parse path: "+err.Error())
		return
	}

	print("----------------- 000.2")

	if req.Pair == "" {
		fail(w, 400, "Pair is required")
		return
	}

	print("----------------- 000.3")

	price, err := h.tradingService.GetPrice(req.Pair)

	print("----------------- 000.31")

	if err != nil {
		fail(w, 500, "Failedddddd to get price: "+err.Error())
		return
	}

	print("----------------- 000.4")

	success(w, map[string]float64{"price": price})
}

func (h *Handler) GetTradablePairs(w http.ResponseWriter, r *http.Request) {
	pairs, err := h.tradingService.GetTradablePairs()
	if err != nil {
		fail(w, 500, "Failed to get tradable pairs: "+err.Error())
		return
	}

	success(w, pairs)
}

type PlaceOrderRequest struct {
	Pair      string  `json:"pair"`
	OrderType string  `json:"order_type"`
	Side      string  `json:"side"`
	Amount    float64 `json:"amount"`
	Price     float64 `json:"price"`
}

func (h *Handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		fail(w, 401, "Unauthorized: "+err.Error())
		return
	}

	var req PlaceOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fail(w, 400, "Invalid request: "+err.Error())
		return
	}

	txID, err := h.tradingService.PlaceOrder(userID, req.Pair, req.OrderType, req.Side, req.Amount, req.Price)
	if err != nil {
		fail(w, 500, "Failed to place order: "+err.Error())
		return
	}

	success(w, map[string]int64{"transaction_id": txID})
}

func (h *Handler) GetUserTransactions(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		fail(w, 401, "Unauthorized: "+err.Error())
		return
	}

	transactions, err := h.tradingService.GetUserTransactions(userID)
	if err != nil {
		fail(w, 500, "Failed to get transactions: "+err.Error())
		return
	}

	success(w, transactions)
}

func (h *Handler) GetTransactionStatus(w http.ResponseWriter, r *http.Request) {
	_, err := h.getUserIDFromRequest(r)
	if err != nil {
		fail(w, 401, "Unauthorized: "+err.Error())
		return
	}

	idStr, ok := r.Context().Value("id").(string)
	if !ok || idStr == "" {
		fail(w, 400, "Invalid transaction ID")
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		fail(w, 400, "Invalid transaction ID")
		return
	}

	status, err := h.tradingService.GetTransactionStatus(id)
	if err != nil {
		fail(w, 500, "Failed to get transaction status: "+err.Error())
		return
	}

	success(w, map[string]string{"status": status})
}
