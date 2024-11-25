package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
)

type AuthRepository interface {
	CreateRefreshToken(userID string, refreshToken string, expiration time.Time) error
	SelectRefresToken(refreshToken string) (bool, error)
	UpdateRefreshToken(userID string, token string, expiration time.Time) error
	GetIDByLoginAndPassword(email, password string) (string, error)
}

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type AuthService struct {
	repo AuthRepository
	log  *logrus.Logger
	conf Config
}

type Config struct {
	Salt       string
	SigningKey string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func NewService(conf Config, repo AuthRepository, log *logrus.Logger) *AuthService {
	return &AuthService{
		repo: repo,
		log:  log,
		conf: conf,
	}
}

func (s *AuthService) GenerateAccessToken(userID string) (string, error) {
	expirationTime := time.Now().Add(s.conf.AccessTTL)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "auth-server",
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	sign, err := token.SignedString([]byte(s.conf.SigningKey))
	if err != nil {
		s.log.Errorf("failed to sign access: %v", err)
		return "", err
	}

	return sign, err
}

func (s *AuthService) GenerateRefreshToken(userID string) (string, *time.Time, error) {
	expirationTime := time.Now().Add(s.conf.AccessTTL)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "auth-server",
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	sign, err := token.SignedString([]byte(s.conf.SigningKey))
	if err != nil {
		s.log.Errorf("failed to sign refresh: %v", err)
		return "", nil, err
	}

	return sign, &expirationTime, nil
}

func (s *AuthService) ValidateRefreshToken(refreshToken string) (*string, error) {
	token, err := jwt.ParseWithClaims(refreshToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.conf.SigningKey), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid refresh token")
	}

	return &claims.UserID, nil
}

func (s *AuthService) SaveRefreshToken(userID string, refreshToken string, expiration time.Time) error {
	return s.repo.CreateRefreshToken(userID, refreshToken, expiration)
}

func (s *AuthService) GetUserID(login, password string) (string, error) {
	return s.repo.GetIDByLoginAndPassword(login, password)
}
