package service

import (
	"crypto-trading-platform/internal/models"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	db           *sql.DB
	accessSecret string
	accessExpire int64
}

// NewUserService 创建新的用户服务
func NewUserService(db *sql.DB, accessSecret string, accessExpire int64) *UserService {
	return &UserService{
		db:           db,
		accessSecret: accessSecret,
		accessExpire: accessExpire,
	}
}

// Register 注册新用户
func (s *UserService) Register(username, password, email string) (int64, error) {
	// 检查用户名是否已存在
	_, err := models.GetUserByUsername(s.db, username)
	if err == nil {
		return 0, errors.New("username already exists")
	} else if err != sql.ErrNoRows {
		return 0, err
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	// 创建用户
	userID, err := models.CreateUser(s.db, username, string(hashedPassword), email)
	if err != nil {
		return 0, err
	}

	// 为新用户创建几个常用货币的钱包
	currencies := []string{"BTC", "ETH", "USD"}
	for _, currency := range currencies {
		_, err = models.CreateWallet(s.db, userID, currency)
		if err != nil {
			return 0, err
		}
	}

	return userID, nil
}

// RegisterHandler 处理 HTTP 注册请求
func (s *UserService) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// 解析请求体
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 调用 Register 方法
	userID, err := s.Register(req.Username, req.Password, req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 返回成功响应
	resp := struct {
		UserID int64  `json:"userId"`
		Msg    string `json:"msg"`
	}{
		UserID: userID,
		Msg:    "User registered successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Login 用户登录
func (s *UserService) Login(username, password string) (string, error) {
	// 获取用户
	user, err := models.GetUserByUsername(s.db, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("invalid username or password")
		}
		return "", err
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid username or password")
	}

	// 生成 JWT 令牌
	now := time.Now().Unix()
	claims := make(jwt.MapClaims)
	claims["exp"] = now + s.accessExpire
	claims["iat"] = now
	claims["userId"] = user.ID
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims

	return token.SignedString([]byte(s.accessSecret))
}

// LoginHandler 处理 HTTP 登录请求
func (s *UserService) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// 解析请求体
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 调用 Login 方法
	token, err := s.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// 返回成功响应
	resp := struct {
		Token string `json:"token"`
		Msg   string `json:"msg"`
	}{
		Token: token,
		Msg:   "Login successful",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(userID int64) (*models.User, error) {
	return models.GetUserByID(s.db, userID)
}

// GetUserInfoHandler 处理 HTTP 获取用户信息请求
func (s *UserService) GetUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// 假设 userID 从 JWT 令牌中获取，这里简化处理
	userID := int64(1) // 实际中需要从 JWT 解析
	user, err := s.GetUserInfo(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// 返回用户信息
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// UpdateUserPassword 更新用户密码
func (s *UserService) UpdateUserPassword(userID int64, currentPassword, newPassword string) error {
	// 获取用户
	user, err := models.GetUserByID(s.db, userID)
	if err != nil {
		return err
	}

	// 验证当前密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword))
	if err != nil {
		return errors.New("current password is incorrect")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 更新密码
	user.Password = string(hashedPassword)
	return models.UpdateUser(s.db, user)
}

// UpdateUserPasswordHandler 处理 HTTP 更新密码请求
func (s *UserService) UpdateUserPasswordHandler(w http.ResponseWriter, r *http.Request) {
	// 解析请求体
	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 假设 userID 从 JWT 令牌中获取
	userID := int64(1) // 实际中需要从 JWT 解析
	err := s.UpdateUserPassword(userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 返回成功响应
	resp := struct {
		Msg string `json:"msg"`
	}{
		Msg: "Password updated successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
