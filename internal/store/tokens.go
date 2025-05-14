package store

import (
	"database/sql"
	"time"

	"github.com/ruhan/internal/tokens"
)

type PostgersTokenStore struct {
	db *sql.DB
}

func NewPostgresTokenStore(db *sql.DB) *PostgersTokenStore {
	return &PostgersTokenStore{
		db,
	}
}

type TokenStore interface {
	Insert(token *tokens.Token) error
	CreateNewToken(userID int, ttl time.Duration, scope string) (*tokens.Token, error)
	DeleteAllTokens(userId int, scope string) error
}

func (t *PostgersTokenStore) CreateNewToken(userID int, ttl time.Duration, scope string) (*tokens.Token, error) {
	token, err := tokens.GenerateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}
	err = t.Insert(token)
	return token, err
}

func (t *PostgersTokenStore) Insert(token *tokens.Token) error {
	query := `
			INSERT INTO tokens(hash, user_id, expiry, scope)
			VALUES ($1, $2, $3, $4)
		`

	_, err := t.db.Exec(query, token.Hash, token.UserID, token.Expiry, token.Scope)
	if err != nil {
		return err
	}
	return nil
}

func (t *PostgersTokenStore) DeleteAllTokens(userId int, scope string) error {
	query := `
		DELETE FROM tokens
		WHERE scope = $1 AND user_id = $2
	`

	_, err := t.db.Exec(query, scope, userId)
	return err
}
