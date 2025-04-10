package models

import (
	"database/sql"
	"time"
)

type Wallet struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Currency  string    `json:"currency"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateWallet 创建新钱包
func CreateWallet(db *sql.DB, userID int64, currency string) (int64, error) {
	stmt, err := db.Prepare(`
		INSERT INTO wallets(user_id, currency) 
		VALUES(?, ?)
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(userID, currency)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// GetWalletByUserAndCurrency 获取用户特定货币的钱包
func GetWalletByUserAndCurrency(db *sql.DB, userID int64, currency string) (*Wallet, error) {
	var wallet Wallet
	err := db.QueryRow(`
		SELECT id, user_id, currency, balance, created_at, updated_at 
		FROM wallets 
		WHERE user_id = ? AND currency = ?
	`, userID, currency).Scan(&wallet.ID, &wallet.UserID, &wallet.Currency, &wallet.Balance, &wallet.CreatedAt, &wallet.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// GetWalletsByUser 获取用户的所有钱包
func GetWalletsByUser(db *sql.DB, userID int64) ([]*Wallet, error) {
	rows, err := db.Query(`
		SELECT id, user_id, currency, balance, created_at, updated_at 
		FROM wallets 
		WHERE user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallets []*Wallet
	for rows.Next() {
		var wallet Wallet
		if err := rows.Scan(&wallet.ID, &wallet.UserID, &wallet.Currency, &wallet.Balance, &wallet.CreatedAt, &wallet.UpdatedAt); err != nil {
			return nil, err
		}
		wallets = append(wallets, &wallet)
	}
	
	return wallets, nil
}

// UpdateWalletBalance 更新钱包余额
func UpdateWalletBalance(db *sql.DB, id int64, balance float64) error {
	stmt, err := db.Prepare(`
		UPDATE wallets 
		SET balance = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(balance, id)
	return err
}