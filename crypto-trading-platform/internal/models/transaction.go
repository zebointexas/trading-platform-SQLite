package models

import (
	"database/sql"
	"time"
)

const (
	TransactionTypeBuy  = "buy"
	TransactionTypeSell = "sell"
	
	TransactionStatusPending   = "pending"
	TransactionStatusCompleted = "completed"
	TransactionStatusFailed    = "failed"
)

type Transaction struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	TransactionType string    `json:"transaction_type"`
	Currency        string    `json:"currency"`
	Amount          float64   `json:"amount"`
	Price           float64   `json:"price"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CreateTransaction 创建新交易记录
func CreateTransaction(db *sql.DB, userID int64, transactionType, currency string, amount, price float64, status string) (int64, error) {
	stmt, err := db.Prepare(`
		INSERT INTO transactions(user_id, transaction_type, currency, amount, price, status) 
		VALUES(?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(userID, transactionType, currency, amount, price, status)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// GetTransactionByID 通过ID获取交易记录
func GetTransactionByID(db *sql.DB, id int64) (*Transaction, error) {
	var tx Transaction
	err := db.QueryRow(`
		SELECT id, user_id, transaction_type, currency, amount, price, status, created_at, updated_at 
		FROM transactions 
		WHERE id = ?
	`, id).Scan(&tx.ID, &tx.UserID, &tx.TransactionType, &tx.Currency, &tx.Amount, &tx.Price, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// GetTransactionsByUser 获取用户的所有交易记录
func GetTransactionsByUser(db *sql.DB, userID int64) ([]*Transaction, error) {
	rows, err := db.Query(`
		SELECT id, user_id, transaction_type, currency, amount, price, status, created_at, updated_at 
		FROM transactions 
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(&tx.ID, &tx.UserID, &tx.TransactionType, &tx.Currency, &tx.Amount, &tx.Price, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, &tx)
	}
	
	return transactions, nil
}

// UpdateTransactionStatus 更新交易记录状态
func UpdateTransactionStatus(db *sql.DB, id int64, status string) error {
	stmt, err := db.Prepare(`
		UPDATE transactions 
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(status, id)
	return err
}