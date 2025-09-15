package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID         int
	Email      string
	Name       string
	Picture    string
	Provider   string
	ProviderID string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (db *Database) FindOrCreateUser(user *User) (*User, error) {
	var existingUser User
	err := db.QueryRow(`
		SELECT id, email, name, picture, provider, provider_id, created_at, updated_at 
		FROM users 
		WHERE provider = $1 AND provider_id = $2`,
		user.Provider, user.ProviderID).Scan(
		&existingUser.ID, &existingUser.Email, &existingUser.Name,
		&existingUser.Picture, &existingUser.Provider, &existingUser.ProviderID,
		&existingUser.CreatedAt, &existingUser.UpdatedAt,
	)

	if err == nil {
		return &existingUser, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	// User doesn't exist, create a new one
	err = db.QueryRow(`
		INSERT INTO users (email, name, picture, provider, provider_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`,
		user.Email, user.Name, user.Picture, user.Provider, user.ProviderID,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (db *Database) GetUserByID(id int) (*User, error) {
	var user User
	err := db.QueryRow(`
		SELECT id, email, name, picture, provider, provider_id, created_at, updated_at 
		FROM users 
		WHERE id = $1`, id).Scan(
		&user.ID, &user.Email, &user.Name,
		&user.Picture, &user.Provider, &user.ProviderID,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}