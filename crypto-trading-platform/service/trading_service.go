package service

import (
	"crypto-trading-platform/internal/kraken"
	"crypto-trading-platform/internal/models"
	"database/sql"
	"errors"
)

const (
	OrderTypeMarket = "market"
	OrderTypeLimit  = "limit"

	OrderSideBuy  = "buy"
	OrderSideSell = "sell"
)

type TradingService struct {
	db            *sql.DB
	krakenClient  *kraken.Client
	walletService *WalletService
}

// NewTradingService 创建新的交易服务
func NewTradingService(db *sql.DB, krakenClient *kraken.Client, walletService *WalletService) *TradingService {
	return &TradingService{
		db:            db,
		krakenClient:  krakenClient,
		walletService: walletService,
	}
}

// GetPrice 获取指定货币对的价格
func (s *TradingService) GetPrice(pair string) (float64, error) {
	return s.krakenClient.GetPrice(pair)
}

// GetTradablePairs 获取可交易的货币对
func (s *TradingService) GetTradablePairs() ([]string, error) {
	return s.krakenClient.GetTradablePairs()
}

// PlaceOrder 下单
func (s *TradingService) PlaceOrder(userID int64, pair, orderType, side string, amount, price float64) (int64, error) {
	// 解析货币对
	baseCurrency, quoteCurrency, err := parsePair(pair)
	if err != nil {
		return 0, err
	}

	// 检查余额
	if side == OrderSideBuy {
		balance, err := s.walletService.GetWalletBalance(userID, quoteCurrency)
		if err != nil {
			return 0, err
		}

		totalCost := amount * price
		if balance < totalCost {
			return 0, errors.New("insufficient balance")
		}
	} else {
		balance, err := s.walletService.GetWalletBalance(userID, baseCurrency)
		if err != nil {
			return 0, err
		}

		if balance < amount {
			return 0, errors.New("insufficient balance")
		}
	}

	// 创建交易记录
	txID, err := models.CreateTransaction(s.db, userID, side, pair, amount, price, models.TransactionStatusPending)
	if err != nil {
		return 0, err
	}

	// 通过Kraken API下单
	_, err = s.krakenClient.PlaceOrder(pair, orderType, side, amount, price)
	if err != nil {
		_ = models.UpdateTransactionStatus(s.db, txID, models.TransactionStatusFailed)
		return 0, err
	}

	// 更新钱包余额
	if side == OrderSideBuy {
		err = s.walletService.UpdateWalletBalance(userID, quoteCurrency, -amount*price)
		if err != nil {
			return 0, err
		}

		err = s.walletService.UpdateWalletBalance(userID, baseCurrency, amount)
		if err != nil {
			return 0, err
		}
	} else {
		err = s.walletService.UpdateWalletBalance(userID, baseCurrency, -amount)
		if err != nil {
			return 0, err
		}

		err = s.walletService.UpdateWalletBalance(userID, quoteCurrency, amount*price)
		if err != nil {
			return 0, err
		}
	}

	// 更新交易状态为完成
	err = models.UpdateTransactionStatus(s.db, txID, models.TransactionStatusCompleted)
	if err != nil {
		return 0, err
	}

	return txID, nil
}

// GetUserTransactions 获取用户的交易记录
func (s *TradingService) GetUserTransactions(userID int64) ([]*models.Transaction, error) {
	return models.GetTransactionsByUser(s.db, userID)
}

// GetTransactionStatus 获取交易状态
func (s *TradingService) GetTransactionStatus(txID int64) (string, error) {
	tx, err := models.GetTransactionByID(s.db, txID)
	if err != nil {
		return "", err
	}
	return tx.Status, nil
}

// 辅助函数：解析货币对
func parsePair(pair string) (baseCurrency, quoteCurrency string, err error) {
	if len(pair) < 6 {
		return "", "", errors.New("invalid pair format")
	}

	// 假设后三个字符是报价货币
	quoteCurrency = pair[len(pair)-3:]

	// 前面的是基础货币
	baseCurrency = pair[:len(pair)-3]

	// 特殊处理Kraken的XBT标记
	if baseCurrency == "XBT" {
		baseCurrency = "BTC"
	}

	return baseCurrency, quoteCurrency, nil
}
