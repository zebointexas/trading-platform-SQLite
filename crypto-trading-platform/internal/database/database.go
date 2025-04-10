package database

import (
	"database/sql"
	"log"
	_ "github.com/mattn/go-sqlite3"
)

// InitDB 初始化SQLite数据库连接，并创建必要的表
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	// 创建用户表
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Printf("Error creating users table: %v", err)
		return nil, err
	}

	// 创建钱包表
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS wallets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		currency TEXT NOT NULL,
		balance REAL DEFAULT 0.0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		UNIQUE(user_id, currency)
	)`)
	if err != nil {
		log.Printf("Error creating wallets table: %v", err)
		return nil, err
	}

	// 创建交易记录表
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		transaction_type TEXT NOT NULL,
		currency TEXT NOT NULL,
		amount REAL NOT NULL,
		price REAL NOT NULL,
		status TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	)`)
	if err != nil {
		log.Printf("Error creating transactions table: %v", err)
		return nil, err
	}

	return db, nil
}