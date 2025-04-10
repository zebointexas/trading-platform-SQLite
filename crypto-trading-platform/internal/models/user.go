package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUser 创建新用户
func CreateUser(db *sql.DB, username, password, email string) (int64, error) {
	stmt, err := db.Prepare(`
		INSERT INTO users(username, password, email) 
		VALUES(?, ?, ?)
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(username, password, email)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// GetUserByID 通过ID获取用户
func GetUserByID(db *sql.DB, id int64) (*User, error) {
	var user User
	err := db.QueryRow(`
		SELECT id, username, password, email, created_at, updated_at 
		FROM users 
		WHERE id = ?
	`, id).Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername 通过用户名获取用户
func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	var user User
	err := db.QueryRow(`
		SELECT id, username, password, email, created_at, updated_at 
		FROM users 
		WHERE username = ?
	`, username).Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser 更新用户信息
func UpdateUser(db *sql.DB, user *User) error {
	stmt, err := db.Prepare(`
		UPDATE users 
		SET password = ?, email = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.Password, user.Email, user.ID)
	return err
}