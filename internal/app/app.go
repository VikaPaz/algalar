package app

import (
	"net/http"
	"os"

	"github.com/VikaPaz/algalar/internal/models"
	"github.com/VikaPaz/algalar/internal/repository"
	"github.com/VikaPaz/algalar/internal/server"
	"github.com/VikaPaz/algalar/internal/server/rest"
	"github.com/VikaPaz/algalar/internal/service"
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

	port := os.Getenv("PORT")

	dbConn, err := repository.Connection(confPostgres)
	if err != nil {
		logger.Errorf("Error connecting to database: %v, config: %v", err, confPostgres)
		return
	}
	logger.Infof("Connected to PostgreSQL")

	repo := repository.NewRepository(dbConn, logger)

	svc := service.NewService(repo, logger)

	svr := server.NewServer(svc, logger)

	router := rest.Handler(svr)

	// go func() {
	if err := http.ListenAndServe(":"+port, router); err != nil {
		logger.Errorf("Cann't run server: %v", err)
		return
	}
	logger.Infof("Run server on port: %s", port)
	// }()
}

func NewLogger(level logrus.Level, formatter logrus.Formatter) *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(level)
	logger.SetFormatter(formatter)
	return logger
}
