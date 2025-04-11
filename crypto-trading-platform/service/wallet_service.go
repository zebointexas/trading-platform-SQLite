package service

import (
	"crypto-trading-platform/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	krakenapi "github.com/beldur/kraken-go-api-client" // Corrected import
)

type WalletService struct {
	db           *sql.DB
	krakenClient *krakenapi.KrakenAPI // Updated to use the correct type
}

// NewWalletService creates a new wallet service
func NewWalletService(db *sql.DB, krakenClient *krakenapi.KrakenAPI) *WalletService {
	return &WalletService{
		db:           db,
		krakenClient: krakenClient,
	}
}

// GetUserWallets gets all wallets for a user
func (s *WalletService) GetUserWallets(userID int64) ([]*models.Wallet, error) {
	return models.GetWalletsByUser(s.db, userID)
}

// GetWalletBalance gets the wallet balance for a specific currency
func (s *WalletService) GetWalletBalance(userID int64, currency string) (float64, error) {
	wallet, err := models.GetWalletByUserAndCurrency(s.db, userID, currency)
	if err != nil {
		if err == sql.ErrNoRows {
			// If wallet doesn’t exist, create a new one
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

// UpdateWalletBalance updates the wallet balance
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

// SyncWalletBalances syncs wallet balances with Kraken account
// SyncWalletBalances 使用反射来提取 BalanceResponse 中的字段
func (s *WalletService) SyncWalletBalances(userID int64) error {
	fmt.Printf("***********************\n")

	// Get Kraken account balance
	balanceResp, err := s.krakenClient.Balance() // Returns *BalanceResponse, error
	if err != nil {
		return err
	}

	// Get user wallets
	wallets, err := models.GetWalletsByUser(s.db, userID)
	if err != nil {
		return err
	}

	// Get reflect.Value of balanceResp
	val := reflect.ValueOf(*balanceResp)

	// Update each wallet’s balance
	for _, wallet := range wallets {

		// Map Kraken currency codes to your system's currency codes
		krakenCurrency := wallet.Currency

		// Kraken uses "ZUSD" for USD, "XXBT" for BTC, etc.
		switch wallet.Currency {
		case "USD":
			krakenCurrency = "ZUSD"
		case "BTC":
			krakenCurrency = "XXBT"
		case "ETH":
			krakenCurrency = "XETH"
		// Add more mappings as needed
		default:
			// Skip if currency mapping is not defined
			continue
		}

		// Use reflection to get the field by name
		fieldValue := val.FieldByName(krakenCurrency) // Get the value of the field dynamically
		if fieldValue.IsValid() {
			// If the field exists, extract its value
			if balance, ok := fieldValue.Interface().(float64); ok {
				// Update the wallet balance
				err = models.UpdateWalletBalance(s.db, wallet.ID, balance)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
