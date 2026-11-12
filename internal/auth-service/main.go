package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"auth-service/repository"
	"auth-service/service"
	"auth-service/transport"
	authv1 "database-service/api/auth/v1"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	databaseAddress := strings.TrimSpace(os.Getenv("DATABASE_SERVICE_ADDRESS"))
	privateKeyFile := strings.TrimSpace(os.Getenv("JWT_PRIVATE_KEY_FILE"))
	if databaseAddress == "" || privateKeyFile == "" {
		return errors.New("DATABASE_SERVICE_ADDRESS and JWT_PRIVATE_KEY_FILE are required")
	}
	privateKey, err := service.LoadEd25519PrivateKey(privateKeyFile)
	if err != nil {
		return err
	}
	issuer := envOrDefault("JWT_ISSUER", "taskflow-auth")
	audience := envOrDefault("JWT_AUDIENCE", "taskflow-api")
	accessTTL, err := durationEnv("ACCESS_TOKEN_TTL", 15*time.Minute)
	if err != nil {
		return err
	}
	refreshTTL, err := durationEnv("REFRESH_TOKEN_TTL", 30*24*time.Hour)
	if err != nil {
		return err
	}
	tokens, err := service.NewJWTManager(privateKey, issuer, audience, accessTTL)
	if err != nil {
		return err
	}
	redisDB, err := strconv.Atoi(envOrDefault("REDIS_DB", "0"))
	if err != nil || redisDB < 0 {
		return errors.New("REDIS_DB must be a non-negative integer")
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     envOrDefault("REDIS_ADDRESS", "localhost:6379"),
		Password: strings.TrimSpace(os.Getenv("REDIS_PASSWORD")),
		DB:       redisDB,
	})
	defer redisClient.Close()
	pingContext, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPing()
	if err := redisClient.Ping(pingContext).Err(); err != nil {
		return fmt.Errorf("ping Redis: %w", err)
	}

	connection, err := grpc.NewClient(databaseAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("create database-service gRPC client: %w", err)
	}
	defer connection.Close()

	authRepository := repository.NewGRPCAuthRepository(authv1.NewAuthPersistenceServiceClient(connection))
	sessionRepository := repository.NewRedisSessionRepository(redisClient)
	authService, err := service.NewAuthService(authRepository, sessionRepository, service.NewBcryptHasher(12), tokens, refreshTTL)
	if err != nil {
		return err
	}
	cookieSecure, err := strconv.ParseBool(envOrDefault("AUTH_COOKIE_SECURE", "false"))
	if err != nil {
		return fmt.Errorf("parse AUTH_COOKIE_SECURE: %w", err)
	}
	address := envOrDefault("AUTH_SERVICE_ADDRESS", ":8090")
	server := &http.Server{
		Addr:              address,
		Handler:           transport.NewRouter(authService, cookieSecure),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("auth HTTP listening on %s", address)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve auth HTTP: %w", err)
	}
	return nil
}

func envOrDefault(name string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func durationEnv(name string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return 0, fmt.Errorf("%s must be a positive Go duration", name)
	}
	return duration, nil
}
