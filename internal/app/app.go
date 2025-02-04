package app

import (
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/VikaPaz/algalar/internal/repository"
	authRepository "github.com/VikaPaz/algalar/internal/repository/auth"
	"github.com/VikaPaz/algalar/internal/server"
	"github.com/VikaPaz/algalar/internal/server/rest"
	"github.com/VikaPaz/algalar/internal/service"
	authService "github.com/VikaPaz/algalar/internal/service/auth"
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

	logger.Debugf("config: %v", confPostgres)

	port := os.Getenv("PORT")

	dbConn, err := repository.Connection(confPostgres)
	if err != nil {
		logger.Errorf("Error connecting to database: %v, config: %v", err, confPostgres)
		return
	}
	logger.Infof("Connected to PostgreSQL")

	repo := repository.NewRepository(dbConn, logger)

	authRepo := authRepository.NewRepository(dbConn, logger)

	svc := service.NewService(repo, logger)

	var accessSigningKey, refreshSigningKey string
	var accsessTTL, refreshTTL int
	accessSigningKey = os.Getenv("JWT_ACCESS_SIGNING_KEY")
	accsessTTL, err = strconv.Atoi(os.Getenv("JWT_ACCESS_TTL_SEC"))
	if err != nil {
		logger.Errorf("Error loading JWT_ACCESS_TTL from .env file: %v", err)
		return
	}
	refreshSigningKey = os.Getenv("JWT_REFRESH_SIGNING_KEY")
	refreshTTL, err = strconv.Atoi(os.Getenv("JWT_REFRESH_TTL_MIN"))
	if err != nil {
		logger.Errorf("Error loading JWT_REFRESH_TTL from .env file: %v", err)
		return
	}

	confAuth := authService.Config{
		AccessSigningKey:  accessSigningKey,
		RefreshSigningKey: refreshSigningKey,
		AccessTTL:         time.Duration(accsessTTL) * time.Second,
		RefreshTTL:        time.Duration(refreshTTL) * time.Minute,
	}

	auth := authService.NewService(confAuth, authRepo, logger)

	confServer := server.Config{
		SigningKey: accessSigningKey,
	}

	svr := server.NewServer(confServer, svc, auth, logger)

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	options := rest.ChiServerOptions{
		BaseRouter: r,
	}
	router := rest.HandlerWithOptions(svr, options)

	go func() {
		if err := http.ListenAndServeTLS(":"+port, "env/server.crt", "env/server.key", router); err != nil {
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
