package service

import (
	"time"

	"github.com/VikaPaz/algalar/internal/models"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
)

const (
	salt       = "hjqrhjqw124617ajfhajs"
	signingKey = "qrkjk#4#%35FSFJlja#4353KSFjH"
	accessTTL  = 1 / 3 * time.Minute
	refreshTTL = 1000 * time.Hour
)

type tokenClaims struct {
	jwt.StandardClaims
	Data string `json:"data"`
}

type Repository interface {
	CreateUser(user models.User) (string, error)
	GetById(userID string) (models.User, error)
	ChangePassword(userID, newPassword string) error
	GetIDByEmailAndPassword(email, password string) (string, error)
	CreateCar(car models.Car) (string, error)
	CreateWheel(wheel models.Wheel) (string, error)
	GetWheelById(wheelID string) (models.Wheel, error)
	ChangeWheel(wheelID string, wheel models.Wheel) error
}

type Service struct {
	repo Repository
	log  *logrus.Logger
}

func NewService(repo Repository, log *logrus.Logger) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

func NewToken(data string, ttl time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(ttl).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		data,
	})
	return token.SignedString([]byte(signingKey))
}

func (s *Service) UserLogin(login string, password string) (string, string, error) {
	id, err := s.repo.GetIDByEmailAndPassword(login, password)
	if err != nil {
		s.log.Debugf("Invalid s.login or password: %s", login)
		return "", "", err
	}
	s.log.Debug("created new user")

	access, err := NewToken(id, accessTTL)
	if err != nil {
		return "", "", err
	}

	refresh, err := NewToken(id, refreshTTL)
	if err != nil {
		return "", "", err
	}

	s.log.Debugf("User s.logged in successfully: %s", login)
	return access, refresh, nil
}

func (s *Service) RefreshToken(token string) (string, string, error) {
	id := "59186207-1269-4986-ba64-93f76cb288d4"
	access, err := NewToken(id, accessTTL)
	if err != nil {
		return "", "", err
	}

	refresh, err := NewToken(id, refreshTTL)
	if err != nil {
		return "", "", err
	}
	s.log.Debugf("Token refreshed successfully: %s", token)
	return access, refresh, nil
}

func (s *Service) RegisterUser(user models.User) error {
	// existingUser, err := s.repo.GetById(user.Id)
	// if err == nil && existingUser.ID != "" {
	// 	s.log.Debugf("User already exists: %s", user.Login)
	// 	return errors.New("user already exists")
	// }

	_, err := s.repo.CreateUser(user)
	if err != nil {
		s.log.Debugf("Error creating user: %s", user.Login)
		return err
	}

	s.log.Debugf("User registered successfully: %s", user.Login)
	return nil
}

func (s *Service) RegisterAuto(car models.Car) error {
	_, err := s.repo.CreateCar(car)
	if err != nil {
		s.log.Debugf("Error registering Auto: %v", car)
		return err
	}

	s.log.Debugf("Auto registered successfully: %v", car)
	return nil
}

func (s *Service) UpdateUserPassword(token string, newPassword string) error {
	// TODO: Parsing token

	// userID, err := s.repo.GetIDByEmailAndPassword(login, newPassword)
	// if err != nil {
	// 	s.log.Debugf("Invalid s.login: %s", login)
	// 	return err
	// }

	userID := "f5dd3092-e139-4295-8f53-5d9ff1d6da31"

	err := s.repo.ChangePassword(userID, newPassword)
	if err != nil {
		s.log.Debugf("Error updating user password: %s", userID)
		return err
	}

	s.log.Debugf("User password updated successfully: %s", userID)
	return nil
}

func (s *Service) GetUserDetails(id string) (interface{}, error) {
	userID := "f5dd3092-e139-4295-8f53-5d9ff1d6da31"
	user, err := s.repo.GetById(userID)
	if err != nil {
		s.log.Debugf("User not found: %s", id)
		return nil, err
	}

	s.log.Debugf("User details fetched successfully: %s", id)
	return user, nil
}

func (s *Service) RegisterWheel(wheel models.Wheel) error {
	_, err := s.repo.CreateWheel(wheel)
	if err != nil {
		s.log.Debugf("Error registering wheel: %v", wheel)
		return err
	}

	s.log.Debugf("Wheel registered successfully: %v", wheel)
	return nil
}

func (s *Service) UpdateWheelData(wheel models.Wheel) error {
	err := s.repo.ChangeWheel(wheel.ID, wheel)
	if err != nil {
		s.log.Debugf("Error updating wheel data: %v", wheel)
		return err
	}

	s.log.Debugf("Wheel data updated successfully: %v", wheel)
	return nil
}

func (s *Service) GetWheelData(id string) (interface{}, error) {
	wheel, err := s.repo.GetWheelById(id)
	if err != nil {
		s.log.Debugf("Wheel not found: %s", id)
		return nil, err
	}

	s.log.Debugf("Wheel data fetched successfully: %s", id)
	return wheel, nil
}

func (s *Service) GenerateReport(params models.GetReportParams) (interface{}, error) {
	return nil, nil
}

func (s *Service) GetSensorData(params models.GetSensorParams) (interface{}, error) {
	return nil, nil
}
