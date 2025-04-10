package api

import (
	"crypto-trading-platform/service"
	"encoding/json"
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
	token := r.Context().Value("jwt").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	userID := int64(claims["userId"].(float64))
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fail(w, 400, "Invalid request: "+err.Error())
		return
	}

	userID, err := h.userService.Register(req.Username, req.Password, req.Email)
	if err != nil {
		fail(w, 500, "Registration failed: "+err.Error())
		return
	}

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
		fail(w, 401, "Unauthorized")
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
		fail(w, 401, "Unauthorized")
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
		fail(w, 401, "Unauthorized")
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
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		fail(w, 401, "Unauthorized")
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
		fail(w, 401, "Unauthorized")
		return
	}

	err = h.walletService.SyncWalletBalances(userID)
	if err != nil {
		fail(w, 500, "Failed to sync balances: "+err.Error())
		return
	}

	success(w, nil)
}

func (h *Handler) GetPrice(w http.ResponseWriter, r *http.Request) {
	// Retrieve the 'pair' path parameter using Go-Zero's helper
	pair := httpx.GetPathParam(r, "pair")
	if pair == "" {
		fail(w, 400, "Pair is required")
		return
	}

	price, err := h.tradingService.GetPrice(pair)
	if err != nil {
		fail(w, 500, "Failed to get price: "+err.Error())
		return
	}

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
		fail(w, 401, "Unauthorized")
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
		fail(w, 401, "Unauthorized")
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
		fail(w, 401, "Unauthorized")
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
