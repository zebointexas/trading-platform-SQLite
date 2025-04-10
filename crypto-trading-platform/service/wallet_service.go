package service

import (
	"crypto-trading-platform/internal/kraken"
	"crypto-trading-platform/internal/models"
	"database/sql"
	"errors"
)

type WalletService struct {
	db          *sql.DB
	krakenClient *kraken.Client
}

// NewWalletService 创建新的钱包服务
func NewWalletService(db *sql.DB, krakenClient *kraken.Client) *WalletService {
	return &WalletService{
		db:          db,
		krakenClient: krakenClient,
	}
}

// GetUserWallets 获取用户的所有钱包
func (s *WalletService) GetUserWallets(userID int64) ([]*models.Wallet, error) {
	return models.GetWalletsByUser(s.db, userID)
}

// GetWalletBalance 获取用户特定货币的钱包余额
func (s *WalletService) GetWalletBalance(userID int64, currency string) (float64, error) {
	wallet, err := models.GetWalletByUserAndCurrency(s.db, userID, currency)
	if err != nil {
		if err == sql.ErrNoRows {
			// 如果钱包不存在，创建一个新钱包
			_, err = models.CreateWallet(s.db, userID, currency)
			if err != nil {
				return 0, err
			}
			return 0, nil
		}
		return 0, err
	}

	return wallet.Balance, nil
}

// UpdateWalletBalance 更新钱包余额
func (s *WalletService) UpdateWalletBalance(userID int64, currency string, amount float64) error {
	wallet, err := models.GetWalletByUserAndCurrency(s.db, userID, currency)
	if err != nil {
		return err
	}

	newBalance := wallet.Balance + amount
	if newBalance < 0 {
		return errors.New("insufficient balance")
	}

	return models.UpdateWalletBalance(s.db, wallet.ID, newBalance)
}

// SyncWalletBalances 同步用户钱包余额与Kraken账户
func (s *WalletService) SyncWalletBalances(userID int64) error {
	// 获取Kraken账户余额
	krakenBalances, err := s.krakenClient.GetBalance()
	if err != nil {
		return err
	}

	// 获取用户钱包
	wallets, err := models.GetWalletsByUser(s.db, userID)
	if err != nil {
		return err
	}

	// 更新每个钱包的余额
	for _, wallet := range wallets {
		if balance, ok := krakenBalances[wallet.Currency]; ok {
			err = models.UpdateWalletBalance(s.db, wallet.ID, balance)
			if err != nil {
				return err
			}
		}
	}

	return nil
}