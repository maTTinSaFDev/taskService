FROM golang:1.27-bookworm AS build

WORKDIR /workspace

COPY internal/core/go.mod ./internal/core/go.mod
COPY internal/database-service/go.mod internal/database-service/go.sum ./internal/database-service/
COPY internal/auth-service/go.mod internal/auth-service/go.sum ./internal/auth-service/

WORKDIR /workspace/internal/auth-service
RUN go mod download

WORKDIR /workspace
COPY internal/database-service/api ./internal/database-service/api
COPY internal/auth-service ./internal/auth-service

WORKDIR /workspace/internal/auth-service
RUN CGO_ENABLED=0 go build -o /app/auth-service .

FROM golang:1.27-bookworm
COPY --from=build /app/auth-service /auth-service
EXPOSE 8090
ENTRYPOINT ["/auth-service"]
