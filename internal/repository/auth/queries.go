package auth

import (
	"database/sql"
	"time"

	"github.com/VikaPaz/algalar/internal/models"
	"github.com/sirupsen/logrus"
)

type Repository struct {
	conn *sql.DB
	log  *logrus.Logger
}

func NewRepository(conn *sql.DB, logger *logrus.Logger) *Repository {
	return &Repository{
		conn: conn,
		log:  logger,
	}
}

func (r *Repository) CreateRefreshToken(userID string, refreshToken string, expiration time.Time) error {
	query := `
        INSERT INTO refresh_store (user_id, token, expiration)
        VALUES ($1, $2, $3)
		returning *`

	_, err := r.conn.Exec(query, userID, refreshToken, expiration)
	if err != nil {
		r.log.Errorf("failed to create token: %v", err)
	}
	return nil
}

func (r *Repository) SelectRefresToken(refreshToken string) (bool, error) {
	var storedHash string
	var expiration time.Time

	err := r.conn.QueryRow(`
        SELECT token, expiration
        FROM refresh_store
        WHERE token = $1`,
		refreshToken,
	).Scan(&storedHash, &expiration)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *Repository) UpdateRefreshToken(userID string, token string, expiration time.Time) error {
	query := `
	update refresh_store 
	set token = $1, expiration = $2
	where user_id = $3 
	`

	_, err := r.conn.Exec(query, userID, token, expiration)
	return err
}

func (r *Repository) GetIDByLoginAndPassword(email, password string) (string, error) {
	var userID string
	query := `
        SELECT id
        FROM  users
        WHERE login = $1 AND password = $2`

	err := r.conn.QueryRow(query, email, password).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", models.ErrNoContent
		}
		return "", err
	}
	return userID, nil
}
