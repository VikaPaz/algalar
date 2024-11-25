package auth

import (
	"database/sql"
	"time"

	"github.com/VikaPaz/algalar/internal/models"
	"github.com/google/uuid"
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

func (r *Repository) GetRefresToken(userID string) (string, error) {
	var storedToken string
	var expiration time.Time

	parsedUUID, err := uuid.Parse(userID)
	if err != nil {
		return "", err
	}

	err = r.conn.QueryRow(`
        SELECT token, expiration
        FROM refresh_store
        WHERE user_id = $1`,
		parsedUUID,
	).Scan(&storedToken, &expiration)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", models.ErrNoContent
		}
		return "", err
	}

	return storedToken, nil
}

func (r *Repository) UpdateRefreshToken(userID string, token string, expiration time.Time) error {
	query := `
	update refresh_store 
	set token = $1, expiration = $2
	where user_id = $3 
	`

	_, err := r.conn.Exec(query, token, expiration, userID)
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
