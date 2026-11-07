package main

import (
	boardv1 "database-service/api/board/v1"
	taskv1 "database-service/api/task/v1"
	userv1 "database-service/api/user/v1"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"task-service/repository"
	"task-service/security"
	"task-service/service"
	"task-service/transport"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	databaseServiceAddressEnv = "DATABASE_SERVICE_ADDRESS"
	httpServerAddress         = ":8080"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	databaseServiceAddress := strings.TrimSpace(os.Getenv(databaseServiceAddressEnv))
	if databaseServiceAddress == "" {
		return fmt.Errorf("%s is not set", databaseServiceAddressEnv)
	}
	publicKeyFile := strings.TrimSpace(os.Getenv("JWT_PUBLIC_KEY_FILE"))
	if publicKeyFile == "" {
		return fmt.Errorf("JWT_PUBLIC_KEY_FILE is not set")
	}
	publicKey, err := security.LoadEd25519PublicKey(publicKeyFile)
	if err != nil {
		return err
	}
	verifier, err := security.NewJWTVerifier(
		publicKey,
		envOrDefault("JWT_ISSUER", "taskflow-auth"),
		envOrDefault("JWT_AUDIENCE", "taskflow-api"),
	)
	if err != nil {
		return err
	}

	connection, err := grpc.NewClient(
		databaseServiceAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("create database-service gRPC client: %w", err)
	}
	defer connection.Close()

	grpcClient := taskv1.NewTaskServiceClient(connection)
	taskRepository := repository.NewGRPCTaskRepository(grpcClient)
	taskService := service.NewTaskService(taskRepository)
	taskHandler := transport.NewTaskHandler(taskService)
	router := transport.NewRouter(taskHandler)
	boardClient := boardv1.NewBoardServiceClient(connection)
	boardRepository := repository.NewGRPCBoardRepository(boardClient)
	boardService := service.NewBoardService(boardRepository)
	boardHandler := transport.NewBoardHandler(boardService)
	transport.RegisterBoardRoutes(router, boardHandler)
	userClient := userv1.NewUserServiceClient(connection)
	userRepository := repository.NewGRPCUserRepository(userClient)
	userService := service.NewUserService(userRepository)
	userHandler := transport.NewUserHandler(userService)
	transport.RegisterUserRoutes(router, userHandler)
	router.Use(transport.RequireAuth(verifier))

	log.Printf("HTTP listening on %s", httpServerAddress)
	if err := http.ListenAndServe(httpServerAddress, router); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP: %w", err)
	}
	return nil
}

func envOrDefault(name string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
