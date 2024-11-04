package app

import (
	"fmt"
	"log"
	"os"

	"github.com/VikaPaz/algalar/internal/models"
	"github.com/VikaPaz/algalar/internal/repository"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func Run() {
	logger := NewLogger(logrus.DebugLevel, &logrus.TextFormatter{
		FullTimestamp: true,
	})

	if err := godotenv.Overload(); err != nil {
		logger.Errorf("Error loading .env file: %e", models.ErrLoadEnvFailed)
		return
	}

	confPostgres := repository.Config{
		Host:     os.Getenv("HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		User:     os.Getenv("USER"),
		Password: os.Getenv("PASSWORD"),
		Dbname:   os.Getenv("DB_NAME"),
	}

	dbConn, err := repository.Connection(confPostgres)
	if err != nil {
		logger.Errorf("Error connecting to database: %v, config: %v", err, confPostgres)
		return
	}
	logger.Infof("Connected to PostgreSQL")

	repo := repository.NewRepository(dbConn, logger)

	user := models.User{
		INN:        1234567890,
		Name:       "Имя",
		Surname:    "Фамилия",
		MiddleName: "Отчество",
		Login:      "example_login",
		Password:   "example_password",
		Timezone:   "UTC",
	}

	userID, err := repo.CreateUser(user)
	if err != nil {
		log.Fatalf("Error creating user: %v", err)
	}

	fmt.Println("Создан пользователь с ID:", userID)
}

func NewLogger(level logrus.Level, formatter logrus.Formatter) *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(level)
	logger.SetFormatter(formatter)
	return logger
}
