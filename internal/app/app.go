package app

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/VikaPaz/algalar/internal/repository"
	"github.com/VikaPaz/algalar/internal/server"
	"github.com/VikaPaz/algalar/internal/server/rest"
	"github.com/VikaPaz/algalar/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func Run() {
	logger := NewLogger(logrus.InfoLevel, &logrus.TextFormatter{
		FullTimestamp: true,
	})

	if err := godotenv.Overload("env/.env"); err != nil {
		logger.Errorf("Error loading .env file: %v", err)
		return
	}

	confPostgres := repository.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		User:     os.Getenv("USER"),
		Password: os.Getenv("PASSWORD"),
		Dbname:   os.Getenv("DB_NAME"),
	}

	confService := service.Config{
		Salt:       os.Getenv("JWT_SALT"),
		SigningKey: os.Getenv("JWT_SIGNING_KEY"),
		AccessTTL:  10 * time.Minute,
		RefreshTTL: 1000 * time.Hour,
	}

	logger.Debugf("config: %v", confPostgres)

	port := os.Getenv("PORT")

	dbConn, err := repository.Connection(confPostgres)
	if err != nil {
		logger.Errorf("Error connecting to database: %v, config: %v", err, confPostgres)
		return
	}
	logger.Infof("Connected to PostgreSQL")

	repo := repository.NewRepository(dbConn, logger)

	svc := service.NewService(confService, repo, logger)

	svr := server.NewServer(svc, logger)

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// TODO: registration with options
	options := rest.ChiServerOptions{
		BaseRouter: r,
	}
	router := rest.HandlerWithOptions(svr, options)

	// router := rest.Handler(svr)

	go func() {
		if err := http.ListenAndServe(":"+port, router); err != nil {
			logger.Errorf("Cann't run server: %v", err)
			return
		}
		logger.Infof("Run server on port: %s", port)
	}()
	logger.Infof("Rest server is running on port: %s", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}

func NewLogger(level logrus.Level, formatter logrus.Formatter) *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(level)
	logger.SetFormatter(formatter)
	return logger
}
